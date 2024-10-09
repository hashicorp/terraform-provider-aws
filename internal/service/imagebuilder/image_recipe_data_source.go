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

// @SDKDataSource("aws_imagebuilder_image_recipe")
func DataSourceImageRecipe() *schema.Resource {
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
										Type:     schema.TypeBool,
										Computed: true,
									},
									names.AttrEncrypted: {
										Type:     schema.TypeBool,
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
			names.AttrTags: tftags.TagsSchema(),
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

func dataSourceImageRecipeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &imagebuilder.GetImageRecipeInput{}

	if v, ok := d.GetOk(names.AttrARN); ok {
		input.ImageRecipeArn = aws.String(v.(string))
	}

	output, err := conn.GetImageRecipeWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Image Builder Image Recipe (%s): %s", aws.StringValue(input.ImageRecipeArn), err)
	}

	if output == nil || output.ImageRecipe == nil {
		return sdkdiag.AppendErrorf(diags, "reading Image Builder Image Recipe (%s): empty response", aws.StringValue(input.ImageRecipeArn))
	}

	imageRecipe := output.ImageRecipe

	d.SetId(aws.StringValue(imageRecipe.Arn))
	d.Set(names.AttrARN, imageRecipe.Arn)
	d.Set("block_device_mapping", flattenInstanceBlockDeviceMappings(imageRecipe.BlockDeviceMappings))
	d.Set("component", flattenComponentConfigurations(imageRecipe.Components))
	d.Set("date_created", imageRecipe.DateCreated)
	d.Set(names.AttrDescription, imageRecipe.Description)
	d.Set(names.AttrName, imageRecipe.Name)
	d.Set(names.AttrOwner, imageRecipe.Owner)
	d.Set("parent_image", imageRecipe.ParentImage)
	d.Set("platform", imageRecipe.Platform)
	d.Set(names.AttrTags, KeyValueTags(ctx, imageRecipe.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map())

	if imageRecipe.AdditionalInstanceConfiguration != nil {
		d.Set("user_data_base64", imageRecipe.AdditionalInstanceConfiguration.UserDataOverride)
	}

	d.Set(names.AttrVersion, imageRecipe.Version)
	d.Set("working_directory", imageRecipe.WorkingDirectory)

	return diags
}
