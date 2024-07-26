// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfsync "github.com/hashicorp/terraform-provider-aws/internal/experimental/sync"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccTransitGateway_serial(t *testing.T) {
	t.Parallel()

	semaphore := tfsync.GetSemaphore("TransitGateway", "AWS_EC2_TRANSIT_GATEWAY_LIMIT", 5)
	testCases := map[string]map[string]func(*testing.T, tfsync.Semaphore){
		"Connect": {
			acctest.CtBasic:      testAccTransitGatewayConnect_basic,
			acctest.CtDisappears: testAccTransitGatewayConnect_disappears,
			"tags":               testAccTransitGatewayConnect_tags,
			"TransitGatewayDefaultRouteTableAssociation":                       testAccTransitGatewayConnect_TransitGatewayDefaultRouteTableAssociation,
			"TransitGatewayDefaultRouteTableAssociationAndPropagationDisabled": testAccTransitGatewayConnect_TransitGatewayDefaultRouteTableAssociationAndPropagationDisabled,
			"TransitGatewayDefaultRouteTablePropagation":                       testAccTransitGatewayConnect_TransitGatewayDefaultRouteTablePropagation,
		},
		"ConnectPeer": {
			acctest.CtBasic:         testAccTransitGatewayConnectPeer_basic,
			acctest.CtDisappears:    testAccTransitGatewayConnectPeer_disappears,
			"tags":                  testAccTransitGatewayConnectPeer_tags,
			"BgpAsn":                testAccTransitGatewayConnectPeer_bgpASN,
			"InsideCidrBlocks":      testAccTransitGatewayConnectPeer_insideCIDRBlocks,
			"TransitGatewayAddress": testAccTransitGatewayConnectPeer_TransitGatewayAddress,
		},
		"Gateway": {
			acctest.CtBasic:               testAccTransitGateway_basic,
			acctest.CtDisappears:          testAccTransitGateway_disappears,
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
			acctest.CtBasic:      testAccTransitGatewayMulticastDomain_basic,
			acctest.CtDisappears: testAccTransitGatewayMulticastDomain_disappears,
			"tags":               testAccTransitGatewayMulticastDomain_tags,
			"IGMPv2Support":      testAccTransitGatewayMulticastDomain_igmpv2Support,
		},
		"MulticastDomainAssociation": {
			acctest.CtBasic:      testAccTransitGatewayMulticastDomainAssociation_basic,
			acctest.CtDisappears: testAccTransitGatewayMulticastDomainAssociation_disappears,
			"DomainDisappears":   testAccTransitGatewayMulticastDomainAssociation_Disappears_domain,
			"TwoAssociations":    testAccTransitGatewayMulticastDomainAssociation_twoAssociations,
		},
		"MulticastGroupMember": {
			acctest.CtBasic:      testAccTransitGatewayMulticastGroupMember_basic,
			acctest.CtDisappears: testAccTransitGatewayMulticastGroupMember_disappears,
			"DomainDisappears":   testAccTransitGatewayMulticastGroupMember_Disappears_domain,
			"TwoMembers":         testAccTransitGatewayMulticastGroupMember_twoMembers,
		},
		"MulticastGroupSource": {
			acctest.CtBasic:      testAccTransitGatewayMulticastGroupSource_basic,
			acctest.CtDisappears: testAccTransitGatewayMulticastGroupSource_disappears,
			"DomainDisappears":   testAccTransitGatewayMulticastGroupSource_Disappears_domain,
		},
		"PeeringAttachment": {
			acctest.CtBasic:      testAccTransitGatewayPeeringAttachment_basic,
			acctest.CtDisappears: testAccTransitGatewayPeeringAttachment_disappears,
			"tags":               testAccTransitGatewayPeeringAttachment_tags,
			"DifferentAccount":   testAccTransitGatewayPeeringAttachment_differentAccount,
			"options":            testAccTransitGatewayPeeringAttachment_options,
		},
		"PeeringAttachmentAccepter": {
			acctest.CtBasic:    testAccTransitGatewayPeeringAttachmentAccepter_basic,
			"tags":             testAccTransitGatewayPeeringAttachmentAccepter_tags,
			"DifferentAccount": testAccTransitGatewayPeeringAttachmentAccepter_differentAccount,
		},
		"PolicyTable": {
			acctest.CtBasic:            testAccTransitGatewayPolicyTable_basic,
			acctest.CtDisappears:       testAccTransitGatewayPolicyTable_disappears,
			"disappearsTransitGateway": testAccTransitGatewayPolicyTable_disappears_TransitGateway,
			"tags":                     testAccTransitGatewayPolicyTable_tags,
		},
		"PolicyTableAssociation": {
			acctest.CtBasic:      testAccTransitGatewayPolicyTableAssociation_basic,
			acctest.CtDisappears: testAccTransitGatewayPolicyTableAssociation_disappears,
		},
		"PrefixListReference": {
			acctest.CtBasic:              testAccTransitGatewayPrefixListReference_basic,
			acctest.CtDisappears:         testAccTransitGatewayPrefixListReference_disappears,
			"disappearsTransitGateway":   testAccTransitGatewayPrefixListReference_disappears_TransitGateway,
			"TransitGatewayAttachmentId": testAccTransitGatewayPrefixListReference_TransitGatewayAttachmentID,
		},
		"Route": {
			acctest.CtBasic:                      testAccTransitGatewayRoute_basic,
			"basicIpv6":                          testAccTransitGatewayRoute_basic_ipv6,
			"blackhole":                          testAccTransitGatewayRoute_blackhole,
			acctest.CtDisappears:                 testAccTransitGatewayRoute_disappears,
			"disappearsTransitGatewayAttachment": testAccTransitGatewayRoute_disappears_TransitGatewayAttachment,
		},
		"RouteTable": {
			acctest.CtBasic:            testAccTransitGatewayRouteTable_basic,
			acctest.CtDisappears:       testAccTransitGatewayRouteTable_disappears,
			"disappearsTransitGateway": testAccTransitGatewayRouteTable_disappears_TransitGateway,
			"tags":                     testAccTransitGatewayRouteTable_tags,
		},
		"RouteTableAssociation": {
			acctest.CtBasic:              testAccTransitGatewayRouteTableAssociation_basic,
			acctest.CtDisappears:         testAccTransitGatewayRouteTableAssociation_disappears,
			"ReplaceExistingAssociation": testAccTransitGatewayRouteTableAssociation_replaceExistingAssociation,
		},
		"RouteTablePropagation": {
			acctest.CtBasic:      testAccTransitGatewayRouteTablePropagation_basic,
			acctest.CtDisappears: testAccTransitGatewayRouteTablePropagation_disappears,
		},
		"VpcAttachment": {
			acctest.CtBasic:        testAccTransitGatewayVPCAttachment_basic,
			acctest.CtDisappears:   testAccTransitGatewayVPCAttachment_disappears,
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
			acctest.CtBasic: testAccTransitGatewayVPCAttachmentAccepter_basic,
			"tags":          testAccTransitGatewayVPCAttachmentAccepter_tags,
			"TransitGatewayDefaultRouteTableAssociationAndPropagation": testAccTransitGatewayVPCAttachmentAccepter_TransitGatewayDefaultRouteTableAssociationAndPropagation,
		},
	}

	acctest.RunLimitedConcurrencyTests2Levels(t, semaphore, testCases)
}

func testAccPreCheckTransitGatewaySynchronize(t *testing.T, semaphore tfsync.Semaphore) {
	tfsync.TestAccPreCheckSyncronize(t, semaphore, "TransitGateway")
}

func testAccTransitGateway_basic(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var transitGateway1 awstypes.TransitGateway
	resourceName := "aws_ec2_transit_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, resourceName, &transitGateway1),
					resource.TestCheckResourceAttr(resourceName, "amazon_side_asn", "64512"),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ec2", regexache.MustCompile(`transit-gateway/tgw-.+`)),
					resource.TestCheckResourceAttrSet(resourceName, "association_default_route_table_id"),
					resource.TestCheckResourceAttr(resourceName, "auto_accept_shared_attachments", string(awstypes.AutoAcceptSharedAttachmentsValueDisable)),
					resource.TestCheckResourceAttr(resourceName, "default_route_table_association", string(awstypes.DefaultRouteTableAssociationValueEnable)),
					resource.TestCheckResourceAttr(resourceName, "default_route_table_propagation", string(awstypes.DefaultRouteTablePropagationValueEnable)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "dns_support", string(awstypes.DnsSupportValueEnable)),
					resource.TestCheckResourceAttr(resourceName, "multicast_support", string(awstypes.MulticastSupportValueDisable)),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrSet(resourceName, "propagation_default_route_table_id"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "vpn_ecmp_support", string(awstypes.VpnEcmpSupportValueEnable)),
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

func testAccTransitGateway_disappears(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var transitGateway1 awstypes.TransitGateway
	resourceName := "aws_ec2_transit_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
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

func testAccTransitGateway_AmazonSideASN(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var transitGateway1, transitGateway2 awstypes.TransitGateway
	resourceName := "aws_ec2_transit_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
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

func testAccTransitGateway_AutoAcceptSharedAttachments(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var transitGateway1, transitGateway2 awstypes.TransitGateway
	resourceName := "aws_ec2_transit_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConfig_autoAcceptSharedAttachments(rName, string(awstypes.AutoAcceptSharedAttachmentsValueEnable)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, resourceName, &transitGateway1),
					resource.TestCheckResourceAttr(resourceName, "auto_accept_shared_attachments", string(awstypes.AutoAcceptSharedAttachmentsValueEnable)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayConfig_autoAcceptSharedAttachments(rName, string(awstypes.AutoAcceptSharedAttachmentsValueDisable)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, resourceName, &transitGateway2),
					testAccCheckTransitGatewayNotRecreated(&transitGateway1, &transitGateway2),
					resource.TestCheckResourceAttr(resourceName, "auto_accept_shared_attachments", string(awstypes.AutoAcceptSharedAttachmentsValueDisable)),
				),
			},
		},
	})
}

func testAccTransitGateway_cidrBlocks(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var v1, v2, v3 awstypes.TransitGateway
	resourceName := "aws_ec2_transit_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConfig_cidrBlocks2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_cidr_blocks.#", acctest.Ct2),
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
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_cidr_blocks.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "transit_gateway_cidr_blocks.*", "10.120.0.0/24"),
				),
			},
			{
				Config: testAccTransitGatewayConfig_cidrBlocks2(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, resourceName, &v3),
					testAccCheckTransitGatewayNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_cidr_blocks.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "transit_gateway_cidr_blocks.*", "10.120.0.0/24"),
					resource.TestCheckTypeSetElemAttr(resourceName, "transit_gateway_cidr_blocks.*", "2001:1234:1234::/64"),
				),
			},
		},
	})
}

func testAccTransitGateway_DefaultRouteTableAssociationAndPropagationDisabled(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var transitGateway1 awstypes.TransitGateway
	resourceName := "aws_ec2_transit_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConfig_defaultRouteTableAssociationAndPropagationDisabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, resourceName, &transitGateway1),
					resource.TestCheckResourceAttr(resourceName, "default_route_table_association", string(awstypes.DefaultRouteTableAssociationValueDisable)),
					resource.TestCheckResourceAttr(resourceName, "default_route_table_propagation", string(awstypes.DefaultRouteTablePropagationValueDisable)),
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

func testAccTransitGateway_DefaultRouteTableAssociation(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var transitGateway1, transitGateway2, transitGateway3 awstypes.TransitGateway
	resourceName := "aws_ec2_transit_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConfig_defaultRouteTableAssociation(rName, string(awstypes.DefaultRouteTableAssociationValueDisable)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, resourceName, &transitGateway1),
					resource.TestCheckResourceAttr(resourceName, "default_route_table_association", string(awstypes.DefaultRouteTableAssociationValueDisable)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayConfig_defaultRouteTableAssociation(rName, string(awstypes.DefaultRouteTableAssociationValueEnable)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, resourceName, &transitGateway2),
					testAccCheckTransitGatewayRecreated(&transitGateway1, &transitGateway2),
					resource.TestCheckResourceAttr(resourceName, "default_route_table_association", string(awstypes.DefaultRouteTableAssociationValueEnable)),
				),
			},
			{
				Config: testAccTransitGatewayConfig_defaultRouteTableAssociation(rName, string(awstypes.DefaultRouteTableAssociationValueDisable)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, resourceName, &transitGateway3),
					testAccCheckTransitGatewayNotRecreated(&transitGateway2, &transitGateway3),
					resource.TestCheckResourceAttr(resourceName, "default_route_table_association", string(awstypes.DefaultRouteTableAssociationValueDisable)),
				),
			},
		},
	})
}

func testAccTransitGateway_DefaultRouteTablePropagation(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var transitGateway1, transitGateway2, transitGateway3 awstypes.TransitGateway
	resourceName := "aws_ec2_transit_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConfig_defaultRouteTablePropagation(rName, string(awstypes.DefaultRouteTablePropagationValueDisable)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, resourceName, &transitGateway1),
					resource.TestCheckResourceAttr(resourceName, "default_route_table_propagation", string(awstypes.DefaultRouteTablePropagationValueDisable)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayConfig_defaultRouteTablePropagation(rName, string(awstypes.DefaultRouteTablePropagationValueEnable)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, resourceName, &transitGateway2),
					testAccCheckTransitGatewayRecreated(&transitGateway1, &transitGateway2),
					resource.TestCheckResourceAttr(resourceName, "default_route_table_propagation", string(awstypes.DefaultRouteTablePropagationValueEnable)),
				),
			},
			{
				Config: testAccTransitGatewayConfig_defaultRouteTablePropagation(rName, string(awstypes.DefaultRouteTablePropagationValueDisable)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, resourceName, &transitGateway3),
					testAccCheckTransitGatewayNotRecreated(&transitGateway2, &transitGateway3),
					resource.TestCheckResourceAttr(resourceName, "default_route_table_propagation", string(awstypes.DefaultRouteTablePropagationValueDisable)),
				),
			},
		},
	})
}

func testAccTransitGateway_DNSSupport(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var transitGateway1, transitGateway2 awstypes.TransitGateway
	resourceName := "aws_ec2_transit_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConfig_dnsSupport(rName, string(awstypes.DnsSupportValueDisable)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, resourceName, &transitGateway1),
					resource.TestCheckResourceAttr(resourceName, "dns_support", string(awstypes.DnsSupportValueDisable)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayConfig_dnsSupport(rName, string(awstypes.DnsSupportValueEnable)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, resourceName, &transitGateway2),
					testAccCheckTransitGatewayNotRecreated(&transitGateway1, &transitGateway2),
					resource.TestCheckResourceAttr(resourceName, "dns_support", string(awstypes.DnsSupportValueEnable)),
				),
			},
		},
	})
}

func testAccTransitGateway_VPNECMPSupport(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var transitGateway1, transitGateway2 awstypes.TransitGateway
	resourceName := "aws_ec2_transit_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConfig_vpnECMPSupport(rName, string(awstypes.VpnEcmpSupportValueDisable)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, resourceName, &transitGateway1),
					resource.TestCheckResourceAttr(resourceName, "vpn_ecmp_support", string(awstypes.VpnEcmpSupportValueDisable)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayConfig_vpnECMPSupport(rName, string(awstypes.VpnEcmpSupportValueEnable)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, resourceName, &transitGateway2),
					testAccCheckTransitGatewayNotRecreated(&transitGateway1, &transitGateway2),
					resource.TestCheckResourceAttr(resourceName, "vpn_ecmp_support", string(awstypes.VpnEcmpSupportValueEnable)),
				),
			},
		},
	})
}

func testAccTransitGateway_Description(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var transitGateway1, transitGateway2 awstypes.TransitGateway
	resourceName := "aws_ec2_transit_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConfig_description(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, resourceName, &transitGateway1),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description1"),
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
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "description2"),
				),
			},
		},
	})
}

func testAccTransitGateway_tags(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	var transitGateway1, transitGateway2, transitGateway3 awstypes.TransitGateway
	resourceName := "aws_ec2_transit_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConfig_tags1(acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, resourceName, &transitGateway1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayConfig_tags2(acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, resourceName, &transitGateway2),
					testAccCheckTransitGatewayNotRecreated(&transitGateway1, &transitGateway2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccTransitGatewayConfig_tags1(acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(ctx, resourceName, &transitGateway3),
					testAccCheckTransitGatewayNotRecreated(&transitGateway2, &transitGateway3),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckTransitGatewayExists(ctx context.Context, n string, v *awstypes.TransitGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Transit Gateway ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

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
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

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

func testAccCheckTransitGatewayNotRecreated(i, j *awstypes.TransitGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(i.TransitGatewayId) != aws.ToString(j.TransitGatewayId) {
			return errors.New("EC2 Transit Gateway was recreated")
		}

		return nil
	}
}

func testAccCheckTransitGatewayRecreated(i, j *awstypes.TransitGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToString(i.TransitGatewayId) == aws.ToString(j.TransitGatewayId) {
			return errors.New("EC2 Transit Gateway was not recreated")
		}

		return nil
	}
}

func testAccCheckTransitGatewayAssociationDefaultRouteTableAttachmentAssociated(ctx context.Context, transitGateway *awstypes.TransitGateway, transitGatewayAttachment interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		transitGatewayRouteTableID := aws.ToString(transitGateway.Options.AssociationDefaultRouteTableId)

		if transitGatewayRouteTableID == "" {
			return errors.New("EC2 Transit Gateway Association Default Route Table empty")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		var transitGatewayAttachmentID string
		switch transitGatewayAttachment := transitGatewayAttachment.(type) {
		case *awstypes.TransitGatewayVpcAttachment:
			transitGatewayAttachmentID = aws.ToString(transitGatewayAttachment.TransitGatewayAttachmentId)
		case *awstypes.TransitGatewayConnect:
			transitGatewayAttachmentID = aws.ToString(transitGatewayAttachment.TransitGatewayAttachmentId)
		}

		_, err := tfec2.FindTransitGatewayRouteTableAssociationByTwoPartKey(ctx, conn, transitGatewayRouteTableID, transitGatewayAttachmentID)

		if tfresource.NotFound(err) {
			return errors.New("EC2 Transit Gateway Route Table Association not found")
		}

		return err
	}
}

func testAccCheckTransitGatewayAssociationDefaultRouteTableAttachmentNotAssociated(ctx context.Context, transitGateway *awstypes.TransitGateway, transitGatewayAttachment interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		transitGatewayRouteTableID := aws.ToString(transitGateway.Options.AssociationDefaultRouteTableId)

		if transitGatewayRouteTableID == "" {
			return nil
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		var transitGatewayAttachmentID string
		switch transitGatewayAttachment := transitGatewayAttachment.(type) {
		case *awstypes.TransitGatewayVpcAttachment:
			transitGatewayAttachmentID = aws.ToString(transitGatewayAttachment.TransitGatewayAttachmentId)
		case *awstypes.TransitGatewayConnect:
			transitGatewayAttachmentID = aws.ToString(transitGatewayAttachment.TransitGatewayAttachmentId)
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

func testAccCheckTransitGatewayPropagationDefaultRouteTableAttachmentPropagated(ctx context.Context, transitGateway *awstypes.TransitGateway, transitGatewayAttachment interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		transitGatewayRouteTableID := aws.ToString(transitGateway.Options.PropagationDefaultRouteTableId)

		if transitGatewayRouteTableID == "" {
			return errors.New("EC2 Transit Gateway Propagation Default Route Table empty")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		var transitGatewayAttachmentID string
		switch transitGatewayAttachment := transitGatewayAttachment.(type) {
		case *awstypes.TransitGatewayVpcAttachment:
			transitGatewayAttachmentID = aws.ToString(transitGatewayAttachment.TransitGatewayAttachmentId)
		case *awstypes.TransitGatewayConnect:
			transitGatewayAttachmentID = aws.ToString(transitGatewayAttachment.TransitGatewayAttachmentId)
		}

		_, err := tfec2.FindTransitGatewayRouteTablePropagationByTwoPartKey(ctx, conn, transitGatewayRouteTableID, transitGatewayAttachmentID)

		if tfresource.NotFound(err) {
			return errors.New("EC2 Transit Gateway Route Table Propagation not enabled")
		}

		return err
	}
}

func testAccCheckTransitGatewayPropagationDefaultRouteTableAttachmentNotPropagated(ctx context.Context, transitGateway *awstypes.TransitGateway, transitGatewayAttachment interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		transitGatewayRouteTableID := aws.ToString(transitGateway.Options.PropagationDefaultRouteTableId)

		if transitGatewayRouteTableID == "" {
			return nil
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		var transitGatewayAttachmentID string
		switch transitGatewayAttachment := transitGatewayAttachment.(type) {
		case *awstypes.TransitGatewayVpcAttachment:
			transitGatewayAttachmentID = aws.ToString(transitGatewayAttachment.TransitGatewayAttachmentId)
		case *awstypes.TransitGatewayConnect:
			transitGatewayAttachmentID = aws.ToString(transitGatewayAttachment.TransitGatewayAttachmentId)
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
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeTransitGatewaysInput{
		MaxResults: aws.Int32(5),
	}

	_, err := conn.DescribeTransitGateways(ctx, input)

	if acctest.PreCheckSkipError(err) || tfawserr.ErrCodeEquals(err, "InvalidAction") {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccPreCheckTransitGatewayConnect(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeTransitGatewayConnectsInput{
		MaxResults: aws.Int32(5),
	}

	_, err := conn.DescribeTransitGatewayConnects(ctx, input)

	if acctest.PreCheckSkipError(err) || tfawserr.ErrCodeEquals(err, "InvalidAction") {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccPreCheckTransitGatewayVPCAttachment(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.DescribeTransitGatewayVpcAttachmentsInput{
		MaxResults: aws.Int32(5),
	}

	_, err := conn.DescribeTransitGatewayVpcAttachments(ctx, input)

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
