// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ec2_local_gateway_route_table")
func dataSourceLocalGatewayRouteTable() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceLocalGatewayRouteTableRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"local_gateway_route_table_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"local_gateway_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"outpost_arn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			names.AttrState: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			names.AttrTags: tftags.TagsSchemaComputed(),

			names.AttrFilter: customFiltersSchema(),
		},
	}
}

func dataSourceLocalGatewayRouteTableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	req := &ec2.DescribeLocalGatewayRouteTablesInput{}

	if v, ok := d.GetOk("local_gateway_route_table_id"); ok {
		req.LocalGatewayRouteTableIds = []string{v.(string)}
	}

	req.Filters = newAttributeFilterListV2(
		map[string]string{
			"local-gateway-id": d.Get("local_gateway_id").(string),
			"outpost-arn":      d.Get("outpost_arn").(string),
			names.AttrState:    d.Get(names.AttrState).(string),
		},
	)

	req.Filters = append(req.Filters, newTagFilterListV2(
		TagsV2(tftags.New(ctx, d.Get(names.AttrTags).(map[string]interface{}))),
	)...)

	req.Filters = append(req.Filters, newCustomFilterListV2(
		d.Get(names.AttrFilter).(*schema.Set),
	)...)
	if len(req.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		req.Filters = nil
	}

	log.Printf("[DEBUG] Reading AWS Local Gateway Route Table: %#v", req)
	resp, err := conn.DescribeLocalGatewayRouteTables(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing EC2 Local Gateway Route Tables: %s", err)
	}
	if resp == nil || len(resp.LocalGatewayRouteTables) == 0 {
		return sdkdiag.AppendErrorf(diags, "no matching Local Gateway Route Table found")
	}
	if len(resp.LocalGatewayRouteTables) > 1 {
		return sdkdiag.AppendErrorf(diags, "multiple Local Gateway Route Tables matched; use additional constraints to reduce matches to a single Local Gateway Route Table")
	}

	localgatewayroutetable := resp.LocalGatewayRouteTables[0]

	d.SetId(aws.ToString(localgatewayroutetable.LocalGatewayRouteTableId))
	d.Set("local_gateway_id", localgatewayroutetable.LocalGatewayId)
	d.Set("local_gateway_route_table_id", localgatewayroutetable.LocalGatewayRouteTableId)
	d.Set("outpost_arn", localgatewayroutetable.OutpostArn)
	d.Set(names.AttrState, localgatewayroutetable.State)

	if err := d.Set(names.AttrTags, keyValueTagsV2(ctx, localgatewayroutetable.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
