// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_ec2_transit_gateway_route_table_routes")
func DataSourceTransitGatewayRouteTableRoutes() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceTransitGatewayRouteTableRoutesRead,

		Schema: map[string]*schema.Schema{
			"filter": customRequiredFiltersSchema(),
			"routes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"destination_cidr_block": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"prefix_list_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrState: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"transit_gateway_route_table_announcement_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrType: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"transit_gateway_route_table_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.NoZeroValues,
			},
		},
	}
}

func dataSourceTransitGatewayRouteTableRoutesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn(ctx)

	tgwRouteTableID := d.Get("transit_gateway_route_table_id").(string)
	input := &ec2.SearchTransitGatewayRoutesInput{
		Filters:                    newCustomFilterList(d.Get("filter").(*schema.Set)),
		TransitGatewayRouteTableId: aws.String(tgwRouteTableID),
	}

	output, err := FindTransitGatewayRoutes(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Transit Gateway Route Table (%s) Routes: %s", tgwRouteTableID, err)
	}

	d.SetId(tgwRouteTableID)

	routes := []interface{}{}
	for _, route := range output {
		routes = append(routes, map[string]interface{}{
			"destination_cidr_block": aws.StringValue(route.DestinationCidrBlock),
			"prefix_list_id":         aws.StringValue(route.PrefixListId),
			names.AttrState:          aws.StringValue(route.State),
			"transit_gateway_route_table_announcement_id": aws.StringValue(route.TransitGatewayRouteTableAnnouncementId),
			names.AttrType: aws.StringValue(route.Type),
		})
	}

	if err := d.Set("routes", routes); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting routes: %s", err)
	}

	return diags
}
