// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKDataSource("aws_ec2_transit_gateway_route_table")
func DataSourceTransitGatewayRouteTable() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceTransitGatewayRouteTableRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_association_route_table": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"default_propagation_route_table": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"filter": CustomFiltersSchema(),
			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"transit_gateway_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceTransitGatewayRouteTableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &ec2.DescribeTransitGatewayRouteTablesInput{}

	input.Filters = append(input.Filters, BuildCustomFilterList(
		d.Get("filter").(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		input.Filters = nil
	}

	if v, ok := d.GetOk("id"); ok {
		input.TransitGatewayRouteTableIds = aws.StringSlice([]string{v.(string)})
	}

	transitGatewayRouteTable, err := FindTransitGatewayRouteTable(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("EC2 Transit Gateway Route Table", err))
	}

	d.SetId(aws.StringValue(transitGatewayRouteTable.TransitGatewayRouteTableId))
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("transit-gateway-route-table/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("default_association_route_table", transitGatewayRouteTable.DefaultAssociationRouteTable)
	d.Set("default_propagation_route_table", transitGatewayRouteTable.DefaultPropagationRouteTable)
	d.Set("transit_gateway_id", transitGatewayRouteTable.TransitGatewayId)

	if err := d.Set("tags", KeyValueTags(ctx, transitGatewayRouteTable.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
