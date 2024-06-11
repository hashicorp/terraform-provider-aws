// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package imagebuilder

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_imagebuilder_image")
func DataSourceImage() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceImageRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"build_version_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"container_recipe_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"date_created": {
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
			"os_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"output_resources": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"amis": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrAccountID: {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrDescription: {
										Type:     schema.TypeString,
										Computed: true,
									},
									"image": {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrName: {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrRegion: {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"containers": {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"image_uris": {
										Type:     schema.TypeSet,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									names.AttrRegion: {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"platform": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			names.AttrVersion: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceImageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderConn(ctx)

	input := &imagebuilder.GetImageInput{}

	if v, ok := d.GetOk(names.AttrARN); ok {
		input.ImageBuildVersionArn = aws.String(v.(string))
	}

	output, err := conn.GetImageWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Image Builder Image: %s", err)
	}

	if output == nil || output.Image == nil {
		return sdkdiag.AppendErrorf(diags, "getting Image Builder Image: empty response")
	}

	image := output.Image

	d.SetId(aws.StringValue(image.Arn))

	// To prevent Terraform errors, only reset arn if not configured.
	// The configured ARN may contain x.x.x wildcards while the API returns
	// the full build version #.#.#/# suffix.
	if _, ok := d.GetOk(names.AttrARN); !ok {
		d.Set(names.AttrARN, image.Arn)
	}

	d.Set("build_version_arn", image.Arn)
	d.Set("date_created", image.DateCreated)

	if image.ContainerRecipe != nil {
		d.Set("container_recipe_arn", image.ContainerRecipe.Arn)
	}

	if image.DistributionConfiguration != nil {
		d.Set("distribution_configuration_arn", image.DistributionConfiguration.Arn)
	}

	d.Set("enhanced_image_metadata_enabled", image.EnhancedImageMetadataEnabled)

	if image.ImageRecipe != nil {
		d.Set("image_recipe_arn", image.ImageRecipe.Arn)
	}

	if image.ImageScanningConfiguration != nil {
		d.Set("image_scanning_configuration", []interface{}{flattenImageScanningConfiguration(image.ImageScanningConfiguration)})
	} else {
		d.Set("image_scanning_configuration", nil)
	}

	if image.ImageTestsConfiguration != nil {
		d.Set("image_tests_configuration", []interface{}{flattenImageTestsConfiguration(image.ImageTestsConfiguration)})
	} else {
		d.Set("image_tests_configuration", nil)
	}

	if image.InfrastructureConfiguration != nil {
		d.Set("infrastructure_configuration_arn", image.InfrastructureConfiguration.Arn)
	}

	d.Set(names.AttrName, image.Name)
	d.Set("platform", image.Platform)
	d.Set("os_version", image.OsVersion)

	if image.OutputResources != nil {
		d.Set("output_resources", []interface{}{flattenOutputResources(image.OutputResources)})
	} else {
		d.Set("output_resources", nil)
	}

	d.Set(names.AttrTags, KeyValueTags(ctx, image.Tags).IgnoreAWS().IgnoreConfig(meta.(*conns.AWSClient).IgnoreTagsConfig).Map())
	d.Set(names.AttrVersion, image.Version)

	return diags
}
