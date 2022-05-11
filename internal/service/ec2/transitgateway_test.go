package ec2_test

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

func TestAccTransitGateway_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Connect": {
			"basic":      testAccTransitGatewayConnect_basic,
			"disappears": testAccTransitGatewayConnect_disappears,
			"Tags":       testAccTransitGatewayConnect_tags,
			"TransitGatewayDefaultRouteTableAssociation":                       testAccTransitGatewayConnect_TransitGatewayDefaultRouteTableAssociation,
			"TransitGatewayDefaultRouteTableAssociationAndPropagationDisabled": testAccTransitGatewayConnect_TransitGatewayDefaultRouteTableAssociationAndPropagationDisabled,
			"TransitGatewayDefaultRouteTablePropagation":                       testAccTransitGatewayConnect_TransitGatewayDefaultRouteTablePropagation,
		},
		"ConnectPeer": {
			"basic":                 testAccTransitGatewayConnectPeer_basic,
			"disappears":            testAccTransitGatewayConnectPeer_disappears,
			"BgpAsn":                testAccTransitGatewayConnectPeer_BgpAsn,
			"InsideCidrBlocks":      testAccTransitGatewayConnectPeer_InsideCidrBlocks,
			"Tags":                  testAccTransitGatewayConnectPeer_tags,
			"TransitGatewayAddress": testAccTransitGatewayConnectPeer_TransitGatewayAddress,
		},
		"Gateway": {
			"basic":                       testAccTransitGateway_basic,
			"disappears":                  testAccTransitGateway_disappears,
			"AmazonSideASN":               testAccTransitGateway_AmazonSideASN,
			"AutoAcceptSharedAttachments": testAccTransitGateway_AutoAcceptSharedAttachments,
			"CidrBlocks":                  testAccTransitGateway_CidrBlocks,
			"DefaultRouteTableAssociationAndPropagationDisabled": testAccTransitGateway_DefaultRouteTableAssociationAndPropagationDisabled,
			"DefaultRouteTableAssociation":                       testAccTransitGateway_DefaultRouteTableAssociation,
			"DefaultRouteTablePropagation":                       testAccTransitGateway_DefaultRouteTablePropagation,
			"Description":                                        testAccTransitGateway_Description,
			"DnsSupport":                                         testAccTransitGateway_DNSSupport,
			"Tags":                                               testAccTransitGateway_Tags,
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
			"DifferentAccount": testAccTransitGatewayPeeringAttachment_differentAccount,
			"TagsSameAccount":  testAccTransitGatewayPeeringAttachment_Tags_sameAccount,
		},
		"PeeringAttachmentAccepter": {
			"basicSameAccount":      testAccTransitGatewayPeeringAttachmentAccepter_basic_sameAccount,
			"TagsSameAccount":       testAccTransitGatewayPeeringAttachmentAccepter_Tags_sameAccount,
			"basicDifferentAccount": testAccTransitGatewayPeeringAttachmentAccepter_basic_differentAccount,
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
			"Tags":                     testAccTransitGatewayRouteTable_Tags,
		},
		"RouteTableAssociation": {
			"basic": testAccTransitGatewayRouteTableAssociation_basic,
		},
		"RouteTablePropagation": {
			"basic": testAccTransitGatewayRouteTablePropagation_basic,
		},
		"VpcAttachment": {
			"basic":                testAccTransitGatewayVPCAttachment_basic,
			"disappears":           testAccTransitGatewayVPCAttachment_disappears,
			"ApplianceModeSupport": testAccTransitGatewayVPCAttachment_ApplianceModeSupport,
			"DnsSupport":           testAccTransitGatewayVPCAttachment_DNSSupport,
			"Ipv6Support":          testAccTransitGatewayVPCAttachment_IPv6Support,
			"SharedTransitGateway": testAccTransitGatewayVPCAttachment_SharedTransitGateway,
			"SubnetIds":            testAccTransitGatewayVPCAttachment_SubnetIDs,
			"Tags":                 testAccTransitGatewayVPCAttachment_Tags,
			"TransitGatewayDefaultRouteTableAssociation":                       testAccTransitGatewayVPCAttachment_TransitGatewayDefaultRouteTableAssociation,
			"TransitGatewayDefaultRouteTableAssociationAndPropagationDisabled": testAccTransitGatewayVPCAttachment_TransitGatewayDefaultRouteTableAssociationAndPropagationDisabled,
			"TransitGatewayDefaultRouteTablePropagation":                       testAccTransitGatewayVPCAttachment_TransitGatewayDefaultRouteTablePropagation,
		},
		"VpcAttachmentAccepter": {
			"basic": testAccTransitGatewayVPCAttachmentAccepter_basic,
			"Tags":  testAccTransitGatewayVPCAttachmentAccepter_Tags,
			"TransitGatewayDefaultRouteTableAssociationAndPropagation": testAccTransitGatewayVPCAttachmentAccepter_TransitGatewayDefaultRouteTableAssociationAndPropagation,
		},
	}

	for group, m := range testCases {
		m := m
		t.Run(group, func(t *testing.T) {
			for name, tc := range m {
				tc := tc
				t.Run(name, func(t *testing.T) {
					tc(t)
				})
			}
		})
	}
}

func testAccTransitGateway_basic(t *testing.T) {
	var transitGateway1 ec2.TransitGateway
	resourceName := "aws_ec2_transit_gateway.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(resourceName, &transitGateway1),
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
	var transitGateway1 ec2.TransitGateway
	resourceName := "aws_ec2_transit_gateway.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(resourceName, &transitGateway1),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceTransitGateway(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccTransitGateway_AmazonSideASN(t *testing.T) {
	var transitGateway1, transitGateway2 ec2.TransitGateway
	resourceName := "aws_ec2_transit_gateway.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayAmazonSideASNConfig(64513),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(resourceName, &transitGateway1),
					resource.TestCheckResourceAttr(resourceName, "amazon_side_asn", "64513"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayAmazonSideASNConfig(64514),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(resourceName, &transitGateway2),
					testAccCheckTransitGatewayRecreated(&transitGateway1, &transitGateway2),
					resource.TestCheckResourceAttr(resourceName, "amazon_side_asn", "64514"),
				),
			},
		},
	})
}

func testAccTransitGateway_AutoAcceptSharedAttachments(t *testing.T) {
	var transitGateway1, transitGateway2 ec2.TransitGateway
	resourceName := "aws_ec2_transit_gateway.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayAutoAcceptSharedAttachmentsConfig(ec2.AutoAcceptSharedAttachmentsValueEnable),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(resourceName, &transitGateway1),
					resource.TestCheckResourceAttr(resourceName, "auto_accept_shared_attachments", ec2.AutoAcceptSharedAttachmentsValueEnable),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayAutoAcceptSharedAttachmentsConfig(ec2.AutoAcceptSharedAttachmentsValueDisable),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(resourceName, &transitGateway2),
					testAccCheckTransitGatewayNotRecreated(&transitGateway1, &transitGateway2),
					resource.TestCheckResourceAttr(resourceName, "auto_accept_shared_attachments", ec2.AutoAcceptSharedAttachmentsValueDisable),
				),
			},
		},
	})
}

func testAccTransitGateway_CidrBlocks(t *testing.T) {
	var v1, v2, v3 ec2.TransitGateway
	resourceName := "aws_ec2_transit_gateway.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayCidrBlocks2Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(resourceName, &v1),
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
				Config: testAccTransitGatewayCidrBlocks1Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(resourceName, &v2),
					testAccCheckTransitGatewayNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_cidr_blocks.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "transit_gateway_cidr_blocks.*", "10.120.0.0/24"),
				),
			},
			{
				Config: testAccTransitGatewayCidrBlocks2Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(resourceName, &v3),
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
	var transitGateway1 ec2.TransitGateway
	resourceName := "aws_ec2_transit_gateway.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayDefaultRouteTableAssociationAndPropagationDisabledConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(resourceName, &transitGateway1),
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
	var transitGateway1, transitGateway2, transitGateway3 ec2.TransitGateway
	resourceName := "aws_ec2_transit_gateway.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayDefaultRouteTableAssociationConfig(ec2.DefaultRouteTableAssociationValueDisable),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(resourceName, &transitGateway1),
					resource.TestCheckResourceAttr(resourceName, "default_route_table_association", ec2.DefaultRouteTableAssociationValueDisable),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayDefaultRouteTableAssociationConfig(ec2.DefaultRouteTableAssociationValueEnable),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(resourceName, &transitGateway2),
					testAccCheckTransitGatewayRecreated(&transitGateway1, &transitGateway2),
					resource.TestCheckResourceAttr(resourceName, "default_route_table_association", ec2.DefaultRouteTableAssociationValueEnable),
				),
			},
			{
				Config: testAccTransitGatewayDefaultRouteTableAssociationConfig(ec2.DefaultRouteTableAssociationValueDisable),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(resourceName, &transitGateway3),
					testAccCheckTransitGatewayNotRecreated(&transitGateway2, &transitGateway3),
					resource.TestCheckResourceAttr(resourceName, "default_route_table_association", ec2.DefaultRouteTableAssociationValueDisable),
				),
			},
		},
	})
}

func testAccTransitGateway_DefaultRouteTablePropagation(t *testing.T) {
	var transitGateway1, transitGateway2, transitGateway3 ec2.TransitGateway
	resourceName := "aws_ec2_transit_gateway.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayDefaultRouteTablePropagationConfig(ec2.DefaultRouteTablePropagationValueDisable),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(resourceName, &transitGateway1),
					resource.TestCheckResourceAttr(resourceName, "default_route_table_propagation", ec2.DefaultRouteTablePropagationValueDisable),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayDefaultRouteTablePropagationConfig(ec2.DefaultRouteTablePropagationValueEnable),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(resourceName, &transitGateway2),
					testAccCheckTransitGatewayRecreated(&transitGateway1, &transitGateway2),
					resource.TestCheckResourceAttr(resourceName, "default_route_table_propagation", ec2.DefaultRouteTablePropagationValueEnable),
				),
			},
			{
				Config: testAccTransitGatewayDefaultRouteTablePropagationConfig(ec2.DefaultRouteTablePropagationValueDisable),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(resourceName, &transitGateway3),
					testAccCheckTransitGatewayNotRecreated(&transitGateway2, &transitGateway3),
					resource.TestCheckResourceAttr(resourceName, "default_route_table_propagation", ec2.DefaultRouteTablePropagationValueDisable),
				),
			},
		},
	})
}

func testAccTransitGateway_DNSSupport(t *testing.T) {
	var transitGateway1, transitGateway2 ec2.TransitGateway
	resourceName := "aws_ec2_transit_gateway.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayDNSSupportConfig(ec2.DnsSupportValueDisable),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(resourceName, &transitGateway1),
					resource.TestCheckResourceAttr(resourceName, "dns_support", ec2.DnsSupportValueDisable),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayDNSSupportConfig(ec2.DnsSupportValueEnable),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(resourceName, &transitGateway2),
					testAccCheckTransitGatewayNotRecreated(&transitGateway1, &transitGateway2),
					resource.TestCheckResourceAttr(resourceName, "dns_support", ec2.DnsSupportValueEnable),
				),
			},
		},
	})
}

func testAccTransitGateway_VPNECMPSupport(t *testing.T) {
	var transitGateway1, transitGateway2 ec2.TransitGateway
	resourceName := "aws_ec2_transit_gateway.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayVPNECMPSupportConfig(ec2.VpnEcmpSupportValueDisable),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(resourceName, &transitGateway1),
					resource.TestCheckResourceAttr(resourceName, "vpn_ecmp_support", ec2.VpnEcmpSupportValueDisable),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayVPNECMPSupportConfig(ec2.VpnEcmpSupportValueEnable),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(resourceName, &transitGateway2),
					testAccCheckTransitGatewayNotRecreated(&transitGateway1, &transitGateway2),
					resource.TestCheckResourceAttr(resourceName, "vpn_ecmp_support", ec2.VpnEcmpSupportValueEnable),
				),
			},
		},
	})
}

func testAccTransitGateway_Description(t *testing.T) {
	var transitGateway1, transitGateway2 ec2.TransitGateway
	resourceName := "aws_ec2_transit_gateway.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayDescriptionConfig("description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(resourceName, &transitGateway1),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayDescriptionConfig("description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(resourceName, &transitGateway2),
					testAccCheckTransitGatewayNotRecreated(&transitGateway1, &transitGateway2),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
		},
	})
}

func testAccTransitGateway_Tags(t *testing.T) {
	var transitGateway1, transitGateway2, transitGateway3 ec2.TransitGateway
	resourceName := "aws_ec2_transit_gateway.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayTags1Config("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(resourceName, &transitGateway1),
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
				Config: testAccTransitGatewayTags2Config("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(resourceName, &transitGateway2),
					testAccCheckTransitGatewayNotRecreated(&transitGateway1, &transitGateway2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccTransitGatewayTags1Config("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(resourceName, &transitGateway3),
					testAccCheckTransitGatewayNotRecreated(&transitGateway2, &transitGateway3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckTransitGatewayExists(resourceName string, transitGateway *ec2.TransitGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Transit Gateway ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		gateway, err := tfec2.DescribeTransitGateway(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if gateway == nil {
			return fmt.Errorf("EC2 Transit Gateway not found")
		}

		if aws.StringValue(gateway.State) != ec2.TransitGatewayStateAvailable {
			return fmt.Errorf("EC2 Transit Gateway (%s) exists in non-available (%s) state", rs.Primary.ID, aws.StringValue(gateway.State))
		}

		*transitGateway = *gateway

		return nil
	}
}

func testAccCheckTransitGatewayDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_transit_gateway" {
			continue
		}

		transitGateway, err := tfec2.DescribeTransitGateway(conn, rs.Primary.ID)

		if tfawserr.ErrCodeEquals(err, "InvalidTransitGatewayID.NotFound") {
			continue
		}

		if err != nil {
			return err
		}

		if transitGateway == nil {
			continue
		}

		if aws.StringValue(transitGateway.State) != ec2.TransitGatewayStateDeleted {
			return fmt.Errorf("EC2 Transit Gateway (%s) still exists in non-deleted (%s) state", rs.Primary.ID, aws.StringValue(transitGateway.State))
		}
	}

	return nil
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

func testAccCheckTransitGatewayAssociationDefaultRouteTableAttachmentAssociated(transitGateway *ec2.TransitGateway, transitGatewayAttachment interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn
		var attachmentID string
		switch transitGatewayAttachment := transitGatewayAttachment.(type) {
		case *ec2.TransitGatewayVpcAttachment:
			attachmentID = aws.StringValue(transitGatewayAttachment.TransitGatewayAttachmentId)
		case *ec2.TransitGatewayConnect:
			attachmentID = aws.StringValue(transitGatewayAttachment.TransitGatewayAttachmentId)
		}
		routeTableID := aws.StringValue(transitGateway.Options.AssociationDefaultRouteTableId)
		association, err := tfec2.DescribeTransitGatewayRouteTableAssociation(conn, routeTableID, attachmentID)

		if err != nil {
			return err
		}

		if association == nil {
			return errors.New("EC2 Transit Gateway Route Table Association not found")
		}

		return nil
	}
}

func testAccCheckTransitGatewayAssociationDefaultRouteTableAttachmentNotAssociated(transitGateway *ec2.TransitGateway, transitGatewayAttachment interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn
		var attachmentID string
		switch transitGatewayAttachment := transitGatewayAttachment.(type) {
		case *ec2.TransitGatewayVpcAttachment:
			attachmentID = aws.StringValue(transitGatewayAttachment.TransitGatewayAttachmentId)
		case *ec2.TransitGatewayConnect:
			attachmentID = aws.StringValue(transitGatewayAttachment.TransitGatewayAttachmentId)
		}
		routeTableID := aws.StringValue(transitGateway.Options.AssociationDefaultRouteTableId)
		association, err := tfec2.DescribeTransitGatewayRouteTableAssociation(conn, routeTableID, attachmentID)

		if err != nil {
			return err
		}

		if association != nil {
			return errors.New("EC2 Transit Gateway Route Table Association found")
		}

		return nil
	}
}

func testAccCheckTransitGatewayPropagationDefaultRouteTableAttachmentNotPropagated(transitGateway *ec2.TransitGateway, transitGatewayAttachment interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn
		var attachmentID string
		switch transitGatewayAttachment := transitGatewayAttachment.(type) {
		case *ec2.TransitGatewayVpcAttachment:
			attachmentID = aws.StringValue(transitGatewayAttachment.TransitGatewayAttachmentId)
		case *ec2.TransitGatewayConnect:
			attachmentID = aws.StringValue(transitGatewayAttachment.TransitGatewayAttachmentId)
		}
		routeTableID := aws.StringValue(transitGateway.Options.AssociationDefaultRouteTableId)
		propagation, err := tfec2.FindTransitGatewayRouteTablePropagation(conn, routeTableID, attachmentID)

		if err != nil {
			return err
		}

		if propagation != nil {
			return errors.New("EC2 Transit Gateway Route Table Propagation enabled")
		}

		return nil
	}
}

func testAccCheckTransitGatewayPropagationDefaultRouteTableAttachmentPropagated(transitGateway *ec2.TransitGateway, transitGatewayAttachment interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn
		var attachmentID string
		switch transitGatewayAttachment := transitGatewayAttachment.(type) {
		case *ec2.TransitGatewayVpcAttachment:
			attachmentID = aws.StringValue(transitGatewayAttachment.TransitGatewayAttachmentId)
		case *ec2.TransitGatewayConnect:
			attachmentID = aws.StringValue(transitGatewayAttachment.TransitGatewayAttachmentId)
		}
		routeTableID := aws.StringValue(transitGateway.Options.AssociationDefaultRouteTableId)
		propagation, err := tfec2.FindTransitGatewayRouteTablePropagation(conn, routeTableID, attachmentID)

		if err != nil {
			return err
		}

		if propagation == nil {
			return errors.New("EC2 Transit Gateway Route Table Propagation not enabled")
		}

		return nil
	}
}

func testAccPreCheckTransitGateway(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	input := &ec2.DescribeTransitGatewaysInput{
		MaxResults: aws.Int64(5),
	}

	_, err := conn.DescribeTransitGateways(input)

	if acctest.PreCheckSkipError(err) || tfawserr.ErrCodeEquals(err, "InvalidAction") {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccTransitGatewayConfig() string {
	return `
resource "aws_ec2_transit_gateway" "test" {}
`
}

func testAccTransitGatewayAmazonSideASNConfig(amazonSideASN int) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  amazon_side_asn = %d
}
`, amazonSideASN)
}

func testAccTransitGatewayAutoAcceptSharedAttachmentsConfig(autoAcceptSharedAttachments string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  auto_accept_shared_attachments = %q
}
`, autoAcceptSharedAttachments)
}

func testAccTransitGatewayDefaultRouteTableAssociationAndPropagationDisabledConfig() string {
	return `
resource "aws_ec2_transit_gateway" "test" {
  default_route_table_association = "disable"
  default_route_table_propagation = "disable"
}
`
}

func testAccTransitGatewayDefaultRouteTableAssociationConfig(defaultRouteTableAssociation string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  default_route_table_association = %q
}
`, defaultRouteTableAssociation)
}

func testAccTransitGatewayDefaultRouteTablePropagationConfig(defaultRouteTablePropagation string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  default_route_table_propagation = %q
}
`, defaultRouteTablePropagation)
}

func testAccTransitGatewayDNSSupportConfig(dnsSupport string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  dns_support = %q
}
`, dnsSupport)
}

func testAccTransitGatewayVPNECMPSupportConfig(vpnEcmpSupport string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  vpn_ecmp_support = %q
}
`, vpnEcmpSupport)
}

func testAccTransitGatewayDescriptionConfig(description string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  description = %q
}
`, description)
}

func testAccTransitGatewayTags1Config(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  tags = {
    %q = %q
  }
}
`, tagKey1, tagValue1)
}

func testAccTransitGatewayTags2Config(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  tags = {
    %q = %q
    %q = %q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccTransitGatewayCidrBlocks1Config(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  transit_gateway_cidr_blocks = ["10.120.0.0/24"]

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccTransitGatewayCidrBlocks2Config(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  transit_gateway_cidr_blocks = ["10.120.0.0/24", "2001:1234:1234::/64"]

  tags = {
    Name = %[1]q
  }
}
`, rName)
}
