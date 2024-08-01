// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmanager

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_networkmanager_site")
func DataSourceSite() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceSiteRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"global_network_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrLocation: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrAddress: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"latitude": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"longitude": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"site_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceSiteRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	globalNetworkID := d.Get("global_network_id").(string)
	siteID := d.Get("site_id").(string)
	site, err := FindSiteByTwoPartKey(ctx, conn, globalNetworkID, siteID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Network Manager Site (%s): %s", siteID, err)
	}

	d.SetId(siteID)
	d.Set(names.AttrARN, site.SiteArn)
	d.Set(names.AttrDescription, site.Description)
	d.Set("global_network_id", site.GlobalNetworkId)
	if site.Location != nil {
		if err := d.Set(names.AttrLocation, []interface{}{flattenLocation(site.Location)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting location: %s", err)
		}
	} else {
		d.Set(names.AttrLocation, nil)
	}
	d.Set("site_id", site.SiteId)

	if err := d.Set(names.AttrTags, KeyValueTags(ctx, site.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
