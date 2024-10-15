// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_backup_framework", name="Framework")
func dataSourceFramework() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceFrameworkRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"control": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"input_parameter": {
							Type:     schema.TypeList,
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
						names.AttrName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrScope: {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"compliance_resource_ids": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"compliance_resource_types": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									names.AttrTags: tftags.TagsSchemaComputed(),
								},
							},
						},
					},
				},
			},
			names.AttrCreationTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"deployment_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceFrameworkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).BackupClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	name := d.Get(names.AttrName).(string)
	output, err := findFrameworkByName(ctx, conn, name)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Backup Framework (%s): %s", name, err)
	}

	d.SetId(name)
	d.Set(names.AttrARN, output.FrameworkArn)
	if err := d.Set("control", flattenFrameworkControls(ctx, output.FrameworkControls)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting control: %s", err)
	}
	d.Set(names.AttrCreationTime, output.CreationTime.Format(time.RFC3339))
	d.Set("deployment_status", output.DeploymentStatus)
	d.Set(names.AttrDescription, output.FrameworkDescription)
	d.Set(names.AttrName, output.FrameworkName)
	d.Set(names.AttrStatus, output.FrameworkStatus)

	tags, err := listTags(ctx, conn, d.Get(names.AttrARN).(string))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for Backup Framework (%s): %s", d.Id(), err)
	}

	if err := d.Set(names.AttrTags, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
