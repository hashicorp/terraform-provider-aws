// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// IPv4 to Internet Gateway.
func TestAccVPCRoute_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var route awstypes.Route
	var routeTable awstypes.RouteTable
	resourceName := "aws_route.test"
	igwResourceName := "aws_internet_gateway.test"
	rtResourceName := "aws_route_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	destinationCidr := "10.3.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteConfig_ipv4InternetGateway(rName, destinationCidr),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRouteTableExists(ctx, rtResourceName, &routeTable),
					testAccCheckRouteTableNumberOfRoutes(&routeTable, 2),
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "core_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_id", igwResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceID, ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrNetworkInterfaceID, ""),
					resource.TestCheckResourceAttr(resourceName, "origin", string(awstypes.RouteOriginCreateRoute)),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.RouteStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrTransitGatewayID, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrVPCEndpointID, ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCRoute_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var route awstypes.Route
	resourceName := "aws_route.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	destinationCidr := "10.3.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteConfig_ipv4InternetGateway(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceRoute(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCRoute_Disappears_routeTable(t *testing.T) {
	ctx := acctest.Context(t)
	var route awstypes.Route
	resourceName := "aws_route.test"
	rtResourceName := "aws_route_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	destinationCidr := "10.3.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteConfig_ipv4InternetGateway(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceRouteTable(), rtResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCRoute_ipv6ToEgressOnlyInternetGateway(t *testing.T) {
	ctx := acctest.Context(t)
	var route awstypes.Route
	resourceName := "aws_route.test"
	eoigwResourceName := "aws_egress_only_internet_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	destinationCidr := "::/0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteConfig_ipv6EgressOnlyInternetGateway(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "core_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "egress_only_gateway_id", eoigwResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceID, ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrNetworkInterfaceID, ""),
					resource.TestCheckResourceAttr(resourceName, "origin", string(awstypes.RouteOriginCreateRoute)),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.RouteStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrTransitGatewayID, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrVPCEndpointID, ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				// Verify that expanded form of the destination CIDR causes no diff.
				Config:   testAccVPCRouteConfig_ipv6EgressOnlyInternetGateway(rName, "::0/0"),
				PlanOnly: true,
			},
		},
	})
}

func TestAccVPCRoute_ipv6ToInternetGateway(t *testing.T) {
	ctx := acctest.Context(t)
	var route awstypes.Route
	resourceName := "aws_route.test"
	igwResourceName := "aws_internet_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	destinationCidr := "::/0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteConfig_ipv6InternetGateway(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "core_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_id", igwResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceID, ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrNetworkInterfaceID, ""),
					resource.TestCheckResourceAttr(resourceName, "origin", string(awstypes.RouteOriginCreateRoute)),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.RouteStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrTransitGatewayID, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrVPCEndpointID, ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCRoute_ipv6ToInstance(t *testing.T) {
	ctx := acctest.Context(t)
	var route awstypes.Route
	resourceName := "aws_route.test"
	instanceResourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	destinationCidr := "::/0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteConfig_ipv6Instance(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "core_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrInstanceID, instanceResourceName, names.AttrID),
					acctest.CheckResourceAttrAccountID(resourceName, "instance_owner_id"),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrNetworkInterfaceID, instanceResourceName, "primary_network_interface_id"),
					resource.TestCheckResourceAttr(resourceName, "origin", string(awstypes.RouteOriginCreateRoute)),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.RouteStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrTransitGatewayID, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrVPCEndpointID, ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCRoute_IPv6ToNetworkInterface_unattached(t *testing.T) {
	ctx := acctest.Context(t)
	var route awstypes.Route
	resourceName := "aws_route.test"
	eniResourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	destinationCidr := "::/0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteConfig_ipv6NetworkInterfaceUnattached(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "core_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceID, ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrNetworkInterfaceID, eniResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "origin", string(awstypes.RouteOriginCreateRoute)),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.RouteStateBlackhole)),
					resource.TestCheckResourceAttr(resourceName, names.AttrTransitGatewayID, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrVPCEndpointID, ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCRoute_ipv6ToVPCPeeringConnection(t *testing.T) {
	ctx := acctest.Context(t)
	var route awstypes.Route
	resourceName := "aws_route.test"
	pcxResourceName := "aws_vpc_peering_connection.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	destinationCidr := "::/0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteConfig_ipv6PeeringConnection(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "core_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceID, ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrNetworkInterfaceID, ""),
					resource.TestCheckResourceAttr(resourceName, "origin", string(awstypes.RouteOriginCreateRoute)),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.RouteStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrTransitGatewayID, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrVPCEndpointID, ""),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_peering_connection_id", pcxResourceName, names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCRoute_ipv6ToVPNGateway(t *testing.T) {
	ctx := acctest.Context(t)
	var route awstypes.Route
	resourceName := "aws_route.test"
	vgwResourceName := "aws_vpn_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	destinationCidr := "::/0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteConfig_ipv6VPNGateway(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "core_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_id", vgwResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceID, ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrNetworkInterfaceID, ""),
					resource.TestCheckResourceAttr(resourceName, "origin", string(awstypes.RouteOriginCreateRoute)),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.RouteStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrTransitGatewayID, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrVPCEndpointID, ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCRoute_ipv4ToVPNGateway(t *testing.T) {
	ctx := acctest.Context(t)
	var route awstypes.Route
	resourceName := "aws_route.test"
	vgwResourceName := "aws_vpn_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	destinationCidr := "10.3.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteConfig_ipv4VPNGateway(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "core_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_id", vgwResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceID, ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrNetworkInterfaceID, ""),
					resource.TestCheckResourceAttr(resourceName, "origin", string(awstypes.RouteOriginCreateRoute)),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.RouteStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrTransitGatewayID, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrVPCEndpointID, ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCRoute_ipv4ToInstance(t *testing.T) {
	ctx := acctest.Context(t)
	var route awstypes.Route
	resourceName := "aws_route.test"
	instanceResourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	destinationCidr := "10.3.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteConfig_ipv4Instance(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "core_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrInstanceID, instanceResourceName, names.AttrID),
					acctest.CheckResourceAttrAccountID(resourceName, "instance_owner_id"),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrNetworkInterfaceID, instanceResourceName, "primary_network_interface_id"),
					resource.TestCheckResourceAttr(resourceName, "origin", string(awstypes.RouteOriginCreateRoute)),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.RouteStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrTransitGatewayID, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrVPCEndpointID, ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCRoute_IPv4ToNetworkInterface_unattached(t *testing.T) {
	ctx := acctest.Context(t)
	var route awstypes.Route
	resourceName := "aws_route.test"
	eniResourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	destinationCidr := "10.3.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteConfig_ipv4NetworkInterfaceUnattached(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "core_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceID, ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrNetworkInterfaceID, eniResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "origin", string(awstypes.RouteOriginCreateRoute)),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.RouteStateBlackhole)),
					resource.TestCheckResourceAttr(resourceName, names.AttrTransitGatewayID, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrVPCEndpointID, ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCRoute_IPv4ToNetworkInterface_attached(t *testing.T) {
	ctx := acctest.Context(t)
	var route awstypes.Route
	resourceName := "aws_route.test"
	eniResourceName := "aws_network_interface.test"
	instanceResourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	destinationCidr := "10.3.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteConfig_ipv4NetworkInterfaceAttached(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "core_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrInstanceID, instanceResourceName, names.AttrID),
					acctest.CheckResourceAttrAccountID(resourceName, "instance_owner_id"),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrNetworkInterfaceID, eniResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "origin", string(awstypes.RouteOriginCreateRoute)),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.RouteStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrTransitGatewayID, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrVPCEndpointID, ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCRoute_IPv4ToNetworkInterface_twoAttachments(t *testing.T) {
	ctx := acctest.Context(t)
	var route awstypes.Route
	resourceName := "aws_route.test"
	eni1ResourceName := "aws_network_interface.test1"
	eni2ResourceName := "aws_network_interface.test2"
	instanceResourceName := "aws_instance.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	destinationCidr := "10.3.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteConfig_ipv4NetworkInterfaceTwoAttachments(rName, destinationCidr, eni1ResourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "core_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrInstanceID, instanceResourceName, names.AttrID),
					acctest.CheckResourceAttrAccountID(resourceName, "instance_owner_id"),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrNetworkInterfaceID, eni1ResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "origin", string(awstypes.RouteOriginCreateRoute)),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.RouteStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrTransitGatewayID, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrVPCEndpointID, ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				Config: testAccVPCRouteConfig_ipv4NetworkInterfaceTwoAttachments(rName, destinationCidr, eni2ResourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "core_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrInstanceID, instanceResourceName, names.AttrID),
					acctest.CheckResourceAttrAccountID(resourceName, "instance_owner_id"),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrNetworkInterfaceID, eni2ResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "origin", string(awstypes.RouteOriginCreateRoute)),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.RouteStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrTransitGatewayID, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrVPCEndpointID, ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCRoute_ipv4ToVPCPeeringConnection(t *testing.T) {
	ctx := acctest.Context(t)
	var route awstypes.Route
	resourceName := "aws_route.test"
	pcxResourceName := "aws_vpc_peering_connection.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	destinationCidr := "10.3.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteConfig_ipv4PeeringConnection(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "core_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceID, ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrNetworkInterfaceID, ""),
					resource.TestCheckResourceAttr(resourceName, "origin", string(awstypes.RouteOriginCreateRoute)),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.RouteStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrTransitGatewayID, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrVPCEndpointID, ""),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_peering_connection_id", pcxResourceName, names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCRoute_ipv4ToNatGateway(t *testing.T) {
	ctx := acctest.Context(t)
	var route awstypes.Route
	resourceName := "aws_route.test"
	ngwResourceName := "aws_nat_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	destinationCidr := "10.3.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteConfig_ipv4NATGateway(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "core_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceID, ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "nat_gateway_id", ngwResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrNetworkInterfaceID, ""),
					resource.TestCheckResourceAttr(resourceName, "origin", string(awstypes.RouteOriginCreateRoute)),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.RouteStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrTransitGatewayID, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrVPCEndpointID, ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCRoute_ipv6ToNatGateway(t *testing.T) {
	ctx := acctest.Context(t)
	var route awstypes.Route
	resourceName := "aws_route.test"
	ngwResourceName := "aws_nat_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	destinationCidr := "64:ff9b::/96"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteConfig_ipv6NATGateway(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "core_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceID, ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "nat_gateway_id", ngwResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrNetworkInterfaceID, ""),
					resource.TestCheckResourceAttr(resourceName, "origin", string(awstypes.RouteOriginCreateRoute)),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.RouteStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrTransitGatewayID, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrVPCEndpointID, ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCRoute_doesNotCrashWithVPCEndpoint(t *testing.T) {
	ctx := acctest.Context(t)
	var route awstypes.Route
	var routeTable awstypes.RouteTable
	resourceName := "aws_route.test"
	rtResourceName := "aws_route_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteConfig_endpoint(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(ctx, rtResourceName, &routeTable),
					testAccCheckRouteTableNumberOfRoutes(&routeTable, 3),
					testAccCheckRouteExists(ctx, resourceName, &route),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCRoute_ipv4ToTransitGateway(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var route awstypes.Route
	resourceName := "aws_route.test"
	tgwResourceName := "aws_ec2_transit_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	destinationCidr := "10.3.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteConfig_ipv4TransitGateway(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "core_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceID, ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrNetworkInterfaceID, ""),
					resource.TestCheckResourceAttr(resourceName, "origin", string(awstypes.RouteOriginCreateRoute)),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.RouteStateActive)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTransitGatewayID, tgwResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrVPCEndpointID, ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCRoute_ipv6ToTransitGateway(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var route awstypes.Route
	resourceName := "aws_route.test"
	tgwResourceName := "aws_ec2_transit_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	destinationCidr := "::/0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteConfig_ipv6TransitGateway(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "core_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceID, ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrNetworkInterfaceID, ""),
					resource.TestCheckResourceAttr(resourceName, "origin", string(awstypes.RouteOriginCreateRoute)),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.RouteStateActive)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTransitGatewayID, tgwResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrVPCEndpointID, ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCRoute_ipv4ToCarrierGateway(t *testing.T) {
	ctx := acctest.Context(t)
	var route awstypes.Route
	resourceName := "aws_route.test"
	cgwResourceName := "aws_ec2_carrier_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	destinationCidr := "172.16.1.0/24"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckWavelengthZoneAvailable(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteConfig_ipv4CarrierGateway(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttrPair(resourceName, "carrier_gateway_id", cgwResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "core_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceID, ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrNetworkInterfaceID, ""),
					resource.TestCheckResourceAttr(resourceName, "origin", string(awstypes.RouteOriginCreateRoute)),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.RouteStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrTransitGatewayID, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrVPCEndpointID, ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCRoute_ipv4ToLocalGateway(t *testing.T) {
	ctx := acctest.Context(t)
	var route awstypes.Route
	resourceName := "aws_route.test"
	localGatewayDataSourceName := "data.aws_ec2_local_gateway.first"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	destinationCidr := "172.16.1.0/24"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteConfig_resourceIPv4LocalGateway(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "core_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceID, ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "local_gateway_id", localGatewayDataSourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrNetworkInterfaceID, ""),
					resource.TestCheckResourceAttr(resourceName, "origin", string(awstypes.RouteOriginCreateRoute)),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.RouteStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrTransitGatewayID, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrVPCEndpointID, ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCRoute_ipv6ToLocalGateway(t *testing.T) {
	ctx := acctest.Context(t)
	var route awstypes.Route
	resourceName := "aws_route.test"
	localGatewayDataSourceName := "data.aws_ec2_local_gateway.first"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	destinationCidr := "2002:bc9:1234:1a00::/56"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteConfig_resourceIPv6LocalGateway(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "core_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceID, ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "local_gateway_id", localGatewayDataSourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrNetworkInterfaceID, ""),
					resource.TestCheckResourceAttr(resourceName, "origin", string(awstypes.RouteOriginCreateRoute)),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.RouteStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrTransitGatewayID, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrVPCEndpointID, ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCRoute_conditionalCIDRBlock(t *testing.T) {
	ctx := acctest.Context(t)
	var route awstypes.Route
	resourceName := "aws_route.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	destinationCidr := "10.2.0.0/16"
	destinationIpv6Cidr := "::/0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteConfig_conditionalIPv4IPv6(rName, destinationCidr, destinationIpv6Cidr, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
				),
			},
			{
				Config: testAccVPCRouteConfig_conditionalIPv4IPv6(rName, destinationCidr, destinationIpv6Cidr, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", destinationIpv6Cidr),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCRoute_IPv4Update_target(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var route awstypes.Route
	resourceName := "aws_route.test"
	vgwResourceName := "aws_vpn_gateway.test"
	igwResourceName := "aws_internet_gateway.test"
	eniResourceName := "aws_network_interface.test"
	pcxResourceName := "aws_vpc_peering_connection.test"
	ngwResourceName := "aws_nat_gateway.test"
	tgwResourceName := "aws_ec2_transit_gateway.test"
	vpcEndpointResourceName := "aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	destinationCidr := "10.3.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckELBv2GatewayLoadBalancer(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID, "elasticloadbalancing"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteConfig_ipv4FlexiTarget(rName, destinationCidr, "gateway_id", vgwResourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "core_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_id", vgwResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceID, ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrNetworkInterfaceID, ""),
					resource.TestCheckResourceAttr(resourceName, "origin", string(awstypes.RouteOriginCreateRoute)),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.RouteStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrTransitGatewayID, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrVPCEndpointID, ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				Config: testAccVPCRouteConfig_ipv4FlexiTarget(rName, destinationCidr, "gateway_id", igwResourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "core_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_id", igwResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceID, ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrNetworkInterfaceID, ""),
					resource.TestCheckResourceAttr(resourceName, "origin", string(awstypes.RouteOriginCreateRoute)),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.RouteStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrTransitGatewayID, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrVPCEndpointID, ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				Config: testAccVPCRouteConfig_ipv4FlexiTarget(rName, destinationCidr, "nat_gateway_id", ngwResourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "core_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceID, ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "nat_gateway_id", ngwResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrNetworkInterfaceID, ""),
					resource.TestCheckResourceAttr(resourceName, "origin", string(awstypes.RouteOriginCreateRoute)),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.RouteStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrTransitGatewayID, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrVPCEndpointID, ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				Config: testAccVPCRouteConfig_ipv4FlexiTarget(rName, destinationCidr, names.AttrNetworkInterfaceID, eniResourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "core_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceID, ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrNetworkInterfaceID, eniResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "origin", string(awstypes.RouteOriginCreateRoute)),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.RouteStateBlackhole)),
					resource.TestCheckResourceAttr(resourceName, names.AttrTransitGatewayID, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrVPCEndpointID, ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				Config: testAccVPCRouteConfig_ipv4FlexiTarget(rName, destinationCidr, names.AttrTransitGatewayID, tgwResourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "core_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceID, ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrNetworkInterfaceID, ""),
					resource.TestCheckResourceAttr(resourceName, "origin", string(awstypes.RouteOriginCreateRoute)),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.RouteStateActive)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTransitGatewayID, tgwResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrVPCEndpointID, ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				Config: testAccVPCRouteConfig_ipv4FlexiTarget(rName, destinationCidr, names.AttrVPCEndpointID, vpcEndpointResourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "core_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceID, ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrNetworkInterfaceID, ""),
					resource.TestCheckResourceAttr(resourceName, "origin", string(awstypes.RouteOriginCreateRoute)),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.RouteStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrTransitGatewayID, ""),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVPCEndpointID, vpcEndpointResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				Config: testAccVPCRouteConfig_ipv4FlexiTarget(rName, destinationCidr, "vpc_peering_connection_id", pcxResourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "core_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceID, ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrNetworkInterfaceID, ""),
					resource.TestCheckResourceAttr(resourceName, "origin", string(awstypes.RouteOriginCreateRoute)),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.RouteStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrTransitGatewayID, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrVPCEndpointID, ""),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_peering_connection_id", pcxResourceName, names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCRoute_IPv6Update_target(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var route awstypes.Route
	resourceName := "aws_route.test"
	vgwResourceName := "aws_vpn_gateway.test"
	igwResourceName := "aws_internet_gateway.test"
	eniResourceName := "aws_network_interface.test"
	pcxResourceName := "aws_vpc_peering_connection.test"
	eoigwResourceName := "aws_egress_only_internet_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	destinationCidr := "::/0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteConfig_ipv6FlexiTarget(rName, destinationCidr, "gateway_id", vgwResourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "core_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_id", vgwResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceID, ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrNetworkInterfaceID, ""),
					resource.TestCheckResourceAttr(resourceName, "origin", string(awstypes.RouteOriginCreateRoute)),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.RouteStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrTransitGatewayID, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrVPCEndpointID, ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				Config: testAccVPCRouteConfig_ipv6FlexiTarget(rName, destinationCidr, "gateway_id", igwResourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "core_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_id", igwResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceID, ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrNetworkInterfaceID, ""),
					resource.TestCheckResourceAttr(resourceName, "origin", string(awstypes.RouteOriginCreateRoute)),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.RouteStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrTransitGatewayID, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrVPCEndpointID, ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				Config: testAccVPCRouteConfig_ipv6FlexiTarget(rName, destinationCidr, "egress_only_gateway_id", eoigwResourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "core_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "egress_only_gateway_id", eoigwResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceID, ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrNetworkInterfaceID, ""),
					resource.TestCheckResourceAttr(resourceName, "origin", string(awstypes.RouteOriginCreateRoute)),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.RouteStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrTransitGatewayID, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrVPCEndpointID, ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				Config: testAccVPCRouteConfig_ipv6FlexiTarget(rName, destinationCidr, names.AttrNetworkInterfaceID, eniResourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "core_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceID, ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrNetworkInterfaceID, eniResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "origin", string(awstypes.RouteOriginCreateRoute)),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.RouteStateBlackhole)),
					resource.TestCheckResourceAttr(resourceName, names.AttrTransitGatewayID, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrVPCEndpointID, ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				Config: testAccVPCRouteConfig_ipv6FlexiTarget(rName, destinationCidr, "vpc_peering_connection_id", pcxResourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "core_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceID, ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrNetworkInterfaceID, ""),
					resource.TestCheckResourceAttr(resourceName, "origin", string(awstypes.RouteOriginCreateRoute)),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.RouteStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrTransitGatewayID, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrVPCEndpointID, ""),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_peering_connection_id", pcxResourceName, names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCRoute_ipv4ToVPCEndpoint(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var route awstypes.Route
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route.test"
	vpcEndpointResourceName := "aws_vpc_endpoint.test"
	destinationCidr := "172.16.1.0/24"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckELBv2GatewayLoadBalancer(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID, "elasticloadbalancing"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteConfig_resourceIPv4Endpoint(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "core_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceID, ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrNetworkInterfaceID, ""),
					resource.TestCheckResourceAttr(resourceName, "origin", string(awstypes.RouteOriginCreateRoute)),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.RouteStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrTransitGatewayID, ""),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVPCEndpointID, vpcEndpointResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCRoute_ipv6ToVPCEndpoint(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var route awstypes.Route
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route.test"
	vpcEndpointResourceName := "aws_vpc_endpoint.test"
	destinationIpv6Cidr := "::/0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckELBv2GatewayLoadBalancer(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID, "elasticloadbalancing"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteConfig_resourceIPv6Endpoint(rName, destinationIpv6Cidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "core_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", destinationIpv6Cidr),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceID, ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrNetworkInterfaceID, ""),
					resource.TestCheckResourceAttr(resourceName, "origin", string(awstypes.RouteOriginCreateRoute)),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.RouteStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrTransitGatewayID, ""),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVPCEndpointID, vpcEndpointResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCRoute_localRouteCreateError(t *testing.T) {
	ctx := acctest.Context(t)
	var routeTable awstypes.RouteTable
	var vpc awstypes.Vpc
	rtResourceName := "aws_route_table.test"
	vpcResourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteConfig_ipv4NoRoute(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(ctx, vpcResourceName, &vpc),
					testAccCheckRouteTableExists(ctx, rtResourceName, &routeTable),
					testAccCheckRouteTableNumberOfRoutes(&routeTable, 1),
				),
			},
			{
				Config:      testAccVPCRouteConfig_ipv4Local(rName),
				ExpectError: regexache.MustCompile("cannot create local Route, use `terraform import` to manage existing local Routes"),
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/11455.
func TestAccVPCRoute_localRouteImport(t *testing.T) {
	ctx := acctest.Context(t)
	var routeTable awstypes.RouteTable
	var vpc awstypes.Vpc
	resourceName := "aws_route.test"
	rtResourceName := "aws_route_table.test"
	vpcResourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteConfig_ipv4NoRoute(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(ctx, vpcResourceName, &vpc),
					testAccCheckRouteTableExists(ctx, rtResourceName, &routeTable),
					testAccCheckRouteTableNumberOfRoutes(&routeTable, 1),
				),
			},
			{
				Config:       testAccVPCRouteConfig_ipv4Local(rName),
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(rt *awstypes.RouteTable, v *awstypes.Vpc) resource.ImportStateIdFunc {
					return func(s *terraform.State) (string, error) {
						return fmt.Sprintf("%s_%s", aws.ToString(rt.RouteTableId), aws.ToString(v.CidrBlock)), nil
					}
				}(&routeTable, &vpc),
				// Don't verify the state as the local route isn't actually in the pre-import state.
				// Just running ImportState verifies that we can import a local route.
				ImportStateVerify: false,
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/21350.
func TestAccVPCRoute_localRouteImportAndUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var routeTable awstypes.RouteTable
	var vpc awstypes.Vpc
	resourceName := "aws_route.test"
	rtResourceName := "aws_route_table.test"
	vpcResourceName := "aws_vpc.test"
	eniResourceName := "aws_network_interface.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteConfig_ipv4NoRoute(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(ctx, vpcResourceName, &vpc),
					testAccCheckRouteTableExists(ctx, rtResourceName, &routeTable),
					testAccCheckRouteTableNumberOfRoutes(&routeTable, 1),
				),
			},
			{
				Config:       testAccVPCRouteConfig_ipv4Local(rName),
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(rt *awstypes.RouteTable, v *awstypes.Vpc) resource.ImportStateIdFunc {
					return func(s *terraform.State) (string, error) {
						return fmt.Sprintf("%s_%s", aws.ToString(rt.RouteTableId), aws.ToString(v.CidrBlock)), nil
					}
				}(&routeTable, &vpc),
				ImportStatePersist: true,
				// Don't verify the state as the local route isn't actually in the pre-import state.
				// Just running ImportState verifies that we can import a local route.
				ImportStateVerify: false,
			},
			{
				Config: testAccVPCRouteConfig_ipv4LocalToNetworkInterface(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(ctx, vpcResourceName, &vpc),
					testAccCheckRouteTableExists(ctx, rtResourceName, &routeTable),
					testAccCheckRouteTableNumberOfRoutes(&routeTable, 1),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrNetworkInterfaceID, eniResourceName, names.AttrID),
				),
			},
			{
				Config: testAccVPCRouteConfig_ipv4LocalRestore(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckVPCExists(ctx, vpcResourceName, &vpc),
					testAccCheckRouteTableExists(ctx, rtResourceName, &routeTable),
					testAccCheckRouteTableNumberOfRoutes(&routeTable, 1),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", "local"),
				),
			},
		},
	})
}

func TestAccVPCRoute_prefixListToInternetGateway(t *testing.T) {
	ctx := acctest.Context(t)
	var route awstypes.Route
	resourceName := "aws_route.test"
	igwResourceName := "aws_internet_gateway.test"
	plResourceName := "aws_ec2_managed_prefix_list.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckManagedPrefixList(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteConfig_prefixListInternetGateway(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "core_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttrPair(resourceName, "destination_prefix_list_id", plResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_id", igwResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceID, ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrNetworkInterfaceID, ""),
					resource.TestCheckResourceAttr(resourceName, "origin", string(awstypes.RouteOriginCreateRoute)),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.RouteStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrTransitGatewayID, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrVPCEndpointID, ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCRoute_prefixListToVPNGateway(t *testing.T) {
	ctx := acctest.Context(t)
	var route awstypes.Route
	resourceName := "aws_route.test"
	vgwResourceName := "aws_vpn_gateway.test"
	plResourceName := "aws_ec2_managed_prefix_list.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckManagedPrefixList(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteConfig_prefixListVPNGateway(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "core_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttrPair(resourceName, "destination_prefix_list_id", plResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_id", vgwResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceID, ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrNetworkInterfaceID, ""),
					resource.TestCheckResourceAttr(resourceName, "origin", string(awstypes.RouteOriginCreateRoute)),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.RouteStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrTransitGatewayID, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrVPCEndpointID, ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCRoute_prefixListToInstance(t *testing.T) {
	ctx := acctest.Context(t)
	var route awstypes.Route
	resourceName := "aws_route.test"
	instanceResourceName := "aws_instance.test"
	plResourceName := "aws_ec2_managed_prefix_list.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckManagedPrefixList(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteConfig_prefixListInstance(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "core_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttrPair(resourceName, "destination_prefix_list_id", plResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrInstanceID, instanceResourceName, names.AttrID),
					acctest.CheckResourceAttrAccountID(resourceName, "instance_owner_id"),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrNetworkInterfaceID, instanceResourceName, "primary_network_interface_id"),
					resource.TestCheckResourceAttr(resourceName, "origin", string(awstypes.RouteOriginCreateRoute)),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.RouteStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrTransitGatewayID, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrVPCEndpointID, ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCRoute_PrefixListToNetworkInterface_unattached(t *testing.T) {
	ctx := acctest.Context(t)
	var route awstypes.Route
	resourceName := "aws_route.test"
	eniResourceName := "aws_network_interface.test"
	plResourceName := "aws_ec2_managed_prefix_list.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckManagedPrefixList(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteConfig_prefixListNetworkInterfaceUnattached(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "core_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttrPair(resourceName, "destination_prefix_list_id", plResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceID, ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrNetworkInterfaceID, eniResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "origin", string(awstypes.RouteOriginCreateRoute)),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.RouteStateBlackhole)),
					resource.TestCheckResourceAttr(resourceName, names.AttrTransitGatewayID, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrVPCEndpointID, ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCRoute_PrefixListToNetworkInterface_attached(t *testing.T) {
	ctx := acctest.Context(t)
	var route awstypes.Route
	resourceName := "aws_route.test"
	eniResourceName := "aws_network_interface.test"
	instanceResourceName := "aws_instance.test"
	plResourceName := "aws_ec2_managed_prefix_list.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckManagedPrefixList(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteConfig_prefixListNetworkInterfaceAttached(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "core_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttrPair(resourceName, "destination_prefix_list_id", plResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrInstanceID, instanceResourceName, names.AttrID),
					acctest.CheckResourceAttrAccountID(resourceName, "instance_owner_id"),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrNetworkInterfaceID, eniResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "origin", string(awstypes.RouteOriginCreateRoute)),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.RouteStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrTransitGatewayID, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrVPCEndpointID, ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCRoute_prefixListToVPCPeeringConnection(t *testing.T) {
	ctx := acctest.Context(t)
	var route awstypes.Route
	resourceName := "aws_route.test"
	pcxResourceName := "aws_vpc_peering_connection.test"
	plResourceName := "aws_ec2_managed_prefix_list.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckManagedPrefixList(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteConfig_prefixListPeeringConnection(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "core_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttrPair(resourceName, "destination_prefix_list_id", plResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceID, ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrNetworkInterfaceID, ""),
					resource.TestCheckResourceAttr(resourceName, "origin", string(awstypes.RouteOriginCreateRoute)),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.RouteStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrTransitGatewayID, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrVPCEndpointID, ""),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_peering_connection_id", pcxResourceName, names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCRoute_prefixListToNatGateway(t *testing.T) {
	ctx := acctest.Context(t)
	var route awstypes.Route
	resourceName := "aws_route.test"
	ngwResourceName := "aws_nat_gateway.test"
	plResourceName := "aws_ec2_managed_prefix_list.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckManagedPrefixList(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteConfig_prefixListNATGateway(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "core_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttrPair(resourceName, "destination_prefix_list_id", plResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceID, ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "nat_gateway_id", ngwResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrNetworkInterfaceID, ""),
					resource.TestCheckResourceAttr(resourceName, "origin", string(awstypes.RouteOriginCreateRoute)),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.RouteStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrTransitGatewayID, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrVPCEndpointID, ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCRoute_prefixListToTransitGateway(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var route awstypes.Route
	resourceName := "aws_route.test"
	tgwResourceName := "aws_ec2_transit_gateway.test"
	plResourceName := "aws_ec2_managed_prefix_list.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckManagedPrefixList(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteConfig_prefixListTransitGateway(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "core_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttrPair(resourceName, "destination_prefix_list_id", plResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceID, ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrNetworkInterfaceID, ""),
					resource.TestCheckResourceAttr(resourceName, "origin", string(awstypes.RouteOriginCreateRoute)),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.RouteStateActive)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrTransitGatewayID, tgwResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, names.AttrVPCEndpointID, ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCRoute_prefixListToCarrierGateway(t *testing.T) {
	ctx := acctest.Context(t)
	var route awstypes.Route
	resourceName := "aws_route.test"
	cgwResourceName := "aws_ec2_carrier_gateway.test"
	plResourceName := "aws_ec2_managed_prefix_list.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckManagedPrefixList(ctx, t)
			testAccPreCheckWavelengthZoneAvailable(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteConfig_prefixListCarrierGateway(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttrPair(resourceName, "carrier_gateway_id", cgwResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "core_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttrPair(resourceName, "destination_prefix_list_id", plResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceID, ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrNetworkInterfaceID, ""),
					resource.TestCheckResourceAttr(resourceName, "origin", string(awstypes.RouteOriginCreateRoute)),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.RouteStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrTransitGatewayID, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrVPCEndpointID, ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCRoute_prefixListToLocalGateway(t *testing.T) {
	ctx := acctest.Context(t)
	var route awstypes.Route
	resourceName := "aws_route.test"
	localGatewayDataSourceName := "data.aws_ec2_local_gateway.first"
	plResourceName := "aws_ec2_managed_prefix_list.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckManagedPrefixList(ctx, t)
			acctest.PreCheckOutpostsOutposts(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteConfig_prefixListLocalGateway(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "core_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttrPair(resourceName, "destination_prefix_list_id", plResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceID, ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "local_gateway_id", localGatewayDataSourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrNetworkInterfaceID, ""),
					resource.TestCheckResourceAttr(resourceName, "origin", string(awstypes.RouteOriginCreateRoute)),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.RouteStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrTransitGatewayID, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrVPCEndpointID, ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCRoute_prefixListToEgressOnlyInternetGateway(t *testing.T) {
	ctx := acctest.Context(t)
	var route awstypes.Route
	resourceName := "aws_route.test"
	eoigwResourceName := "aws_egress_only_internet_gateway.test"
	plResourceName := "aws_ec2_managed_prefix_list.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckManagedPrefixList(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCRouteConfig_prefixListEgressOnlyInternetGateway(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteExists(ctx, resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "core_network_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttrPair(resourceName, "destination_prefix_list_id", plResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "egress_only_gateway_id", eoigwResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrInstanceID, ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrNetworkInterfaceID, ""),
					resource.TestCheckResourceAttr(resourceName, "origin", string(awstypes.RouteOriginCreateRoute)),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.RouteStateActive)),
					resource.TestCheckResourceAttr(resourceName, names.AttrTransitGatewayID, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrVPCEndpointID, ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCRoute_duplicate(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	destinationCidr := "10.3.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccVPCRouteConfig_ipv4InternetGatewayDuplicate(rName, destinationCidr),
				ExpectError: regexache.MustCompile(`RouteAlreadyExists: Route .* already exists`),
			},
		},
	})
}

func testAccCheckRouteExists(ctx context.Context, n string, v *awstypes.Route) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		var route *awstypes.Route
		var err error
		if v := rs.Primary.Attributes["destination_cidr_block"]; v != "" {
			route, err = tfec2.FindRouteByIPv4Destination(ctx, conn, rs.Primary.Attributes["route_table_id"], v)
		} else if v := rs.Primary.Attributes["destination_ipv6_cidr_block"]; v != "" {
			route, err = tfec2.FindRouteByIPv6Destination(ctx, conn, rs.Primary.Attributes["route_table_id"], v)
		} else if v := rs.Primary.Attributes["destination_prefix_list_id"]; v != "" {
			route, err = tfec2.FindRouteByPrefixListIDDestination(ctx, conn, rs.Primary.Attributes["route_table_id"], v)
		}

		if err != nil {
			return err
		}

		*v = *route

		return nil
	}
}

func testAccCheckRouteDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_route" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

			var err error
			if v := rs.Primary.Attributes["destination_cidr_block"]; v != "" {
				_, err = tfec2.FindRouteByIPv4Destination(ctx, conn, rs.Primary.Attributes["route_table_id"], v)
			} else if v := rs.Primary.Attributes["destination_ipv6_cidr_block"]; v != "" {
				_, err = tfec2.FindRouteByIPv6Destination(ctx, conn, rs.Primary.Attributes["route_table_id"], v)
			} else if v := rs.Primary.Attributes["destination_prefix_list_id"]; v != "" {
				_, err = tfec2.FindRouteByPrefixListIDDestination(ctx, conn, rs.Primary.Attributes["route_table_id"], v)
			}

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Route still exists")
		}

		return nil
	}
}

func testAccRouteImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}

		destination := rs.Primary.Attributes["destination_cidr_block"]
		if v, ok := rs.Primary.Attributes["destination_ipv6_cidr_block"]; ok && v != "" {
			destination = v
		}
		if v, ok := rs.Primary.Attributes["destination_prefix_list_id"]; ok && v != "" {
			destination = v
		}

		return fmt.Sprintf("%s_%s", rs.Primary.Attributes["route_table_id"], destination), nil
	}
}

func testAccVPCRouteConfig_ipv4InternetGateway(rName, destinationCidr string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id         = aws_route_table.test.id
  destination_cidr_block = %[2]q
  gateway_id             = aws_internet_gateway.test.id
}
`, rName, destinationCidr)
}

func testAccVPCRouteConfig_ipv4InternetGatewayDuplicate(rName, destinationCidr string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test1" {
  route_table_id         = aws_route_table.test.id
  destination_cidr_block = %[2]q
  gateway_id             = aws_internet_gateway.test.id
}

resource "aws_route" "test2" {
  route_table_id         = aws_route.test1.route_table_id
  destination_cidr_block = %[2]q
  gateway_id             = aws_internet_gateway.test.id
}
`, rName, destinationCidr)
}

func testAccVPCRouteConfig_ipv6InternetGateway(rName, destinationCidr string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block                       = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_egress_only_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id              = aws_route_table.test.id
  destination_ipv6_cidr_block = %[2]q
  gateway_id                  = aws_internet_gateway.test.id
}
`, rName, destinationCidr)
}

func testAccVPCRouteConfig_ipv6NetworkInterfaceUnattached(rName, destinationCidr string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block                       = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block        = "10.1.1.0/24"
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]
  ipv6_cidr_block   = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, 1)

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id              = aws_route_table.test.id
  destination_ipv6_cidr_block = %[2]q
  network_interface_id        = aws_network_interface.test.id
}
`, rName, destinationCidr))
}

func testAccVPCRouteConfig_ipv6Instance(rName, destinationCidr string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("data.aws_availability_zones.available.names[0]", "t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block                       = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block        = "10.1.1.0/24"
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]
  ipv6_cidr_block   = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, 1)

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test.id

  ipv6_address_count = 1

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id              = aws_route_table.test.id
  destination_ipv6_cidr_block = %[2]q
  network_interface_id        = aws_instance.test.primary_network_interface_id
}
`, rName, destinationCidr))
}

func testAccVPCRouteConfig_ipv6PeeringConnection(rName, destinationCidr string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block                       = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "target" {
  cidr_block                       = "10.0.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_peering_connection" "test" {
  vpc_id      = aws_vpc.test.id
  peer_vpc_id = aws_vpc.target.id
  auto_accept = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id              = aws_route_table.test.id
  destination_ipv6_cidr_block = %[2]q
  vpc_peering_connection_id   = aws_vpc_peering_connection.test.id
}
`, rName, destinationCidr)
}

func testAccVPCRouteConfig_ipv6EgressOnlyInternetGateway(rName, destinationCidr string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block                       = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_egress_only_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id              = aws_route_table.test.id
  destination_ipv6_cidr_block = %[2]q
  egress_only_gateway_id      = aws_egress_only_internet_gateway.test.id
}
`, rName, destinationCidr)
}

func testAccVPCRouteConfig_endpoint(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id         = aws_route_table.test.id
  destination_cidr_block = "10.3.0.0/16"
  gateway_id             = aws_internet_gateway.test.id

  # Forcing endpoint to create before route - without this the crash is a race.
  depends_on = [aws_vpc_endpoint.test]
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id          = aws_vpc.test.id
  service_name    = "com.amazonaws.${data.aws_region.current.name}.s3"
  route_table_ids = [aws_route_table.test.id]
}
`, rName)
}

func testAccVPCRouteConfig_ipv4TransitGateway(rName, destinationCidr string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.1.1.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  destination_cidr_block = %[2]q
  route_table_id         = aws_route_table.test.id
  transit_gateway_id     = aws_ec2_transit_gateway_vpc_attachment.test.transit_gateway_id
}
`, rName, destinationCidr))
}

func testAccVPCRouteConfig_ipv6TransitGateway(rName, destinationCidr string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block                       = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.1.1.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  destination_ipv6_cidr_block = %[2]q
  route_table_id              = aws_route_table.test.id
  transit_gateway_id          = aws_ec2_transit_gateway_vpc_attachment.test.transit_gateway_id
}
`, rName, destinationCidr))
}

func testAccVPCRouteConfig_conditionalIPv4IPv6(rName, destinationCidr, destinationIpv6Cidr string, ipv6Route bool) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block                       = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

locals {
  ipv6             = %[4]t
  destination      = %[2]q
  destination_ipv6 = %[3]q
}

resource "aws_route" "test" {
  route_table_id = aws_route_table.test.id
  gateway_id     = aws_internet_gateway.test.id

  destination_cidr_block      = local.ipv6 ? null : local.destination
  destination_ipv6_cidr_block = local.ipv6 ? local.destination_ipv6 : null
}
`, rName, destinationCidr, destinationIpv6Cidr, ipv6Route)
}

func testAccVPCRouteConfig_ipv4Instance(rName, destinationCidr string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("data.aws_availability_zones.available.names[0]", "t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block        = "10.1.1.0/24"
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id         = aws_route_table.test.id
  destination_cidr_block = %[2]q
  network_interface_id   = aws_instance.test.primary_network_interface_id
}
`, rName, destinationCidr))
}

func testAccVPCRouteConfig_ipv4NetworkInterfaceUnattached(rName, destinationCidr string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block        = "10.1.1.0/24"
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id         = aws_route_table.test.id
  destination_cidr_block = %[2]q
  network_interface_id   = aws_network_interface.test.id
}
`, rName, destinationCidr))
}

func testAccVPCRouteConfig_resourceIPv4LocalGateway(rName, destinationCidr string) string {
	return fmt.Sprintf(`
data "aws_ec2_local_gateways" "all" {}

data "aws_ec2_local_gateway" "first" {
  id = tolist(data.aws_ec2_local_gateways.all.ids)[0]
}

data "aws_ec2_local_gateway_route_tables" "all" {}

data "aws_ec2_local_gateway_route_table" "first" {
  local_gateway_route_table_id = tolist(data.aws_ec2_local_gateway_route_tables.all.ids)[0]
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_local_gateway_route_table_vpc_association" "example" {
  local_gateway_route_table_id = data.aws_ec2_local_gateway_route_table.first.id
  vpc_id                       = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_ec2_local_gateway_route_table_vpc_association.example]
}

resource "aws_route" "test" {
  route_table_id         = aws_route_table.test.id
  destination_cidr_block = %[2]q
  local_gateway_id       = data.aws_ec2_local_gateway.first.id
}
`, rName, destinationCidr)
}

func testAccVPCRouteConfig_resourceIPv6LocalGateway(rName, destinationCidr string) string {
	return fmt.Sprintf(`
data "aws_ec2_local_gateways" "all" {}

data "aws_ec2_local_gateway" "first" {
  id = tolist(data.aws_ec2_local_gateways.all.ids)[0]
}

data "aws_ec2_local_gateway_route_tables" "all" {}

data "aws_ec2_local_gateway_route_table" "first" {
  local_gateway_route_table_id = tolist(data.aws_ec2_local_gateway_route_tables.all.ids)[0]
}

resource "aws_vpc" "test" {
  cidr_block                       = "10.0.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_local_gateway_route_table_vpc_association" "example" {
  local_gateway_route_table_id = data.aws_ec2_local_gateway_route_table.first.id
  vpc_id                       = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_ec2_local_gateway_route_table_vpc_association.example]
}

resource "aws_route" "test" {
  route_table_id              = aws_route_table.test.id
  destination_ipv6_cidr_block = %[2]q
  local_gateway_id            = data.aws_ec2_local_gateway.first.id
}
`, rName, destinationCidr)
}

func testAccVPCRouteConfig_ipv4NetworkInterfaceAttached(rName, destinationCidr string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("data.aws_availability_zones.available.names[0]", "t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block        = "10.1.1.0/24"
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type

  network_interface {
    device_index         = 0
    network_interface_id = aws_network_interface.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id         = aws_route_table.test.id
  destination_cidr_block = %[2]q
  network_interface_id   = aws_network_interface.test.id

  # Wait for the ENI attachment.
  depends_on = [aws_instance.test]
}
`, rName, destinationCidr))
}

func testAccVPCRouteConfig_ipv4NetworkInterfaceTwoAttachments(rName, destinationCidr, targetResourceName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("data.aws_availability_zones.available.names[0]", "t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block        = "10.1.1.0/24"
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test1" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test2" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type

  network_interface {
    device_index         = 0
    network_interface_id = aws_network_interface.test1.id
  }

  network_interface {
    device_index         = 1
    network_interface_id = aws_network_interface.test2.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id         = aws_route_table.test.id
  destination_cidr_block = %[2]q
  network_interface_id   = %[3]s.id

  # Wait for the ENI attachment.
  depends_on = [aws_instance.test]
}
`, rName, destinationCidr, targetResourceName))
}

func testAccVPCRouteConfig_ipv4PeeringConnection(rName, destinationCidr string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "target" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_peering_connection" "test" {
  vpc_id      = aws_vpc.test.id
  peer_vpc_id = aws_vpc.target.id
  auto_accept = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id            = aws_route_table.test.id
  destination_cidr_block    = %[2]q
  vpc_peering_connection_id = aws_vpc_peering_connection.test.id
}
`, rName, destinationCidr)
}

func testAccVPCRouteConfig_ipv4NATGateway(rName, destinationCidr string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.1.1.0/24"
  vpc_id     = aws_vpc.test.id

  map_public_ip_on_launch = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_eip" "test" {
  domain = "vpc"

  tags = {
    Name = %[1]q
  }
}

resource "aws_nat_gateway" "test" {
  allocation_id = aws_eip.test.id
  subnet_id     = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_internet_gateway.test]
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id         = aws_route_table.test.id
  destination_cidr_block = %[2]q
  nat_gateway_id         = aws_nat_gateway.test.id
}
`, rName, destinationCidr)
}

func testAccVPCRouteConfig_ipv6NATGateway(rName, destinationCidr string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block                       = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id                          = aws_vpc.test.id
  cidr_block                      = "10.1.1.0/24"
  ipv6_cidr_block                 = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, 1)
  assign_ipv6_address_on_creation = true

  enable_resource_name_dns_aaaa_record_on_launch = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_nat_gateway" "test" {
  connectivity_type = "private"
  subnet_id         = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id              = aws_route_table.test.id
  destination_ipv6_cidr_block = %[2]q
  nat_gateway_id              = aws_nat_gateway.test.id
}
`, rName, destinationCidr)
}

func testAccVPCRouteConfig_ipv4VPNGateway(rName, destinationCidr string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id         = aws_route_table.test.id
  destination_cidr_block = %[2]q
  gateway_id             = aws_vpn_gateway.test.id
}
`, rName, destinationCidr)
}

func testAccVPCRouteConfig_ipv6VPNGateway(rName, destinationCidr string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block                       = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id              = aws_route_table.test.id
  destination_ipv6_cidr_block = %[2]q
  gateway_id                  = aws_vpn_gateway.test.id
}
`, rName, destinationCidr)
}

func testAccVPCRouteConfig_resourceIPv4Endpoint(rName, destinationCidr string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_vpc" "test" {
  cidr_block = "10.10.10.0/25"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 2, 0)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb" "test" {
  load_balancer_type = "gateway"
  name               = %[1]q

  subnet_mapping {
    subnet_id = aws_subnet.test.id
  }
}

resource "aws_vpc_endpoint_service" "test" {
  acceptance_required        = false
  allowed_principals         = [data.aws_iam_session_context.current.issuer_arn]
  gateway_load_balancer_arns = [aws_lb.test.arn]

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  service_name      = aws_vpc_endpoint_service.test.service_name
  subnet_ids        = [aws_subnet.test.id]
  vpc_endpoint_type = aws_vpc_endpoint_service.test.service_type
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id         = aws_route_table.test.id
  destination_cidr_block = %[2]q
  vpc_endpoint_id        = aws_vpc_endpoint.test.id
}
`, rName, destinationCidr))
}

func testAccVPCRouteConfig_resourceIPv6Endpoint(rName, destinationIpv6Cidr string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_vpc" "test" {
  cidr_block = "10.10.10.0/25"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 2, 0)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb" "test" {
  load_balancer_type = "gateway"
  name               = %[1]q

  subnet_mapping {
    subnet_id = aws_subnet.test.id
  }
}

resource "aws_vpc_endpoint_service" "test" {
  acceptance_required        = false
  allowed_principals         = [data.aws_iam_session_context.current.issuer_arn]
  gateway_load_balancer_arns = [aws_lb.test.arn]

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  service_name      = aws_vpc_endpoint_service.test.service_name
  subnet_ids        = [aws_subnet.test.id]
  vpc_endpoint_type = aws_vpc_endpoint_service.test.service_type
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id              = aws_route_table.test.id
  destination_ipv6_cidr_block = %[2]q
  vpc_endpoint_id             = aws_vpc_endpoint.test.id
}
`, rName, destinationIpv6Cidr))
}

func testAccVPCRouteConfig_ipv4FlexiTarget(rName, destinationCidr, targetAttribute, targetValue string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("data.aws_availability_zones.available.names[0]", "t3.micro", "t2.micro"),
		fmt.Sprintf(`
locals {
  target_attr  = %[3]q
  target_value = %[4]s.id
}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block        = "10.1.1.0/24"
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]

  map_public_ip_on_launch = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "target" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_peering_connection" "test" {
  vpc_id      = aws_vpc.test.id
  peer_vpc_id = aws_vpc.target.id
  auto_accept = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_eip" "test" {
  domain = "vpc"

  tags = {
    Name = %[1]q
  }
}

resource "aws_nat_gateway" "test" {
  allocation_id = aws_eip.test.id
  subnet_id     = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_internet_gateway.test]
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

resource "aws_lb" "test" {
  load_balancer_type = "gateway"
  name               = %[1]q

  subnet_mapping {
    subnet_id = aws_subnet.test.id
  }
}

resource "aws_vpc_endpoint_service" "test" {
  acceptance_required        = false
  allowed_principals         = [data.aws_iam_session_context.current.issuer_arn]
  gateway_load_balancer_arns = [aws_lb.test.arn]

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  service_name      = aws_vpc_endpoint_service.test.service_name
  subnet_ids        = [aws_subnet.test.id]
  vpc_endpoint_type = aws_vpc_endpoint_service.test.service_type
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id         = aws_route_table.test.id
  destination_cidr_block = %[2]q

  egress_only_gateway_id    = (local.target_attr == "egress_only_gateway_id") ? local.target_value : null
  gateway_id                = (local.target_attr == "gateway_id") ? local.target_value : null
  local_gateway_id          = (local.target_attr == "local_gateway_id") ? local.target_value : null
  nat_gateway_id            = (local.target_attr == "nat_gateway_id") ? local.target_value : null
  network_interface_id      = (local.target_attr == "network_interface_id") ? local.target_value : null
  transit_gateway_id        = (local.target_attr == "transit_gateway_id") ? local.target_value : null
  vpc_endpoint_id           = (local.target_attr == "vpc_endpoint_id") ? local.target_value : null
  vpc_peering_connection_id = (local.target_attr == "vpc_peering_connection_id") ? local.target_value : null
}
`, rName, destinationCidr, targetAttribute, targetValue))
}

func testAccVPCRouteConfig_ipv6FlexiTarget(rName, destinationCidr, targetAttribute, targetValue string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("data.aws_availability_zones.available.names[0]", "t3.micro", "t2.micro"),
		fmt.Sprintf(`
locals {
  target_attr  = %[3]q
  target_value = %[4]s.id
}

resource "aws_vpc" "test" {
  cidr_block                       = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block        = "10.1.1.0/24"
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]
  ipv6_cidr_block   = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, 1)

  map_public_ip_on_launch = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test.id

  ipv6_address_count = 1

  tags = {
    Name = %[1]q
  }
}

resource "aws_egress_only_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "target" {
  cidr_block                       = "10.0.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_peering_connection" "test" {
  vpc_id      = aws_vpc.test.id
  peer_vpc_id = aws_vpc.target.id
  auto_accept = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id              = aws_route_table.test.id
  destination_ipv6_cidr_block = %[2]q

  egress_only_gateway_id    = (local.target_attr == "egress_only_gateway_id") ? local.target_value : null
  gateway_id                = (local.target_attr == "gateway_id") ? local.target_value : null
  local_gateway_id          = (local.target_attr == "local_gateway_id") ? local.target_value : null
  nat_gateway_id            = (local.target_attr == "nat_gateway_id") ? local.target_value : null
  network_interface_id      = (local.target_attr == "network_interface_id") ? local.target_value : null
  transit_gateway_id        = (local.target_attr == "transit_gateway_id") ? local.target_value : null
  vpc_endpoint_id           = (local.target_attr == "vpc_endpoint_id") ? local.target_value : null
  vpc_peering_connection_id = (local.target_attr == "vpc_peering_connection_id") ? local.target_value : null
}
`, rName, destinationCidr, targetAttribute, targetValue))
}

func testAccVPCRouteConfig_ipv4NoRoute(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCRouteConfig_ipv4Local(rName string) string {
	return acctest.ConfigCompose(testAccVPCRouteConfig_ipv4NoRoute(rName), `
resource "aws_route" "test" {
  route_table_id         = aws_route_table.test.id
  destination_cidr_block = aws_vpc.test.cidr_block
  gateway_id             = "local"
}
`)
}

func testAccVPCRouteConfig_ipv4LocalToNetworkInterface(rName string) string {
	return acctest.ConfigCompose(testAccVPCRouteConfig_ipv4NoRoute(rName), fmt.Sprintf(`
resource "aws_route" "test" {
  route_table_id         = aws_route_table.test.id
  destination_cidr_block = aws_vpc.test.cidr_block
  network_interface_id   = aws_network_interface.test.id
}

resource "aws_subnet" "test" {
  cidr_block = "10.1.1.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCRouteConfig_ipv4LocalRestore(rName string) string {
	return acctest.ConfigCompose(testAccVPCRouteConfig_ipv4NoRoute(rName), fmt.Sprintf(`
resource "aws_route" "test" {
  route_table_id         = aws_route_table.test.id
  destination_cidr_block = aws_vpc.test.cidr_block
  gateway_id             = "local"
}

resource "aws_subnet" "test" {
  cidr_block = "10.1.1.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCRouteConfig_prefixListInternetGateway(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_managed_prefix_list" "test" {
  address_family = "IPv4"
  max_entries    = 1
  name           = %[1]q
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id             = aws_route_table.test.id
  destination_prefix_list_id = aws_ec2_managed_prefix_list.test.id
  gateway_id                 = aws_internet_gateway.test.id
}
`, rName)
}

func testAccVPCRouteConfig_prefixListVPNGateway(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_managed_prefix_list" "test" {
  address_family = "IPv4"
  max_entries    = 1
  name           = %[1]q
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id             = aws_route_table.test.id
  destination_prefix_list_id = aws_ec2_managed_prefix_list.test.id
  gateway_id                 = aws_vpn_gateway.test.id
}
`, rName)
}

func testAccVPCRouteConfig_prefixListInstance(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("data.aws_availability_zones.available.names[0]", "t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block        = "10.1.1.0/24"
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_managed_prefix_list" "test" {
  address_family = "IPv4"
  max_entries    = 1
  name           = %[1]q
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id             = aws_route_table.test.id
  destination_prefix_list_id = aws_ec2_managed_prefix_list.test.id
  network_interface_id       = aws_instance.test.primary_network_interface_id
}
`, rName))
}

func testAccVPCRouteConfig_prefixListNetworkInterfaceUnattached(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block        = "10.1.1.0/24"
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_managed_prefix_list" "test" {
  address_family = "IPv4"
  max_entries    = 1
  name           = %[1]q
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id             = aws_route_table.test.id
  destination_prefix_list_id = aws_ec2_managed_prefix_list.test.id
  network_interface_id       = aws_network_interface.test.id
}
`, rName))
}

func testAccVPCRouteConfig_prefixListNetworkInterfaceAttached(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(),
		acctest.ConfigAvailableAZsNoOptIn(),
		acctest.AvailableEC2InstanceTypeForAvailabilityZone("data.aws_availability_zones.available.names[0]", "t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block        = "10.1.1.0/24"
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type

  network_interface {
    device_index         = 0
    network_interface_id = aws_network_interface.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_managed_prefix_list" "test" {
  address_family = "IPv4"
  max_entries    = 1
  name           = %[1]q
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id             = aws_route_table.test.id
  destination_prefix_list_id = aws_ec2_managed_prefix_list.test.id
  network_interface_id       = aws_network_interface.test.id

  # Wait for the ENI attachment.
  depends_on = [aws_instance.test]
}
`, rName))
}

func testAccVPCRouteConfig_prefixListPeeringConnection(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "target" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_peering_connection" "test" {
  vpc_id      = aws_vpc.test.id
  peer_vpc_id = aws_vpc.target.id
  auto_accept = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_managed_prefix_list" "test" {
  address_family = "IPv4"
  max_entries    = 1
  name           = %[1]q
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id             = aws_route_table.test.id
  destination_prefix_list_id = aws_ec2_managed_prefix_list.test.id
  vpc_peering_connection_id  = aws_vpc_peering_connection.test.id
}
`, rName)
}

func testAccVPCRouteConfig_prefixListNATGateway(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.1.1.0/24"
  vpc_id     = aws_vpc.test.id

  map_public_ip_on_launch = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_eip" "test" {
  domain = "vpc"

  tags = {
    Name = %[1]q
  }
}

resource "aws_nat_gateway" "test" {
  allocation_id = aws_eip.test.id
  subnet_id     = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_internet_gateway.test]
}

resource "aws_ec2_managed_prefix_list" "test" {
  address_family = "IPv4"
  max_entries    = 1
  name           = %[1]q
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id             = aws_route_table.test.id
  destination_prefix_list_id = aws_ec2_managed_prefix_list.test.id
  nat_gateway_id             = aws_nat_gateway.test.id
}
`, rName)
}

func testAccVPCRouteConfig_prefixListTransitGateway(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.1.1.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_managed_prefix_list" "test" {
  address_family = "IPv4"
  max_entries    = 1
  name           = %[1]q
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id             = aws_route_table.test.id
  destination_prefix_list_id = aws_ec2_managed_prefix_list.test.id
  transit_gateway_id         = aws_ec2_transit_gateway_vpc_attachment.test.transit_gateway_id
}
`, rName))
}

func testAccVPCRouteConfig_prefixListCarrierGateway(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_carrier_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_managed_prefix_list" "test" {
  address_family = "IPv4"
  max_entries    = 1
  name           = %[1]q
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id             = aws_route_table.test.id
  destination_prefix_list_id = aws_ec2_managed_prefix_list.test.id
  carrier_gateway_id         = aws_ec2_carrier_gateway.test.id
}
`, rName)
}

func testAccVPCRouteConfig_prefixListLocalGateway(rName string) string {
	return fmt.Sprintf(`
data "aws_ec2_local_gateways" "all" {}

data "aws_ec2_local_gateway" "first" {
  id = tolist(data.aws_ec2_local_gateways.all.ids)[0]
}

data "aws_ec2_local_gateway_route_tables" "all" {}

data "aws_ec2_local_gateway_route_table" "first" {
  local_gateway_route_table_id = tolist(data.aws_ec2_local_gateway_route_tables.all.ids)[0]
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_local_gateway_route_table_vpc_association" "example" {
  local_gateway_route_table_id = data.aws_ec2_local_gateway_route_table.first.id
  vpc_id                       = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_managed_prefix_list" "test" {
  address_family = "IPv4"
  max_entries    = 1
  name           = %[1]q
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_ec2_local_gateway_route_table_vpc_association.example]
}

resource "aws_route" "test" {
  route_table_id             = aws_route_table.test.id
  destination_prefix_list_id = aws_ec2_managed_prefix_list.test.id
  local_gateway_id           = data.aws_ec2_local_gateway.first.id
}
`, rName)
}

func testAccVPCRouteConfig_prefixListEgressOnlyInternetGateway(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block                       = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_egress_only_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_managed_prefix_list" "test" {
  address_family = "IPv6"
  max_entries    = 1
  name           = %[1]q
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id             = aws_route_table.test.id
  destination_prefix_list_id = aws_ec2_managed_prefix_list.test.id
  egress_only_gateway_id     = aws_egress_only_internet_gateway.test.id
}
`, rName)
}

func testAccVPCRouteConfig_ipv4CarrierGateway(rName, destinationCidr string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_carrier_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  destination_cidr_block = %[2]q
  route_table_id         = aws_route_table.test.id
  carrier_gateway_id     = aws_ec2_carrier_gateway.test.id
}
`, rName, destinationCidr)
}
