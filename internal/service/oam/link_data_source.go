// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package oam

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_oam_link", name="Link")
// @Tags(identifierAttribute="arn")
func dataSourceLink() *schema.Resource {
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

func dataSourceLinkRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ObservabilityAccessManagerClient(ctx)

	linkIdentifier := d.Get("link_identifier").(string)
	out, err := findLinkByID(ctx, conn, linkIdentifier)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ObservabilityAccessManager Link (%s): %s", linkIdentifier, err)
	}

	arn := aws.ToString(out.Arn)
	d.SetId(arn)
	d.Set(names.AttrARN, arn)
	d.Set("label", out.Label)
	d.Set("label_template", out.LabelTemplate)
	d.Set("link_configuration", flattenLinkConfiguration(out.LinkConfiguration))
	d.Set("link_id", out.Id)
	d.Set("resource_types", out.ResourceTypes)
	d.Set("sink_arn", out.SinkArn)

	return diags
}
