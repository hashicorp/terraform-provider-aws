package aws

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/ec2/finder"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func init() {
	resource.AddTestSweepers("aws_ec2_transit_gateway", &resource.Sweeper{
		Name: "aws_ec2_transit_gateway",
		F:    testSweepEc2TransitGateways,
		Dependencies: []string{
			"aws_dx_gateway_association",
			"aws_ec2_transit_gateway_vpc_attachment",
			"aws_ec2_transit_gateway_peering_attachment",
			"aws_vpn_connection",
		},
	})
}

func testSweepEc2TransitGateways(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).ec2conn
	input := &ec2.DescribeTransitGatewaysInput{}

	for {
		output, err := conn.DescribeTransitGateways(input)

		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 Transit Gateway sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error retrieving EC2 Transit Gateways: %s", err)
		}

		for _, transitGateway := range output.TransitGateways {
			if aws.StringValue(transitGateway.State) == ec2.TransitGatewayStateDeleted {
				continue
			}

			id := aws.StringValue(transitGateway.TransitGatewayId)

			input := &ec2.DeleteTransitGatewayInput{
				TransitGatewayId: aws.String(id),
			}

			log.Printf("[INFO] Deleting EC2 Transit Gateway: %s", id)
			err := resource.Retry(2*time.Minute, func() *resource.RetryError {
				_, err := conn.DeleteTransitGateway(input)

				if tfawserr.ErrMessageContains(err, "IncorrectState", "has non-deleted Transit Gateway Attachments") {
					return resource.RetryableError(err)
				}

				if tfawserr.ErrMessageContains(err, "IncorrectState", "has non-deleted DirectConnect Gateway Attachments") {
					return resource.RetryableError(err)
				}

				if tfawserr.ErrMessageContains(err, "IncorrectState", "has non-deleted VPN Attachments") {
					return resource.RetryableError(err)
				}

				if tfawserr.ErrMessageContains(err, "InvalidTransitGatewayID.NotFound", "") {
					return nil
				}

				if err != nil {
					return resource.NonRetryableError(err)
				}

				return nil
			})

			if tfresource.TimedOut(err) {
				_, err = conn.DeleteTransitGateway(input)
			}

			if err != nil {
				return fmt.Errorf("error deleting EC2 Transit Gateway (%s): %s", id, err)
			}

			if err := waitForEc2TransitGatewayDeletion(conn, id); err != nil {
				return fmt.Errorf("error waiting for EC2 Transit Gateway (%s) deletion: %s", id, err)
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil
}

func TestAccAWSEc2TransitGateway_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Gateway": {
			"basic":                       testAccAWSEc2TransitGateway_basic,
			"disappears":                  testAccAWSEc2TransitGateway_disappears,
			"AmazonSideASN":               testAccAWSEc2TransitGateway_AmazonSideASN,
			"AutoAcceptSharedAttachments": testAccAWSEc2TransitGateway_AutoAcceptSharedAttachments,
			"DefaultRouteTableAssociationAndPropagationDisabled": testAccAWSEc2TransitGateway_DefaultRouteTableAssociationAndPropagationDisabled,
			"DefaultRouteTableAssociation":                       testAccAWSEc2TransitGateway_DefaultRouteTableAssociation,
			"DefaultRouteTablePropagation":                       testAccAWSEc2TransitGateway_DefaultRouteTablePropagation,
			"Description":                                        testAccAWSEc2TransitGateway_Description,
			"DnsSupport":                                         testAccAWSEc2TransitGateway_DnsSupport,
			"Tags":                                               testAccAWSEc2TransitGateway_Tags,
			"VpnEcmpSupport":                                     testAccAWSEc2TransitGateway_VpnEcmpSupport,
		},
		"PeeringAttachment": {
			"basic":            testAccAWSEc2TransitGatewayPeeringAttachment_basic,
			"disappears":       testAccAWSEc2TransitGatewayPeeringAttachment_disappears,
			"DifferentAccount": testAccAWSEc2TransitGatewayPeeringAttachment_differentAccount,
			"TagsSameAccount":  testAccAWSEc2TransitGatewayPeeringAttachment_Tags_sameAccount,
		},
		"PeeringAttachmentAccepter": {
			"basicSameAccount":      testAccAWSEc2TransitGatewayPeeringAttachmentAccepter_basic_sameAccount,
			"TagsSameAccount":       testAccAWSEc2TransitGatewayPeeringAttachmentAccepter_Tags_sameAccount,
			"basicDifferentAccount": testAccAWSEc2TransitGatewayPeeringAttachmentAccepter_basic_differentAccount,
		},
		"PrefixListReference": {
			"basic":                      testAccAwsEc2TransitGatewayPrefixListReference_basic,
			"disappears":                 testAccAwsEc2TransitGatewayPrefixListReference_disappears,
			"disappearsTransitGateway":   testAccAwsEc2TransitGatewayPrefixListReference_disappears_TransitGateway,
			"TransitGatewayAttachmentId": testAccAwsEc2TransitGatewayPrefixListReference_TransitGatewayAttachmentId,
		},
		"Route": {
			"basic":                              testAccAWSEc2TransitGatewayRoute_basic,
			"basicIpv6":                          testAccAWSEc2TransitGatewayRoute_basic_ipv6,
			"blackhole":                          testAccAWSEc2TransitGatewayRoute_blackhole,
			"disappears":                         testAccAWSEc2TransitGatewayRoute_disappears,
			"disappearsTransitGatewayAttachment": testAccAWSEc2TransitGatewayRoute_disappears_TransitGatewayAttachment,
		},
		"RouteTable": {
			"basic":                    testAccAWSEc2TransitGatewayRouteTable_basic,
			"disappears":               testAccAWSEc2TransitGatewayRouteTable_disappears,
			"disappearsTransitGateway": testAccAWSEc2TransitGatewayRouteTable_disappears_TransitGateway,
			"Tags":                     testAccAWSEc2TransitGatewayRouteTable_Tags,
		},
		"RouteTableAssociation": {
			"basic": testAccAWSEc2TransitGatewayRouteTableAssociation_basic,
		},
		"RouteTablePropagation": {
			"basic": testAccAWSEc2TransitGatewayRouteTablePropagation_basic,
		},
		"VpcAttachment": {
			"basic":                testAccAWSEc2TransitGatewayVpcAttachment_basic,
			"disappears":           testAccAWSEc2TransitGatewayVpcAttachment_disappears,
			"ApplianceModeSupport": testAccAWSEc2TransitGatewayVpcAttachment_ApplianceModeSupport,
			"DnsSupport":           testAccAWSEc2TransitGatewayVpcAttachment_DnsSupport,
			"Ipv6Support":          testAccAWSEc2TransitGatewayVpcAttachment_Ipv6Support,
			"SharedTransitGateway": testAccAWSEc2TransitGatewayVpcAttachment_SharedTransitGateway,
			"SubnetIds":            testAccAWSEc2TransitGatewayVpcAttachment_SubnetIds,
			"Tags":                 testAccAWSEc2TransitGatewayVpcAttachment_Tags,
			"TransitGatewayDefaultRouteTableAssociation":                       testAccAWSEc2TransitGatewayVpcAttachment_TransitGatewayDefaultRouteTableAssociation,
			"TransitGatewayDefaultRouteTableAssociationAndPropagationDisabled": testAccAWSEc2TransitGatewayVpcAttachment_TransitGatewayDefaultRouteTableAssociationAndPropagationDisabled,
			"TransitGatewayDefaultRouteTablePropagation":                       testAccAWSEc2TransitGatewayVpcAttachment_TransitGatewayDefaultRouteTablePropagation,
		},
		"VpcAttachmentAccepter": {
			"basic": testAccAWSEc2TransitGatewayVpcAttachmentAccepter_basic,
			"Tags":  testAccAWSEc2TransitGatewayVpcAttachmentAccepter_Tags,
			"TransitGatewayDefaultRouteTableAssociationAndPropagation": testAccAWSEc2TransitGatewayVpcAttachmentAccepter_TransitGatewayDefaultRouteTableAssociationAndPropagation,
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

func testAccAWSEc2TransitGateway_basic(t *testing.T) {
	var transitGateway1 ec2.TransitGateway
	resourceName := "aws_ec2_transit_gateway.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayExists(resourceName, &transitGateway1),
					resource.TestCheckResourceAttr(resourceName, "amazon_side_asn", "64512"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`transit-gateway/tgw-.+`)),
					resource.TestCheckResourceAttrSet(resourceName, "association_default_route_table_id"),
					resource.TestCheckResourceAttr(resourceName, "auto_accept_shared_attachments", ec2.AutoAcceptSharedAttachmentsValueDisable),
					resource.TestCheckResourceAttr(resourceName, "default_route_table_association", ec2.DefaultRouteTableAssociationValueEnable),
					resource.TestCheckResourceAttr(resourceName, "default_route_table_propagation", ec2.DefaultRouteTablePropagationValueEnable),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "dns_support", ec2.DnsSupportValueEnable),
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

func testAccAWSEc2TransitGateway_disappears(t *testing.T) {
	var transitGateway1 ec2.TransitGateway
	resourceName := "aws_ec2_transit_gateway.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayExists(resourceName, &transitGateway1),
					acctest.CheckResourceDisappears(acctest.Provider, resourceAwsEc2TransitGateway(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccAWSEc2TransitGateway_AmazonSideASN(t *testing.T) {
	var transitGateway1, transitGateway2 ec2.TransitGateway
	resourceName := "aws_ec2_transit_gateway.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayConfigAmazonSideASN(64513),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayExists(resourceName, &transitGateway1),
					resource.TestCheckResourceAttr(resourceName, "amazon_side_asn", "64513"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEc2TransitGatewayConfigAmazonSideASN(64514),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayExists(resourceName, &transitGateway2),
					testAccCheckAWSEc2TransitGatewayRecreated(&transitGateway1, &transitGateway2),
					resource.TestCheckResourceAttr(resourceName, "amazon_side_asn", "64514"),
				),
			},
		},
	})
}

func testAccAWSEc2TransitGateway_AutoAcceptSharedAttachments(t *testing.T) {
	var transitGateway1, transitGateway2 ec2.TransitGateway
	resourceName := "aws_ec2_transit_gateway.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayConfigAutoAcceptSharedAttachments(ec2.AutoAcceptSharedAttachmentsValueEnable),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayExists(resourceName, &transitGateway1),
					resource.TestCheckResourceAttr(resourceName, "auto_accept_shared_attachments", ec2.AutoAcceptSharedAttachmentsValueEnable),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEc2TransitGatewayConfigAutoAcceptSharedAttachments(ec2.AutoAcceptSharedAttachmentsValueDisable),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayExists(resourceName, &transitGateway2),
					testAccCheckAWSEc2TransitGatewayNotRecreated(&transitGateway1, &transitGateway2),
					resource.TestCheckResourceAttr(resourceName, "auto_accept_shared_attachments", ec2.AutoAcceptSharedAttachmentsValueDisable),
				),
			},
		},
	})
}

func testAccAWSEc2TransitGateway_DefaultRouteTableAssociationAndPropagationDisabled(t *testing.T) {
	var transitGateway1 ec2.TransitGateway
	resourceName := "aws_ec2_transit_gateway.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayConfigDefaultRouteTableAssociationAndPropagationDisabled(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayExists(resourceName, &transitGateway1),
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

func testAccAWSEc2TransitGateway_DefaultRouteTableAssociation(t *testing.T) {
	var transitGateway1, transitGateway2, transitGateway3 ec2.TransitGateway
	resourceName := "aws_ec2_transit_gateway.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayConfigDefaultRouteTableAssociation(ec2.DefaultRouteTableAssociationValueDisable),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayExists(resourceName, &transitGateway1),
					resource.TestCheckResourceAttr(resourceName, "default_route_table_association", ec2.DefaultRouteTableAssociationValueDisable),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEc2TransitGatewayConfigDefaultRouteTableAssociation(ec2.DefaultRouteTableAssociationValueEnable),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayExists(resourceName, &transitGateway2),
					testAccCheckAWSEc2TransitGatewayRecreated(&transitGateway1, &transitGateway2),
					resource.TestCheckResourceAttr(resourceName, "default_route_table_association", ec2.DefaultRouteTableAssociationValueEnable),
				),
			},
			{
				Config: testAccAWSEc2TransitGatewayConfigDefaultRouteTableAssociation(ec2.DefaultRouteTableAssociationValueDisable),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayExists(resourceName, &transitGateway3),
					testAccCheckAWSEc2TransitGatewayNotRecreated(&transitGateway2, &transitGateway3),
					resource.TestCheckResourceAttr(resourceName, "default_route_table_association", ec2.DefaultRouteTableAssociationValueDisable),
				),
			},
		},
	})
}

func testAccAWSEc2TransitGateway_DefaultRouteTablePropagation(t *testing.T) {
	var transitGateway1, transitGateway2, transitGateway3 ec2.TransitGateway
	resourceName := "aws_ec2_transit_gateway.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayConfigDefaultRouteTablePropagation(ec2.DefaultRouteTablePropagationValueDisable),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayExists(resourceName, &transitGateway1),
					resource.TestCheckResourceAttr(resourceName, "default_route_table_propagation", ec2.DefaultRouteTablePropagationValueDisable),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEc2TransitGatewayConfigDefaultRouteTablePropagation(ec2.DefaultRouteTablePropagationValueEnable),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayExists(resourceName, &transitGateway2),
					testAccCheckAWSEc2TransitGatewayRecreated(&transitGateway1, &transitGateway2),
					resource.TestCheckResourceAttr(resourceName, "default_route_table_propagation", ec2.DefaultRouteTablePropagationValueEnable),
				),
			},
			{
				Config: testAccAWSEc2TransitGatewayConfigDefaultRouteTablePropagation(ec2.DefaultRouteTablePropagationValueDisable),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayExists(resourceName, &transitGateway3),
					testAccCheckAWSEc2TransitGatewayNotRecreated(&transitGateway2, &transitGateway3),
					resource.TestCheckResourceAttr(resourceName, "default_route_table_propagation", ec2.DefaultRouteTablePropagationValueDisable),
				),
			},
		},
	})
}

func testAccAWSEc2TransitGateway_DnsSupport(t *testing.T) {
	var transitGateway1, transitGateway2 ec2.TransitGateway
	resourceName := "aws_ec2_transit_gateway.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayConfigDnsSupport(ec2.DnsSupportValueDisable),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayExists(resourceName, &transitGateway1),
					resource.TestCheckResourceAttr(resourceName, "dns_support", ec2.DnsSupportValueDisable),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEc2TransitGatewayConfigDnsSupport(ec2.DnsSupportValueEnable),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayExists(resourceName, &transitGateway2),
					testAccCheckAWSEc2TransitGatewayNotRecreated(&transitGateway1, &transitGateway2),
					resource.TestCheckResourceAttr(resourceName, "dns_support", ec2.DnsSupportValueEnable),
				),
			},
		},
	})
}

func testAccAWSEc2TransitGateway_VpnEcmpSupport(t *testing.T) {
	var transitGateway1, transitGateway2 ec2.TransitGateway
	resourceName := "aws_ec2_transit_gateway.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayConfigVpnEcmpSupport(ec2.VpnEcmpSupportValueDisable),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayExists(resourceName, &transitGateway1),
					resource.TestCheckResourceAttr(resourceName, "vpn_ecmp_support", ec2.VpnEcmpSupportValueDisable),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEc2TransitGatewayConfigVpnEcmpSupport(ec2.VpnEcmpSupportValueEnable),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayExists(resourceName, &transitGateway2),
					testAccCheckAWSEc2TransitGatewayNotRecreated(&transitGateway1, &transitGateway2),
					resource.TestCheckResourceAttr(resourceName, "vpn_ecmp_support", ec2.VpnEcmpSupportValueEnable),
				),
			},
		},
	})
}

func testAccAWSEc2TransitGateway_Description(t *testing.T) {
	var transitGateway1, transitGateway2 ec2.TransitGateway
	resourceName := "aws_ec2_transit_gateway.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayConfigDescription("description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayExists(resourceName, &transitGateway1),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEc2TransitGatewayConfigDescription("description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayExists(resourceName, &transitGateway2),
					testAccCheckAWSEc2TransitGatewayNotRecreated(&transitGateway1, &transitGateway2),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
		},
	})
}

func testAccAWSEc2TransitGateway_Tags(t *testing.T) {
	var transitGateway1, transitGateway2, transitGateway3 ec2.TransitGateway
	resourceName := "aws_ec2_transit_gateway.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayConfigTags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayExists(resourceName, &transitGateway1),
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
				Config: testAccAWSEc2TransitGatewayConfigTags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayExists(resourceName, &transitGateway2),
					testAccCheckAWSEc2TransitGatewayNotRecreated(&transitGateway1, &transitGateway2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSEc2TransitGatewayConfigTags1("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayExists(resourceName, &transitGateway3),
					testAccCheckAWSEc2TransitGatewayNotRecreated(&transitGateway2, &transitGateway3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAWSEc2TransitGatewayExists(resourceName string, transitGateway *ec2.TransitGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Transit Gateway ID is set")
		}

		conn := acctest.Provider.Meta().(*AWSClient).ec2conn

		gateway, err := ec2DescribeTransitGateway(conn, rs.Primary.ID)

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

func testAccCheckAWSEc2TransitGatewayDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_transit_gateway" {
			continue
		}

		transitGateway, err := ec2DescribeTransitGateway(conn, rs.Primary.ID)

		if tfawserr.ErrMessageContains(err, "InvalidTransitGatewayID.NotFound", "") {
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

func testAccCheckAWSEc2TransitGatewayNotRecreated(i, j *ec2.TransitGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.TransitGatewayId) != aws.StringValue(j.TransitGatewayId) {
			return errors.New("EC2 Transit Gateway was recreated")
		}

		return nil
	}
}

func testAccCheckAWSEc2TransitGatewayRecreated(i, j *ec2.TransitGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.TransitGatewayId) == aws.StringValue(j.TransitGatewayId) {
			return errors.New("EC2 Transit Gateway was not recreated")
		}

		return nil
	}
}

func testAccCheckAWSEc2TransitGatewayAssociationDefaultRouteTableVpcAttachmentAssociated(transitGateway *ec2.TransitGateway, transitGatewayVpcAttachment *ec2.TransitGatewayVpcAttachment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*AWSClient).ec2conn

		attachmentID := aws.StringValue(transitGatewayVpcAttachment.TransitGatewayAttachmentId)
		routeTableID := aws.StringValue(transitGateway.Options.AssociationDefaultRouteTableId)
		association, err := ec2DescribeTransitGatewayRouteTableAssociation(conn, routeTableID, attachmentID)

		if err != nil {
			return err
		}

		if association == nil {
			return errors.New("EC2 Transit Gateway Route Table Association not found")
		}

		return nil
	}
}

func testAccCheckAWSEc2TransitGatewayAssociationDefaultRouteTableVpcAttachmentNotAssociated(transitGateway *ec2.TransitGateway, transitGatewayVpcAttachment *ec2.TransitGatewayVpcAttachment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*AWSClient).ec2conn

		attachmentID := aws.StringValue(transitGatewayVpcAttachment.TransitGatewayAttachmentId)
		routeTableID := aws.StringValue(transitGateway.Options.AssociationDefaultRouteTableId)
		association, err := ec2DescribeTransitGatewayRouteTableAssociation(conn, routeTableID, attachmentID)

		if err != nil {
			return err
		}

		if association != nil {
			return errors.New("EC2 Transit Gateway Route Table Association found")
		}

		return nil
	}
}

func testAccCheckAWSEc2TransitGatewayPropagationDefaultRouteTableVpcAttachmentNotPropagated(transitGateway *ec2.TransitGateway, transitGatewayVpcAttachment *ec2.TransitGatewayVpcAttachment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*AWSClient).ec2conn

		attachmentID := aws.StringValue(transitGatewayVpcAttachment.TransitGatewayAttachmentId)
		routeTableID := aws.StringValue(transitGateway.Options.AssociationDefaultRouteTableId)
		propagation, err := finder.TransitGatewayRouteTablePropagation(conn, routeTableID, attachmentID)

		if err != nil {
			return err
		}

		if propagation != nil {
			return errors.New("EC2 Transit Gateway Route Table Propagation enabled")
		}

		return nil
	}
}

func testAccCheckAWSEc2TransitGatewayPropagationDefaultRouteTableVpcAttachmentPropagated(transitGateway *ec2.TransitGateway, transitGatewayVpcAttachment *ec2.TransitGatewayVpcAttachment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*AWSClient).ec2conn

		attachmentID := aws.StringValue(transitGatewayVpcAttachment.TransitGatewayAttachmentId)
		routeTableID := aws.StringValue(transitGateway.Options.AssociationDefaultRouteTableId)
		propagation, err := finder.TransitGatewayRouteTablePropagation(conn, routeTableID, attachmentID)

		if err != nil {
			return err
		}

		if propagation == nil {
			return errors.New("EC2 Transit Gateway Route Table Propagation not enabled")
		}

		return nil
	}
}

func testAccPreCheckAWSEc2TransitGateway(t *testing.T) {
	conn := acctest.Provider.Meta().(*AWSClient).ec2conn

	input := &ec2.DescribeTransitGatewaysInput{
		MaxResults: aws.Int64(5),
	}

	_, err := conn.DescribeTransitGateways(input)

	if acctest.PreCheckSkipError(err) || tfawserr.ErrMessageContains(err, "InvalidAction", "") {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAWSEc2TransitGatewayConfig() string {
	return `
resource "aws_ec2_transit_gateway" "test" {}
`
}

func testAccAWSEc2TransitGatewayConfigAmazonSideASN(amazonSideASN int) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  amazon_side_asn = %d
}
`, amazonSideASN)
}

func testAccAWSEc2TransitGatewayConfigAutoAcceptSharedAttachments(autoAcceptSharedAttachments string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  auto_accept_shared_attachments = %q
}
`, autoAcceptSharedAttachments)
}

func testAccAWSEc2TransitGatewayConfigDefaultRouteTableAssociationAndPropagationDisabled() string {
	return `
resource "aws_ec2_transit_gateway" "test" {
  default_route_table_association = "disable"
  default_route_table_propagation = "disable"
}
`
}

func testAccAWSEc2TransitGatewayConfigDefaultRouteTableAssociation(defaultRouteTableAssociation string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  default_route_table_association = %q
}
`, defaultRouteTableAssociation)
}

func testAccAWSEc2TransitGatewayConfigDefaultRouteTablePropagation(defaultRouteTablePropagation string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  default_route_table_propagation = %q
}
`, defaultRouteTablePropagation)
}

func testAccAWSEc2TransitGatewayConfigDnsSupport(dnsSupport string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  dns_support = %q
}
`, dnsSupport)
}

func testAccAWSEc2TransitGatewayConfigVpnEcmpSupport(vpnEcmpSupport string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  vpn_ecmp_support = %q
}
`, vpnEcmpSupport)
}

func testAccAWSEc2TransitGatewayConfigDescription(description string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  description = %q
}
`, description)
}

func testAccAWSEc2TransitGatewayConfigTags1(tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  tags = {
    %q = %q
  }
}
`, tagKey1, tagValue1)
}

func testAccAWSEc2TransitGatewayConfigTags2(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  tags = {
    %q = %q
    %q = %q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2)
}
