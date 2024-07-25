// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ec2_public_ipv4_pool", name="Public IPv4 Pool")
func dataSourcePublicIPv4Pool() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourcePublicIPv4PoolRead,

		Schema: map[string]*schema.Schema{
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"network_border_group": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"pool_address_ranges": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"available_address_count": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"first_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"last_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"pool_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			"total_address_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"total_available_address_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourcePublicIPv4PoolRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	poolID := d.Get("pool_id").(string)
	pool, err := findPublicIPv4PoolByID(ctx, conn, poolID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Public IPv4 Pool (%s): %s", poolID, err)
	}

	d.SetId(poolID)
	d.Set(names.AttrDescription, pool.Description)
	d.Set("network_border_group", pool.NetworkBorderGroup)
	if err := d.Set("pool_address_ranges", flattenPublicIPv4PoolRanges(pool.PoolAddressRanges)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting pool_address_ranges: %s", err)
	}
	if err := d.Set(names.AttrTags, keyValueTags(ctx, pool.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}
	d.Set("total_address_count", pool.TotalAddressCount)
	d.Set("total_available_address_count", pool.TotalAvailableAddressCount)

	return diags
}

func flattenPublicIPv4PoolRange(apiObject awstypes.PublicIpv4PoolRange) map[string]interface{} {
	tfMap := map[string]interface{}{}

	if v := apiObject.AddressCount; v != nil {
		tfMap["address_count"] = aws.ToInt32(v)
	}

	if v := apiObject.AvailableAddressCount; v != nil {
		tfMap["available_address_count"] = aws.ToInt32(v)
	}

	if v := apiObject.FirstAddress; v != nil {
		tfMap["first_address"] = aws.ToString(v)
	}

	if v := apiObject.LastAddress; v != nil {
		tfMap["last_address"] = aws.ToString(v)
	}

	return tfMap
}

func flattenPublicIPv4PoolRanges(apiObjects []awstypes.PublicIpv4PoolRange) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenPublicIPv4PoolRange(apiObject))
	}

	return tfList
}
