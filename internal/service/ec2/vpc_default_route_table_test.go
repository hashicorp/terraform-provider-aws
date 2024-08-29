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
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCDefaultRouteTable_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var routeTable awstypes.RouteTable
	resourceName := "aws_default_route_table.test"
	vpcResourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteTableDestroy(ctx),
		Steps: []resource.TestStep{
			// Verify non-existent Route Table ID behavior
			{
				Config:      testAccVPCDefaultRouteTableConfig_id("rtb-00000000"),
				ExpectError: regexache.MustCompile(`EC2 Default Route Table \(rtb-00000000\): couldn't find resource`),
			},
			// Verify invalid Route Table ID behavior
			{
				Config:      testAccVPCDefaultRouteTableConfig_id("vpc-00000000"),
				ExpectError: regexache.MustCompile(`EC2 Default Route Table \(vpc-00000000\): couldn't find resource`),
			},
			{
				Config: testAccVPCDefaultRouteTableConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(ctx, resourceName, &routeTable),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`route-table/.+$`)),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrVPCID, vpcResourceName, names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDefaultRouteTableImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCDefaultRouteTable_Disappears_vpc(t *testing.T) {
	ctx := acctest.Context(t)
	var routeTable awstypes.RouteTable
	var vpc awstypes.Vpc
	resourceName := "aws_default_route_table.test"
	vpcResourceName := "aws_vpc.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultRouteTableConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(ctx, resourceName, &routeTable),
					acctest.CheckVPCExists(ctx, vpcResourceName, &vpc),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceVPC(), vpcResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCDefaultRouteTable_Route_mode(t *testing.T) {
	ctx := acctest.Context(t)
	var routeTable awstypes.RouteTable
	resourceName := "aws_default_route_table.test"
	igwResourceName := "aws_internet_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	destinationCidr := "10.2.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDefaultRouteTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultRouteTableConfig_ipv4InternetGateway(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(ctx, resourceName, &routeTable),
					testAccCheckRouteTableNumberOfRoutes(&routeTable, 2),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`route-table/.+$`)),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "route.#", acctest.Ct1),
					testAccCheckRouteTableRoute(resourceName, names.AttrCIDRBlock, destinationCidr, "gateway_id", igwResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDefaultRouteTableImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccVPCDefaultRouteTableConfig_noBlock(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(ctx, resourceName, &routeTable),
					testAccCheckRouteTableNumberOfRoutes(&routeTable, 2),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`route-table/.+$`)),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", acctest.Ct0),
					// The route block from the previous step should still be
					// present, because no blocks means "ignore existing blocks".
					resource.TestCheckResourceAttr(resourceName, "route.#", acctest.Ct1),
					testAccCheckRouteTableRoute(resourceName, names.AttrCIDRBlock, destinationCidr, "gateway_id", igwResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				Config: testAccVPCDefaultRouteTableConfig_blocksExplicitZero(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(ctx, resourceName, &routeTable),
					testAccCheckRouteTableNumberOfRoutes(&routeTable, 1),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`route-table/.+$`)),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", acctest.Ct0),
					// This config uses attribute syntax to set zero routes
					// explicitly, so should remove the one we created before.
					resource.TestCheckResourceAttr(resourceName, "route.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
		},
	})
}

func TestAccVPCDefaultRouteTable_swap(t *testing.T) {
	ctx := acctest.Context(t)
	var routeTable awstypes.RouteTable
	resourceName := "aws_default_route_table.test"
	igwResourceName := "aws_internet_gateway.test"
	rtResourceName := "aws_route_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	destinationCidr1 := "10.2.0.0/16"
	destinationCidr2 := "10.3.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDefaultRouteTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultRouteTableConfig_ipv4InternetGateway(rName, destinationCidr1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(ctx, resourceName, &routeTable),
					testAccCheckRouteTableNumberOfRoutes(&routeTable, 2),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`route-table/.+$`)),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "route.#", acctest.Ct1),
					testAccCheckRouteTableRoute(resourceName, names.AttrCIDRBlock, destinationCidr1, "gateway_id", igwResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDefaultRouteTableImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},

			// This config will swap out the original Default Route Table and replace
			// it with the custom route table. While this is not advised, it's a
			// behavior that may happen, in which case a follow up plan will show (in
			// this case) a diff as the table now needs to be updated to match the
			// config
			{
				Config: testAccVPCDefaultRouteTableConfig_swap(rName, destinationCidr1, destinationCidr2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(ctx, resourceName, &routeTable),
					testAccCheckRouteTableNumberOfRoutes(&routeTable, 2),
					resource.TestCheckResourceAttr(resourceName, "route.#", acctest.Ct1),
					testAccCheckRouteTableRoute(resourceName, names.AttrCIDRBlock, destinationCidr1, "gateway_id", igwResourceName, names.AttrID),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccVPCDefaultRouteTableConfig_swap(rName, destinationCidr1, destinationCidr2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(ctx, resourceName, &routeTable),
					testAccCheckRouteTableNumberOfRoutes(&routeTable, 2),
					resource.TestCheckResourceAttr(resourceName, "route.#", acctest.Ct1),
					testAccCheckRouteTableRoute(resourceName, names.AttrCIDRBlock, destinationCidr1, "gateway_id", igwResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, rtResourceName, names.AttrID),
				),
				// Follow up plan will now show a diff as the destination CIDR on the aws_route_table
				// (now also the aws_default_route_table) will change from destinationCidr1 to destinationCidr2.
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccVPCDefaultRouteTable_ipv4ToTransitGateway(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var routeTable awstypes.RouteTable
	resourceName := "aws_default_route_table.test"
	tgwResourceName := "aws_ec2_transit_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	destinationCidr := "10.2.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultRouteTableConfig_ipv4TransitGateway(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(ctx, resourceName, &routeTable),
					testAccCheckRouteTableNumberOfRoutes(&routeTable, 2),
					resource.TestCheckResourceAttr(resourceName, "route.#", acctest.Ct1),
					testAccCheckRouteTableRoute(resourceName, names.AttrCIDRBlock, destinationCidr, names.AttrTransitGatewayID, tgwResourceName, names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDefaultRouteTableImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCDefaultRouteTable_ipv4ToVPCEndpoint(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var routeTable awstypes.RouteTable
	resourceName := "aws_default_route_table.test"
	vpceResourceName := "aws_vpc_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	destinationCidr := "0.0.0.0/0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckELBv2GatewayLoadBalancer(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID, "elasticloadbalancing"),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultRouteTableConfig_ipv4Endpoint(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(ctx, resourceName, &routeTable),
					testAccCheckRouteTableNumberOfRoutes(&routeTable, 2),
					resource.TestCheckResourceAttr(resourceName, "route.#", acctest.Ct1),
					testAccCheckRouteTableRoute(resourceName, names.AttrCIDRBlock, destinationCidr, names.AttrVPCEndpointID, vpceResourceName, names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDefaultRouteTableImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			// Default route tables do not currently have a method to remove routes during deletion.
			// VPC Endpoints will not delete unless the route is removed prior, otherwise will error:
			// InvalidParameter: Endpoint must be removed from route table before deletion
			{
				Config: testAccVPCDefaultRouteTableConfig_ipv4EndpointNo(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(ctx, resourceName, &routeTable),
				),
			},
		},
	})
}

func TestAccVPCDefaultRouteTable_vpcEndpointAssociation(t *testing.T) {
	ctx := acctest.Context(t)
	var routeTable awstypes.RouteTable
	resourceName := "aws_default_route_table.test"
	igwResourceName := "aws_internet_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	destinationCidr := "10.2.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDefaultRouteTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultRouteTableConfig_endpointAssociation(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(ctx, resourceName, &routeTable),
					testAccCheckRouteTableNumberOfRoutes(&routeTable, 3),
					resource.TestCheckResourceAttr(resourceName, "route.#", acctest.Ct1),
					testAccCheckRouteTableRoute(resourceName, names.AttrCIDRBlock, destinationCidr, "gateway_id", igwResourceName, names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDefaultRouteTableImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCDefaultRouteTable_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var routeTable awstypes.RouteTable
	resourceName := "aws_default_route_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultRouteTableConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(ctx, resourceName, &routeTable),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config: testAccVPCDefaultRouteTableConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(ctx, resourceName, &routeTable),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccVPCDefaultRouteTableConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(ctx, resourceName, &routeTable),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccVPCDefaultRouteTable_conditionalCIDRBlock(t *testing.T) {
	ctx := acctest.Context(t)
	var routeTable awstypes.RouteTable
	resourceName := "aws_default_route_table.test"
	igwResourceName := "aws_internet_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	destinationCidr := "10.2.0.0/16"
	destinationIpv6Cidr := "::/0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultRouteTableConfig_conditionalIPv4v6(rName, destinationCidr, destinationIpv6Cidr, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(ctx, resourceName, &routeTable),
					testAccCheckRouteTableRoute(resourceName, names.AttrCIDRBlock, destinationCidr, "gateway_id", igwResourceName, names.AttrID),
				),
			},
			{
				Config: testAccVPCDefaultRouteTableConfig_conditionalIPv4v6(rName, destinationCidr, destinationIpv6Cidr, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(ctx, resourceName, &routeTable),
					testAccCheckRouteTableRoute(resourceName, "ipv6_cidr_block", destinationIpv6Cidr, "gateway_id", igwResourceName, names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDefaultRouteTableImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccVPCDefaultRouteTable_prefixListToInternetGateway(t *testing.T) {
	ctx := acctest.Context(t)
	var routeTable awstypes.RouteTable
	resourceName := "aws_default_route_table.test"
	igwResourceName := "aws_internet_gateway.test"
	plResourceName := "aws_ec2_managed_prefix_list.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckManagedPrefixList(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultRouteTableConfig_prefixListInternetGateway(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(ctx, resourceName, &routeTable),
					testAccCheckRouteTableNumberOfRoutes(&routeTable, 2),
					resource.TestCheckResourceAttr(resourceName, "route.#", acctest.Ct1),
					testAccCheckRouteTablePrefixListRoute(resourceName, plResourceName, "gateway_id", igwResourceName, names.AttrID),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccDefaultRouteTableImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			// Default route tables do not currently have a method to remove routes during deletion.
			// Managed prefix lists will not delete unless the route is removed prior, otherwise will error:
			// "unexpected state 'delete-failed', wanted target 'delete-complete'"
			{
				Config: testAccVPCDefaultRouteTableConfig_prefixListInternetGatewayNo(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(ctx, resourceName, &routeTable),
				),
			},
		},
	})
}

func TestAccVPCDefaultRouteTable_revokeExistingRules(t *testing.T) {
	ctx := acctest.Context(t)
	var routeTable awstypes.RouteTable
	resourceName := "aws_default_route_table.test"
	rtResourceName := "aws_route_table.test"
	eoigwResourceName := "aws_egress_only_internet_gateway.test"
	igwResourceName := "aws_internet_gateway.test"
	vgwResourceName := "aws_vpn_gateway.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccVPCDefaultRouteTableConfig_revokeExistingRulesCustom(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(ctx, rtResourceName, &routeTable),
					testAccCheckRouteTableNumberOfRoutes(&routeTable, 3),
					resource.TestCheckResourceAttr(rtResourceName, "propagating_vgws.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(rtResourceName, "propagating_vgws.*", vgwResourceName, names.AttrID),
					resource.TestCheckResourceAttr(rtResourceName, "route.#", acctest.Ct1),
					testAccCheckRouteTableRoute(rtResourceName, "ipv6_cidr_block", "::/0", "egress_only_gateway_id", eoigwResourceName, names.AttrID),
					resource.TestCheckResourceAttr(rtResourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(rtResourceName, "tags.Name", rName),
				),
			},
			{
				Config: testAccVPCDefaultRouteTableConfig_revokeExistingRulesCustomToMain(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(ctx, rtResourceName, &routeTable),
					testAccCheckRouteTableNumberOfRoutes(&routeTable, 3),
					resource.TestCheckResourceAttr(rtResourceName, "propagating_vgws.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(rtResourceName, "propagating_vgws.*", vgwResourceName, names.AttrID),
					resource.TestCheckResourceAttr(rtResourceName, "route.#", acctest.Ct1),
					testAccCheckRouteTableRoute(rtResourceName, "ipv6_cidr_block", "::/0", "egress_only_gateway_id", eoigwResourceName, names.AttrID),
					resource.TestCheckResourceAttr(rtResourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(rtResourceName, "tags.Name", rName),
				),
			},
			{
				Config: testAccVPCDefaultRouteTableConfig_revokeExistingRulesOverlaysCustom(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(ctx, resourceName, &routeTable),
					testAccCheckRouteTableNumberOfRoutes(&routeTable, 3),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "route.#", acctest.Ct1),
					testAccCheckRouteTableRoute(resourceName, names.AttrCIDRBlock, "0.0.0.0/0", "gateway_id", igwResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
				// The plan on refresh will not be empty as the custom route table resource's routes and propagating VGWs have
				// been modified since the default route table's routes and propagating VGWs now overlay the custom route table.
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckDefaultRouteTableDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_default_route_table" {
				continue
			}

			_, err := tfec2.FindRouteTableByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Route table %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccDefaultRouteTableImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes[names.AttrVPCID], nil
	}
}

func testAccVPCDefaultRouteTableConfig_id(defaultRouteTableId string) string {
	return fmt.Sprintf(`
resource "aws_default_route_table" "test" {
  default_route_table_id = %[1]q
}
`, defaultRouteTableId)
}

func testAccVPCDefaultRouteTableConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id
}
`, rName)
}

func testAccVPCDefaultRouteTableConfig_ipv4InternetGateway(rName, destinationCidr string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id

  route {
    cidr_block = %[2]q
    gateway_id = aws_internet_gateway.test.id
  }

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
`, rName, destinationCidr)
}

func testAccVPCDefaultRouteTableConfig_noBlock(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}`, rName)
}

func testAccVPCDefaultRouteTableConfig_blocksExplicitZero(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id

  route = []

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}`, rName)
}

func testAccVPCDefaultRouteTableConfig_swap(rName, destinationCidr1, destinationCidr2 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id

  route {
    cidr_block = %[2]q
    gateway_id = aws_internet_gateway.test.id
  }

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

  route {
    cidr_block = %[3]q
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_main_route_table_association" "test" {
  vpc_id         = aws_vpc.test.id
  route_table_id = aws_route_table.test.id
}
`, rName, destinationCidr1, destinationCidr2)
}

func testAccVPCDefaultRouteTableConfig_ipv4TransitGateway(rName, destinationCidr string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(), fmt.Sprintf(`
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

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id

  route {
    cidr_block         = %[2]q
    transit_gateway_id = aws_ec2_transit_gateway_vpc_attachment.test.transit_gateway_id
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, destinationCidr))
}

func testAccVPCDefaultRouteTableConfig_ipv4Endpoint(rName, destinationCidr string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.10.10.0/25"

  tags = {
    Name = %[1]q
  }
}

# Another route destination for update
resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

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
  allowed_principals         = [data.aws_caller_identity.current.arn]
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

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id

  route {
    cidr_block      = %[2]q
    vpc_endpoint_id = aws_vpc_endpoint.test.id
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, destinationCidr))
}

func testAccVPCDefaultRouteTableConfig_ipv4EndpointNo(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.10.10.0/25"

  tags = {
    Name = %[1]q
  }
}

# Another route destination for update
resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

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
  allowed_principals         = [data.aws_caller_identity.current.arn]
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

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVPCDefaultRouteTableConfig_endpointAssociation(rName, destinationCidr string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

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

resource "aws_vpc_endpoint" "test" {
  vpc_id          = aws_vpc.test.id
  service_name    = "com.amazonaws.${data.aws_region.current.name}.s3"
  route_table_ids = [aws_vpc.test.default_route_table_id]

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id

  tags = {
    Name = %[1]q
  }

  route {
    cidr_block = %[2]q
    gateway_id = aws_internet_gateway.test.id
  }
}
`, rName, destinationCidr)
}

func testAccVPCDefaultRouteTableConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccVPCDefaultRouteTableConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccVPCDefaultRouteTableConfig_conditionalIPv4v6(rName, destinationCidr, destinationIpv6Cidr string, ipv6Route bool) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

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

locals {
  ipv6             = %[4]t
  destination      = %[2]q
  destination_ipv6 = %[3]q
}

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id

  route {
    cidr_block      = local.ipv6 ? null : local.destination
    ipv6_cidr_block = local.ipv6 ? local.destination_ipv6 : null
    gateway_id      = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, destinationCidr, destinationIpv6Cidr, ipv6Route)
}

func testAccVPCDefaultRouteTableConfig_prefixListInternetGateway(rName string) string {
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

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id

  route {
    destination_prefix_list_id = aws_ec2_managed_prefix_list.test.id
    gateway_id                 = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCDefaultRouteTableConfig_prefixListInternetGatewayNo(rName string) string {
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

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCDefaultRouteTableConfig_revokeExistingRulesCustom(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway_attachment" "test" {
  vpc_id         = aws_vpc.test.id
  vpn_gateway_id = aws_vpn_gateway.test.id
}

resource "aws_egress_only_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  propagating_vgws = [aws_vpn_gateway_attachment.test.vpn_gateway_id]

  route {
    ipv6_cidr_block        = "::/0"
    egress_only_gateway_id = aws_egress_only_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVPCDefaultRouteTableConfig_revokeExistingRulesCustomToMain(rName string) string {
	return acctest.ConfigCompose(
		testAccVPCDefaultRouteTableConfig_revokeExistingRulesCustom(rName),
		`
resource "aws_main_route_table_association" "test" {
  vpc_id         = aws_vpc.test.id
  route_table_id = aws_route_table.test.id
}
`)
}

func testAccVPCDefaultRouteTableConfig_revokeExistingRulesOverlaysCustom(rName string) string {
	return acctest.ConfigCompose(
		testAccVPCDefaultRouteTableConfig_revokeExistingRulesCustomToMain(rName),
		fmt.Sprintf(`
resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_route_table.test.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccPreCheckELBv2GatewayLoadBalancer(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ELBV2Client(ctx)

	input := &elasticloadbalancingv2.DescribeAccountLimitsInput{}

	output, err := conn.DescribeAccountLimits(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected ELBv2 Gateway Load Balancer PreCheck error: %s", err)
	}

	if output == nil {
		t.Fatal("unexpected ELBv2 Gateway Load Balancer PreCheck error: empty response")
	}

	for _, limit := range output.Limits {
		if aws.ToString(limit.Name) == "gateway-load-balancers" {
			return
		}
	}

	t.Skip("skipping acceptance testing: region does not support ELBv2 Gateway Load Balancers")
}
