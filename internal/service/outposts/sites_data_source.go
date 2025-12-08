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

// @SDKDataSource("aws_outposts_sites", name="Sites")
func dataSourceSites() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSitesRead,

		Schema: map[string]*schema.Schema{
			names.AttrIDs: {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceSitesRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OutpostsClient(ctx)

	input := &outposts.ListSitesInput{}

	var ids []string

	pages := outposts.NewListSitesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "listing Outposts Sites: %s", err)
		}

		for _, site := range page.Sites {
			ids = append(ids, aws.ToString(site.SiteId))
		}
	}

	if err := d.Set(names.AttrIDs, ids); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ids: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region(ctx))

	return diags
}
