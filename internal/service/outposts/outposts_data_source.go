// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package outposts

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/outposts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_outposts_outposts", name="Outposts")
func dataSourceOutposts() *schema.Resource { // nosemgrep:ci.outposts-in-func-name
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceOutpostsRead,

		Schema: map[string]*schema.Schema{
			names.AttrARNs: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrAvailabilityZone: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"availability_zone_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrIDs: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"site_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func dataSourceOutpostsRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics { // nosemgrep:ci.outposts-in-func-name
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OutpostsClient(ctx)

	input := &outposts.ListOutpostsInput{}

	var arns, ids []string

	pages := outposts.NewListOutpostsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "listing Outposts Outposts: %s", err)
		}

		for _, outpost := range page.Outposts {
			if v, ok := d.GetOk(names.AttrAvailabilityZone); ok && v.(string) != aws.ToString(outpost.AvailabilityZone) {
				continue
			}

			if v, ok := d.GetOk("availability_zone_id"); ok && v.(string) != aws.ToString(outpost.AvailabilityZoneId) {
				continue
			}

			if v, ok := d.GetOk("site_id"); ok && v.(string) != aws.ToString(outpost.SiteId) {
				continue
			}

			if v, ok := d.GetOk(names.AttrOwnerID); ok && v.(string) != aws.ToString(outpost.OwnerId) {
				continue
			}

			arns = append(arns, aws.ToString(outpost.OutpostArn))
			ids = append(ids, aws.ToString(outpost.OutpostId))
		}
	}

	if err := d.Set(names.AttrARNs, arns); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting arns: %s", err)
	}

	if err := d.Set(names.AttrIDs, ids); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ids: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region(ctx))

	return diags
}
