// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudwatch

import (
	"context"
	"errors"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudwatch_metric_alarm", name="Metric Alarm")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/cloudwatch/types;awstypes;awstypes.MetricAlarm")
func resourceMetricAlarm() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		CreateWithoutTimeout: resourceMetricAlarmCreate,
		ReadWithoutTimeout:   resourceMetricAlarmRead,
		UpdateWithoutTimeout: resourceMetricAlarmUpdate,
		DeleteWithoutTimeout: resourceMetricAlarmDelete,

		SchemaVersion: 1,
		MigrateState:  MetricAlarmMigrateState,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
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
				Required:         true,
				ValidateDiagFunc: enum.Validate[types.ComparisonOperator](),
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
			"evaluation_periods": {
				Type:         schema.TypeInt,
				Required:     true,
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
											validation.IntInSlice([]int{1, 5, 10, 30}),
											validation.IntDivisibleBy(60),
										),
									},
									"stat": {
										Type:     schema.TypeString,
										Required: true,
										ValidateDiagFunc: validation.AnyDiag(
											enum.Validate[types.Statistic](),
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
										ValidateDiagFunc: enum.Validate[types.StandardUnit](),
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
								validation.IntInSlice([]int{1, 5, 10, 30}),
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
					validation.IntInSlice([]int{10, 30}),
					validation.IntDivisibleBy(60),
				),
			},
			"statistic": {
				Type:             schema.TypeString,
				Optional:         true,
				ConflictsWith:    []string{"extended_statistic", "metric_query"},
				ValidateDiagFunc: enum.Validate[types.Statistic](),
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
				ValidateDiagFunc: enum.Validate[types.StandardUnit](),
			},
		},

		CustomizeDiff: customdiff.All(
			verify.SetTagsDiff,
			func(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
				_, metricNameOk := diff.GetOk(names.AttrMetricName)
				_, statisticOk := diff.GetOk("statistic")
				_, extendedStatisticOk := diff.GetOk("extended_statistic")

				if metricNameOk && ((!statisticOk && !extendedStatisticOk) || (statisticOk && extendedStatisticOk)) {
					return errors.New("One of `statistic` or `extended_statistic` must be set for a cloudwatch metric alarm")
				}

				if v := diff.Get("metric_query"); v != nil {
					for _, v := range v.(*schema.Set).List() {
						tfMap := v.(map[string]interface{})
						if v, ok := tfMap[names.AttrExpression]; ok && v.(string) != "" {
							if v := tfMap["metric"]; v != nil {
								if len(v.([]interface{})) > 0 {
									return errors.New("No metric_query may have both `expression` and a `metric` specified")
								}
							}
						}
					}
				}

				return nil
			},
		),
	}
}

func resourceMetricAlarmCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudWatchClient(ctx)

	name := d.Get("alarm_name").(string)
	input := expandPutMetricAlarmInput(ctx, d)

	_, err := conn.PutMetricAlarm(ctx, input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(meta.(*conns.AWSClient).Partition, err) {
		input.Tags = nil

		_, err = conn.PutMetricAlarm(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudWatch Metric Alarm (%s): %s", name, err)
	}

	d.SetId(name)

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		alarm, err := findMetricAlarmByName(ctx, conn, d.Id())

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading CloudWatch Metric Alarm (%s): %s", d.Id(), err)
		}

		err = createTags(ctx, conn, aws.ToString(alarm.AlarmArn), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(meta.(*conns.AWSClient).Partition, err) {
			return append(diags, resourceMetricAlarmRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting CloudWatch Metric Alarm (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceMetricAlarmRead(ctx, d, meta)...)
}

func resourceMetricAlarmRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudWatchClient(ctx)

	alarm, err := findMetricAlarmByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudWatch Metric Alarm %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudWatch Metric Alarm (%s): %s", d.Id(), err)
	}

	d.Set("actions_enabled", alarm.ActionsEnabled)
	d.Set("alarm_actions", alarm.AlarmActions)
	d.Set("alarm_description", alarm.AlarmDescription)
	d.Set("alarm_name", alarm.AlarmName)
	d.Set(names.AttrARN, alarm.AlarmArn)
	d.Set("comparison_operator", alarm.ComparisonOperator)
	d.Set("datapoints_to_alarm", alarm.DatapointsToAlarm)
	if err := d.Set("dimensions", flattenMetricAlarmDimensions(alarm.Dimensions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting dimensions: %s", err)
	}
	d.Set("evaluate_low_sample_count_percentiles", alarm.EvaluateLowSampleCountPercentile)
	d.Set("evaluation_periods", alarm.EvaluationPeriods)
	d.Set("extended_statistic", alarm.ExtendedStatistic)
	d.Set("insufficient_data_actions", alarm.InsufficientDataActions)
	d.Set(names.AttrMetricName, alarm.MetricName)
	if len(alarm.Metrics) > 0 {
		if err := d.Set("metric_query", flattenMetricAlarmMetrics(alarm.Metrics)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting metric_query: %s", err)
		}
	}
	d.Set(names.AttrNamespace, alarm.Namespace)
	d.Set("ok_actions", alarm.OKActions)
	d.Set("period", alarm.Period)
	d.Set("statistic", alarm.Statistic)
	d.Set("threshold", alarm.Threshold)
	d.Set("threshold_metric_id", alarm.ThresholdMetricId)
	if alarm.TreatMissingData != nil { // nosemgrep: ci.helper-schema-ResourceData-Set-extraneous-nil-check
		d.Set("treat_missing_data", alarm.TreatMissingData)
	} else {
		d.Set("treat_missing_data", missingDataMissing)
	}
	d.Set(names.AttrUnit, alarm.Unit)

	return diags
}

func resourceMetricAlarmUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudWatchClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := expandPutMetricAlarmInput(ctx, d)

		_, err := conn.PutMetricAlarm(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating CloudWatch Metric Alarm (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceMetricAlarmRead(ctx, d, meta)...)
}

func resourceMetricAlarmDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudWatchClient(ctx)

	log.Printf("[INFO] Deleting CloudWatch Metric Alarm: %s", d.Id())
	_, err := conn.DeleteAlarms(ctx, &cloudwatch.DeleteAlarmsInput{
		AlarmNames: []string{d.Id()},
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudWatch Metric Alarm (%s): %s", d.Id(), err)
	}

	return diags
}

func findMetricAlarmByName(ctx context.Context, conn *cloudwatch.Client, name string) (*types.MetricAlarm, error) {
	input := &cloudwatch.DescribeAlarmsInput{
		AlarmNames: []string{name},
		AlarmTypes: []types.AlarmType{types.AlarmTypeMetricAlarm},
	}

	output, err := conn.DescribeAlarms(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return tfresource.AssertSingleValueResult(output.MetricAlarms)
}

func expandPutMetricAlarmInput(ctx context.Context, d *schema.ResourceData) *cloudwatch.PutMetricAlarmInput {
	apiObject := &cloudwatch.PutMetricAlarmInput{
		AlarmName:          aws.String(d.Get("alarm_name").(string)),
		ComparisonOperator: types.ComparisonOperator(d.Get("comparison_operator").(string)),
		EvaluationPeriods:  aws.Int32(int32(d.Get("evaluation_periods").(int))),
		Tags:               getTagsIn(ctx),
		TreatMissingData:   aws.String(d.Get("treat_missing_data").(string)),
	}

	if v := d.Get("actions_enabled"); v != nil {
		apiObject.ActionsEnabled = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("alarm_actions"); ok && v.(*schema.Set).Len() > 0 {
		apiObject.AlarmActions = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("alarm_description"); ok {
		apiObject.AlarmDescription = aws.String(v.(string))
	}

	if v, ok := d.GetOk("datapoints_to_alarm"); ok {
		apiObject.DatapointsToAlarm = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("dimensions"); ok && len(v.(map[string]interface{})) > 0 {
		apiObject.Dimensions = expandMetricAlarmDimensions(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("evaluate_low_sample_count_percentiles"); ok {
		apiObject.EvaluateLowSampleCountPercentile = aws.String(v.(string))
	}

	if v, ok := d.GetOk("extended_statistic"); ok {
		apiObject.ExtendedStatistic = aws.String(v.(string))
	}

	if v, ok := d.GetOk("insufficient_data_actions"); ok && v.(*schema.Set).Len() > 0 {
		apiObject.InsufficientDataActions = flex.ExpandStringValueSet(v.(*schema.Set))
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

	if v, ok := d.GetOk("ok_actions"); ok && v.(*schema.Set).Len() > 0 {
		apiObject.OKActions = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("period"); ok {
		apiObject.Period = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("statistic"); ok {
		apiObject.Statistic = types.Statistic(v.(string))
	}

	if v, ok := d.GetOk("threshold_metric_id"); ok {
		apiObject.ThresholdMetricId = aws.String(v.(string))
	} else {
		apiObject.Threshold = aws.Float64(d.Get("threshold").(float64))
	}

	if v, ok := d.GetOk(names.AttrUnit); ok {
		apiObject.Unit = types.StandardUnit(v.(string))
	}

	return apiObject
}

func flattenMetricAlarmDimensions(apiObjects []types.Dimension) map[string]interface{} {
	tfMap := map[string]interface{}{}

	for _, apiObject := range apiObjects {
		tfMap[aws.ToString(apiObject.Name)] = aws.ToString(apiObject.Value)
	}

	return tfMap
}

func flattenMetricAlarmMetrics(apiObjects []types.MetricDataQuery) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
			names.AttrAccountID:  aws.ToString(apiObject.AccountId),
			names.AttrExpression: aws.ToString(apiObject.Expression),
			names.AttrID:         aws.ToString(apiObject.Id),
			"label":              aws.ToString(apiObject.Label),
			"return_data":        aws.ToBool(apiObject.ReturnData),
		}

		if v := apiObject.MetricStat; v != nil {
			tfMap["metric"] = []interface{}{flattenMetricAlarmMetricsMetricStat(v)}
		}

		if apiObject.Period != nil {
			tfMap["period"] = aws.ToInt32(apiObject.Period)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func flattenMetricAlarmMetricsMetricStat(apiObject *types.MetricStat) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{
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

func expandMetricAlarmMetrics(tfList []interface{}) []types.MetricDataQuery {
	var apiObjects []types.MetricDataQuery

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		id := tfMap[names.AttrID].(string)
		if id == "" {
			continue
		}

		apiObject := types.MetricDataQuery{
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

		if v, ok := tfMap["metric"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			apiObject.MetricStat = expandMetricAlarmMetricsMetric(v[0].(map[string]interface{}))
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

func expandMetricAlarmMetricsMetric(tfMap map[string]interface{}) *types.MetricStat {
	if tfMap == nil {
		return nil
	}

	apiObject := &types.MetricStat{
		Metric: &types.Metric{
			MetricName: aws.String(tfMap[names.AttrMetricName].(string)),
		},
		Stat: aws.String(tfMap["stat"].(string)),
	}

	if v, ok := tfMap["dimensions"].(map[string]interface{}); ok && len(v) > 0 {
		apiObject.Metric.Dimensions = expandMetricAlarmDimensions(v)
	}

	if v, ok := tfMap[names.AttrNamespace]; ok && v.(string) != "" {
		apiObject.Metric.Namespace = aws.String(v.(string))
	}

	if v, ok := tfMap["period"]; ok {
		apiObject.Period = aws.Int32(int32(v.(int)))
	}

	if v, ok := tfMap[names.AttrUnit]; ok && v.(string) != "" {
		apiObject.Unit = types.StandardUnit(v.(string))
	}

	return apiObject
}

func expandMetricAlarmDimensions(tfMap map[string]interface{}) []types.Dimension {
	if len(tfMap) == 0 {
		return nil
	}

	var apiObjects []types.Dimension

	for k, v := range tfMap {
		apiObjects = append(apiObjects, types.Dimension{
			Name:  aws.String(k),
			Value: aws.String(v.(string)),
		})
	}

	return apiObjects
}
