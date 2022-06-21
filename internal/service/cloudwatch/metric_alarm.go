package cloudwatch

import (
	"errors"
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceMetricAlarm() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		Create:        resourceMetricAlarmCreate,
		Read:          resourceMetricAlarmRead,
		Update:        resourceMetricAlarmUpdate,
		Delete:        resourceMetricAlarmDelete,
		SchemaVersion: 1,
		MigrateState:  MetricAlarmMigrateState,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
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
			"evaluation_periods": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtLeast(1),
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
				Set: schema.HashString,
			},
			"alarm_description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"datapoints_to_alarm": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"dimensions": {
				Type:          schema.TypeMap,
				Optional:      true,
				ConflictsWith: []string{"metric_query"},
				Elem:          &schema.Schema{Type: schema.TypeString},
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
			"unit": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(cloudwatch.StandardUnit_Values(), false),
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
			"treat_missing_data": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      missingDataMissing,
				ValidateFunc: validation.StringInSlice(missingData_Values(), true),
			},
			"evaluate_low_sample_count_percentiles": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(lowSampleCountPercentiles_Values(), true),
			},

			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
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

func resourceMetricAlarmCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudWatchConn

	err := validMetricAlarm(d)
	if err != nil {
		return err
	}
	params := getPutMetricAlarmInput(d, meta)

	log.Printf("[DEBUG] Creating CloudWatch Metric Alarm: %#v", params)
	_, err = conn.PutMetricAlarm(&params)

	// Some partitions (i.e., ISO) may not support tag-on-create
	if params.Tags != nil && verify.CheckISOErrorTagsUnsupported(err) {
		log.Printf("[WARN] failed creating CloudWatch Metric Alarm (%s) with tags: %s. Trying create without tags.", d.Get("alarm_name").(string), err)
		params.Tags = nil

		_, err = conn.PutMetricAlarm(&params)
	}

	if err != nil {
		return fmt.Errorf("failed creating CloudWatch Metric Alarm (%s): %w", d.Get("alarm_name").(string), err)
	}

	d.SetId(d.Get("alarm_name").(string))
	log.Println("[INFO] CloudWatch Metric Alarm created")

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	// Some partitions (i.e., ISO) may not support tag-on-create, attempt tag after create
	if params.Tags == nil && len(tags) > 0 {
		resp, err := FindMetricAlarmByName(conn, d.Id())

		if err != nil {
			return fmt.Errorf("while finding metric alarm (%s): %w", d.Id(), err)
		}

		err = UpdateTags(conn, aws.StringValue(resp.AlarmArn), nil, tags)

		// If default tags only, log and continue. Otherwise, error.
		if v, ok := d.GetOk("tags"); (!ok || len(v.(map[string]interface{})) == 0) && verify.CheckISOErrorTagsUnsupported(err) {
			log.Printf("[WARN] failed adding tags after create for CloudWatch Metric Alarm (%s): %s", d.Id(), err)
			return resourceMetricAlarmRead(d, meta)
		}

		if err != nil {
			return fmt.Errorf("failed adding tags after create for CloudWatch Metric Alarm (%s): %w", d.Id(), err)
		}
	}

	return resourceMetricAlarmRead(d, meta)
}

func resourceMetricAlarmRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudWatchConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	resp, err := FindMetricAlarmByName(conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		names.LogNotFoundRemoveState(names.CloudWatch, names.ErrActionReading, ResMetricAlarm, d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return names.Error(names.CloudWatch, names.ErrActionReading, ResMetricAlarm, d.Id(), err)
	}

	if !d.IsNewResource() && resp == nil {
		names.LogNotFoundRemoveState(names.CloudWatch, names.ErrActionReading, ResMetricAlarm, d.Id())
		d.SetId("")
		return nil
	}

	if resp == nil {
		return names.Error(names.CloudWatch, names.ErrActionReading, ResMetricAlarm, d.Id(), errors.New("not found after create"))
	}

	log.Printf("[DEBUG] Reading CloudWatch Metric Alarm: %s", d.Id())

	d.Set("actions_enabled", resp.ActionsEnabled)

	if err := d.Set("alarm_actions", flex.FlattenStringSet(resp.AlarmActions)); err != nil {
		return fmt.Errorf("error setting Alarm Actions: %w", err)
	}
	arn := aws.StringValue(resp.AlarmArn)
	d.Set("alarm_description", resp.AlarmDescription)
	d.Set("alarm_name", resp.AlarmName)
	d.Set("arn", arn)
	d.Set("comparison_operator", resp.ComparisonOperator)
	d.Set("datapoints_to_alarm", resp.DatapointsToAlarm)
	if err := d.Set("dimensions", flattenMetricAlarmDimensions(resp.Dimensions)); err != nil {
		return fmt.Errorf("error setting dimensions: %w", err)
	}
	d.Set("evaluation_periods", resp.EvaluationPeriods)

	if err := d.Set("insufficient_data_actions", flex.FlattenStringSet(resp.InsufficientDataActions)); err != nil {
		return fmt.Errorf("error setting Insufficient Data Actions: %w", err)
	}
	d.Set("metric_name", resp.MetricName)
	d.Set("namespace", resp.Namespace)

	if resp.Metrics != nil && len(resp.Metrics) > 0 {
		if err := d.Set("metric_query", flattenMetricAlarmMetrics(resp.Metrics)); err != nil {
			return fmt.Errorf("error setting metric_query: %w", err)
		}
	}

	if err := d.Set("ok_actions", flex.FlattenStringSet(resp.OKActions)); err != nil {
		return fmt.Errorf("error setting OK Actions: %w", err)
	}

	d.Set("period", resp.Period)
	d.Set("statistic", resp.Statistic)
	d.Set("threshold", resp.Threshold)
	d.Set("threshold_metric_id", resp.ThresholdMetricId)
	d.Set("unit", resp.Unit)
	d.Set("extended_statistic", resp.ExtendedStatistic)
	if resp.TreatMissingData != nil {
		d.Set("treat_missing_data", resp.TreatMissingData)
	} else {
		d.Set("treat_missing_data", missingDataMissing)
	}
	d.Set("evaluate_low_sample_count_percentiles", resp.EvaluateLowSampleCountPercentile)

	tags, err := ListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for CloudWatch Metric Alarm (%s): %w", arn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	// Some partitions (i.e., ISO) may not support tagging, giving error
	if verify.CheckISOErrorTagsUnsupported(err) {
		log.Printf("[WARN] failed listing tags for CloudWatch Metric Alarm (%s): %s", d.Id(), err)
		return nil
	}

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceMetricAlarmUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudWatchConn
	params := getPutMetricAlarmInput(d, meta)

	log.Printf("[DEBUG] Updating CloudWatch Metric Alarm: %#v", params)
	_, err := conn.PutMetricAlarm(&params)
	if err != nil {
		return fmt.Errorf("Updating metric alarm failed: %w", err)
	}
	log.Println("[INFO] CloudWatch Metric Alarm updated")

	arn := d.Get("arn").(string)
	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		err := UpdateTags(conn, arn, o, n)

		// Some partitions (i.e., ISO) may not support tagging, giving error
		if verify.CheckISOErrorTagsUnsupported(err) {
			log.Printf("[WARN] failed updating tags for CloudWatch Metric Alarm (%s): %s", d.Id(), err)
			return resourceMetricAlarmRead(d, meta)
		}

		if err != nil {
			return fmt.Errorf("failed updating tags for CloudWatch Metric Alarm (%s): %w", d.Id(), err)
		}
	}

	return resourceMetricAlarmRead(d, meta)
}

func resourceMetricAlarmDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudWatchConn
	params := cloudwatch.DeleteAlarmsInput{
		AlarmNames: []*string{aws.String(d.Id())},
	}

	log.Printf("[INFO] Deleting CloudWatch Metric Alarm: %s", d.Id())

	if _, err := conn.DeleteAlarms(&params); err != nil {
		if tfawserr.ErrCodeEquals(err, cloudwatch.ErrCodeResourceNotFoundException) {
			return nil
		}
		return fmt.Errorf("Error deleting CloudWatch Metric Alarm: %w", err)
	}
	log.Println("[INFO] CloudWatch Metric Alarm deleted")

	return nil
}

func getPutMetricAlarmInput(d *schema.ResourceData, meta interface{}) cloudwatch.PutMetricAlarmInput {
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	params := cloudwatch.PutMetricAlarmInput{
		AlarmName:          aws.String(d.Get("alarm_name").(string)),
		ComparisonOperator: aws.String(d.Get("comparison_operator").(string)),
		EvaluationPeriods:  aws.Int64(int64(d.Get("evaluation_periods").(int))),
		TreatMissingData:   aws.String(d.Get("treat_missing_data").(string)),
	}

	if len(tags) > 0 {
		params.Tags = Tags(tags.IgnoreAWS())
	}

	if v := d.Get("actions_enabled"); v != nil {
		params.ActionsEnabled = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("alarm_description"); ok {
		params.AlarmDescription = aws.String(v.(string))
	}

	if v, ok := d.GetOk("datapoints_to_alarm"); ok {
		params.DatapointsToAlarm = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("metric_name"); ok {
		params.MetricName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("namespace"); ok {
		params.Namespace = aws.String(v.(string))
	}
	if v, ok := d.GetOk("period"); ok {
		params.Period = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("unit"); ok {
		params.Unit = aws.String(v.(string))
	}

	if v, ok := d.GetOk("statistic"); ok {
		params.Statistic = aws.String(v.(string))
	}

	if v, ok := d.GetOk("extended_statistic"); ok {
		params.ExtendedStatistic = aws.String(v.(string))
	}

	if v, ok := d.GetOk("evaluate_low_sample_count_percentiles"); ok {
		params.EvaluateLowSampleCountPercentile = aws.String(v.(string))
	}

	if v, ok := d.GetOk("threshold_metric_id"); ok {
		params.ThresholdMetricId = aws.String(v.(string))
	} else {
		params.Threshold = aws.Float64(d.Get("threshold").(float64))
	}

	if v, ok := d.GetOk("alarm_actions"); ok {
		params.AlarmActions = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("insufficient_data_actions"); ok {
		params.InsufficientDataActions = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v := d.Get("metric_query"); v != nil {
		params.Metrics = expandMetricAlarmMetrics(v.(*schema.Set))
	}

	if v, ok := d.GetOk("ok_actions"); ok {
		params.OKActions = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("dimensions"); ok {
		params.Dimensions = expandMetricAlarmDimensions(v.(map[string]interface{}))
	}

	return params
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
