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

// @SDKDataSource("aws_imagebuilder_container_recipe")
func DataSourceContainerRecipe() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceContainerRecipeRead,
		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
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
			"container_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"date_created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dockerfile_template_data": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEncrypted: {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"instance_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
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
						"image": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrKMSKeyID: {
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
			"target_repository": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrRepositoryName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"service": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
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

func dataSourceContainerRecipeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ImageBuilderConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &imagebuilder.GetContainerRecipeInput{}

	if v, ok := d.GetOk(names.AttrARN); ok {
		input.ContainerRecipeArn = aws.String(v.(string))
	}

	output, err := conn.GetContainerRecipeWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Image Builder Container Recipe (%s): %s", aws.StringValue(input.ContainerRecipeArn), err)
	}

	if output == nil || output.ContainerRecipe == nil {
		return sdkdiag.AppendErrorf(diags, "reading Image Builder Container Recipe (%s): empty response", aws.StringValue(input.ContainerRecipeArn))
	}

	containerRecipe := output.ContainerRecipe

	d.SetId(aws.StringValue(containerRecipe.Arn))
	d.Set(names.AttrARN, containerRecipe.Arn)
	d.Set("component", flattenComponentConfigurations(containerRecipe.Components))
	d.Set("container_type", containerRecipe.ContainerType)
	d.Set("date_created", containerRecipe.DateCreated)
	d.Set(names.AttrDescription, containerRecipe.Description)
	d.Set("dockerfile_template_data", containerRecipe.DockerfileTemplateData)
	d.Set(names.AttrEncrypted, containerRecipe.Encrypted)

	if containerRecipe.InstanceConfiguration != nil {
		d.Set("instance_configuration", []interface{}{flattenInstanceConfiguration(containerRecipe.InstanceConfiguration)})
	} else {
		d.Set("instance_configuration", nil)
	}

	d.Set(names.AttrKMSKeyID, containerRecipe.KmsKeyId)
	d.Set(names.AttrName, containerRecipe.Name)
	d.Set(names.AttrOwner, containerRecipe.Owner)
	d.Set("parent_image", containerRecipe.ParentImage)
	d.Set("platform", containerRecipe.Platform)
	d.Set(names.AttrTags, KeyValueTags(ctx, containerRecipe.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map())
	d.Set("target_repository", []interface{}{flattenTargetContainerRepository(containerRecipe.TargetRepository)})
	d.Set(names.AttrVersion, containerRecipe.Version)
	d.Set("working_directory", containerRecipe.WorkingDirectory)

	return diags
}
