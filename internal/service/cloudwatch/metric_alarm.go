// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudwatch

import (
	"context"
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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
func ResourceMetricAlarm() *schema.Resource {
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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"comparison_operator": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(cloudwatch.ComparisonOperator_Values(), false),
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
					regexp.MustCompile(`^((p|(tm)|(wm)|(tc)|(ts))((\d{1,2}(\.\d{1,2})?)|(100))|(IQM)|(((TM)|(WM)|(PR)|(TC)|(TS)))\((\d+(\.\d+)?%?)?:(\d+(\.\d+)?%?)?\))$`),
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
			"metric_name": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"metric_query"},
				ValidateFunc:  validation.StringLenBetween(1, 255),
			},
			"metric_query": {
				Type:          schema.TypeSet,
				Optional:      true,
				ConflictsWith: []string{"metric_name"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
						"account_id": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
						"expression": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(1, 1024),
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
									"metric_name": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 255),
									},
									"namespace": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 255),
											validation.StringMatch(regexp.MustCompile(`[^:].*`), "must not contain colon characters"),
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
										ValidateFunc: validation.Any(
											validation.StringInSlice(cloudwatch.Statistic_Values(), false),
											validation.StringMatch(
												// doesn't catch: PR with %-values provided, TM/WM/PR/TC/TS with no values provided
												regexp.MustCompile(`^((p|(tm)|(wm)|(tc)|(ts))((\d{1,2}(\.\d{1,2})?)|(100))|(IQM)|(((TM)|(WM)|(PR)|(TC)|(TS)))\((\d+(\.\d+)?%?)?:(\d+(\.\d+)?%?)?\))$`),
												"invalid statistic, see: https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/Statistics-definitions.html",
											),
										),
									},
									"unit": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(cloudwatch.StandardUnit_Values(), false),
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
			"namespace": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"metric_query"},
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringMatch(regexp.MustCompile(`[^:].*`), "must not contain colon characters"),
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
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"extended_statistic", "metric_query"},
				ValidateFunc:  validation.StringInSlice(cloudwatch.Statistic_Values(), false),
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
			"unit": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(cloudwatch.StandardUnit_Values(), false),
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func validMetricAlarm(d *schema.ResourceData) error {
	_, metricNameOk := d.GetOk("metric_name")
	_, statisticOk := d.GetOk("statistic")
	_, extendedStatisticOk := d.GetOk("extended_statistic")

	if metricNameOk && ((!statisticOk && !extendedStatisticOk) || (statisticOk && extendedStatisticOk)) {
		return fmt.Errorf("One of `statistic` or `extended_statistic` must be set for a cloudwatch metric alarm")
	}

	if v := d.Get("metric_query"); v != nil {
		for _, v := range v.(*schema.Set).List() {
			metricQueryResource := v.(map[string]interface{})
			if v, ok := metricQueryResource["expression"]; ok && v.(string) != "" {
				if v := metricQueryResource["metric"]; v != nil {
					if len(v.([]interface{})) > 0 {
						return fmt.Errorf("No metric_query may have both `expression` and a `metric` specified")
					}
				}
			}
		}
	}

	return nil
}

func resourceMetricAlarmCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudWatchConn(ctx)

	err := validMetricAlarm(d)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudWatch Metric Alarm (%s): %s", d.Get("alarm_name").(string), err)
	}

	name := d.Get("alarm_name").(string)
	input := expandPutMetricAlarmInput(ctx, d)

	_, err = conn.PutMetricAlarmWithContext(ctx, input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
		input.Tags = nil

		_, err = conn.PutMetricAlarmWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudWatch Metric Alarm (%s): %s", name, err)
	}

	d.SetId(name)

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		alarm, err := FindMetricAlarmByName(ctx, conn, d.Id())

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading CloudWatch Metric Alarm (%s): %s", d.Id(), err)
		}

		err = createTags(ctx, conn, aws.StringValue(alarm.AlarmArn), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
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
	conn := meta.(*conns.AWSClient).CloudWatchConn(ctx)

	alarm, err := FindMetricAlarmByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudWatch Metric Alarm %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return diag.Errorf("reading CloudWatch Metric Alarm (%s): %s", d.Id(), err)
	}

	d.Set("actions_enabled", alarm.ActionsEnabled)
	d.Set("alarm_actions", aws.StringValueSlice(alarm.AlarmActions))
	d.Set("alarm_description", alarm.AlarmDescription)
	d.Set("alarm_name", alarm.AlarmName)
	d.Set("arn", alarm.AlarmArn)
	d.Set("comparison_operator", alarm.ComparisonOperator)
	d.Set("datapoints_to_alarm", alarm.DatapointsToAlarm)
	if err := d.Set("dimensions", flattenMetricAlarmDimensions(alarm.Dimensions)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting dimensions: %s", err)
	}
	d.Set("evaluate_low_sample_count_percentiles", alarm.EvaluateLowSampleCountPercentile)
	d.Set("evaluation_periods", alarm.EvaluationPeriods)
	d.Set("extended_statistic", alarm.ExtendedStatistic)
	d.Set("insufficient_data_actions", aws.StringValueSlice(alarm.InsufficientDataActions))
	d.Set("metric_name", alarm.MetricName)
	if len(alarm.Metrics) > 0 {
		if err := d.Set("metric_query", flattenMetricAlarmMetrics(alarm.Metrics)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting metric_query: %s", err)
		}
	}
	d.Set("namespace", alarm.Namespace)
	d.Set("ok_actions", aws.StringValueSlice(alarm.OKActions))
	d.Set("period", alarm.Period)
	d.Set("statistic", alarm.Statistic)
	d.Set("threshold", alarm.Threshold)
	d.Set("threshold_metric_id", alarm.ThresholdMetricId)
	if alarm.TreatMissingData != nil { // nosemgrep: ci.helper-schema-ResourceData-Set-extraneous-nil-check
		d.Set("treat_missing_data", alarm.TreatMissingData)
	} else {
		d.Set("treat_missing_data", missingDataMissing)
	}
	d.Set("unit", alarm.Unit)

	return diags
}

func resourceMetricAlarmUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudWatchConn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := expandPutMetricAlarmInput(ctx, d)

		_, err := conn.PutMetricAlarmWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating CloudWatch Metric Alarm (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceMetricAlarmRead(ctx, d, meta)...)
}

func resourceMetricAlarmDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudWatchConn(ctx)

	log.Printf("[INFO] Deleting CloudWatch Metric Alarm: %s", d.Id())
	_, err := conn.DeleteAlarmsWithContext(ctx, &cloudwatch.DeleteAlarmsInput{
		AlarmNames: aws.StringSlice([]string{d.Id()}),
	})

	if tfawserr.ErrCodeEquals(err, cloudwatch.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudWatch Metric Alarm (%s): %s", d.Id(), err)
	}

	return diags
}

func FindMetricAlarmByName(ctx context.Context, conn *cloudwatch.CloudWatch, name string) (*cloudwatch.MetricAlarm, error) {
	input := &cloudwatch.DescribeAlarmsInput{
		AlarmNames: aws.StringSlice([]string{name}),
		AlarmTypes: aws.StringSlice([]string{cloudwatch.AlarmTypeMetricAlarm}),
	}

	output, err := conn.DescribeAlarmsWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, cloudwatch.ErrCodeResourceNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.MetricAlarms) == 0 || output.MetricAlarms[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.MetricAlarms); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.MetricAlarms[0], nil
}

func expandPutMetricAlarmInput(ctx context.Context, d *schema.ResourceData) *cloudwatch.PutMetricAlarmInput {
	apiObject := &cloudwatch.PutMetricAlarmInput{
		AlarmName:          aws.String(d.Get("alarm_name").(string)),
		ComparisonOperator: aws.String(d.Get("comparison_operator").(string)),
		EvaluationPeriods:  aws.Int64(int64(d.Get("evaluation_periods").(int))),
		Tags:               getTagsIn(ctx),
		TreatMissingData:   aws.String(d.Get("treat_missing_data").(string)),
	}

	if v := d.Get("actions_enabled"); v != nil {
		apiObject.ActionsEnabled = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("alarm_actions"); ok {
		apiObject.AlarmActions = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("alarm_description"); ok {
		apiObject.AlarmDescription = aws.String(v.(string))
	}

	if v, ok := d.GetOk("datapoints_to_alarm"); ok {
		apiObject.DatapointsToAlarm = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("dimensions"); ok {
		apiObject.Dimensions = expandMetricAlarmDimensions(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("evaluate_low_sample_count_percentiles"); ok {
		apiObject.EvaluateLowSampleCountPercentile = aws.String(v.(string))
	}

	if v, ok := d.GetOk("extended_statistic"); ok {
		apiObject.ExtendedStatistic = aws.String(v.(string))
	}

	if v, ok := d.GetOk("insufficient_data_actions"); ok {
		apiObject.InsufficientDataActions = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("metric_name"); ok {
		apiObject.MetricName = aws.String(v.(string))
	}

	if v := d.Get("metric_query"); v != nil {
		apiObject.Metrics = expandMetricAlarmMetrics(v.(*schema.Set))
	}

	if v, ok := d.GetOk("namespace"); ok {
		apiObject.Namespace = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ok_actions"); ok {
		apiObject.OKActions = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("period"); ok {
		apiObject.Period = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("statistic"); ok {
		apiObject.Statistic = aws.String(v.(string))
	}

	if v, ok := d.GetOk("threshold_metric_id"); ok {
		apiObject.ThresholdMetricId = aws.String(v.(string))
	} else {
		apiObject.Threshold = aws.Float64(d.Get("threshold").(float64))
	}

	if v, ok := d.GetOk("unit"); ok {
		apiObject.Unit = aws.String(v.(string))
	}

	return apiObject
}

func flattenMetricAlarmDimensions(dims []*cloudwatch.Dimension) map[string]interface{} {
	flatDims := make(map[string]interface{})
	for _, d := range dims {
		flatDims[aws.StringValue(d.Name)] = aws.StringValue(d.Value)
	}
	return flatDims
}

func flattenMetricAlarmMetrics(metrics []*cloudwatch.MetricDataQuery) []map[string]interface{} {
	metricQueries := make([]map[string]interface{}, 0)
	for _, mq := range metrics {
		metricQuery := map[string]interface{}{
			"account_id":  aws.StringValue(mq.AccountId),
			"expression":  aws.StringValue(mq.Expression),
			"id":          aws.StringValue(mq.Id),
			"label":       aws.StringValue(mq.Label),
			"return_data": aws.BoolValue(mq.ReturnData),
		}
		if mq.MetricStat != nil {
			metric := flattenMetricAlarmMetricsMetricStat(mq.MetricStat)
			metricQuery["metric"] = []interface{}{metric}
		}
		if mq.Period != nil {
			metricQuery["period"] = aws.Int64Value(mq.Period)
		}
		metricQueries = append(metricQueries, metricQuery)
	}

	return metricQueries
}

func flattenMetricAlarmMetricsMetricStat(ms *cloudwatch.MetricStat) map[string]interface{} {
	msm := ms.Metric
	metric := map[string]interface{}{
		"metric_name": aws.StringValue(msm.MetricName),
		"namespace":   aws.StringValue(msm.Namespace),
		"period":      int(aws.Int64Value(ms.Period)),
		"stat":        aws.StringValue(ms.Stat),
		"unit":        aws.StringValue(ms.Unit),
		"dimensions":  flattenMetricAlarmDimensions(msm.Dimensions),
	}

	return metric
}

func expandMetricAlarmMetrics(v *schema.Set) []*cloudwatch.MetricDataQuery {
	var metrics []*cloudwatch.MetricDataQuery

	for _, v := range v.List() {
		metricQueryResource := v.(map[string]interface{})
		id := metricQueryResource["id"].(string)
		if id == "" {
			continue
		}
		metricQuery := cloudwatch.MetricDataQuery{
			Id: aws.String(id),
		}
		if v, ok := metricQueryResource["expression"]; ok && v.(string) != "" {
			metricQuery.Expression = aws.String(v.(string))
		}
		if v, ok := metricQueryResource["label"]; ok && v.(string) != "" {
			metricQuery.Label = aws.String(v.(string))
		}
		if v, ok := metricQueryResource["return_data"]; ok {
			metricQuery.ReturnData = aws.Bool(v.(bool))
		}
		if v := metricQueryResource["metric"]; v != nil && len(v.([]interface{})) > 0 {
			metricQuery.MetricStat = expandMetricAlarmMetricsMetric(v.([]interface{}))
		}
		if v, ok := metricQueryResource["period"]; ok && v.(int) != 0 {
			metricQuery.Period = aws.Int64(int64(v.(int)))
		}
		if v, ok := metricQueryResource["account_id"]; ok && v.(string) != "" {
			metricQuery.AccountId = aws.String(v.(string))
		}
		metrics = append(metrics, &metricQuery)
	}
	return metrics
}

func expandMetricAlarmMetricsMetric(v []interface{}) *cloudwatch.MetricStat {
	metricResource := v[0].(map[string]interface{})
	metric := cloudwatch.Metric{
		MetricName: aws.String(metricResource["metric_name"].(string)),
	}
	metricStat := cloudwatch.MetricStat{
		Metric: &metric,
		Stat:   aws.String(metricResource["stat"].(string)),
	}
	if v, ok := metricResource["namespace"]; ok && v.(string) != "" {
		metric.Namespace = aws.String(v.(string))
	}
	if v, ok := metricResource["period"]; ok {
		metricStat.Period = aws.Int64(int64(v.(int)))
	}
	if v, ok := metricResource["unit"]; ok && v.(string) != "" {
		metricStat.Unit = aws.String(v.(string))
	}
	if v, ok := metricResource["dimensions"]; ok {
		metric.Dimensions = expandMetricAlarmDimensions(v.(map[string]interface{}))
	}

	return &metricStat
}

func expandMetricAlarmDimensions(dims map[string]interface{}) []*cloudwatch.Dimension {
	var dimensions []*cloudwatch.Dimension
	for k, v := range dims {
		dimensions = append(dimensions, &cloudwatch.Dimension{
			Name:  aws.String(k),
			Value: aws.String(v.(string)),
		})
	}
	return dimensions
}
