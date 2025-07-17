// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package imagebuilder

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_imagebuilder_image_pipeline", name="Image Pipeline")
// @Tags
func dataSourceImagePipeline() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceImagePipelineRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"container_recipe_arn": {
				Type:     schema.TypeString,
				Computed: true,
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
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"distribution_configuration_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enhanced_image_metadata_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"image_recipe_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"image_scanning_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ecr_configuration": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"container_tags": {
										Type:     schema.TypeSet,
										Computed: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									names.AttrRepositoryName: {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"image_scanning_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"image_tests_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"image_tests_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"timeout_minutes": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			"infrastructure_configuration_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"platform": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrSchedule: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"pipeline_execution_start_condition": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrScheduleExpression: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceImagePipelineRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderClient(ctx)

	arn := d.Get(names.AttrARN).(string)
	imagePipeline, err := findImagePipelineByARN(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Image Builder Image Pipeline (%s): %s", arn, err)
	}

	d.SetId(aws.ToString(imagePipeline.Arn))
	d.Set(names.AttrARN, imagePipeline.Arn)
	d.Set("container_recipe_arn", imagePipeline.ContainerRecipeArn)
	d.Set("date_created", imagePipeline.DateCreated)
	d.Set("date_last_run", imagePipeline.DateLastRun)
	d.Set("date_next_run", imagePipeline.DateNextRun)
	d.Set("date_updated", imagePipeline.DateUpdated)
	d.Set(names.AttrDescription, imagePipeline.Description)
	d.Set("distribution_configuration_arn", imagePipeline.DistributionConfigurationArn)
	d.Set("enhanced_image_metadata_enabled", imagePipeline.EnhancedImageMetadataEnabled)
	d.Set("image_recipe_arn", imagePipeline.ImageRecipeArn)
	if imagePipeline.ImageScanningConfiguration != nil {
		if err := d.Set("image_scanning_configuration", []any{flattenImageScanningConfiguration(imagePipeline.ImageScanningConfiguration)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting image_scanning_configuration: %s", err)
		}
	} else {
		d.Set("image_scanning_configuration", nil)
	}
	if imagePipeline.ImageTestsConfiguration != nil {
		if err := d.Set("image_tests_configuration", []any{flattenImageTestsConfiguration(imagePipeline.ImageTestsConfiguration)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting image_tests_configuration: %s", err)
		}
	} else {
		d.Set("image_tests_configuration", nil)
	}
	d.Set("infrastructure_configuration_arn", imagePipeline.InfrastructureConfigurationArn)
	d.Set(names.AttrName, imagePipeline.Name)
	d.Set("platform", imagePipeline.Platform)
	if imagePipeline.Schedule != nil {
		if err := d.Set(names.AttrSchedule, []any{flattenSchedule(imagePipeline.Schedule)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting schedule: %s", err)
		}
	} else {
		d.Set(names.AttrSchedule, nil)
	}
	d.Set(names.AttrStatus, imagePipeline.Status)

	setTagsOut(ctx, imagePipeline.Tags)

	return diags
}
