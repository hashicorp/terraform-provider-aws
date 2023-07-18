// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @SDKDataSource("aws_vpc_ipam_pools")
func DataSourceIPAMPools() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceIPAMPoolsRead,

		Schema: map[string]*schema.Schema{
			"filter": DataSourceFiltersSchema(),
			"ipam_pools": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"address_family": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"allocation_default_netmask_length": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"allocation_max_netmask_length": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"allocation_min_netmask_length": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"allocation_resource_tags": tftags.TagsSchemaComputed(),
						"arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"auto_import": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"aws_service": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"ipam_scope_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ipam_scope_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ipam_pool_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"locale": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"publicly_advertisable": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"pool_depth": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"source_ipam_pool_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"state": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"tags": tftags.TagsSchemaComputed(),
					},
				},
			},
		},
	}
}

func dataSourceIPAMPoolsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &ec2.DescribeIpamPoolsInput{}

	input.Filters = append(input.Filters, BuildFiltersDataSource(
		d.Get("filter").(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		input.Filters = nil
	}

	pools, err := FindIPAMPools(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IPAM Pools: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("ipam_pools", flattenIPAMPools(ctx, pools, ignoreTagsConfig))

	return diags
}

func flattenIPAMPools(ctx context.Context, c []*ec2.IpamPool, ignoreTagsConfig *tftags.IgnoreConfig) []interface{} {
	pools := []interface{}{}
	for _, pool := range c {
		pools = append(pools, flattenIPAMPool(ctx, pool, ignoreTagsConfig))
	}
	return pools
}

func flattenIPAMPool(ctx context.Context, p *ec2.IpamPool, ignoreTagsConfig *tftags.IgnoreConfig) map[string]interface{} {
	pool := make(map[string]interface{})

	pool["address_family"] = aws.StringValue(p.AddressFamily)
	pool["allocation_default_netmask_length"] = aws.Int64Value(p.AllocationDefaultNetmaskLength)
	pool["allocation_max_netmask_length"] = aws.Int64Value(p.AllocationMaxNetmaskLength)
	pool["allocation_min_netmask_length"] = aws.Int64Value(p.AllocationMinNetmaskLength)
	pool["allocation_resource_tags"] = KeyValueTags(ctx, tagsFromIPAMAllocationTags(p.AllocationResourceTags)).Map()
	pool["arn"] = aws.StringValue(p.IpamPoolArn)
	pool["auto_import"] = aws.BoolValue(p.AutoImport)
	pool["aws_service"] = aws.StringValue(p.AwsService)
	pool["description"] = aws.StringValue(p.Description)
	pool["ipam_scope_id"] = strings.Split(aws.StringValue(p.IpamScopeArn), "/")[1]
	pool["ipam_scope_type"] = aws.StringValue(p.IpamScopeType)
	pool["locale"] = aws.StringValue(p.Locale)
	pool["pool_depth"] = aws.Int64Value(p.PoolDepth)
	pool["publicly_advertisable"] = aws.BoolValue(p.PubliclyAdvertisable)
	pool["source_ipam_pool_id"] = aws.StringValue(p.SourceIpamPoolId)
	pool["state"] = aws.StringValue(p.State)

	if v := p.Tags; v != nil {
		pool["tags"] = KeyValueTags(ctx, v).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()
	}

	return pool
}
