package cloudwatchevidently

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchevidently"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceProject() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceProjectCreate,
		ReadContext:   resourceProjectRead,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"active_experiment_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"active_launch_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"data_delivery": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				// while there is an API for UpdateProjectDataDelivery, ForceNew because there is a bug in the service API
				// A bug in the service API for UpdateProjectDataDelivery has been reported
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cloudwatch_logs": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							// You can't specify both cloudWatchLogs and s3Destination in the same operation.
							// https://docs.aws.amazon.com/cloudwatchevidently/latest/APIReference/API_UpdateProjectDataDelivery.html
							ConflictsWith: []string{"data_delivery.0.s3_destination"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"log_group": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 512),
											validation.StringMatch(regexp.MustCompile(`^[-a-zA-Z0-9._/]+$`), "must be a valid CloudWatch Log Group name"),
										),
									},
								},
							},
						},
						"s3_destination": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							// You can't specify both cloudWatchLogs and s3Destination in the same operation.
							// https://docs.aws.amazon.com/cloudwatchevidently/latest/APIReference/API_UpdateProjectDataDelivery.html
							ConflictsWith: []string{"data_delivery.0.cloudwatch_logs"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"bucket": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(3, 63),
											validation.StringMatch(regexp.MustCompile(`^[a-z0-9][-a-z0-9]*[a-z0-9]$`), "must be a valid Bucket name"),
										),
									},
									"prefix": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 1024),
											validation.StringMatch(regexp.MustCompile(`^[-a-zA-Z0-9!_.*'()/]*$`), "must be a valid prefix name"),
										),
									},
								},
							},
						},
					},
				},
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 160),
			},
			"experiment_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"feature_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"last_updated_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"launch_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 127),
					validation.StringMatch(regexp.MustCompile(`^[-a-zA-Z0-9._]*$`), "alphanumeric and can contain hyphens, underscores, and periods"),
				),
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceProjectCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CloudWatchEvidentlyConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)

	input := &cloudwatchevidently.CreateProjectInput{
		Name: aws.String(name),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("data_delivery"); ok && len(v.([]interface{})) > 0 {
		input.DataDelivery = expandDataDelivery(v.([]interface{}))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating CloudWatch Evidently Project %s", input)
	output, err := conn.CreateProjectWithContext(ctx, input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating CloudWatch Evidently Project (%s): %w", name, err))
	}

	if output == nil {
		return diag.FromErr(fmt.Errorf("error creating CloudWatch Evidently Project (%s): empty output", name))
	}

	d.SetId(aws.StringValue(output.Project.Arn))

	return resourceProjectRead(ctx, d, meta)
}

func resourceProjectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CloudWatchEvidentlyConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	arn := d.Id()

	resp, err := conn.GetProjectWithContext(ctx, &cloudwatchevidently.GetProjectInput{
		Project: aws.String(arn),
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, cloudwatchevidently.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] CloudWatch Evidently Project (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting CloudWatch Evidently Project (%s): %w", d.Id(), err))
	}

	if resp == nil || resp.Project == nil {
		return diag.FromErr(fmt.Errorf("error getting CloudWatch Evidently Project (%s): empty response", d.Id()))
	}

	project := resp.Project

	if err := d.Set("data_delivery", flattenDataDelivery(project.DataDelivery)); err != nil {
		return diag.FromErr(err)
	}

	d.Set("active_experiment_count", project.ActiveExperimentCount)
	d.Set("active_launch_count", project.ActiveLaunchCount)
	d.Set("arn", project.Arn)
	d.Set("created_time", aws.TimeValue(project.CreatedTime).Format(time.RFC3339))
	d.Set("description", project.Description)
	d.Set("experiment_count", project.ExperimentCount)
	d.Set("feature_count", project.FeatureCount)
	d.Set("last_updated_time", aws.TimeValue(project.LastUpdatedTime).Format(time.RFC3339))
	d.Set("launch_count", project.LaunchCount)
	d.Set("name", project.Name)
	d.Set("status", project.Status)

	tags := KeyValueTags(resp.Project.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %w", err))
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags_all: %w", err))
	}

	return nil
}

func expandDataDelivery(dataDelivery []interface{}) *cloudwatchevidently.ProjectDataDeliveryConfig {
	if len(dataDelivery) == 0 || dataDelivery[0] == nil {
		return nil
	}

	tfMap, ok := dataDelivery[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &cloudwatchevidently.ProjectDataDeliveryConfig{}

	if v, ok := tfMap["cloudwatch_logs"]; ok && len(v.([]interface{})) > 0 {
		result.CloudWatchLogs = expandCloudWatchLogs(v.([]interface{}))
	}

	if v, ok := tfMap["s3_destination"]; ok && len(v.([]interface{})) > 0 {
		result.S3Destination = expandS3Destination(v.([]interface{}))
	}

	return result
}

func expandCloudWatchLogs(cloudWatchLogs []interface{}) *cloudwatchevidently.CloudWatchLogsDestinationConfig {
	if len(cloudWatchLogs) == 0 || cloudWatchLogs[0] == nil {
		return nil
	}

	tfMap, ok := cloudWatchLogs[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &cloudwatchevidently.CloudWatchLogsDestinationConfig{}

	if v, ok := tfMap["log_group"].(string); ok && v != "" {
		result.LogGroup = aws.String(v)
	}

	return result
}

func expandS3Destination(s3Destination []interface{}) *cloudwatchevidently.S3DestinationConfig {
	if len(s3Destination) == 0 || s3Destination[0] == nil {
		return nil
	}

	tfMap, ok := s3Destination[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &cloudwatchevidently.S3DestinationConfig{}

	if v, ok := tfMap["bucket"].(string); ok && v != "" {
		result.Bucket = aws.String(v)
	}

	if v, ok := tfMap["prefix"].(string); ok && v != "" {
		result.Prefix = aws.String(v)
	}

	return result
}

func flattenDataDelivery(dataDelivery *cloudwatchevidently.ProjectDataDelivery) []interface{} {
	if dataDelivery == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{}

	if dataDelivery.CloudWatchLogs != nil {
		values["cloudwatch_logs"] = flattenCloudWatchLogs(dataDelivery.CloudWatchLogs)
	}

	if dataDelivery.S3Destination != nil {
		values["s3_destination"] = flattenS3Destination(dataDelivery.S3Destination)
	}

	return []interface{}{values}
}

func flattenCloudWatchLogs(cloudWatchLogs *cloudwatchevidently.CloudWatchLogsDestination) []interface{} {
	if cloudWatchLogs == nil || cloudWatchLogs.LogGroup == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{}

	if cloudWatchLogs.LogGroup != nil {
		values["log_group"] = aws.StringValue(cloudWatchLogs.LogGroup)
	}

	return []interface{}{values}
}

func flattenS3Destination(s3Destination *cloudwatchevidently.S3Destination) []interface{} {
	if s3Destination == nil || s3Destination.Bucket == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{}

	if s3Destination.Bucket != nil {
		values["bucket"] = aws.StringValue(s3Destination.Bucket)
	}

	if s3Destination.Prefix != nil {
		values["prefix"] = aws.StringValue(s3Destination.Prefix)
	}

	return []interface{}{values}
}
