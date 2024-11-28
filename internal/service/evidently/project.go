// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package evidently

import (
	"context"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/evidently"
	awstypes "github.com/aws/aws-sdk-go-v2/service/evidently/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_evidently_project", name="Project")
// @Tags(identifierAttribute="arn")
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
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreatedTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"data_delivery": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrCloudWatchLogs: {
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
											validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_./-]+$`), "must be a valid CloudWatch Log Group name"),
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
									names.AttrBucket: {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(3, 63),
											validation.StringMatch(regexache.MustCompile(`^[0-9a-z][0-9a-z-]*[0-9a-z]$`), "must be a valid Bucket name"),
										),
									},
									names.AttrPrefix: {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(1, 1024),
											validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_!.*'()/-]*$`), "must be a valid prefix name"),
										),
									},
								},
							},
						},
					},
				},
			},
			names.AttrDescription: {
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
			names.AttrLastUpdatedTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"launch_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 127),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_.-]*$`), "alphanumeric and can contain hyphens, underscores, and periods"),
				),
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceProjectCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EvidentlyClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &evidently.CreateProjectInput{
		Name: aws.String(name),
		Tags: getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("data_delivery"); ok && len(v.([]interface{})) > 0 {
		input.DataDelivery = expandDataDelivery(v.([]interface{}))
	}

	output, err := conn.CreateProject(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudWatch Evidently Project (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Project.Name))

	if _, err := waitProjectCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for CloudWatch Evidently Project (%s) creation: %s", d.Id(), err)
	}

	return append(diags, resourceProjectRead(ctx, d, meta)...)
}

func resourceProjectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EvidentlyClient(ctx)

	project, err := FindProjectByNameOrARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CloudWatch Evidently Project (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudWatch Evidently Project (%s): %s", d.Id(), err)
	}

	if err := d.Set("data_delivery", flattenDataDelivery(project.DataDelivery)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting data_delivery: %s", err)
	}

	d.Set("active_experiment_count", project.ActiveExperimentCount)
	d.Set("active_launch_count", project.ActiveLaunchCount)
	d.Set(names.AttrARN, project.Arn)
	d.Set(names.AttrCreatedTime, aws.ToTime(project.CreatedTime).Format(time.RFC3339))
	d.Set(names.AttrDescription, project.Description)
	d.Set("experiment_count", project.ExperimentCount)
	d.Set("feature_count", project.FeatureCount)
	d.Set(names.AttrLastUpdatedTime, aws.ToTime(project.LastUpdatedTime).Format(time.RFC3339))
	d.Set("launch_count", project.LaunchCount)
	d.Set(names.AttrName, project.Name)
	d.Set(names.AttrStatus, project.Status)

	setTagsOut(ctx, project.Tags)

	return diags
}

func resourceProjectUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EvidentlyClient(ctx)

	// Project has 2 update APIs
	// UpdateProjectWithContext: Updates the description of an existing project.
	// UpdateProjectDataDeliveryWithContext: Updates the data storage options for this project.

	if d.HasChanges(names.AttrDescription) {
		_, err := conn.UpdateProject(ctx, &evidently.UpdateProjectInput{
			Description: aws.String(d.Get(names.AttrDescription).(string)),
			Project:     aws.String(d.Id()),
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating CloudWatch Evidently Project (%s): %s", d.Id(), err)
		}

		if _, err := waitProjectUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for CloudWatch Evidently Project (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChange("data_delivery") {
		input := &evidently.UpdateProjectDataDeliveryInput{
			Project: aws.String(d.Id()),
		}

		dataDelivery := d.Get("data_delivery").([]interface{})

		tfMap, ok := dataDelivery[0].(map[string]interface{})

		if !ok {
			return sdkdiag.AppendErrorf(diags, "updating Project (%s)", d.Id())
		}

		// You can't specify both cloudWatchLogs and s3Destination in the same operation.
		if v, ok := tfMap[names.AttrCloudWatchLogs]; ok && len(v.([]interface{})) > 0 {
			input.CloudWatchLogs = expandCloudWatchLogs(v.([]interface{}))
		}

		if v, ok := tfMap["s3_destination"]; ok && len(v.([]interface{})) > 0 {
			input.S3Destination = expandS3Destination(v.([]interface{}))
		}

		_, err := conn.UpdateProjectDataDelivery(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating CloudWatch Evidently Project (%s) data delivery: %s", d.Id(), err)
		}

		if _, err := waitProjectUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for CloudWatch Evidently Project (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceProjectRead(ctx, d, meta)...)
}

func resourceProjectDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EvidentlyClient(ctx)

	log.Printf("[DEBUG] Deleting CloudWatch Evidently Project: %s", d.Id())
	_, err := conn.DeleteProject(ctx, &evidently.DeleteProjectInput{
		Project: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting CloudWatch Evidently Project (%s): %s", d.Id(), err)
	}

	if _, err := waitProjectDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for CloudWatch Evidently Project (%s) deletion: %s", d.Id(), err)
	}

	return diags
}

func expandDataDelivery(dataDelivery []interface{}) *awstypes.ProjectDataDeliveryConfig {
	if len(dataDelivery) == 0 || dataDelivery[0] == nil {
		return nil
	}

	tfMap, ok := dataDelivery[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &awstypes.ProjectDataDeliveryConfig{}

	if v, ok := tfMap[names.AttrCloudWatchLogs]; ok && len(v.([]interface{})) > 0 {
		result.CloudWatchLogs = expandCloudWatchLogs(v.([]interface{}))
	}

	if v, ok := tfMap["s3_destination"]; ok && len(v.([]interface{})) > 0 {
		result.S3Destination = expandS3Destination(v.([]interface{}))
	}

	return result
}

func expandCloudWatchLogs(cloudWatchLogs []interface{}) *awstypes.CloudWatchLogsDestinationConfig {
	if len(cloudWatchLogs) == 0 || cloudWatchLogs[0] == nil {
		return nil
	}

	tfMap, ok := cloudWatchLogs[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &awstypes.CloudWatchLogsDestinationConfig{}

	if v, ok := tfMap["log_group"].(string); ok && v != "" {
		result.LogGroup = aws.String(v)
	}

	return result
}

func expandS3Destination(s3Destination []interface{}) *awstypes.S3DestinationConfig {
	if len(s3Destination) == 0 || s3Destination[0] == nil {
		return nil
	}

	tfMap, ok := s3Destination[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &awstypes.S3DestinationConfig{}

	if v, ok := tfMap[names.AttrBucket].(string); ok && v != "" {
		result.Bucket = aws.String(v)
	}

	if v, ok := tfMap[names.AttrPrefix].(string); ok && v != "" {
		result.Prefix = aws.String(v)
	}

	return result
}

func flattenDataDelivery(dataDelivery *awstypes.ProjectDataDelivery) []interface{} {
	if dataDelivery == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{}

	if dataDelivery.CloudWatchLogs != nil {
		values[names.AttrCloudWatchLogs] = flattenCloudWatchLogs(dataDelivery.CloudWatchLogs)
	}

	if dataDelivery.S3Destination != nil {
		values["s3_destination"] = flattenS3Destination(dataDelivery.S3Destination)
	}

	return []interface{}{values}
}

func flattenCloudWatchLogs(cloudWatchLogs *awstypes.CloudWatchLogsDestination) []interface{} {
	if cloudWatchLogs == nil || cloudWatchLogs.LogGroup == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{}

	if cloudWatchLogs.LogGroup != nil {
		values["log_group"] = aws.ToString(cloudWatchLogs.LogGroup)
	}

	return []interface{}{values}
}

func flattenS3Destination(s3Destination *awstypes.S3Destination) []interface{} {
	if s3Destination == nil || s3Destination.Bucket == nil {
		return []interface{}{}
	}

	values := map[string]interface{}{}

	if s3Destination.Bucket != nil {
		values[names.AttrBucket] = aws.ToString(s3Destination.Bucket)
	}

	if s3Destination.Prefix != nil {
		values[names.AttrPrefix] = aws.ToString(s3Destination.Prefix)
	}

	return []interface{}{values}
}
