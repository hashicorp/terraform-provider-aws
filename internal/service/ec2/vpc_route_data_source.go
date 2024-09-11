// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_route", name="Route")
func dataSourceRoute() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceRouteRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"route_table_id": {
				Type:     schema.TypeString,
				Required: true,
			},

			///
			// Destinations.
			///
			"destination_cidr_block": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"destination_ipv6_cidr_block": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"destination_prefix_list_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			//
			// Targets.
			//
			"carrier_gateway_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"core_network_arn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"egress_only_gateway_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"gateway_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrInstanceID: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"local_gateway_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"nat_gateway_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrNetworkInterfaceID: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrTransitGatewayID: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"vpc_peering_connection_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func dataSourceRouteRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	routeTableID := d.Get("route_table_id").(string)

	routeTable, err := findRouteTableByID(ctx, conn, routeTableID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route Table (%s): %s", routeTableID, err)
	}

	routes := []awstypes.Route{}

	for _, r := range routeTable.Routes {
		if r.Origin == awstypes.RouteOriginEnableVgwRoutePropagation {
			continue
		}

		if r.DestinationPrefixListId != nil && strings.HasPrefix(aws.ToString(r.GatewayId), "vpce-") {
			// Skipping because VPC endpoint routes are handled separately
			// See aws_vpc_endpoint
			continue
		}

		if v, ok := d.GetOk("destination_cidr_block"); ok && aws.ToString(r.DestinationCidrBlock) != v.(string) {
			continue
		}

		if v, ok := d.GetOk("destination_ipv6_cidr_block"); ok && aws.ToString(r.DestinationIpv6CidrBlock) != v.(string) {
			continue
		}

		if v, ok := d.GetOk("destination_prefix_list_id"); ok && aws.ToString(r.DestinationPrefixListId) != v.(string) {
			continue
		}

		if v, ok := d.GetOk("carrier_gateway_id"); ok && aws.ToString(r.CarrierGatewayId) != v.(string) {
			continue
		}

		if v, ok := d.GetOk("core_network_arn"); ok && aws.ToString(r.CoreNetworkArn) != v.(string) {
			continue
		}

		if v, ok := d.GetOk("egress_only_gateway_id"); ok && aws.ToString(r.EgressOnlyInternetGatewayId) != v.(string) {
			continue
		}

		if v, ok := d.GetOk("gateway_id"); ok && aws.ToString(r.GatewayId) != v.(string) {
			continue
		}

		if v, ok := d.GetOk(names.AttrInstanceID); ok && aws.ToString(r.InstanceId) != v.(string) {
			continue
		}

		if v, ok := d.GetOk("local_gateway_id"); ok && aws.ToString(r.LocalGatewayId) != v.(string) {
			continue
		}

		if v, ok := d.GetOk("nat_gateway_id"); ok && aws.ToString(r.NatGatewayId) != v.(string) {
			continue
		}

		if v, ok := d.GetOk(names.AttrNetworkInterfaceID); ok && aws.ToString(r.NetworkInterfaceId) != v.(string) {
			continue
		}

		if v, ok := d.GetOk(names.AttrTransitGatewayID); ok && aws.ToString(r.TransitGatewayId) != v.(string) {
			continue
		}

		if v, ok := d.GetOk("vpc_peering_connection_id"); ok && aws.ToString(r.VpcPeeringConnectionId) != v.(string) {
			continue
		}

		routes = append(routes, r)
	}

	if len(routes) == 0 {
		return sdkdiag.AppendErrorf(diags, "No routes matching supplied arguments found in Route Table (%s)", routeTableID)
	}

	if len(routes) > 1 {
		return sdkdiag.AppendErrorf(diags, "%d routes matched in Route Table (%s); use additional constraints to reduce matches to a single route", len(routes), routeTableID)
	}

	route := routes[0]

	if destination := aws.ToString(route.DestinationCidrBlock); destination != "" {
		d.SetId(routeCreateID(routeTableID, destination))
	} else if destination := aws.ToString(route.DestinationIpv6CidrBlock); destination != "" {
		d.SetId(routeCreateID(routeTableID, destination))
	} else if destination := aws.ToString(route.DestinationPrefixListId); destination != "" {
		d.SetId(routeCreateID(routeTableID, destination))
	}

	d.Set("carrier_gateway_id", route.CarrierGatewayId)
	d.Set("core_network_arn", route.CoreNetworkArn)
	d.Set("destination_cidr_block", route.DestinationCidrBlock)
	d.Set("destination_ipv6_cidr_block", route.DestinationIpv6CidrBlock)
	d.Set("destination_prefix_list_id", route.DestinationPrefixListId)
	d.Set("egress_only_gateway_id", route.EgressOnlyInternetGatewayId)
	d.Set("gateway_id", route.GatewayId)
	d.Set(names.AttrInstanceID, route.InstanceId)
	d.Set("local_gateway_id", route.LocalGatewayId)
	d.Set("nat_gateway_id", route.NatGatewayId)
	d.Set(names.AttrNetworkInterfaceID, route.NetworkInterfaceId)
	d.Set(names.AttrTransitGatewayID, route.TransitGatewayId)
	d.Set("vpc_peering_connection_id", route.VpcPeeringConnectionId)

	return diags
}
