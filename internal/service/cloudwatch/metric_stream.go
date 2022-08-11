package cloudwatch

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceMetricStream() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMetricStreamCreate,
		ReadContext:   resourceMetricStreamRead,
		UpdateContext: resourceMetricStreamCreate,
		DeleteContext: resourceMetricStreamDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(MetricStreamReadyTimeout),
			Delete: schema.DefaultTimeout(MetricStreamDeleteTimeout),
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"creation_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"exclude_filter": {
				Type:          schema.TypeSet,
				Optional:      true,
				ConflictsWith: []string{"include_filter"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"namespace": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
					},
				},
			},
			"firehose_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"include_filter": {
				Type:          schema.TypeSet,
				Optional:      true,
				ConflictsWith: []string{"exclude_filter"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"namespace": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
					},
				},
			},
			"last_update_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validateMetricStreamName,
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validateMetricStreamName,
			},
			"output_format": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"statistics_configuration": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"additional_statistics": {
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
								ValidateFunc: validation.All(
									validation.Any(
										validation.StringMatch(
											regexp.MustCompile(`(^IQM$)|(^(p|tc|tm|ts|wm)(100|\d{1,2})(\.\d{0,10})?$)|(^[ou]\d+(\.\d*)?$)`),
											"invalid statistic, see: https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/Statistics-definitions.html",
										),
										validation.StringMatch(
											regexp.MustCompile(`^(TM|TC|TS|WM)\(((((\d{1,2})(\.\d{0,10})?|100(\.0{0,10})?)%)?:((\d{1,2})(\.\d{0,10})?|100(\.0{0,10})?)%|((\d{1,2})(\.\d{0,10})?|100(\.0{0,10})?)%:(((\d{1,2})(\.\d{0,10})?|100(\.0{0,10})?)%)?)\)|(TM|TC|TS|WM|PR)\(((\d+(\.\d{0,10})?|(\d+(\.\d{0,10})?[Ee][+-]?\d+)):((\d+(\.\d{0,10})?|(\d+(\.\d{0,10})?[Ee][+-]?\d+)))?|((\d+(\.\d{0,10})?|(\d+(\.\d{0,10})?[Ee][+-]?\d+)))?:(\d+(\.\d{0,10})?|(\d+(\.\d{0,10})?[Ee][+-]?\d+)))\)$`),
											"invalid statistic, see: https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/Statistics-definitions.html",
										),
									),
									validation.StringDoesNotMatch(
										regexp.MustCompile(`^p0(\.0{0,10})?|p100(\.\d{0,10})?$`),
										"invalid statistic, see: https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/Statistics-definitions.html",
									),
								),
							},
						},
						"include_metric": {
							Type:     schema.TypeSet,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"metric_name": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 255),
									},
									"namespace": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 255),
									},
								},
							},
						},
					},
				},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
	}
}

func resourceMetricStreamCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CloudWatchConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := create.Name(d.Get("name").(string), d.Get("name_prefix").(string))

	params := cloudwatch.PutMetricStreamInput{
		Name:         aws.String(name),
		FirehoseArn:  aws.String(d.Get("firehose_arn").(string)),
		RoleArn:      aws.String(d.Get("role_arn").(string)),
		OutputFormat: aws.String(d.Get("output_format").(string)),
	}

	if len(tags) > 0 {
		params.Tags = Tags(tags.IgnoreAWS())
	}

	if v, ok := d.GetOk("include_filter"); ok && v.(*schema.Set).Len() > 0 {
		params.IncludeFilters = expandMetricStreamFilters(v.(*schema.Set))
	}

	if v, ok := d.GetOk("exclude_filter"); ok && v.(*schema.Set).Len() > 0 {
		params.ExcludeFilters = expandMetricStreamFilters(v.(*schema.Set))
	}

	if v, ok := d.GetOk("statistics_configuration"); ok && v.(*schema.Set).Len() > 0 {
		params.StatisticsConfigurations = expandMetricStreamStatisticsConfigurations(v.(*schema.Set))
	}

	log.Printf("[DEBUG] Putting CloudWatch Metric Stream: %#v", params)
	output, err := conn.PutMetricStreamWithContext(ctx, &params)

	// Some partitions (i.e., ISO) may not support tag-on-create
	if params.Tags != nil && verify.CheckISOErrorTagsUnsupported(conn.PartitionID, err) {
		log.Printf("[WARN] failed creating CloudWatch Metric Stream (%s) with tags: %s. Trying create without tags.", name, err)
		params.Tags = nil

		output, err = conn.PutMetricStreamWithContext(ctx, &params)
	}

	if err != nil {
		return diag.Errorf("failed creating CloudWatch Metric Stream (%s): %s", name, err)
	}

	d.SetId(name)
	log.Println("[INFO] CloudWatch Metric Stream put finished")

	// Some partitions (i.e., ISO) may not support tag-on-create, attempt tag after create
	if params.Tags == nil && len(tags) > 0 {
		err := UpdateTags(conn, aws.StringValue(output.Arn), nil, tags)

		// If default tags only, log and continue. Otherwise, error.
		if v, ok := d.GetOk("tags"); (!ok || len(v.(map[string]interface{})) == 0) && verify.CheckISOErrorTagsUnsupported(conn.PartitionID, err) {
			log.Printf("[WARN] failed adding tags after create for CloudWatch Metric Stream (%s): %s", d.Id(), err)
			return resourceMetricStreamRead(ctx, d, meta)
		}

		if err != nil {
			return diag.Errorf("failed adding tags after create for CloudWatch Metric Stream (%s): %s", d.Id(), err)
		}
	}

	return resourceMetricStreamRead(ctx, d, meta)
}

func resourceMetricStreamRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CloudWatchConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	output, err := WaitMetricStreamReady(ctx, conn, d.Id())

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, cloudwatch.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] CloudWatch Metric Stream (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting CloudWatch Metric Stream (%s): %w", d.Id(), err))
	}

	if output == nil {
		return diag.FromErr(fmt.Errorf("error getting CloudWatch Metric Stream (%s): empty response", d.Id()))
	}

	d.Set("arn", output.Arn)
	d.Set("creation_date", output.CreationDate.Format(time.RFC3339))
	d.Set("firehose_arn", output.FirehoseArn)
	d.Set("last_update_date", output.CreationDate.Format(time.RFC3339))
	d.Set("name", output.Name)
	d.Set("name_prefix", create.NamePrefixFromName(aws.StringValue(output.Name)))
	d.Set("output_format", output.OutputFormat)
	d.Set("role_arn", output.RoleArn)
	d.Set("state", output.State)

	if output.IncludeFilters != nil {
		if err := d.Set("include_filter", flattenMetricStreamFilters(output.IncludeFilters)); err != nil {
			return diag.FromErr(fmt.Errorf("error setting include_filter error: %w", err))
		}
	}

	if output.ExcludeFilters != nil {
		if err := d.Set("exclude_filter", flattenMetricStreamFilters(output.ExcludeFilters)); err != nil {
			return diag.FromErr(fmt.Errorf("error setting exclude_filter error: %w", err))
		}
	}

	if output.StatisticsConfigurations != nil {
		if err := d.Set("statistics_configuration", flattenMetricStreamStatisticsConfigurations(output.StatisticsConfigurations)); err != nil {
			return diag.FromErr(fmt.Errorf("error setting statistics_configuration error: %w", err))
		}
	}

	tags, err := ListTags(conn, aws.StringValue(output.Arn))

	// Some partitions (i.e., ISO) may not support tagging, giving error
	if verify.CheckISOErrorTagsUnsupported(conn.PartitionID, err) {
		log.Printf("[WARN] failed listing tags for CloudWatch Metric Stream (%s): %s", d.Id(), err)
		return nil
	}

	if err != nil {
		return diag.Errorf("failed listing tags for CloudWatch Metric Stream (%s): %s", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %w", err))
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags_all: %w", err))
	}

	return nil
}

func resourceMetricStreamDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[INFO] Deleting CloudWatch Metric Stream %s", d.Id())
	conn := meta.(*conns.AWSClient).CloudWatchConn
	params := cloudwatch.DeleteMetricStreamInput{
		Name: aws.String(d.Id()),
	}

	if _, err := conn.DeleteMetricStreamWithContext(ctx, &params); err != nil {
		return diag.FromErr(fmt.Errorf("error deleting CloudWatch Metric Stream: %s", err))
	}

	if _, err := WaitMetricStreamDeleted(ctx, conn, d.Id()); err != nil {
		return diag.FromErr(fmt.Errorf("error while waiting for CloudWatch Metric Stream (%s) to become deleted: %w", d.Id(), err))
	}

	log.Printf("[INFO] CloudWatch Metric Stream %s deleted", d.Id())

	return nil
}

func validateMetricStreamName(v interface{}, k string) (ws []string, errors []error) {
	return validation.All(
		validation.StringLenBetween(1, 255),
		validation.StringMatch(regexp.MustCompile(`^[\-_A-Za-z0-9]*$`), "must match [\\-_A-Za-z0-9]"),
	)(v, k)
}

func expandMetricStreamFilters(s *schema.Set) []*cloudwatch.MetricStreamFilter {
	var filters []*cloudwatch.MetricStreamFilter

	for _, filterRaw := range s.List() {
		filter := &cloudwatch.MetricStreamFilter{}
		mFilter := filterRaw.(map[string]interface{})

		if v, ok := mFilter["namespace"].(string); ok && v != "" {
			filter.Namespace = aws.String(v)
		}

		filters = append(filters, filter)
	}

	return filters
}

func flattenMetricStreamFilters(s []*cloudwatch.MetricStreamFilter) []map[string]interface{} {
	filters := make([]map[string]interface{}, 0)

	for _, bd := range s {
		if bd.Namespace != nil {
			stage := make(map[string]interface{})
			stage["namespace"] = aws.StringValue(bd.Namespace)

			filters = append(filters, stage)
		}
	}

	if len(filters) > 0 {
		return filters
	}

	return nil
}

func expandMetricStreamStatisticsConfigurations(s *schema.Set) []*cloudwatch.MetricStreamStatisticsConfiguration {
	var configurations []*cloudwatch.MetricStreamStatisticsConfiguration

	for _, configurationRaw := range s.List() {
		configuration := &cloudwatch.MetricStreamStatisticsConfiguration{}
		mConfiguration := configurationRaw.(map[string]interface{})

		if v, ok := mConfiguration["additional_statistics"].(*schema.Set); ok && v.Len() > 0 {
			log.Printf("[DEBUG] CloudWatch Metric Stream StatisticsConfigurations additional_statistics: %#v", v)
			configuration.AdditionalStatistics = flex.ExpandStringSet(v)
		}

		if v, ok := mConfiguration["include_metric"].(*schema.Set); ok && v.Len() > 0 {
			log.Printf("[DEBUG] CloudWatch Metric Stream StatisticsConfigurations include_metrics: %#v", v)
			configuration.IncludeMetrics = expandMetricStreamStatisticsConfigurationsIncludeMetrics(v)
		}

		configurations = append(configurations, configuration)

	}

	log.Printf("[DEBUG] statistics_configurations: %#v", configurations)

	if len(configurations) > 0 {
		return configurations
	}

	return nil
}

func expandMetricStreamStatisticsConfigurationsIncludeMetrics(metrics *schema.Set) []*cloudwatch.MetricStreamStatisticsMetric {
	var includeMetrics []*cloudwatch.MetricStreamStatisticsMetric

	for _, metricRaw := range metrics.List() {
		metric := &cloudwatch.MetricStreamStatisticsMetric{}
		mMetric := metricRaw.(map[string]interface{})

		if v, ok := mMetric["metric_name"].(string); ok && v != "" {
			metric.MetricName = aws.String(v)
		}

		if v, ok := mMetric["namespace"].(string); ok && v != "" {
			metric.Namespace = aws.String(v)
		}

		includeMetrics = append(includeMetrics, metric)
	}

	if len(includeMetrics) > 0 {
		return includeMetrics
	}

	return nil
}

func flattenMetricStreamStatisticsConfigurations(configurations []*cloudwatch.MetricStreamStatisticsConfiguration) []map[string]interface{} {
	flatConfigurations := make([]map[string]interface{}, len(configurations))

	for i, configuration := range configurations {
		flatConfiguration := map[string]interface{}{
			"additional_statistics": flex.FlattenStringSet(configuration.AdditionalStatistics),
			"include_metric":        flattenMetricStreamStatisticsConfigurationsIncludeMetrics(configuration.IncludeMetrics),
		}

		flatConfigurations[i] = flatConfiguration
	}

	return flatConfigurations
}

func flattenMetricStreamStatisticsConfigurationsIncludeMetrics(metrics []*cloudwatch.MetricStreamStatisticsMetric) []map[string]interface{} {
	flatMetrics := make([]map[string]interface{}, len(metrics))

	for i, metric := range metrics {
		flatMetric := map[string]interface{}{
			"metric_name": aws.StringValue(metric.MetricName),
			"namespace":   aws.StringValue(metric.Namespace),
		}

		flatMetrics[i] = flatMetric
	}

	return flatMetrics
}
