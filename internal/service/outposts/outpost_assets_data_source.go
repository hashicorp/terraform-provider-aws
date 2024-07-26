// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package outposts

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/outposts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_outposts_assets")
func DataSourceOutpostAssets() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: DataSourceOutpostAssetsRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"asset_ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"host_id_filter": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 50),
						validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z-]*$`), "must match [0-9A-Za-z-]"),
					),
				},
			},
			"status_id_filter": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 2,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.All(
						validation.StringInSlice(outposts.AssetState_Values(), false),
					),
				},
			},
		},
	}
}

func DataSourceOutpostAssetsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OutpostsConn(ctx)
	outpost_id := aws.String(d.Get(names.AttrARN).(string))

	input := &outposts.ListAssetsInput{
		OutpostIdentifier: outpost_id,
	}

	if _, ok := d.GetOk("host_id_filter"); ok {
		input.HostIdFilter = flex.ExpandStringSet(d.Get("host_id_filter").(*schema.Set))
	}

	if _, ok := d.GetOk("status_id_filter"); ok {
		input.StatusFilter = flex.ExpandStringSet(d.Get("status_id_filter").(*schema.Set))
	}

	var asset_ids []string
	err := conn.ListAssetsPagesWithContext(ctx, input, func(page *outposts.ListAssetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}
		for _, asset := range page.Assets {
			if asset == nil {
				continue
			}
			asset_ids = append(asset_ids, aws.StringValue(asset.AssetId))
		}
		return !lastPage
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Outposts Assets: %s", err)
	}
	if len(asset_ids) == 0 {
		return sdkdiag.AppendErrorf(diags, "no Outposts Assets found matching criteria; try different search")
	}

	d.SetId(aws.StringValue(outpost_id))
	d.Set("asset_ids", asset_ids)

	return diags
}
