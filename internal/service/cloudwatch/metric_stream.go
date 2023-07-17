// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudwatch

import (
	"context"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudwatch_metric_stream", name="Metric Alarm")
// @Tags(identifierAttribute="arn")
func ResourceMetricStream() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMetricStreamCreate,
		ReadWithoutTimeout:   resourceMetricStreamRead,
		UpdateWithoutTimeout: resourceMetricStreamUpdate,
		DeleteWithoutTimeout: resourceMetricStreamDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(1 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
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
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"metric_names": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(1, 255),
							},
						},
						"namespace": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
					},
				},
				ConflictsWith: []string{"include_filter"},
			},
			"firehose_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"include_filter": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"metric_names": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(1, 255),
							},
						},
						"namespace": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 255),
						},
					},
				},
				ConflictsWith: []string{"exclude_filter"},
			},
			"include_linked_accounts_metrics": {
				Type:     schema.TypeBool,
				Optional: true,
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
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceMetricStreamCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CloudWatchConn(ctx)

	name := create.Name(d.Get("name").(string), d.Get("name_prefix").(string))
	input := &cloudwatch.PutMetricStreamInput{
		FirehoseArn:                  aws.String(d.Get("firehose_arn").(string)),
		IncludeLinkedAccountsMetrics: aws.Bool(d.Get("include_linked_accounts_metrics").(bool)),
		Name:                         aws.String(name),
		OutputFormat:                 aws.String(d.Get("output_format").(string)),
		RoleArn:                      aws.String(d.Get("role_arn").(string)),
		Tags:                         getTagsIn(ctx),
	}

	if v, ok := d.GetOk("exclude_filter"); ok && v.(*schema.Set).Len() > 0 {
		input.ExcludeFilters = expandMetricStreamFilters(v.(*schema.Set))
	}

	if v, ok := d.GetOk("include_filter"); ok && v.(*schema.Set).Len() > 0 {
		input.IncludeFilters = expandMetricStreamFilters(v.(*schema.Set))
	}

	if v, ok := d.GetOk("statistics_configuration"); ok && v.(*schema.Set).Len() > 0 {
		input.StatisticsConfigurations = expandMetricStreamStatisticsConfigurations(v.(*schema.Set))
	}

	output, err := conn.PutMetricStreamWithContext(ctx, input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
		input.Tags = nil

		output, err = conn.PutMetricStreamWithContext(ctx, input)
	}

	if err != nil {
		return diag.Errorf("creating CloudWatch Metric Stream (%s): %s", name, err)
	}

	d.SetId(name)

	if _, err := waitMetricStreamRunning(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("waiting for CloudWatch Metric Stream (%s) create: %s", d.Id(), err)
	}

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		err := createTags(ctx, conn, aws.StringValue(output.Arn), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
			return resourceMetricStreamRead(ctx, d, meta)
		}

		if err != nil {
			return diag.Errorf("setting CloudWatch Metric Stream (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceMetricStreamRead(ctx, d, meta)
}

func resourceMetricStreamRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CloudWatchConn(ctx)

	output, err := FindMetricStreamByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudWatch Metric Stream (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading CloudWatch Metric Stream (%s): %s", d.Id(), err)
	}

	d.Set("arn", output.Arn)
	d.Set("creation_date", output.CreationDate.Format(time.RFC3339))
	d.Set("firehose_arn", output.FirehoseArn)
	d.Set("include_linked_accounts_metrics", output.IncludeLinkedAccountsMetrics)
	d.Set("last_update_date", output.CreationDate.Format(time.RFC3339))
	d.Set("name", output.Name)
	d.Set("name_prefix", create.NamePrefixFromName(aws.StringValue(output.Name)))
	d.Set("output_format", output.OutputFormat)
	d.Set("role_arn", output.RoleArn)
	d.Set("state", output.State)

	if output.IncludeFilters != nil {
		if err := d.Set("include_filter", flattenMetricStreamFilters(output.IncludeFilters)); err != nil {
			return diag.Errorf("setting include_filter: %s", err)
		}
	}

	if output.ExcludeFilters != nil {
		if err := d.Set("exclude_filter", flattenMetricStreamFilters(output.ExcludeFilters)); err != nil {
			return diag.Errorf("setting exclude_filter: %s", err)
		}
	}

	if output.StatisticsConfigurations != nil {
		if err := d.Set("statistics_configuration", flattenMetricStreamStatisticsConfigurations(output.StatisticsConfigurations)); err != nil {
			return diag.Errorf("setting statistics_configuration: %s", err)
		}
	}

	return nil
}

func resourceMetricStreamUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CloudWatchConn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &cloudwatch.PutMetricStreamInput{
			FirehoseArn:                  aws.String(d.Get("firehose_arn").(string)),
			IncludeLinkedAccountsMetrics: aws.Bool(d.Get("include_linked_accounts_metrics").(bool)),
			Name:                         aws.String(d.Id()),
			OutputFormat:                 aws.String(d.Get("output_format").(string)),
			RoleArn:                      aws.String(d.Get("role_arn").(string)),
		}

		if v, ok := d.GetOk("exclude_filter"); ok && v.(*schema.Set).Len() > 0 {
			input.ExcludeFilters = expandMetricStreamFilters(v.(*schema.Set))
		}

		if v, ok := d.GetOk("include_filter"); ok && v.(*schema.Set).Len() > 0 {
			input.IncludeFilters = expandMetricStreamFilters(v.(*schema.Set))
		}

		if v, ok := d.GetOk("statistics_configuration"); ok && v.(*schema.Set).Len() > 0 {
			input.StatisticsConfigurations = expandMetricStreamStatisticsConfigurations(v.(*schema.Set))
		}

		_, err := conn.PutMetricStreamWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("updating CloudWatch Metric Stream (%s): %s", d.Id(), err)
		}

		if _, err := waitMetricStreamRunning(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.Errorf("waiting for CloudWatch Metric Stream (%s) update: %s", d.Id(), err)
		}
	}

	return resourceMetricStreamRead(ctx, d, meta)
}

func resourceMetricStreamDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CloudWatchConn(ctx)

	log.Printf("[INFO] Deleting CloudWatch Metric Stream: %s", d.Id())
	_, err := conn.DeleteMetricStreamWithContext(ctx, &cloudwatch.DeleteMetricStreamInput{
		Name: aws.String(d.Id()),
	})

	if err != nil {
		return diag.Errorf("deleting CloudWatch Metric Stream (%s): %s", d.Id(), err)
	}

	if _, err := waitMetricStreamDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.Errorf("waiting for CloudWatch Metric Stream (%s) delete: %s", d.Id(), err)
	}

	return nil
}

func FindMetricStreamByName(ctx context.Context, conn *cloudwatch.CloudWatch, name string) (*cloudwatch.GetMetricStreamOutput, error) {
	input := &cloudwatch.GetMetricStreamInput{
		Name: aws.String(name),
	}

	output, err := conn.GetMetricStreamWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, cloudwatch.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func statusMetricStream(ctx context.Context, conn *cloudwatch.CloudWatch, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindMetricStreamByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

const (
	metricStreamStateRunning = "running"
	metricStreamStateStopped = "stopped"
)

func waitMetricStreamDeleted(ctx context.Context, conn *cloudwatch.CloudWatch, name string, timeout time.Duration) (*cloudwatch.GetMetricStreamOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{metricStreamStateRunning, metricStreamStateStopped},
		Target:  []string{},
		Refresh: statusMetricStream(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*cloudwatch.GetMetricStreamOutput); ok {
		return output, err
	}

	return nil, err
}

func waitMetricStreamRunning(ctx context.Context, conn *cloudwatch.CloudWatch, name string, timeout time.Duration) (*cloudwatch.GetMetricStreamOutput, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{metricStreamStateStopped},
		Target:  []string{metricStreamStateRunning},
		Refresh: statusMetricStream(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*cloudwatch.GetMetricStreamOutput); ok {
		return output, err
	}

	return nil, err
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

		if v, ok := mFilter["metric_names"].(*schema.Set); ok && v.Len() > 0 {
			filter.MetricNames = flex.ExpandStringSet(v)
		}
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
			stage["metric_names"] = aws.StringValueSlice(bd.MetricNames)
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
			configuration.AdditionalStatistics = flex.ExpandStringSet(v)
		}

		if v, ok := mConfiguration["include_metric"].(*schema.Set); ok && v.Len() > 0 {
			configuration.IncludeMetrics = expandMetricStreamStatisticsConfigurationsIncludeMetrics(v)
		}

		configurations = append(configurations, configuration)
	}

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
