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
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2/types/nullable"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_imagebuilder_image_recipe", name="Image Recipe")
// @Tags
func dataSourceImageRecipe() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceImageRecipeRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"block_device_mapping": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrDeviceName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ebs": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrDeleteOnTermination: {
										Type:     nullable.TypeNullableBool,
										Computed: true,
									},
									names.AttrEncrypted: {
										Type:     nullable.TypeNullableBool,
										Computed: true,
									},
									names.AttrIOPS: {
										Type:     schema.TypeInt,
										Computed: true,
									},
									names.AttrKMSKeyID: {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrSnapshotID: {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrThroughput: {
										Type:     schema.TypeInt,
										Computed: true,
									},
									names.AttrVolumeSize: {
										Type:     schema.TypeInt,
										Computed: true,
									},
									names.AttrVolumeType: {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"no_device": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrVirtualName: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"component": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"component_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrParameter: {
							Type:     schema.TypeSet,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrName: {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrValue: {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"date_created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrOwner: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"parent_image": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"platform": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			"user_data_base64": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrVersion: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"working_directory": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceImageRecipeRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderClient(ctx)

	arn := d.Get(names.AttrARN).(string)
	imageRecipe, err := findImageRecipeByARN(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Image Builder Image Recipe (%s): %s", arn, err)
	}

	arn = aws.ToString(imageRecipe.Arn)
	d.SetId(arn)
	d.Set(names.AttrARN, arn)
	if err := d.Set("block_device_mapping", flattenInstanceBlockDeviceMappings(imageRecipe.BlockDeviceMappings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting block_device_mapping: %s", err)
	}
	if err := d.Set("component", flattenComponentConfigurations(imageRecipe.Components)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting component: %s", err)
	}
	d.Set("date_created", imageRecipe.DateCreated)
	d.Set(names.AttrDescription, imageRecipe.Description)
	d.Set(names.AttrName, imageRecipe.Name)
	d.Set(names.AttrOwner, imageRecipe.Owner)
	d.Set("parent_image", imageRecipe.ParentImage)
	d.Set("platform", imageRecipe.Platform)
	if imageRecipe.AdditionalInstanceConfiguration != nil {
		d.Set("user_data_base64", imageRecipe.AdditionalInstanceConfiguration.UserDataOverride)
	}
	d.Set(names.AttrVersion, imageRecipe.Version)
	d.Set("working_directory", imageRecipe.WorkingDirectory)

	setTagsOut(ctx, imageRecipe.Tags)

	return diags
}
