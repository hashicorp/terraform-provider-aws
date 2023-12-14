// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @SDKDataSource("aws_ec2_public_ipv4_pool")
func DataSourcePublicIPv4Pool() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourcePublicIPv4PoolRead,

		Schema: map[string]*schema.Schema{
			"description": {
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
			"tags": tftags.TagsSchemaComputed(),
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
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	poolID := d.Get("pool_id").(string)
	pool, err := FindPublicIPv4PoolByID(ctx, conn, poolID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Public IPv4 Pool (%s): %s", poolID, err)
	}

	d.SetId(poolID)
	d.Set("description", pool.Description)
	d.Set("network_border_group", pool.NetworkBorderGroup)
	if err := d.Set("pool_address_ranges", flattenPublicIPv4PoolRanges(pool.PoolAddressRanges)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting pool_address_ranges: %s", err)
	}
	if err := d.Set("tags", KeyValueTags(ctx, pool.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}
	d.Set("total_address_count", pool.TotalAddressCount)
	d.Set("total_available_address_count", pool.TotalAvailableAddressCount)

	return diags
}

func flattenPublicIPv4PoolRange(apiObject *ec2.PublicIpv4PoolRange) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.AddressCount; v != nil {
		tfMap["address_count"] = aws.Int64Value(v)
	}

	if v := apiObject.AvailableAddressCount; v != nil {
		tfMap["available_address_count"] = aws.Int64Value(v)
	}

	if v := apiObject.FirstAddress; v != nil {
		tfMap["first_address"] = aws.StringValue(v)
	}

	if v := apiObject.LastAddress; v != nil {
		tfMap["last_address"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenPublicIPv4PoolRanges(apiObjects []*ec2.PublicIpv4PoolRange) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenPublicIPv4PoolRange(apiObject))
	}

	return tfList
}
