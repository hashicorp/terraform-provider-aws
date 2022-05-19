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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

func testAccTransitGatewayVPCAttachment_basic(t *testing.T) {
	var transitGatewayVpcAttachment1 ec2.TransitGatewayVpcAttachment
	resourceName := "aws_ec2_transit_gateway_vpc_attachment.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	vpcResourceName := "aws_vpc.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTransitGatewayVPCAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayVPCAttachmentConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayVPCAttachmentExists(resourceName, &transitGatewayVpcAttachment1),
					resource.TestCheckResourceAttr(resourceName, "dns_support", ec2.DnsSupportValueEnable),
					resource.TestCheckResourceAttr(resourceName, "ipv6_support", ec2.Ipv6SupportValueDisable),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_association", "true"),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_propagation", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", transitGatewayResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", vpcResourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "vpc_owner_id"),
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

func testAccTransitGatewayVPCAttachment_disappears(t *testing.T) {
	var transitGatewayVpcAttachment1 ec2.TransitGatewayVpcAttachment
	resourceName := "aws_ec2_transit_gateway_vpc_attachment.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTransitGatewayVPCAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayVPCAttachmentConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayVPCAttachmentExists(resourceName, &transitGatewayVpcAttachment1),
					testAccCheckTransitGatewayVPCAttachmentDisappears(&transitGatewayVpcAttachment1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccTransitGatewayVPCAttachment_ApplianceModeSupport(t *testing.T) {
	var transitGatewayVpcAttachment1, transitGatewayVpcAttachment2 ec2.TransitGatewayVpcAttachment
	resourceName := "aws_ec2_transit_gateway_vpc_attachment.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTransitGatewayVPCAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccTransitGatewayVPCAttachmentApplianceModeSupportConfig("false"),
				ExpectError: regexp.MustCompile(`expected appliance_mode_support to be one of`),
			},
			{
				Config:      testAccTransitGatewayVPCAttachmentApplianceModeSupportConfig("true"),
				ExpectError: regexp.MustCompile(`expected appliance_mode_support to be one of`),
			},
			{
				Config: testAccTransitGatewayVPCAttachmentApplianceModeSupportConfig(ec2.ApplianceModeSupportValueDisable),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayVPCAttachmentExists(resourceName, &transitGatewayVpcAttachment1),
					resource.TestCheckResourceAttr(resourceName, "appliance_mode_support", ec2.ApplianceModeSupportValueDisable),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayVPCAttachmentApplianceModeSupportConfig(ec2.ApplianceModeSupportValueEnable),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayVPCAttachmentExists(resourceName, &transitGatewayVpcAttachment2),
					testAccCheckTransitGatewayVPCAttachmentNotRecreated(&transitGatewayVpcAttachment1, &transitGatewayVpcAttachment2),
					resource.TestCheckResourceAttr(resourceName, "appliance_mode_support", ec2.ApplianceModeSupportValueEnable),
				),
			},
		},
	})
}

func testAccTransitGatewayVPCAttachment_DNSSupport(t *testing.T) {
	var transitGatewayVpcAttachment1, transitGatewayVpcAttachment2 ec2.TransitGatewayVpcAttachment
	resourceName := "aws_ec2_transit_gateway_vpc_attachment.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTransitGatewayVPCAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayVPCAttachmentDNSSupportConfig(ec2.DnsSupportValueDisable),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayVPCAttachmentExists(resourceName, &transitGatewayVpcAttachment1),
					resource.TestCheckResourceAttr(resourceName, "dns_support", ec2.DnsSupportValueDisable),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayVPCAttachmentDNSSupportConfig(ec2.DnsSupportValueEnable),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayVPCAttachmentExists(resourceName, &transitGatewayVpcAttachment2),
					testAccCheckTransitGatewayVPCAttachmentNotRecreated(&transitGatewayVpcAttachment1, &transitGatewayVpcAttachment2),
					resource.TestCheckResourceAttr(resourceName, "dns_support", ec2.DnsSupportValueEnable),
				),
			},
		},
	})
}

func testAccTransitGatewayVPCAttachment_IPv6Support(t *testing.T) {
	var transitGatewayVpcAttachment1, transitGatewayVpcAttachment2 ec2.TransitGatewayVpcAttachment
	resourceName := "aws_ec2_transit_gateway_vpc_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTransitGatewayVPCAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayVPCAttachmentIPv6SupportConfig(rName, ec2.Ipv6SupportValueEnable),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayVPCAttachmentExists(resourceName, &transitGatewayVpcAttachment1),
					resource.TestCheckResourceAttr(resourceName, "ipv6_support", ec2.Ipv6SupportValueEnable),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayVPCAttachmentIPv6SupportConfig(rName, ec2.Ipv6SupportValueDisable),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayVPCAttachmentExists(resourceName, &transitGatewayVpcAttachment2),
					testAccCheckTransitGatewayVPCAttachmentNotRecreated(&transitGatewayVpcAttachment1, &transitGatewayVpcAttachment2),
					resource.TestCheckResourceAttr(resourceName, "ipv6_support", ec2.Ipv6SupportValueDisable),
				),
			},
		},
	})
}

func testAccTransitGatewayVPCAttachment_SharedTransitGateway(t *testing.T) {
	var providers []*schema.Provider
	var transitGatewayVpcAttachment1 ec2.TransitGatewayVpcAttachment
	resourceName := "aws_ec2_transit_gateway_vpc_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckAlternateAccount(t)
			testAccPreCheckTransitGateway(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckTransitGatewayVPCAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayVPCAttachmentSharedTransitGatewayConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayVPCAttachmentExists(resourceName, &transitGatewayVpcAttachment1),
				),
			},
			{
				Config:            testAccTransitGatewayVPCAttachmentSharedTransitGatewayConfig(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccTransitGatewayVPCAttachment_SubnetIDs(t *testing.T) {
	var transitGatewayVpcAttachment1, transitGatewayVpcAttachment2, transitGatewayVpcAttachment3 ec2.TransitGatewayVpcAttachment
	resourceName := "aws_ec2_transit_gateway_vpc_attachment.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTransitGatewayVPCAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayVPCAttachmentSubnetIds2Config(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayVPCAttachmentExists(resourceName, &transitGatewayVpcAttachment1),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayVPCAttachmentSubnetIds1Config(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayVPCAttachmentExists(resourceName, &transitGatewayVpcAttachment2),
					testAccCheckTransitGatewayVPCAttachmentNotRecreated(&transitGatewayVpcAttachment1, &transitGatewayVpcAttachment2),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "1"),
				),
			},
			{
				Config: testAccTransitGatewayVPCAttachmentSubnetIds2Config(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayVPCAttachmentExists(resourceName, &transitGatewayVpcAttachment3),
					testAccCheckTransitGatewayVPCAttachmentNotRecreated(&transitGatewayVpcAttachment2, &transitGatewayVpcAttachment3),
					resource.TestCheckResourceAttr(resourceName, "subnet_ids.#", "2"),
				),
			},
		},
	})
}

func testAccTransitGatewayVPCAttachment_Tags(t *testing.T) {
	var transitGatewayVpcAttachment1, transitGatewayVpcAttachment2, transitGatewayVpcAttachment3 ec2.TransitGatewayVpcAttachment
	resourceName := "aws_ec2_transit_gateway_vpc_attachment.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTransitGatewayVPCAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayVPCAttachmentTags1Config("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayVPCAttachmentExists(resourceName, &transitGatewayVpcAttachment1),
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
				Config: testAccTransitGatewayVPCAttachmentTags2Config("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayVPCAttachmentExists(resourceName, &transitGatewayVpcAttachment2),
					testAccCheckTransitGatewayVPCAttachmentNotRecreated(&transitGatewayVpcAttachment1, &transitGatewayVpcAttachment2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccTransitGatewayVPCAttachmentTags1Config("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayVPCAttachmentExists(resourceName, &transitGatewayVpcAttachment3),
					testAccCheckTransitGatewayVPCAttachmentNotRecreated(&transitGatewayVpcAttachment2, &transitGatewayVpcAttachment3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccTransitGatewayVPCAttachment_TransitGatewayDefaultRouteTableAssociationAndPropagationDisabled(t *testing.T) {
	var transitGateway1 ec2.TransitGateway
	var transitGatewayVpcAttachment1 ec2.TransitGatewayVpcAttachment
	resourceName := "aws_ec2_transit_gateway_vpc_attachment.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTransitGatewayVPCAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayVPCAttachmentTransitGatewayDefaultRouteTableAssociationAndPropagationDisabledConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(transitGatewayResourceName, &transitGateway1),
					testAccCheckTransitGatewayVPCAttachmentExists(resourceName, &transitGatewayVpcAttachment1),
					testAccCheckTransitGatewayAssociationDefaultRouteTableAttachmentNotAssociated(&transitGateway1, &transitGatewayVpcAttachment1),
					testAccCheckTransitGatewayPropagationDefaultRouteTableAttachmentNotPropagated(&transitGateway1, &transitGatewayVpcAttachment1),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_association", "false"),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_propagation", "false"),
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

func testAccTransitGatewayVPCAttachment_TransitGatewayDefaultRouteTableAssociation(t *testing.T) {
	var transitGateway1, transitGateway2, transitGateway3 ec2.TransitGateway
	var transitGatewayVpcAttachment1, transitGatewayVpcAttachment2, transitGatewayVpcAttachment3 ec2.TransitGatewayVpcAttachment
	resourceName := "aws_ec2_transit_gateway_vpc_attachment.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTransitGatewayVPCAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayVPCAttachmentTransitGatewayDefaultRouteTableAssociationConfig(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(transitGatewayResourceName, &transitGateway1),
					testAccCheckTransitGatewayVPCAttachmentExists(resourceName, &transitGatewayVpcAttachment1),
					testAccCheckTransitGatewayAssociationDefaultRouteTableAttachmentNotAssociated(&transitGateway1, &transitGatewayVpcAttachment1),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_association", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayVPCAttachmentTransitGatewayDefaultRouteTableAssociationConfig(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(transitGatewayResourceName, &transitGateway2),
					testAccCheckTransitGatewayVPCAttachmentExists(resourceName, &transitGatewayVpcAttachment2),
					testAccCheckTransitGatewayVPCAttachmentNotRecreated(&transitGatewayVpcAttachment1, &transitGatewayVpcAttachment2),
					testAccCheckTransitGatewayAssociationDefaultRouteTableAttachmentAssociated(&transitGateway2, &transitGatewayVpcAttachment2),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_association", "true"),
				),
			},
			{
				Config: testAccTransitGatewayVPCAttachmentTransitGatewayDefaultRouteTableAssociationConfig(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(transitGatewayResourceName, &transitGateway3),
					testAccCheckTransitGatewayVPCAttachmentExists(resourceName, &transitGatewayVpcAttachment3),
					testAccCheckTransitGatewayVPCAttachmentNotRecreated(&transitGatewayVpcAttachment2, &transitGatewayVpcAttachment3),
					testAccCheckTransitGatewayAssociationDefaultRouteTableAttachmentNotAssociated(&transitGateway3, &transitGatewayVpcAttachment3),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_association", "false"),
				),
			},
		},
	})
}

func testAccTransitGatewayVPCAttachment_TransitGatewayDefaultRouteTablePropagation(t *testing.T) {
	var transitGateway1, transitGateway2, transitGateway3 ec2.TransitGateway
	var transitGatewayVpcAttachment1, transitGatewayVpcAttachment2, transitGatewayVpcAttachment3 ec2.TransitGatewayVpcAttachment
	resourceName := "aws_ec2_transit_gateway_vpc_attachment.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTransitGatewayVPCAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayVPCAttachmentTransitGatewayDefaultRouteTablePropagationConfig(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(transitGatewayResourceName, &transitGateway1),
					testAccCheckTransitGatewayVPCAttachmentExists(resourceName, &transitGatewayVpcAttachment1),
					testAccCheckTransitGatewayPropagationDefaultRouteTableAttachmentNotPropagated(&transitGateway1, &transitGatewayVpcAttachment1),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_propagation", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayVPCAttachmentTransitGatewayDefaultRouteTablePropagationConfig(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(transitGatewayResourceName, &transitGateway2),
					testAccCheckTransitGatewayVPCAttachmentExists(resourceName, &transitGatewayVpcAttachment2),
					testAccCheckTransitGatewayVPCAttachmentNotRecreated(&transitGatewayVpcAttachment1, &transitGatewayVpcAttachment2),
					testAccCheckTransitGatewayPropagationDefaultRouteTableAttachmentPropagated(&transitGateway2, &transitGatewayVpcAttachment2),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_propagation", "true"),
				),
			},
			{
				Config: testAccTransitGatewayVPCAttachmentTransitGatewayDefaultRouteTablePropagationConfig(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(transitGatewayResourceName, &transitGateway3),
					testAccCheckTransitGatewayVPCAttachmentExists(resourceName, &transitGatewayVpcAttachment3),
					testAccCheckTransitGatewayVPCAttachmentNotRecreated(&transitGatewayVpcAttachment2, &transitGatewayVpcAttachment3),
					testAccCheckTransitGatewayPropagationDefaultRouteTableAttachmentNotPropagated(&transitGateway3, &transitGatewayVpcAttachment3),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_propagation", "false"),
				),
			},
		},
	})
}

func testAccCheckTransitGatewayVPCAttachmentExists(resourceName string, transitGatewayVpcAttachment *ec2.TransitGatewayVpcAttachment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Transit Gateway VPC Attachment ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		attachment, err := tfec2.DescribeTransitGatewayVPCAttachment(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if attachment == nil {
			return fmt.Errorf("EC2 Transit Gateway VPC Attachment not found")
		}

		if aws.StringValue(attachment.State) != ec2.TransitGatewayAttachmentStateAvailable && aws.StringValue(attachment.State) != ec2.TransitGatewayAttachmentStatePendingAcceptance {
			return fmt.Errorf("EC2 Transit Gateway VPC Attachment (%s) exists in non-available/pending acceptance (%s) state", rs.Primary.ID, aws.StringValue(attachment.State))
		}

		*transitGatewayVpcAttachment = *attachment

		return nil
	}
}

func testAccCheckTransitGatewayVPCAttachmentDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_transit_gateway_route_table" {
			continue
		}

		vpcAttachment, err := tfec2.DescribeTransitGatewayVPCAttachment(conn, rs.Primary.ID)

		if tfawserr.ErrCodeEquals(err, "InvalidTransitGatewayAttachmentID.NotFound") {
			continue
		}

		if err != nil {
			return err
		}

		if vpcAttachment == nil {
			continue
		}

		if aws.StringValue(vpcAttachment.State) != ec2.TransitGatewayAttachmentStateDeleted {
			return fmt.Errorf("EC2 Transit Gateway VPC Attachment (%s) still exists in non-deleted (%s) state", rs.Primary.ID, aws.StringValue(vpcAttachment.State))
		}
	}

	return nil
}

func testAccCheckTransitGatewayVPCAttachmentDisappears(transitGatewayVpcAttachment *ec2.TransitGatewayVpcAttachment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		input := &ec2.DeleteTransitGatewayVpcAttachmentInput{
			TransitGatewayAttachmentId: transitGatewayVpcAttachment.TransitGatewayAttachmentId,
		}

		if _, err := conn.DeleteTransitGatewayVpcAttachment(input); err != nil {
			return err
		}

		return tfec2.WaitForTransitGatewayAttachmentDeletion(conn, aws.StringValue(transitGatewayVpcAttachment.TransitGatewayAttachmentId))
	}
}

func testAccCheckTransitGatewayVPCAttachmentNotRecreated(i, j *ec2.TransitGatewayVpcAttachment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.TransitGatewayAttachmentId) != aws.StringValue(j.TransitGatewayAttachmentId) {
			return errors.New("EC2 Transit Gateway VPC Attachment was recreated")
		}

		return nil
	}
}

func testAccPreCheckTransitGatewayVPCAttachment(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	input := &ec2.DescribeTransitGatewayVpcAttachmentsInput{
		MaxResults: aws.Int64(5),
	}

	_, err := conn.DescribeTransitGatewayVpcAttachments(input)

	if acctest.PreCheckSkipError(err) || tfawserr.ErrCodeEquals(err, "InvalidAction") {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccTransitGatewayVPCAttachmentConfig() string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(), `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-vpc-attachment"
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-vpc-attachment"
  }
}

resource "aws_ec2_transit_gateway" "test" {}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id
}
`)
}

func testAccTransitGatewayVPCAttachmentApplianceModeSupportConfig(appModeSupport string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-vpc-attachment"
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-vpc-attachment"
  }
}

resource "aws_ec2_transit_gateway" "test" {
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  appliance_mode_support = %[1]q
  subnet_ids             = [aws_subnet.test.id]
  transit_gateway_id     = aws_ec2_transit_gateway.test.id
  vpc_id                 = aws_vpc.test.id
}
`, appModeSupport))
}

func testAccTransitGatewayVPCAttachmentDNSSupportConfig(dnsSupport string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-vpc-attachment"
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-vpc-attachment"
  }
}

resource "aws_ec2_transit_gateway" "test" {
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  dns_support        = %[1]q
  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id
}
`, dnsSupport))
}

func testAccTransitGatewayVPCAttachmentIPv6SupportConfig(rName, ipv6Support string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  assign_generated_ipv6_cidr_block = true
  cidr_block                       = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  ipv6_cidr_block   = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, 1)
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
  ipv6_support       = %[2]q
  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName, ipv6Support))
}

func testAccTransitGatewayVPCAttachmentSharedTransitGatewayConfig(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAlternateAccountProvider(),
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		fmt.Sprintf(`
data "aws_organizations_organization" "test" {}

resource "aws_ec2_transit_gateway" "test" {
  provider = "awsalternate"
}

resource "aws_ram_resource_share" "test" {
  provider = "awsalternate"

  name = %[1]q
}

resource "aws_ram_resource_association" "test" {
  provider = "awsalternate"

  resource_arn       = aws_ec2_transit_gateway.test.arn
  resource_share_arn = aws_ram_resource_share.test.id
}

resource "aws_ram_principal_association" "test" {
  provider = "awsalternate"

  principal          = data.aws_organizations_organization.test.arn
  resource_share_arn = aws_ram_resource_share.test.id
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-vpc-attachment"
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-vpc-attachment"
  }
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  depends_on = [aws_ram_principal_association.test, aws_ram_resource_association.test]

  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id
}
`, rName))
}

func testAccTransitGatewayVPCAttachmentSubnetIds1Config() string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(), `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-vpc-attachment"
  }
}

resource "aws_subnet" "test" {
  count = "2"

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-vpc-attachment"
  }
}

resource "aws_ec2_transit_gateway" "test" {}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = [aws_subnet.test[0].id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id
}
`)
}

func testAccTransitGatewayVPCAttachmentSubnetIds2Config() string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(), `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-vpc-attachment"
  }
}

resource "aws_subnet" "test" {
  count = "2"

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-vpc-attachment"
  }
}

resource "aws_ec2_transit_gateway" "test" {}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = aws_subnet.test[*].id
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id
}
`)
}

func testAccTransitGatewayVPCAttachmentTags1Config(tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-vpc-attachment"
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-vpc-attachment"
  }
}

resource "aws_ec2_transit_gateway" "test" {}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccTransitGatewayVPCAttachmentTags2Config(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-vpc-attachment"
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-vpc-attachment"
  }
}

resource "aws_ec2_transit_gateway" "test" {}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccTransitGatewayVPCAttachmentTransitGatewayDefaultRouteTableAssociationAndPropagationDisabledConfig() string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(), `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-vpc-attachment"
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-vpc-attachment"
  }
}

resource "aws_ec2_transit_gateway" "test" {
  default_route_table_association = "disable"
  default_route_table_propagation = "disable"
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids                                      = [aws_subnet.test.id]
  transit_gateway_default_route_table_association = false
  transit_gateway_default_route_table_propagation = false
  transit_gateway_id                              = aws_ec2_transit_gateway.test.id
  vpc_id                                          = aws_vpc.test.id
}
`)
}

func testAccTransitGatewayVPCAttachmentTransitGatewayDefaultRouteTableAssociationConfig(transitGatewayDefaultRouteTableAssociation bool) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-vpc-attachment"
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-vpc-attachment"
  }
}

resource "aws_ec2_transit_gateway" "test" {
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids                                      = [aws_subnet.test.id]
  transit_gateway_default_route_table_association = %[1]t
  transit_gateway_id                              = aws_ec2_transit_gateway.test.id
  vpc_id                                          = aws_vpc.test.id
}
`, transitGatewayDefaultRouteTableAssociation))
}

func testAccTransitGatewayVPCAttachmentTransitGatewayDefaultRouteTablePropagationConfig(transitGatewayDefaultRouteTablePropagation bool) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-vpc-attachment"
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-vpc-attachment"
  }
}

resource "aws_ec2_transit_gateway" "test" {
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids                                      = [aws_subnet.test.id]
  transit_gateway_default_route_table_propagation = %[1]t
  transit_gateway_id                              = aws_ec2_transit_gateway.test.id
  vpc_id                                          = aws_vpc.test.id
}
`, transitGatewayDefaultRouteTablePropagation))
}
