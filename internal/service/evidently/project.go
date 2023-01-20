package evidently

import (
	"context"
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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceProject() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceProjectCreate,
		ReadWithoutTimeout:   resourceProjectRead,
		UpdateWithoutTimeout: resourceProjectUpdate,
		DeleteWithoutTimeout: resourceProjectDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(2 * time.Minute),
			Update: schema.DefaultTimeout(2 * time.Minute),
			Delete: schema.DefaultTimeout(2 * time.Minute),
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
	conn := meta.(*conns.AWSClient).EvidentlyConn()
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

	log.Printf("[DEBUG] Creating CloudWatch Evidently Project: %s", input)
	output, err := conn.CreateProjectWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating CloudWatch Evidently Project (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.Project.Name))

	if _, err := waitProjectCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("waiting for CloudWatch Evidently Project (%s) creation: %s", d.Id(), err)
	}

	return resourceProjectRead(ctx, d, meta)
}

func resourceProjectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EvidentlyConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	project, err := FindProjectByNameOrARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudWatch Evidently Project (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading CloudWatch Evidently Project (%s): %s", d.Id(), err)
	}

	if err := d.Set("data_delivery", flattenDataDelivery(project.DataDelivery)); err != nil {
		return diag.Errorf("setting data_delivery: %s", err)
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

	tags := KeyValueTags(project.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("setting tags_all: %s", err)
	}

	return nil
}

func resourceProjectUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EvidentlyConn()

	// Project has 2 update APIs
	// UpdateProjectWithContext: Updates the description of an existing project.
	// UpdateProjectDataDeliveryWithContext: Updates the data storage options for this project.

	if d.HasChanges("description") {
		_, err := conn.UpdateProjectWithContext(ctx, &cloudwatchevidently.UpdateProjectInput{
			Description: aws.String(d.Get("description").(string)),
			Project:     aws.String(d.Id()),
		})

		if err != nil {
			return diag.Errorf("updating CloudWatch Evidently Project (%s): %s", d.Id(), err)
		}

		if _, err := waitProjectUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.Errorf("waiting for CloudWatch Evidently Project (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChange("data_delivery") {
		input := &cloudwatchevidently.UpdateProjectDataDeliveryInput{
			Project: aws.String(d.Id()),
		}

		dataDelivery := d.Get("data_delivery").([]interface{})

		tfMap, ok := dataDelivery[0].(map[string]interface{})

		if !ok {
			return diag.Errorf("updating Project (%s)", d.Id())
		}

		// You can't specify both cloudWatchLogs and s3Destination in the same operation.
		if v, ok := tfMap["cloudwatch_logs"]; ok && len(v.([]interface{})) > 0 {
			input.CloudWatchLogs = expandCloudWatchLogs(v.([]interface{}))
		}

		if v, ok := tfMap["s3_destination"]; ok && len(v.([]interface{})) > 0 {
			input.S3Destination = expandS3Destination(v.([]interface{}))
		}

		_, err := conn.UpdateProjectDataDeliveryWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("updating CloudWatch Evidently Project (%s) data delivery: %s", d.Id(), err)
		}

		if _, err := waitProjectUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return diag.Errorf("waiting for CloudWatch Evidently Project (%s) update: %s", d.Id(), err)
		}
	}

	// updates to tags
	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return diag.Errorf("updating tags: %s", err)
		}
	}

	return resourceProjectRead(ctx, d, meta)
}

func resourceProjectDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EvidentlyConn()

	log.Printf("[DEBUG] Deleting CloudWatch Evidently Project: %s", d.Id())
	_, err := conn.DeleteProjectWithContext(ctx, &cloudwatchevidently.DeleteProjectInput{
		Project: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, cloudwatchevidently.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting CloudWatch Evidently Project (%s): %s", d.Id(), err)
	}

	if _, err := waitProjectDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.Errorf("waiting for CloudWatch Evidently Project (%s) deletion: %s", d.Id(), err)
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
