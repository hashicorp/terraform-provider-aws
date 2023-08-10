// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccTransitGateway_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Connect": {
			"basic":      testAccTransitGatewayConnect_basic,
			"disappears": testAccTransitGatewayConnect_disappears,
			"tags":       testAccTransitGatewayConnect_tags,
			"TransitGatewayDefaultRouteTableAssociation":                       testAccTransitGatewayConnect_TransitGatewayDefaultRouteTableAssociation,
			"TransitGatewayDefaultRouteTableAssociationAndPropagationDisabled": testAccTransitGatewayConnect_TransitGatewayDefaultRouteTableAssociationAndPropagationDisabled,
			"TransitGatewayDefaultRouteTablePropagation":                       testAccTransitGatewayConnect_TransitGatewayDefaultRouteTablePropagation,
		},
		"ConnectPeer": {
			"basic":                 testAccTransitGatewayConnectPeer_basic,
			"disappears":            testAccTransitGatewayConnectPeer_disappears,
			"tags":                  testAccTransitGatewayConnectPeer_tags,
			"BgpAsn":                testAccTransitGatewayConnectPeer_bgpASN,
			"InsideCidrBlocks":      testAccTransitGatewayConnectPeer_insideCIDRBlocks,
			"TransitGatewayAddress": testAccTransitGatewayConnectPeer_TransitGatewayAddress,
		},
		"Gateway": {
			"basic":                       testAccTransitGateway_basic,
			"disappears":                  testAccTransitGateway_disappears,
			"tags":                        testAccTransitGateway_tags,
			"AmazonSideASN":               testAccTransitGateway_AmazonSideASN,
			"AutoAcceptSharedAttachments": testAccTransitGateway_AutoAcceptSharedAttachments,
			"CidrBlocks":                  testAccTransitGateway_cidrBlocks,
			"DefaultRouteTableAssociationAndPropagationDisabled": testAccTransitGateway_DefaultRouteTableAssociationAndPropagationDisabled,
			"DefaultRouteTableAssociation":                       testAccTransitGateway_DefaultRouteTableAssociation,
			"DefaultRouteTablePropagation":                       testAccTransitGateway_DefaultRouteTablePropagation,
			"Description":                                        testAccTransitGateway_Description,
			"DnsSupport":                                         testAccTransitGateway_DNSSupport,
			"VpnEcmpSupport":                                     testAccTransitGateway_VPNECMPSupport,
		},
		"MulticastDomain": {
			"basic":         testAccTransitGatewayMulticastDomain_basic,
			"disappears":    testAccTransitGatewayMulticastDomain_disappears,
			"tags":          testAccTransitGatewayMulticastDomain_tags,
			"IGMPv2Support": testAccTransitGatewayMulticastDomain_igmpv2Support,
		},
		"MulticastDomainAssociation": {
			"basic":            testAccTransitGatewayMulticastDomainAssociation_basic,
			"disappears":       testAccTransitGatewayMulticastDomainAssociation_disappears,
			"DomainDisappears": testAccTransitGatewayMulticastDomainAssociation_Disappears_domain,
			"TwoAssociations":  testAccTransitGatewayMulticastDomainAssociation_twoAssociations,
		},
		"MulticastGroupMember": {
			"basic":            testAccTransitGatewayMulticastGroupMember_basic,
			"disappears":       testAccTransitGatewayMulticastGroupMember_disappears,
			"DomainDisappears": testAccTransitGatewayMulticastGroupMember_Disappears_domain,
			"TwoMembers":       testAccTransitGatewayMulticastGroupMember_twoMembers,
		},
		"MulticastGroupSource": {
			"basic":            testAccTransitGatewayMulticastGroupSource_basic,
			"disappears":       testAccTransitGatewayMulticastGroupSource_disappears,
			"DomainDisappears": testAccTransitGatewayMulticastGroupSource_Disappears_domain,
		},
		"PeeringAttachment": {
			"basic":            testAccTransitGatewayPeeringAttachment_basic,
			"disappears":       testAccTransitGatewayPeeringAttachment_disappears,
			"tags":             testAccTransitGatewayPeeringAttachment_tags,
			"DifferentAccount": testAccTransitGatewayPeeringAttachment_differentAccount,
		},
		"PeeringAttachmentAccepter": {
			"basic":            testAccTransitGatewayPeeringAttachmentAccepter_basic,
			"tags":             testAccTransitGatewayPeeringAttachmentAccepter_tags,
			"DifferentAccount": testAccTransitGatewayPeeringAttachmentAccepter_differentAccount,
		},
		"PolicyTable": {
			"basic":                    testAccTransitGatewayPolicyTable_basic,
			"disappears":               testAccTransitGatewayPolicyTable_disappears,
			"disappearsTransitGateway": testAccTransitGatewayPolicyTable_disappears_TransitGateway,
			"tags":                     testAccTransitGatewayPolicyTable_tags,
		},
		"PolicyTableAssociation": {
			"basic":      testAccTransitGatewayPolicyTableAssociation_basic,
			"disappears": testAccTransitGatewayPolicyTableAssociation_disappears,
		},
		"PrefixListReference": {
			"basic":                      testAccTransitGatewayPrefixListReference_basic,
			"disappears":                 testAccTransitGatewayPrefixListReference_disappears,
			"disappearsTransitGateway":   testAccTransitGatewayPrefixListReference_disappears_TransitGateway,
			"TransitGatewayAttachmentId": testAccTransitGatewayPrefixListReference_TransitGatewayAttachmentID,
		},
		"Route": {
			"basic":                              testAccTransitGatewayRoute_basic,
			"basicIpv6":                          testAccTransitGatewayRoute_basic_ipv6,
			"blackhole":                          testAccTransitGatewayRoute_blackhole,
			"disappears":                         testAccTransitGatewayRoute_disappears,
			"disappearsTransitGatewayAttachment": testAccTransitGatewayRoute_disappears_TransitGatewayAttachment,
		},
		"RouteTable": {
			"basic":                    testAccTransitGatewayRouteTable_basic,
			"disappears":               testAccTransitGatewayRouteTable_disappears,
			"disappearsTransitGateway": testAccTransitGatewayRouteTable_disappears_TransitGateway,
			"tags":                     testAccTransitGatewayRouteTable_tags,
		},
		"RouteTableAssociation": {
			"basic":                      testAccTransitGatewayRouteTableAssociation_basic,
			"disappears":                 testAccTransitGatewayRouteTableAssociation_disappears,
			"ReplaceExistingAssociation": testAccTransitGatewayRouteTableAssociation_replaceExistingAssociation,
		},
		"RouteTablePropagation": {
			"basic":      testAccTransitGatewayRouteTablePropagation_basic,
			"disappears": testAccTransitGatewayRouteTablePropagation_disappears,
		},
		"VpcAttachment": {
			"basic":                testAccTransitGatewayVPCAttachment_basic,
			"disappears":           testAccTransitGatewayVPCAttachment_disappears,
			"tags":                 testAccTransitGatewayVPCAttachment_tags,
			"ApplianceModeSupport": testAccTransitGatewayVPCAttachment_ApplianceModeSupport,
			"DnsSupport":           testAccTransitGatewayVPCAttachment_DNSSupport,
			"Ipv6Support":          testAccTransitGatewayVPCAttachment_IPv6Support,
			"SharedTransitGateway": testAccTransitGatewayVPCAttachment_SharedTransitGateway,
			"SubnetIds":            testAccTransitGatewayVPCAttachment_SubnetIDs,
			"TransitGatewayDefaultRouteTableAssociation":                       testAccTransitGatewayVPCAttachment_TransitGatewayDefaultRouteTableAssociation,
			"TransitGatewayDefaultRouteTableAssociationAndPropagationDisabled": testAccTransitGatewayVPCAttachment_TransitGatewayDefaultRouteTableAssociationAndPropagationDisabled,
			"TransitGatewayDefaultRouteTablePropagation":                       testAccTransitGatewayVPCAttachment_TransitGatewayDefaultRouteTablePropagation,
		},
		"VpcAttachmentAccepter": {
			"basic": testAccTransitGatewayVPCAttachmentAccepter_basic,
			"tags":  testAccTransitGatewayVPCAttachmentAccepter_tags,
			"TransitGatewayDefaultRouteTableAssociationAndPropagation": testAccTransitGatewayVPCAttachmentAccepter_TransitGatewayDefaultRouteTableAssociationAndPropagation,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}

func testAccTransitGateway_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var transitGateway1 ec2.TransitGateway
	resourceName := "aws_ec2_transit_gateway.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckTransitGateway(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, resourceName, &transitGateway1),
					resource.TestCheckResourceAttr(resourceName, "amazon_side_asn", "64512"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`transit-gateway/tgw-.+`)),
					resource.TestCheckResourceAttrSet(resourceName, "association_default_route_table_id"),
					resource.TestCheckResourceAttr(resourceName, "auto_accept_shared_attachments", ec2.AutoAcceptSharedAttachmentsValueDisable),
					resource.TestCheckResourceAttr(resourceName, "default_route_table_association", ec2.DefaultRouteTableAssociationValueEnable),
					resource.TestCheckResourceAttr(resourceName, "default_route_table_propagation", ec2.DefaultRouteTablePropagationValueEnable),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "dns_support", ec2.DnsSupportValueEnable),
					resource.TestCheckResourceAttr(resourceName, "multicast_support", ec2.MulticastSupportValueDisable),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttrSet(resourceName, "propagation_default_route_table_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "vpn_ecmp_support", ec2.VpnEcmpSupportValueEnable),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccTransitGateway_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var transitGateway1 ec2.TransitGateway
	resourceName := "aws_ec2_transit_gateway.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckTransitGateway(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, resourceName, &transitGateway1),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceTransitGateway(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccTransitGateway_AmazonSideASN(t *testing.T) {
	ctx := acctest.Context(t)
	var transitGateway1, transitGateway2 ec2.TransitGateway
	resourceName := "aws_ec2_transit_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckTransitGateway(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConfig_amazonSideASN(rName, 64513),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, resourceName, &transitGateway1),
					resource.TestCheckResourceAttr(resourceName, "amazon_side_asn", "64513"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayConfig_amazonSideASN(rName, 64514),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, resourceName, &transitGateway2),
					testAccCheckTransitGatewayNotRecreated(&transitGateway1, &transitGateway2),
					resource.TestCheckResourceAttr(resourceName, "amazon_side_asn", "64514"),
				),
			},
		},
	})
}

func testAccTransitGateway_AutoAcceptSharedAttachments(t *testing.T) {
	ctx := acctest.Context(t)
	var transitGateway1, transitGateway2 ec2.TransitGateway
	resourceName := "aws_ec2_transit_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckTransitGateway(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConfig_autoAcceptSharedAttachments(rName, ec2.AutoAcceptSharedAttachmentsValueEnable),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, resourceName, &transitGateway1),
					resource.TestCheckResourceAttr(resourceName, "auto_accept_shared_attachments", ec2.AutoAcceptSharedAttachmentsValueEnable),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayConfig_autoAcceptSharedAttachments(rName, ec2.AutoAcceptSharedAttachmentsValueDisable),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, resourceName, &transitGateway2),
					testAccCheckTransitGatewayNotRecreated(&transitGateway1, &transitGateway2),
					resource.TestCheckResourceAttr(resourceName, "auto_accept_shared_attachments", ec2.AutoAcceptSharedAttachmentsValueDisable),
				),
			},
		},
	})
}

func testAccTransitGateway_cidrBlocks(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2, v3 ec2.TransitGateway
	resourceName := "aws_ec2_transit_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckTransitGateway(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConfig_cidrBlocks2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_cidr_blocks.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "transit_gateway_cidr_blocks.*", "10.120.0.0/24"),
					resource.TestCheckTypeSetElemAttr(resourceName, "transit_gateway_cidr_blocks.*", "2001:1234:1234::/64"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayConfig_cidrBlocks1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, resourceName, &v2),
					testAccCheckTransitGatewayNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_cidr_blocks.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "transit_gateway_cidr_blocks.*", "10.120.0.0/24"),
				),
			},
			{
				Config: testAccTransitGatewayConfig_cidrBlocks2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, resourceName, &v3),
					testAccCheckTransitGatewayNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_cidr_blocks.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "transit_gateway_cidr_blocks.*", "10.120.0.0/24"),
					resource.TestCheckTypeSetElemAttr(resourceName, "transit_gateway_cidr_blocks.*", "2001:1234:1234::/64"),
				),
			},
		},
	})
}

func testAccTransitGateway_DefaultRouteTableAssociationAndPropagationDisabled(t *testing.T) {
	ctx := acctest.Context(t)
	var transitGateway1 ec2.TransitGateway
	resourceName := "aws_ec2_transit_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckTransitGateway(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConfig_defaultRouteTableAssociationAndPropagationDisabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, resourceName, &transitGateway1),
					resource.TestCheckResourceAttr(resourceName, "default_route_table_association", ec2.DefaultRouteTableAssociationValueDisable),
					resource.TestCheckResourceAttr(resourceName, "default_route_table_propagation", ec2.DefaultRouteTablePropagationValueDisable),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccTransitGateway_DefaultRouteTableAssociation(t *testing.T) {
	ctx := acctest.Context(t)
	var transitGateway1, transitGateway2, transitGateway3 ec2.TransitGateway
	resourceName := "aws_ec2_transit_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckTransitGateway(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConfig_defaultRouteTableAssociation(rName, ec2.DefaultRouteTableAssociationValueDisable),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, resourceName, &transitGateway1),
					resource.TestCheckResourceAttr(resourceName, "default_route_table_association", ec2.DefaultRouteTableAssociationValueDisable),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayConfig_defaultRouteTableAssociation(rName, ec2.DefaultRouteTableAssociationValueEnable),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, resourceName, &transitGateway2),
					testAccCheckTransitGatewayRecreated(&transitGateway1, &transitGateway2),
					resource.TestCheckResourceAttr(resourceName, "default_route_table_association", ec2.DefaultRouteTableAssociationValueEnable),
				),
			},
			{
				Config: testAccTransitGatewayConfig_defaultRouteTableAssociation(rName, ec2.DefaultRouteTableAssociationValueDisable),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, resourceName, &transitGateway3),
					testAccCheckTransitGatewayNotRecreated(&transitGateway2, &transitGateway3),
					resource.TestCheckResourceAttr(resourceName, "default_route_table_association", ec2.DefaultRouteTableAssociationValueDisable),
				),
			},
		},
	})
}

func testAccTransitGateway_DefaultRouteTablePropagation(t *testing.T) {
	ctx := acctest.Context(t)
	var transitGateway1, transitGateway2, transitGateway3 ec2.TransitGateway
	resourceName := "aws_ec2_transit_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckTransitGateway(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConfig_defaultRouteTablePropagation(rName, ec2.DefaultRouteTablePropagationValueDisable),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, resourceName, &transitGateway1),
					resource.TestCheckResourceAttr(resourceName, "default_route_table_propagation", ec2.DefaultRouteTablePropagationValueDisable),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayConfig_defaultRouteTablePropagation(rName, ec2.DefaultRouteTablePropagationValueEnable),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, resourceName, &transitGateway2),
					testAccCheckTransitGatewayRecreated(&transitGateway1, &transitGateway2),
					resource.TestCheckResourceAttr(resourceName, "default_route_table_propagation", ec2.DefaultRouteTablePropagationValueEnable),
				),
			},
			{
				Config: testAccTransitGatewayConfig_defaultRouteTablePropagation(rName, ec2.DefaultRouteTablePropagationValueDisable),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, resourceName, &transitGateway3),
					testAccCheckTransitGatewayNotRecreated(&transitGateway2, &transitGateway3),
					resource.TestCheckResourceAttr(resourceName, "default_route_table_propagation", ec2.DefaultRouteTablePropagationValueDisable),
				),
			},
		},
	})
}

func testAccTransitGateway_DNSSupport(t *testing.T) {
	ctx := acctest.Context(t)
	var transitGateway1, transitGateway2 ec2.TransitGateway
	resourceName := "aws_ec2_transit_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckTransitGateway(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConfig_dnsSupport(rName, ec2.DnsSupportValueDisable),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, resourceName, &transitGateway1),
					resource.TestCheckResourceAttr(resourceName, "dns_support", ec2.DnsSupportValueDisable),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayConfig_dnsSupport(rName, ec2.DnsSupportValueEnable),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, resourceName, &transitGateway2),
					testAccCheckTransitGatewayNotRecreated(&transitGateway1, &transitGateway2),
					resource.TestCheckResourceAttr(resourceName, "dns_support", ec2.DnsSupportValueEnable),
				),
			},
		},
	})
}

func testAccTransitGateway_VPNECMPSupport(t *testing.T) {
	ctx := acctest.Context(t)
	var transitGateway1, transitGateway2 ec2.TransitGateway
	resourceName := "aws_ec2_transit_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckTransitGateway(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConfig_vpnECMPSupport(rName, ec2.VpnEcmpSupportValueDisable),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, resourceName, &transitGateway1),
					resource.TestCheckResourceAttr(resourceName, "vpn_ecmp_support", ec2.VpnEcmpSupportValueDisable),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayConfig_vpnECMPSupport(rName, ec2.VpnEcmpSupportValueEnable),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, resourceName, &transitGateway2),
					testAccCheckTransitGatewayNotRecreated(&transitGateway1, &transitGateway2),
					resource.TestCheckResourceAttr(resourceName, "vpn_ecmp_support", ec2.VpnEcmpSupportValueEnable),
				),
			},
		},
	})
}

func testAccTransitGateway_Description(t *testing.T) {
	ctx := acctest.Context(t)
	var transitGateway1, transitGateway2 ec2.TransitGateway
	resourceName := "aws_ec2_transit_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckTransitGateway(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, resourceName, &transitGateway1),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayConfig_description(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, resourceName, &transitGateway2),
					testAccCheckTransitGatewayNotRecreated(&transitGateway1, &transitGateway2),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
		},
	})
}

func testAccTransitGateway_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var transitGateway1, transitGateway2, transitGateway3 ec2.TransitGateway
	resourceName := "aws_ec2_transit_gateway.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckTransitGateway(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConfig_tags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, resourceName, &transitGateway1),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayConfig_tags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, resourceName, &transitGateway2),
					testAccCheckTransitGatewayNotRecreated(&transitGateway1, &transitGateway2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccTransitGatewayConfig_tags1("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, resourceName, &transitGateway3),
					testAccCheckTransitGatewayNotRecreated(&transitGateway2, &transitGateway3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckTransitGatewayExists(ctx context.Context, n string, v *ec2.TransitGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Transit Gateway ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn(ctx)

		output, err := tfec2.FindTransitGatewayByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckTransitGatewayDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ec2_transit_gateway" {
				continue
			}

			_, err := tfec2.FindTransitGatewayByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("EC2 Transit Gateway %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckTransitGatewayNotRecreated(i, j *ec2.TransitGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.TransitGatewayId) != aws.StringValue(j.TransitGatewayId) {
			return errors.New("EC2 Transit Gateway was recreated")
		}

		return nil
	}
}

func testAccCheckTransitGatewayRecreated(i, j *ec2.TransitGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.TransitGatewayId) == aws.StringValue(j.TransitGatewayId) {
			return errors.New("EC2 Transit Gateway was not recreated")
		}

		return nil
	}
}

func testAccCheckTransitGatewayAssociationDefaultRouteTableAttachmentAssociated(ctx context.Context, transitGateway *ec2.TransitGateway, transitGatewayAttachment interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		transitGatewayRouteTableID := aws.StringValue(transitGateway.Options.AssociationDefaultRouteTableId)

		if transitGatewayRouteTableID == "" {
			return errors.New("EC2 Transit Gateway Association Default Route Table empty")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn(ctx)

		var transitGatewayAttachmentID string
		switch transitGatewayAttachment := transitGatewayAttachment.(type) {
		case *ec2.TransitGatewayVpcAttachment:
			transitGatewayAttachmentID = aws.StringValue(transitGatewayAttachment.TransitGatewayAttachmentId)
		case *ec2.TransitGatewayConnect:
			transitGatewayAttachmentID = aws.StringValue(transitGatewayAttachment.TransitGatewayAttachmentId)
		}

		_, err := tfec2.FindTransitGatewayRouteTableAssociationByTwoPartKey(ctx, conn, transitGatewayRouteTableID, transitGatewayAttachmentID)

		if tfresource.NotFound(err) {
			return errors.New("EC2 Transit Gateway Route Table Association not found")
		}

		return err
	}
}

func testAccCheckTransitGatewayAssociationDefaultRouteTableAttachmentNotAssociated(ctx context.Context, transitGateway *ec2.TransitGateway, transitGatewayAttachment interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		transitGatewayRouteTableID := aws.StringValue(transitGateway.Options.AssociationDefaultRouteTableId)

		if transitGatewayRouteTableID == "" {
			return nil
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn(ctx)

		var transitGatewayAttachmentID string
		switch transitGatewayAttachment := transitGatewayAttachment.(type) {
		case *ec2.TransitGatewayVpcAttachment:
			transitGatewayAttachmentID = aws.StringValue(transitGatewayAttachment.TransitGatewayAttachmentId)
		case *ec2.TransitGatewayConnect:
			transitGatewayAttachmentID = aws.StringValue(transitGatewayAttachment.TransitGatewayAttachmentId)
		}

		_, err := tfec2.FindTransitGatewayRouteTableAssociationByTwoPartKey(ctx, conn, transitGatewayRouteTableID, transitGatewayAttachmentID)

		if tfresource.NotFound(err) {
			return nil
		}

		if err != nil {
			return err
		}

		return errors.New("EC2 Transit Gateway Route Table Association found")
	}
}

func testAccCheckTransitGatewayPropagationDefaultRouteTableAttachmentPropagated(ctx context.Context, transitGateway *ec2.TransitGateway, transitGatewayAttachment interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		transitGatewayRouteTableID := aws.StringValue(transitGateway.Options.PropagationDefaultRouteTableId)

		if transitGatewayRouteTableID == "" {
			return errors.New("EC2 Transit Gateway Propagation Default Route Table empty")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn(ctx)

		var transitGatewayAttachmentID string
		switch transitGatewayAttachment := transitGatewayAttachment.(type) {
		case *ec2.TransitGatewayVpcAttachment:
			transitGatewayAttachmentID = aws.StringValue(transitGatewayAttachment.TransitGatewayAttachmentId)
		case *ec2.TransitGatewayConnect:
			transitGatewayAttachmentID = aws.StringValue(transitGatewayAttachment.TransitGatewayAttachmentId)
		}

		_, err := tfec2.FindTransitGatewayRouteTablePropagationByTwoPartKey(ctx, conn, transitGatewayRouteTableID, transitGatewayAttachmentID)

		if tfresource.NotFound(err) {
			return errors.New("EC2 Transit Gateway Route Table Propagation not enabled")
		}

		return err
	}
}

func testAccCheckTransitGatewayPropagationDefaultRouteTableAttachmentNotPropagated(ctx context.Context, transitGateway *ec2.TransitGateway, transitGatewayAttachment interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		transitGatewayRouteTableID := aws.StringValue(transitGateway.Options.PropagationDefaultRouteTableId)

		if transitGatewayRouteTableID == "" {
			return nil
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn(ctx)

		var transitGatewayAttachmentID string
		switch transitGatewayAttachment := transitGatewayAttachment.(type) {
		case *ec2.TransitGatewayVpcAttachment:
			transitGatewayAttachmentID = aws.StringValue(transitGatewayAttachment.TransitGatewayAttachmentId)
		case *ec2.TransitGatewayConnect:
			transitGatewayAttachmentID = aws.StringValue(transitGatewayAttachment.TransitGatewayAttachmentId)
		}

		_, err := tfec2.FindTransitGatewayRouteTablePropagationByTwoPartKey(ctx, conn, transitGatewayRouteTableID, transitGatewayAttachmentID)

		if tfresource.NotFound(err) {
			return nil
		}

		if err != nil {
			return err
		}

		return errors.New("EC2 Transit Gateway Route Table Propagation enabled")
	}
}

func testAccPreCheckTransitGateway(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn(ctx)

	input := &ec2.DescribeTransitGatewaysInput{
		MaxResults: aws.Int64(5),
	}

	_, err := conn.DescribeTransitGatewaysWithContext(ctx, input)

	if acctest.PreCheckSkipError(err) || tfawserr.ErrCodeEquals(err, "InvalidAction") {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccPreCheckTransitGatewayConnect(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn(ctx)

	input := &ec2.DescribeTransitGatewayConnectsInput{
		MaxResults: aws.Int64(5),
	}

	_, err := conn.DescribeTransitGatewayConnectsWithContext(ctx, input)

	if acctest.PreCheckSkipError(err) || tfawserr.ErrCodeEquals(err, "InvalidAction") {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccPreCheckTransitGatewayVPCAttachment(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn(ctx)

	input := &ec2.DescribeTransitGatewayVpcAttachmentsInput{
		MaxResults: aws.Int64(5),
	}

	_, err := conn.DescribeTransitGatewayVpcAttachmentsWithContext(ctx, input)

	if acctest.PreCheckSkipError(err) || tfawserr.ErrCodeEquals(err, "InvalidAction") {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccTransitGatewayConfig_basic() string {
	return `
resource "aws_ec2_transit_gateway" "test" {}
`
}

func testAccTransitGatewayConfig_amazonSideASN(rName string, amazonSideASN int) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  amazon_side_asn = %[2]d

  tags = {
    Name = %[1]q
  }
}
`, rName, amazonSideASN)
}

func testAccTransitGatewayConfig_autoAcceptSharedAttachments(rName, autoAcceptSharedAttachments string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  auto_accept_shared_attachments = %[2]q

  tags = {
    Name = %[1]q
  }
}
`, rName, autoAcceptSharedAttachments)
}

func testAccTransitGatewayConfig_defaultRouteTableAssociationAndPropagationDisabled(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  default_route_table_association = "disable"
  default_route_table_propagation = "disable"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccTransitGatewayConfig_defaultRouteTableAssociation(rName, defaultRouteTableAssociation string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  default_route_table_association = %[2]q

  tags = {
    Name = %[1]q
  }
}
`, rName, defaultRouteTableAssociation)
}

func testAccTransitGatewayConfig_defaultRouteTablePropagation(rName, defaultRouteTablePropagation string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  default_route_table_propagation = %[2]q

  tags = {
    Name = %[1]q
  }
}
`, rName, defaultRouteTablePropagation)
}

func testAccTransitGatewayConfig_dnsSupport(rName, dnsSupport string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  dns_support = %[2]q

  tags = {
    Name = %[1]q
  }
}
`, rName, dnsSupport)
}

func testAccTransitGatewayConfig_vpnECMPSupport(rName, vpnEcmpSupport string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  vpn_ecmp_support = %[2]q

  tags = {
    Name = %[1]q
  }
}
`, rName, vpnEcmpSupport)
}

func testAccTransitGatewayConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  description = %[2]q

  tags = {
    Name = %[1]q
  }
}
`, rName, description)
}

func testAccTransitGatewayConfig_tags1(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1)
}

func testAccTransitGatewayConfig_tags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccTransitGatewayConfig_cidrBlocks1(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  transit_gateway_cidr_blocks = ["10.120.0.0/24"]

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccTransitGatewayConfig_cidrBlocks2(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  transit_gateway_cidr_blocks = ["10.120.0.0/24", "2001:1234:1234::/64"]

  tags = {
    Name = %[1]q
  }
}
`, rName)
}
