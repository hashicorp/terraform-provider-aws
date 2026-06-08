// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package cloudwatch

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfcty "github.com/hashicorp/terraform-provider-aws/internal/cty"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudwatch_metric_alarm", name="Metric Alarm")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/cloudwatch/types;awstypes;awstypes.MetricAlarm")
// @IdentityAttribute("alarm_name")
// @Testing(idAttrDuplicates="alarm_name")
// @Testing(preIdentityVersion="v6.7.0")
func resourceMetricAlarm() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		CreateWithoutTimeout: resourceMetricAlarmCreate,
		ReadWithoutTimeout:   resourceMetricAlarmRead,
		UpdateWithoutTimeout: resourceMetricAlarmUpdate,
		DeleteWithoutTimeout: resourceMetricAlarmDelete,

		SchemaVersion: 1,
		MigrateState:  MetricAlarmMigrateState,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"actions_enabled": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  true,
				},
				"alarm_actions": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
						ValidateFunc: validation.Any(
							verify.ValidARN,
							validEC2AutomateARN,
						),
					},
				},
				"alarm_description": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validation.StringLenBetween(0, 1024),
				},
				"alarm_name": {
					Type:         schema.TypeString,
					Required:     true,
					ForceNew:     true,
					ValidateFunc: validation.StringLenBetween(1, 255),
				},
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"comparison_operator": {
					Type:             schema.TypeString,
					Optional:         true,
					ValidateDiagFunc: enum.Validate[awstypes.ComparisonOperator](),
				},
				"datapoints_to_alarm": {
					Type:         schema.TypeInt,
					Optional:     true,
					ValidateFunc: validation.IntAtLeast(1),
				},
				"dimensions": {
					Type:          schema.TypeMap,
					Optional:      true,
					Elem:          &schema.Schema{Type: schema.TypeString},
					ConflictsWith: []string{"metric_query"},
				},
				"evaluate_low_sample_count_percentiles": {
					Type:         schema.TypeString,
					Optional:     true,
					Computed:     true,
					ValidateFunc: validation.StringInSlice(lowSampleCountPercentiles_Values(), true),
				},
				"evaluation_criteria": {
					Type:          schema.TypeList,
					Optional:      true,
					MaxItems:      1,
					ExactlyOneOf:  []string{"evaluation_criteria", names.AttrMetricName, "metric_query"},
					ConflictsWith: []string{names.AttrNamespace, names.AttrMetricName, "dimensions", "period", names.AttrUnit, "statistic", "extended_statistic", "metric_query", "threshold", "comparison_operator", "threshold_metric_id", "evaluation_periods", "datapoints_to_alarm"},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"promql_criteria": {
								Type:     schema.TypeList,
								Required: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"pending_period": {
											Type:         schema.TypeInt,
											Optional:     true,
											ValidateFunc: validation.IntBetween(0, 86400),
										},
										"query": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.StringLenBetween(1, 10000),
										},
										"recovery_period": {
											Type:         schema.TypeInt,
											Optional:     true,
											ValidateFunc: validation.IntBetween(0, 86400),
										},
									},
								},
							},
						},
					},
				},
				"evaluation_interval": {
					Type:          schema.TypeInt,
					Optional:      true,
					ConflictsWith: []string{names.AttrMetricName, "metric_query"},
					RequiredWith:  []string{"evaluation_criteria"},
					ValidateFunc: validation.Any(
						validation.IntInSlice([]int{10, 20, 30}),
						validation.IntDivisibleBy(60),
					),
				},
				"evaluation_periods": {
					Type:         schema.TypeInt,
					Optional:     true,
					ValidateFunc: validation.IntAtLeast(1),
				},
				"extended_statistic": {
					Type:          schema.TypeString,
					Optional:      true,
					ConflictsWith: []string{"statistic", "metric_query"},
					ValidateFunc: validation.StringMatch(
						// doesn't catch: PR with %-values provided, TM/WM/PR/TC/TS with no values provided
						regexache.MustCompile(`^((p|(tm)|(wm)|(tc)|(ts))((\d{1,2}(\.\d{1,2})?)|(100))|(IQM)|(((TM)|(WM)|(PR)|(TC)|(TS)))\((\d+(\.\d+)?%?)?:(\d+(\.\d+)?%?)?\))$`),
						"invalid statistic, see: https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/Statistics-definitions.html",
					),
				},
				"insufficient_data_actions": {
					Type:     schema.TypeSet,
					Optional: true,
					MaxItems: 5,
					Elem: &schema.Schema{
						Type: schema.TypeString,
						ValidateFunc: validation.Any(
							verify.ValidARN,
							validEC2AutomateARN,
						),
					},
				},
				names.AttrMetricName: {
					Type:          schema.TypeString,
					Optional:      true,
					ConflictsWith: []string{"metric_query"},
					ValidateFunc:  validation.StringLenBetween(1, 255),
				},
				"metric_query": {
					Type:          schema.TypeSet,
					Optional:      true,
					ConflictsWith: []string{names.AttrMetricName},
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							names.AttrAccountID: {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: validation.StringLenBetween(1, 255),
							},
							names.AttrExpression: {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: validation.StringLenBetween(1, 1024),
							},
							names.AttrID: {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.StringLenBetween(1, 255),
							},
							"metric": {
								Type:     schema.TypeList,
								MaxItems: 1,
								Optional: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"dimensions": {
											Type:     schema.TypeMap,
											Optional: true,
											Elem:     &schema.Schema{Type: schema.TypeString},
										},
										names.AttrMetricName: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.StringLenBetween(1, 255),
										},
										names.AttrNamespace: {
											Type:     schema.TypeString,
											Optional: true,
											ValidateFunc: validation.All(
												validation.StringLenBetween(1, 255),
												validation.StringMatch(regexache.MustCompile(`[^:].*`), "must not contain colon characters"),
											),
										},
										"period": {
											Type:     schema.TypeInt,
											Required: true,
											ValidateFunc: validation.Any(
												validation.IntInSlice([]int{1, 5, 10, 20, 30}),
												validation.IntDivisibleBy(60),
											),
										},
										"stat": {
											Type:     schema.TypeString,
											Required: true,
											ValidateDiagFunc: validation.AnyDiag(
												enum.Validate[awstypes.Statistic](),
												validation.ToDiagFunc(
													validation.StringMatch(
														// doesn't catch: PR with %-values provided, TM/WM/PR/TC/TS with no values provided
														regexache.MustCompile(`^((p|(tm)|(wm)|(tc)|(ts))((\d{1,2}(\.\d{1,2})?)|(100))|(IQM)|(((TM)|(WM)|(PR)|(TC)|(TS)))\((\d+(\.\d+)?%?)?:(\d+(\.\d+)?%?)?\))$`),
														"invalid statistic, see: https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/Statistics-definitions.html",
													),
												),
											),
										},
										names.AttrUnit: {
											Type:             schema.TypeString,
											Optional:         true,
											ValidateDiagFunc: enum.Validate[awstypes.StandardUnit](),
										},
									},
								},
							},
							"label": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"period": {
								Type:     schema.TypeInt,
								Optional: true,
								ValidateFunc: validation.Any(
									validation.IntInSlice([]int{1, 5, 10, 20, 30}),
									validation.IntDivisibleBy(60),
								),
							},
							"return_data": {
								Type:     schema.TypeBool,
								Optional: true,
								Default:  false,
							},
						},
					},
				},
				names.AttrNamespace: {
					Type:          schema.TypeString,
					Optional:      true,
					ConflictsWith: []string{"metric_query"},
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 255),
						validation.StringMatch(regexache.MustCompile(`[^:].*`), "must not contain colon characters"),
					),
				},
				"ok_actions": {
					Type:     schema.TypeSet,
					Optional: true,
					MaxItems: 5,
					Elem: &schema.Schema{
						Type: schema.TypeString,
						ValidateFunc: validation.Any(
							verify.ValidARN,
							validEC2AutomateARN,
						),
					},
				},
				"period": {
					Type:          schema.TypeInt,
					Optional:      true,
					ConflictsWith: []string{"metric_query"},
					ValidateFunc: validation.Any(
						validation.IntInSlice([]int{10, 20, 30}),
						validation.IntDivisibleBy(60),
					),
				},
				"statistic": {
					Type:             schema.TypeString,
					Optional:         true,
					ConflictsWith:    []string{"extended_statistic", "metric_query"},
					ValidateDiagFunc: enum.Validate[awstypes.Statistic](),
				},
				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaComputed(),
				"threshold": {
					Type:          schema.TypeFloat,
					Optional:      true,
					ConflictsWith: []string{"threshold_metric_id"},
				},
				"threshold_metric_id": {
					Type:          schema.TypeString,
					Optional:      true,
					ConflictsWith: []string{"threshold"},
					ValidateFunc:  validation.StringLenBetween(1, 255),
				},
				"treat_missing_data": {
					Type:         schema.TypeString,
					Optional:     true,
					Default:      missingDataMissing,
					ValidateFunc: validation.StringInSlice(missingData_Values(), true),
				},
				names.AttrUnit: {
					Type:             schema.TypeString,
					Optional:         true,
					ValidateDiagFunc: enum.Validate[awstypes.StandardUnit](),
				},
			}
		},

		CustomizeDiff: resourceMetricAlarmCustomizeDiff,
	}
}

func resourceMetricAlarmCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudWatchClient(ctx)

	name := d.Get("alarm_name").(string)
	input := expandPutMetricAlarmInput(ctx, d)

	_, err := conn.PutMetricAlarm(ctx, input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(meta.(*conns.AWSClient).Partition(ctx), err) {
		input.Tags = nil

		_, err = conn.PutMetricAlarm(ctx, input)
	}

	if err != nil {
		return smerr.Append(ctx, diags, err, smerr.ID, name)
	}

	d.SetId(name)

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		alarm, err := findMetricAlarmByName(ctx, conn, d.Id())

		if err != nil {
			return smerr.Append(ctx, diags, err, smerr.ID, d.Id())
		}

		err = createTags(ctx, conn, aws.ToString(alarm.AlarmArn), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]any)) == 0) && errs.IsUnsupportedOperationInPartitionError(meta.(*conns.AWSClient).Partition(ctx), err) {
			return smerr.AppendEnrich(ctx, diags, resourceMetricAlarmRead(ctx, d, meta))
		}

		if err != nil {
			return smerr.Append(ctx, diags, err, smerr.ID, d.Id())
		}
	}

	return smerr.AppendEnrich(ctx, diags, resourceMetricAlarmRead(ctx, d, meta))
}

func resourceMetricAlarmRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudWatchClient(ctx)

	alarm, err := findMetricAlarmByName(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		smerr.AppendOne(ctx, diags, sdkdiag.NewResourceNotFoundWarningDiagnostic(err), smerr.ID, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return smerr.Append(ctx, diags, err, smerr.ID, d.Id())
	}

	if err := resourceMetricAlarmFlatten(ctx, d, alarm); err != nil {
		return smerr.Append(ctx, diags, err, smerr.ID, d.Id())
	}

	return diags
}

func resourceMetricAlarmUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudWatchClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := expandPutMetricAlarmInput(ctx, d)

		_, err := conn.PutMetricAlarm(ctx, input)

		if err != nil {
			return smerr.Append(ctx, diags, err, smerr.ID, d.Id())
		}
	}

	return smerr.AppendEnrich(ctx, diags, resourceMetricAlarmRead(ctx, d, meta))
}

func resourceMetricAlarmDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudWatchClient(ctx)

	log.Printf("[INFO] Deleting CloudWatch Metric Alarm: %s", d.Id())
	input := cloudwatch.DeleteAlarmsInput{
		AlarmNames: []string{d.Id()},
	}
	_, err := conn.DeleteAlarms(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return smerr.Append(ctx, diags, err, smerr.ID, d.Id())
	}

	return diags
}

func resourceMetricAlarmCustomizeDiff(ctx context.Context, diff *schema.ResourceDiff, v any) error {
	var plan metricAlarmResourceModel
	if err := tfcty.ToFramework(ctx, diff.GetRawPlan(), &plan); err != nil {
		return smarterr.NewError(fmt.Errorf("RawPlan to framework model: %w", err))
	}

	// Traditional metric alarm validation.
	if !plan.MetricName.IsUnknown() && !plan.MetricQuery.IsUnknown() && !plan.Statistic.IsUnknown() && !plan.ExtendedStatistic.IsUnknown() {
		metricNameOk := fwflex.StringValueFromFramework(ctx, plan.MetricName) != ""
		metricQueryOk := plan.MetricQuery.Length(basetypes.CollectionLengthOptions{UnhandledNullAsZero: true}) > 0

		// Traditional metric alarms require comparison_operator and evaluation_periods.
		if metricNameOk || metricQueryOk {
			if fwflex.StringValueFromFramework(ctx, plan.ComparisonOperator) == "" {
				return errors.New("comparison_operator is required for traditional metric alarms")
			}
			if plan.EvaluationPeriods.ValueInt64() == 0 {
				return errors.New("evaluation_periods is required for traditional metric alarms")
			}
		}

		statisticOk := fwflex.StringValueFromFramework(ctx, plan.Statistic) != ""
		extendedStatisticOk := fwflex.StringValueFromFramework(ctx, plan.ExtendedStatistic) != ""
		if metricNameOk && ((!statisticOk && !extendedStatisticOk) || (statisticOk && extendedStatisticOk)) {
			return errors.New("One of `statistic` or `extended_statistic` must be set for a cloudwatch metric alarm")
		}
	}

	if plan.MetricQuery.Length(basetypes.CollectionLengthOptions{UnhandledNullAsZero: true, UnhandledUnknownAsZero: true}) > 0 {
		if mdqs, diags := plan.MetricQuery.ToSlice(ctx); !diags.HasError() {
			for _, mdq := range mdqs {
				if mdq == nil {
					continue
				}
				if fwflex.StringValueFromFramework(ctx, mdq.Expression) != "" {
					if mdq.Metric.Length(basetypes.CollectionLengthOptions{UnhandledNullAsZero: true, UnhandledUnknownAsZero: true}) > 0 {
						return errors.New("No metric_query may have both `expression` and a `metric` specified")
					}
				}
			}
		}
	}

	return nil
}

func findMetricAlarmByName(ctx context.Context, conn *cloudwatch.Client, name string) (*awstypes.MetricAlarm, error) {
	input := &cloudwatch.DescribeAlarmsInput{
		AlarmNames: []string{name},
		AlarmTypes: []awstypes.AlarmType{awstypes.AlarmTypeMetricAlarm},
	}

	output, err := conn.DescribeAlarms(ctx, input)

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if output == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return smarterr.Assert(tfresource.AssertSingleValueResult(output.MetricAlarms))
}

func expandPutMetricAlarmInput(ctx context.Context, d *schema.ResourceData) *cloudwatch.PutMetricAlarmInput {
	apiObject := &cloudwatch.PutMetricAlarmInput{
		AlarmName: aws.String(d.Get("alarm_name").(string)),
		Tags:      getTagsIn(ctx),
	}

	// Set common fields for both PromQL and traditional alarms.
	if v, ok := d.GetOk("alarm_actions"); ok && v.(*schema.Set).Len() > 0 {
		apiObject.AlarmActions = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("alarm_description"); ok {
		apiObject.AlarmDescription = aws.String(v.(string))
	}

	if v, ok := d.GetOk("insufficient_data_actions"); ok && v.(*schema.Set).Len() > 0 {
		apiObject.InsufficientDataActions = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("ok_actions"); ok && v.(*schema.Set).Len() > 0 {
		apiObject.OKActions = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	// Handle evaluation_criteria (PromQL alarms).
	if v, ok := d.GetOk("evaluation_criteria"); ok && len(v.([]any)) > 0 {
		apiObject.EvaluationCriteria = expandEvaluationCriteria(v.([]any)[0].(map[string]any))

		if v, ok := d.GetOk("evaluation_interval"); ok {
			apiObject.EvaluationInterval = aws.Int32(int32(v.(int)))
		}

		return apiObject
	}

	// Handle traditional metric alarms - set fields that are only for traditional alarms.
	if v := d.Get("actions_enabled"); v != nil {
		apiObject.ActionsEnabled = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("comparison_operator"); ok {
		apiObject.ComparisonOperator = awstypes.ComparisonOperator(v.(string))
	}

	if v, ok := d.GetOk("datapoints_to_alarm"); ok {
		apiObject.DatapointsToAlarm = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("dimensions"); ok && len(v.(map[string]any)) > 0 {
		apiObject.Dimensions = expandMetricAlarmDimensions(v.(map[string]any))
	}

	if v, ok := d.GetOk("evaluate_low_sample_count_percentiles"); ok {
		apiObject.EvaluateLowSampleCountPercentile = aws.String(v.(string))
	}

	if v, ok := d.GetOk("evaluation_periods"); ok {
		apiObject.EvaluationPeriods = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("extended_statistic"); ok {
		apiObject.ExtendedStatistic = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrMetricName); ok {
		apiObject.MetricName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("metric_query"); ok && v.(*schema.Set).Len() > 0 {
		apiObject.Metrics = expandMetricAlarmMetrics(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk(names.AttrNamespace); ok {
		apiObject.Namespace = aws.String(v.(string))
	}

	if v, ok := d.GetOk("period"); ok {
		apiObject.Period = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("statistic"); ok {
		apiObject.Statistic = awstypes.Statistic(v.(string))
	}

	if v, ok := d.GetOk("threshold_metric_id"); ok {
		apiObject.ThresholdMetricId = aws.String(v.(string))
	} else {
		apiObject.Threshold = aws.Float64(d.Get("threshold").(float64))
	}

	apiObject.TreatMissingData = aws.String(d.Get("treat_missing_data").(string))

	if v, ok := d.GetOk(names.AttrUnit); ok {
		apiObject.Unit = awstypes.StandardUnit(v.(string))
	}

	return apiObject
}

func flattenEvaluationCriteria(apiObject awstypes.EvaluationCriteria) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	switch v := apiObject.(type) {
	case *awstypes.EvaluationCriteriaMemberPromQLCriteria:
		promqlMap := map[string]any{
			"query": aws.ToString(v.Value.Query),
		}

		if v.Value.PendingPeriod != nil {
			promqlMap["pending_period"] = aws.ToInt32(v.Value.PendingPeriod)
		}

		if v.Value.RecoveryPeriod != nil {
			promqlMap["recovery_period"] = aws.ToInt32(v.Value.RecoveryPeriod)
		}

		tfMap["promql_criteria"] = []any{promqlMap}
	}

	return []any{tfMap}
}

func flattenMetricAlarmDimensions(apiObjects []awstypes.Dimension) map[string]any {
	tfMap := map[string]any{}

	for _, apiObject := range apiObjects {
		tfMap[aws.ToString(apiObject.Name)] = aws.ToString(apiObject.Value)
	}

	return tfMap
}

func flattenMetricAlarmMetrics(apiObjects []awstypes.MetricDataQuery) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfMap := map[string]any{
			names.AttrAccountID:  aws.ToString(apiObject.AccountId),
			names.AttrExpression: aws.ToString(apiObject.Expression),
			names.AttrID:         aws.ToString(apiObject.Id),
			"label":              aws.ToString(apiObject.Label),
			"return_data":        aws.ToBool(apiObject.ReturnData),
		}

		if v := apiObject.MetricStat; v != nil {
			tfMap["metric"] = []any{flattenMetricAlarmMetricsMetricStat(v)}
		}

		if apiObject.Period != nil {
			tfMap["period"] = aws.ToInt32(apiObject.Period)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenMetricAlarmMetricsMetricStat(apiObject *awstypes.MetricStat) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"period":       aws.ToInt32(apiObject.Period),
		"stat":         aws.ToString(apiObject.Stat),
		names.AttrUnit: apiObject.Unit,
	}

	if v := apiObject.Metric; v != nil {
		tfMap["dimensions"] = flattenMetricAlarmDimensions(v.Dimensions)
		tfMap[names.AttrMetricName] = aws.ToString(v.MetricName)
		tfMap[names.AttrNamespace] = aws.ToString(v.Namespace)
	}

	return tfMap
}

func expandMetricAlarmMetrics(tfList []any) []awstypes.MetricDataQuery {
	var apiObjects []awstypes.MetricDataQuery

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		id := tfMap[names.AttrID].(string)
		if id == "" {
			continue
		}

		apiObject := awstypes.MetricDataQuery{
			Id: aws.String(id),
		}

		if v, ok := tfMap[names.AttrAccountID]; ok && v.(string) != "" {
			apiObject.AccountId = aws.String(v.(string))
		}

		if v, ok := tfMap[names.AttrExpression]; ok && v.(string) != "" {
			apiObject.Expression = aws.String(v.(string))
		}

		if v, ok := tfMap["label"]; ok && v.(string) != "" {
			apiObject.Label = aws.String(v.(string))
		}

		if v, ok := tfMap["return_data"]; ok {
			apiObject.ReturnData = aws.Bool(v.(bool))
		}

		if v, ok := tfMap["metric"].([]any); ok && len(v) > 0 && v[0] != nil {
			apiObject.MetricStat = expandMetricAlarmMetricsMetric(v[0].(map[string]any))
		}

		if v, ok := tfMap["period"]; ok && v.(int) != 0 {
			apiObject.Period = aws.Int32(int32(v.(int)))
		}

		apiObjects = append(apiObjects, apiObject)
	}

	if len(apiObjects) == 0 {
		return nil
	}

	return apiObjects
}

func expandMetricAlarmMetricsMetric(tfMap map[string]any) *awstypes.MetricStat {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.MetricStat{
		Metric: &awstypes.Metric{
			MetricName: aws.String(tfMap[names.AttrMetricName].(string)),
		},
		Stat: aws.String(tfMap["stat"].(string)),
	}

	if v, ok := tfMap["dimensions"].(map[string]any); ok && len(v) > 0 {
		apiObject.Metric.Dimensions = expandMetricAlarmDimensions(v)
	}

	if v, ok := tfMap[names.AttrNamespace]; ok && v.(string) != "" {
		apiObject.Metric.Namespace = aws.String(v.(string))
	}

	if v, ok := tfMap["period"]; ok {
		apiObject.Period = aws.Int32(int32(v.(int)))
	}

	if v, ok := tfMap[names.AttrUnit]; ok && v.(string) != "" {
		apiObject.Unit = awstypes.StandardUnit(v.(string))
	}

	return apiObject
}

func expandEvaluationCriteria(tfMap map[string]any) awstypes.EvaluationCriteria {
	if tfMap == nil {
		return nil
	}

	if v, ok := tfMap["promql_criteria"].([]any); ok && len(v) > 0 && v[0] != nil {
		tfMap := v[0].(map[string]any)

		apiObject := awstypes.AlarmPromQLCriteria{
			Query: aws.String(tfMap["query"].(string)),
		}

		if v, ok := tfMap["pending_period"]; ok && v.(int) != 0 {
			apiObject.PendingPeriod = aws.Int32(int32(v.(int)))
		}

		if v, ok := tfMap["recovery_period"]; ok && v.(int) != 0 {
			apiObject.RecoveryPeriod = aws.Int32(int32(v.(int)))
		}

		return &awstypes.EvaluationCriteriaMemberPromQLCriteria{
			Value: apiObject,
		}
	}

	return nil
}

func expandMetricAlarmDimensions(tfMap map[string]any) []awstypes.Dimension {
	if len(tfMap) == 0 {
		return nil
	}

	var apiObjects []awstypes.Dimension

	for k, v := range tfMap {
		apiObjects = append(apiObjects, awstypes.Dimension{
			Name:  aws.String(k),
			Value: aws.String(v.(string)),
		})
	}

	return apiObjects
}

func resourceMetricAlarmFlatten(_ context.Context, d *schema.ResourceData, alarm *awstypes.MetricAlarm) error {
	d.Set("actions_enabled", alarm.ActionsEnabled)
	d.Set("alarm_actions", alarm.AlarmActions)
	d.Set("alarm_description", alarm.AlarmDescription)
	d.Set("alarm_name", alarm.AlarmName)
	d.Set(names.AttrARN, alarm.AlarmArn)
	d.Set("insufficient_data_actions", alarm.InsufficientDataActions)
	d.Set("ok_actions", alarm.OKActions)

	// Handle EvaluationCriteria (PromQL alarms).
	if alarm.EvaluationCriteria != nil {
		if err := d.Set("evaluation_criteria", flattenEvaluationCriteria(alarm.EvaluationCriteria)); err != nil {
			return smarterr.NewError(fmt.Errorf("setting evaluation_criteria: %w", err))
		}
		d.Set("evaluation_interval", alarm.EvaluationInterval)

		// Clear traditional metric alarm fields for PromQL alarms
		d.Set("comparison_operator", nil)
		d.Set("evaluation_periods", nil)
		d.Set("datapoints_to_alarm", nil)
		d.Set("dimensions", nil)
		d.Set("evaluate_low_sample_count_percentiles", nil)
		d.Set("extended_statistic", nil)
		d.Set(names.AttrMetricName, nil)
		d.Set("metric_query", nil)
		d.Set(names.AttrNamespace, nil)
		d.Set("period", nil)
		d.Set("statistic", nil)
		d.Set("threshold", nil)
		d.Set("threshold_metric_id", nil)
		d.Set(names.AttrUnit, nil)
	} else {
		// Handle traditional metric alarms
		d.Set("comparison_operator", alarm.ComparisonOperator)
		d.Set("datapoints_to_alarm", alarm.DatapointsToAlarm)
		if err := d.Set("dimensions", flattenMetricAlarmDimensions(alarm.Dimensions)); err != nil {
			return smarterr.NewError(fmt.Errorf("setting dimensions: %w", err))
		}
		d.Set("evaluate_low_sample_count_percentiles", alarm.EvaluateLowSampleCountPercentile)
		d.Set("evaluation_periods", alarm.EvaluationPeriods)
		d.Set("extended_statistic", alarm.ExtendedStatistic)
		d.Set(names.AttrMetricName, alarm.MetricName)
		if len(alarm.Metrics) > 0 {
			if err := d.Set("metric_query", flattenMetricAlarmMetrics(alarm.Metrics)); err != nil {
				return smarterr.NewError(fmt.Errorf("setting metric_query: %w", err))
			}
		}
		d.Set(names.AttrNamespace, alarm.Namespace)
		d.Set("period", alarm.Period)
		d.Set("statistic", alarm.Statistic)
		d.Set("threshold", alarm.Threshold)
		d.Set("threshold_metric_id", alarm.ThresholdMetricId)
		d.Set(names.AttrUnit, alarm.Unit)

		// Clear PromQL fields for traditional alarms
		d.Set("evaluation_criteria", nil)
		d.Set("evaluation_interval", nil)
	}

	if alarm.TreatMissingData != nil { // nosemgrep: ci.helper-schema-ResourceData-Set-extraneous-nil-check
		d.Set("treat_missing_data", alarm.TreatMissingData)
	} else {
		d.Set("treat_missing_data", missingDataMissing)
	}

	return nil
}

type metricAlarmResourceModel struct {
	framework.WithRegionModel
	ActionsEnabled                    types.Bool                                               `tfsdk:"actions_enabled"`
	AlarmActions                      fwtypes.SetOfString                                      `tfsdk:"alarm_actions"`
	AlarmDescription                  types.String                                             `tfsdk:"alarm_description"`
	AlarmName                         types.String                                             `tfsdk:"alarm_name"`
	ARN                               types.String                                             `tfsdk:"arn"`
	ComparisonOperator                fwtypes.StringEnum[awstypes.ComparisonOperator]          `tfsdk:"comparison_operator"`
	DatapointsToAlarm                 types.Int64                                              `tfsdk:"datapoints_to_alarm"`
	Dimensions                        fwtypes.MapOfString                                      `tfsdk:"dimensions"`
	EvaluateLowSampleCountPercentiles types.String                                             `tfsdk:"evaluate_low_sample_count_percentiles"`
	EvaluationCriteria                fwtypes.ListNestedObjectValueOf[evaluationCriteriaModel] `tfsdk:"evaluation_criteria"`
	EvaluationInterval                types.Int64                                              `tfsdk:"evaluation_interval"`
	EvaluationPeriods                 types.Int64                                              `tfsdk:"evaluation_periods"`
	ExtendedStatistic                 types.String                                             `tfsdk:"extended_statistic"`
	InsufficientDataActions           fwtypes.SetOfString                                      `tfsdk:"insufficient_data_actions"`
	MetricName                        types.String                                             `tfsdk:"metric_name"`
	MetricQuery                       fwtypes.SetNestedObjectValueOf[metricDataQueryModel]     `tfsdk:"metric_query"`
	Namespace                         types.String                                             `tfsdk:"namespace"`
	OKActions                         fwtypes.SetOfString                                      `tfsdk:"ok_actions"`
	Period                            types.Int64                                              `tfsdk:"period"`
	Statistic                         fwtypes.StringEnum[awstypes.Statistic]                   `tfsdk:"statistic"`
	Tags                              tftags.Map                                               `tfsdk:"tags"`
	TagsAll                           tftags.Map                                               `tfsdk:"tags_all"`
	Threshold                         types.Float64                                            `tfsdk:"threshold"`
	ThresholdMetricID                 types.String                                             `tfsdk:"threshold_metric_id"`
	TreatMissingData                  types.String                                             `tfsdk:"treat_missing_data"`
	Unit                              fwtypes.StringEnum[awstypes.StandardUnit]                `tfsdk:"unit"`
}

type evaluationCriteriaModel struct {
	PromQLCriteria fwtypes.ListNestedObjectValueOf[alarmPromQLCriteriaModel] `tfsdk:"promql_criteria"`
}

type alarmPromQLCriteriaModel struct {
	PendingPeriod  types.Int64  `tfsdk:"pending_period"`
	Query          types.String `tfsdk:"query"`
	RecoveryPeriod types.Int64  `tfsdk:"recovery_period"`
}

type metricDataQueryModel struct {
	AccountID  types.String                                     `tfsdk:"account_id"`
	Expression types.String                                     `tfsdk:"expression"`
	ID         types.String                                     `tfsdk:"id"`
	Metric     fwtypes.ListNestedObjectValueOf[metricStatModel] `tfsdk:"metric"`
	Label      types.String                                     `tfsdk:"label"`
	Period     types.Int64                                      `tfsdk:"period"`
	ReturnData types.Bool                                       `tfsdk:"return_data"`
}

type metricStatModel struct {
	Dimensions fwtypes.MapOfString                       `tfsdk:"dimensions"`
	MetricName types.String                              `tfsdk:"metric_name"`
	Namespace  types.String                              `tfsdk:"namespace"`
	Period     types.Int64                               `tfsdk:"period"`
	Stat       types.String                              `tfsdk:"stat"`
	Unit       fwtypes.StringEnum[awstypes.StandardUnit] `tfsdk:"unit"`
}
