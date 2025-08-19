// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"log"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
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

// @SDKResource("aws_sagemaker_pipeline", name="Pipeline")
// @Tags(identifierAttribute="arn")
func resourcePipeline() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePipelineCreate,
		ReadWithoutTimeout:   resourcePipelineRead,
		UpdateWithoutTimeout: resourcePipelineUpdate,
		DeleteWithoutTimeout: resourcePipelineDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"parallelism_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"max_parallel_execution_steps": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntAtLeast(1),
						},
					},
				},
			},
			"pipeline_definition": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"pipeline_definition", "pipeline_definition_s3_location"},
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 1048576),
					validation.StringIsJSON,
				),
			},
			"pipeline_definition_s3_location": {
				Type:         schema.TypeList,
				Optional:     true,
				ExactlyOneOf: []string{"pipeline_definition", "pipeline_definition_s3_location"},
				MaxItems:     1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrBucket: {
							Type:     schema.TypeString,
							Required: true,
						},
						"object_key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"version_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"pipeline_description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 3072),
			},
			"pipeline_display_name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 256),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z]([0-9A-Za-z-])*$`), "Valid characters are a-z, A-Z, 0-9, and - (hyphen)."),
				),
			},
			"pipeline_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 256),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z]([0-9A-Za-z-])*$`), "Valid characters are a-z, A-Z, 0-9, and - (hyphen)."),
				),
			},
			names.AttrRoleARN: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourcePipelineCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	name := d.Get("pipeline_name").(string)
	input := &sagemaker.CreatePipelineInput{
		ClientRequestToken:  aws.String(id.UniqueId()),
		PipelineDisplayName: aws.String(d.Get("pipeline_display_name").(string)),
		PipelineName:        aws.String(name),
		RoleArn:             aws.String(d.Get(names.AttrRoleARN).(string)),
		Tags:                getTagsIn(ctx),
	}

	if v, ok := d.GetOk("parallelism_configuration"); ok {
		input.ParallelismConfiguration = expandParallelismConfiguration(v.([]any))
	}

	if v, ok := d.GetOk("pipeline_definition"); ok {
		input.PipelineDefinition = aws.String(v.(string))
	}

	if v, ok := d.GetOk("pipeline_definition_s3_location"); ok {
		input.PipelineDefinitionS3Location = expandPipelineDefinitionS3Location(v.([]any))
	}

	if v, ok := d.GetOk("pipeline_description"); ok {
		input.PipelineDescription = aws.String(v.(string))
	}

	_, err := conn.CreatePipeline(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker AI Pipeline (%s): %s", name, err)
	}

	d.SetId(name)

	return append(diags, resourcePipelineRead(ctx, d, meta)...)
}

func resourcePipelineRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	pipeline, err := findPipelineByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SageMaker AI Pipeline (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker AI Pipeline (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, pipeline.PipelineArn)
	if err := d.Set("parallelism_configuration", flattenParallelismConfiguration(pipeline.ParallelismConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting parallelism_configuration: %s", err)
	}
	d.Set("pipeline_definition", pipeline.PipelineDefinition)
	d.Set("pipeline_description", pipeline.PipelineDescription)
	d.Set("pipeline_display_name", pipeline.PipelineDisplayName)
	d.Set("pipeline_name", pipeline.PipelineName)
	d.Set(names.AttrRoleARN, pipeline.RoleArn)

	return diags
}

func resourcePipelineUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &sagemaker.UpdatePipelineInput{
			PipelineName: aws.String(d.Id()),
		}

		if d.HasChange("parallelism_configuration") {
			input.ParallelismConfiguration = expandParallelismConfiguration(d.Get("parallelism_configuration").([]any))
		}

		if d.HasChange("pipeline_definition") {
			input.PipelineDefinition = aws.String(d.Get("pipeline_definition").(string))
		}

		if d.HasChange("pipeline_definition_s3_location") {
			input.PipelineDefinitionS3Location = expandPipelineDefinitionS3Location(d.Get("pipeline_definition_s3_location").([]any))
		}

		if d.HasChange("pipeline_description") {
			input.PipelineDescription = aws.String(d.Get("pipeline_description").(string))
		}

		if d.HasChange("pipeline_display_name") {
			input.PipelineDisplayName = aws.String(d.Get("pipeline_display_name").(string))
		}

		if d.HasChange(names.AttrRoleARN) {
			input.RoleArn = aws.String(d.Get(names.AttrRoleARN).(string))
		}

		_, err := conn.UpdatePipeline(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SageMaker AI Pipeline (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourcePipelineRead(ctx, d, meta)...)
}

func resourcePipelineDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	log.Printf("[DEBUG] Deleting SageMaker AI Pipeline: %s", d.Id())
	_, err := conn.DeletePipeline(ctx, &sagemaker.DeletePipelineInput{
		ClientRequestToken: aws.String(id.UniqueId()),
		PipelineName:       aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFound](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker AI Pipeline (%s): %s", d.Id(), err)
	}

	return diags
}

func findPipelineByName(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribePipelineOutput, error) {
	input := &sagemaker.DescribePipelineInput{
		PipelineName: aws.String(name),
	}

	output, err := conn.DescribePipeline(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFound](err) {
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

func expandPipelineDefinitionS3Location(l []any) *awstypes.PipelineDefinitionS3Location {
	if len(l) == 0 || l[0] == nil {
		return &awstypes.PipelineDefinitionS3Location{}
	}

	m := l[0].(map[string]any)

	config := &awstypes.PipelineDefinitionS3Location{
		Bucket:    aws.String(m[names.AttrBucket].(string)),
		ObjectKey: aws.String(m["object_key"].(string)),
	}

	if v, ok := m["version_id"].(string); ok && v != "" {
		config.VersionId = aws.String(v)
	}

	return config
}

func expandParallelismConfiguration(l []any) *awstypes.ParallelismConfiguration {
	if len(l) == 0 || l[0] == nil {
		return &awstypes.ParallelismConfiguration{}
	}

	m := l[0].(map[string]any)

	config := &awstypes.ParallelismConfiguration{
		MaxParallelExecutionSteps: aws.Int32(int32(m["max_parallel_execution_steps"].(int))),
	}

	return config
}

func flattenParallelismConfiguration(config *awstypes.ParallelismConfiguration) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		"max_parallel_execution_steps": aws.ToInt32(config.MaxParallelExecutionSteps),
	}

	return []map[string]any{m}
}
