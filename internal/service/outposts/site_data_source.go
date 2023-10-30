// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package outposts

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/outposts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_outposts_site")
func DataSourceSite() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSiteRead,

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"id", "name"},
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"id", "name"},
			},
		},
	}
}

func dataSourceSiteRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OutpostsConn(ctx)

	input := &outposts.ListSitesInput{}

	var results []*outposts.Site

	err := conn.ListSitesPagesWithContext(ctx, input, func(page *outposts.ListSitesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, site := range page.Sites {
			if site == nil {
				continue
			}

			if v, ok := d.GetOk("id"); ok && v.(string) != aws.StringValue(site.SiteId) {
				continue
			}

			if v, ok := d.GetOk("name"); ok && v.(string) != aws.StringValue(site.Name) {
				continue
			}

			results = append(results, site)
		}

		return !lastPage
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Outposts Sites: %s", err)
	}

	if len(results) == 0 {
		return sdkdiag.AppendErrorf(diags, "no Outposts Site found matching criteria; try different search")
	}

	if len(results) > 1 {
		return sdkdiag.AppendErrorf(diags, "multiple Outposts Sites found matching criteria; try different search")
	}

	site := results[0]

	d.SetId(aws.StringValue(site.SiteId))
	d.Set("account_id", site.AccountId)
	d.Set("description", site.Description)
	d.Set("name", site.Name)

	return diags
}
