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

// @SDKDataSource("aws_networkmanager_link")
func DataSourceLink() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceLinkRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bandwidth": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"download_speed": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"upload_speed": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"global_network_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"link_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrProviderName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"site_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			names.AttrType: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceLinkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerConn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	globalNetworkID := d.Get("global_network_id").(string)
	linkID := d.Get("link_id").(string)
	link, err := FindLinkByTwoPartKey(ctx, conn, globalNetworkID, linkID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Network Manager Link (%s): %s", linkID, err)
	}

	d.SetId(linkID)
	d.Set(names.AttrARN, link.LinkArn)
	if link.Bandwidth != nil {
		if err := d.Set("bandwidth", []interface{}{flattenBandwidth(link.Bandwidth)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting bandwidth: %s", err)
		}
	} else {
		d.Set("bandwidth", nil)
	}
	d.Set(names.AttrDescription, link.Description)
	d.Set("global_network_id", link.GlobalNetworkId)
	d.Set("link_id", link.LinkId)
	d.Set(names.AttrProviderName, link.Provider)
	d.Set("site_id", link.SiteId)
	d.Set(names.AttrType, link.Type)

	if err := d.Set(names.AttrTags, KeyValueTags(ctx, link.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
