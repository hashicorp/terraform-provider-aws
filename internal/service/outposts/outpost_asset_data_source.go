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
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_outposts_asset")
func DataSourceOutpostAsset() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: DataSourceOutpostAssetRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"asset_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"asset_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"host_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"rack_elevation": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"rack_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func DataSourceOutpostAssetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OutpostsConn(ctx)
	outpost_id := aws.String(d.Get(names.AttrARN).(string))

	input := &outposts.ListAssetsInput{
		OutpostIdentifier: outpost_id,
	}

	var results []*outposts.AssetInfo
	err := conn.ListAssetsPagesWithContext(ctx, input, func(page *outposts.ListAssetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}
		for _, asset := range page.Assets {
			if asset == nil {
				continue
			}
			if v, ok := d.GetOk("asset_id"); ok && v.(string) != aws.StringValue(asset.AssetId) {
				continue
			}
			results = append(results, asset)
		}
		return !lastPage
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Outposts Asset: %s", err)
	}
	if len(results) == 0 {
		return sdkdiag.AppendErrorf(diags, "no Outposts Asset found matching criteria; try different search")
	}

	asset := results[0]

	d.SetId(aws.StringValue(outpost_id))
	d.Set("asset_id", asset.AssetId)
	d.Set("asset_type", asset.AssetType)
	d.Set("host_id", asset.ComputeAttributes.HostId)
	d.Set("rack_elevation", asset.AssetLocation.RackElevation)
	d.Set("rack_id", asset.RackId)
	return diags
}
