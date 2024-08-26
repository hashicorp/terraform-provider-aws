// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package oam

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_oam_link")
func DataSourceLink() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceLinkRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"label": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"label_template": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"link_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"log_group_configuration": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrFilter: {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"metric_configuration": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrFilter: {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"link_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"link_identifier": {
				Type:     schema.TypeString,
				Required: true,
			},
			"resource_types": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"sink_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

const (
	DSNameLink = "Link Data Source"
)

func dataSourceLinkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ObservabilityAccessManagerClient(ctx)

	linkIdentifier := d.Get("link_identifier").(string)

	out, err := findLinkByID(ctx, conn, linkIdentifier)
	if err != nil {
		return create.AppendDiagError(diags, names.ObservabilityAccessManager, create.ErrActionReading, DSNameLink, linkIdentifier, err)
	}

	d.SetId(aws.ToString(out.Arn))

	d.Set(names.AttrARN, out.Arn)
	d.Set("label", out.Label)
	d.Set("label_template", out.LabelTemplate)
	d.Set("link_configuration", flattenLinkConfiguration(out.LinkConfiguration))
	d.Set("link_id", out.Id)
	d.Set("resource_types", flex.FlattenStringValueList(out.ResourceTypes))
	d.Set("sink_arn", out.SinkArn)

	tags, err := listTags(ctx, conn, d.Id())
	if err != nil {
		return create.AppendDiagError(diags, names.ObservabilityAccessManager, create.ErrActionReading, DSNameLink, d.Id(), err)
	}

	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	if err := d.Set(names.AttrTags, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return create.AppendDiagError(diags, names.ObservabilityAccessManager, create.ErrActionSetting, DSNameLink, d.Id(), err)
	}

	return nil
}
