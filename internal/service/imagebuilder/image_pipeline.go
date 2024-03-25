// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package imagebuilder

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/imagebuilder"
	awstypes "github.com/aws/aws-sdk-go-v2/service/imagebuilder/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_imagebuilder_image_pipeline", name="Image Pipeline")
// @Tags(identifierAttribute="id")
func ResourceImagePipeline() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceImagePipelineCreate,
		ReadWithoutTimeout:   resourceImagePipelineRead,
		UpdateWithoutTimeout: resourceImagePipelineUpdate,
		DeleteWithoutTimeout: resourceImagePipelineDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"container_recipe_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^arn:aws[^:]*:imagebuilder:[^:]+:(?:\d{12}|aws):container-recipe/[0-9a-z_-]+/\d+\.\d+\.\d+$`), "valid container recipe ARN must be provided"),
				ExactlyOneOf: []string{"container_recipe_arn", "image_recipe_arn"},
			},
			"date_created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"date_last_run": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"date_next_run": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"date_updated": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"distribution_configuration_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^arn:aws[^:]*:imagebuilder:[^:]+:(?:\d{12}|aws):distribution-configuration/[0-9a-z_-]+$`), "valid distribution configuration ARN must be provided"),
			},
			"enhanced_image_metadata_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"image_recipe_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^arn:aws[^:]*:imagebuilder:[^:]+:(?:\d{12}|aws):image-recipe/[0-9a-z_-]+/\d+\.\d+\.\d+$`), "valid image recipe ARN must be provided"),
				ExactlyOneOf: []string{"container_recipe_arn", "image_recipe_arn"},
			},
			"image_scanning_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ecr_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"container_tags": {
										Type:     schema.TypeSet,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"repository_name": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"image_scanning_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"image_tests_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"image_tests_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"timeout_minutes": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      720,
							ValidateFunc: validation.IntBetween(60, 1440),
						},
					},
				},
			},
			"infrastructure_configuration_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^arn:aws[^:]*:imagebuilder:[^:]+:(?:\d{12}|aws):infrastructure-configuration/[0-9a-z_-]+$`), "valid infrastructure configuration ARN must be provided"),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile("^[0-9A-Za-z_-][0-9A-Za-z_ -]{1,126}[0-9A-Za-z_-]$"), "valid name must be provided"),
			},
			"platform": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"schedule": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"pipeline_execution_start_condition": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          string(awstypes.PipelineExecutionStartConditionExpressionMatchAndDependencyUpdatesAvailable),
							ValidateDiagFunc: enum.Validate[awstypes.PipelineExecutionStartCondition](),
						},
						"schedule_expression": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 1024),
						},
						"timezone": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(3, 100),
								validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z]{2,}(?:\/[0-9a-zA-z_+-]+)*`), "")),
						},
					},
				},
			},
			"status": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          string(awstypes.PipelineStatusEnabled),
				ValidateDiagFunc: enum.Validate[awstypes.PipelineStatus](),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceImagePipelineCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderClient(ctx)

	input := &imagebuilder.CreateImagePipelineInput{
		ClientToken:                  aws.String(id.UniqueId()),
		EnhancedImageMetadataEnabled: aws.Bool(d.Get("enhanced_image_metadata_enabled").(bool)),
		Tags:                         getTagsIn(ctx),
	}

	if v, ok := d.GetOk("container_recipe_arn"); ok {
		input.ContainerRecipeArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("distribution_configuration_arn"); ok {
		input.DistributionConfigurationArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("image_recipe_arn"); ok {
		input.ImageRecipeArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("image_scanning_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ImageScanningConfiguration = expandImageScanningConfiguration(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("image_tests_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ImageTestsConfiguration = expandImageTestConfiguration(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("infrastructure_configuration_arn"); ok {
		input.InfrastructureConfigurationArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("name"); ok {
		input.Name = aws.String(v.(string))
	}

	if v, ok := d.GetOk("schedule"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Schedule = expandPipelineSchedule(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("status"); ok {
		input.Status = awstypes.PipelineStatus(v.(string))
	}

	output, err := conn.CreateImagePipeline(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Image Builder Image Pipeline: %s", err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "creating Image Builder Image Pipeline: empty response")
	}

	d.SetId(aws.ToString(output.ImagePipelineArn))

	return append(diags, resourceImagePipelineRead(ctx, d, meta)...)
}

func resourceImagePipelineRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderClient(ctx)

	input := &imagebuilder.GetImagePipelineInput{
		ImagePipelineArn: aws.String(d.Id()),
	}

	output, err := conn.GetImagePipeline(ctx, input)

	if !d.IsNewResource() && errs.MessageContains(err, ResourceNotFoundException, "cannot be found") {
		log.Printf("[WARN] Image Builder Image Pipeline (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Image Builder Image Pipeline (%s): %s", d.Id(), err)
	}

	if output == nil || output.ImagePipeline == nil {
		return sdkdiag.AppendErrorf(diags, "getting Image Builder Image Pipeline (%s): empty response", d.Id())
	}

	imagePipeline := output.ImagePipeline

	d.Set("arn", imagePipeline.Arn)
	d.Set("container_recipe_arn", imagePipeline.ContainerRecipeArn)
	d.Set("date_created", imagePipeline.DateCreated)
	d.Set("date_last_run", imagePipeline.DateLastRun)
	d.Set("date_next_run", imagePipeline.DateNextRun)
	d.Set("date_updated", imagePipeline.DateUpdated)
	d.Set("description", imagePipeline.Description)
	d.Set("distribution_configuration_arn", imagePipeline.DistributionConfigurationArn)
	d.Set("enhanced_image_metadata_enabled", imagePipeline.EnhancedImageMetadataEnabled)
	d.Set("image_recipe_arn", imagePipeline.ImageRecipeArn)
	if imagePipeline.ImageScanningConfiguration != nil {
		d.Set("image_scanning_configuration", []interface{}{flattenImageScanningConfiguration(imagePipeline.ImageScanningConfiguration)})
	} else {
		d.Set("image_scanning_configuration", nil)
	}
	if imagePipeline.ImageTestsConfiguration != nil {
		d.Set("image_tests_configuration", []interface{}{flattenImageTestsConfiguration(imagePipeline.ImageTestsConfiguration)})
	} else {
		d.Set("image_tests_configuration", nil)
	}
	d.Set("infrastructure_configuration_arn", imagePipeline.InfrastructureConfigurationArn)
	d.Set("name", imagePipeline.Name)
	d.Set("platform", imagePipeline.Platform)
	if imagePipeline.Schedule != nil {
		d.Set("schedule", []interface{}{flattenSchedule(imagePipeline.Schedule)})
	} else {
		d.Set("schedule", nil)
	}
	d.Set("status", imagePipeline.Status)

	setTagsOut(ctx, imagePipeline.Tags)

	return diags
}

func resourceImagePipelineUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderClient(ctx)

	if d.HasChanges(
		"description",
		"distribution_configuration_arn",
		"enhanced_image_metadata_enabled",
		"image_scanning_configuration",
		"image_tests_configuration",
		"infrastructure_configuration_arn",
		"schedule",
		"status",
	) {
		input := &imagebuilder.UpdateImagePipelineInput{
			ClientToken:                  aws.String(id.UniqueId()),
			EnhancedImageMetadataEnabled: aws.Bool(d.Get("enhanced_image_metadata_enabled").(bool)),
			ImagePipelineArn:             aws.String(d.Id()),
		}

		if v, ok := d.GetOk("container_recipe_arn"); ok {
			input.ContainerRecipeArn = aws.String(v.(string))
		}

		if v, ok := d.GetOk("description"); ok {
			input.Description = aws.String(v.(string))
		}

		if v, ok := d.GetOk("distribution_configuration_arn"); ok {
			input.DistributionConfigurationArn = aws.String(v.(string))
		}

		if v, ok := d.GetOk("image_recipe_arn"); ok {
			input.ImageRecipeArn = aws.String(v.(string))
		}

		if v, ok := d.GetOk("image_scanning_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.ImageScanningConfiguration = expandImageScanningConfiguration(v.([]interface{})[0].(map[string]interface{}))
		}

		if v, ok := d.GetOk("image_tests_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.ImageTestsConfiguration = expandImageTestConfiguration(v.([]interface{})[0].(map[string]interface{}))
		}

		if v, ok := d.GetOk("infrastructure_configuration_arn"); ok {
			input.InfrastructureConfigurationArn = aws.String(v.(string))
		}

		if v, ok := d.GetOk("schedule"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.Schedule = expandPipelineSchedule(v.([]interface{})[0].(map[string]interface{}))
		}

		if v, ok := d.GetOk("status"); ok {
			input.Status = awstypes.PipelineStatus(v.(string))
		}

		_, err := conn.UpdateImagePipeline(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Image Builder Image Pipeline (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceImagePipelineRead(ctx, d, meta)...)
}

func resourceImagePipelineDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderClient(ctx)

	input := &imagebuilder.DeleteImagePipelineInput{
		ImagePipelineArn: aws.String(d.Id()),
	}

	_, err := conn.DeleteImagePipeline(ctx, input)

	if errs.MessageContains(err, ResourceNotFoundException, "cannot be found") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Image Builder Image Pipeline (%s): %s", d.Id(), err)
	}

	return diags
}

func expandImageScanningConfiguration(tfMap map[string]interface{}) *awstypes.ImageScanningConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ImageScanningConfiguration{}

	if v, ok := tfMap["image_scanning_enabled"].(bool); ok {
		apiObject.ImageScanningEnabled = aws.Bool(v)
	}

	if v, ok := tfMap["ecr_configuration"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.EcrConfiguration = expandECRConfiguration(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandECRConfiguration(tfMap map[string]interface{}) *awstypes.EcrConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.EcrConfiguration{}

	if v, ok := tfMap["container_tags"].(*schema.Set); ok {
		apiObject.ContainerTags = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["repository_name"].(string); ok {
		apiObject.RepositoryName = aws.String(v)
	}

	return apiObject
}

func expandImageTestConfiguration(tfMap map[string]interface{}) *awstypes.ImageTestsConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ImageTestsConfiguration{}

	if v, ok := tfMap["image_tests_enabled"].(bool); ok {
		apiObject.ImageTestsEnabled = aws.Bool(v)
	}

	if v, ok := tfMap["timeout_minutes"].(int); ok && v != 0 {
		apiObject.TimeoutMinutes = aws.Int32(int32(v))
	}

	return apiObject
}

func expandPipelineSchedule(tfMap map[string]interface{}) *awstypes.Schedule {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.Schedule{}

	if v, ok := tfMap["pipeline_execution_start_condition"].(string); ok && v != "" {
		apiObject.PipelineExecutionStartCondition = awstypes.PipelineExecutionStartCondition(v)
	}

	if v, ok := tfMap["schedule_expression"].(string); ok && v != "" {
		apiObject.ScheduleExpression = aws.String(v)
	}

	if v, ok := tfMap["timezone"].(string); ok && v != "" {
		apiObject.Timezone = aws.String(v)
	}

	return apiObject
}

func flattenImageScanningConfiguration(apiObject *awstypes.ImageScanningConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ImageScanningEnabled; v != nil {
		tfMap["image_scanning_enabled"] = aws.ToBool(v)
	}

	if v := apiObject.EcrConfiguration; v != nil {
		tfMap["ecr_configuration"] = []interface{}{flattenECRConfiguration(v)}
	}

	return tfMap
}

func flattenECRConfiguration(apiObject *awstypes.EcrConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.RepositoryName; v != nil {
		tfMap["repository_name"] = aws.ToString(v)
	}

	if v := apiObject.ContainerTags; v != nil {
		tfMap["container_tags"] = aws.StringSlice(v)
	}

	return tfMap
}

func flattenImageTestsConfiguration(apiObject *awstypes.ImageTestsConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.ImageTestsEnabled; v != nil {
		tfMap["image_tests_enabled"] = aws.ToBool(v)
	}

	if v := apiObject.TimeoutMinutes; v != nil {
		tfMap["timeout_minutes"] = aws.ToInt32(v)
	}

	return tfMap
}

func flattenSchedule(apiObject *awstypes.Schedule) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	tfMap["pipeline_execution_start_condition"] = string(apiObject.PipelineExecutionStartCondition)

	if v := apiObject.ScheduleExpression; v != nil {
		tfMap["schedule_expression"] = aws.ToString(v)
	}

	if v := apiObject.Timezone; v != nil {
		tfMap["timezone"] = aws.ToString(v)
	}

	return tfMap
}
