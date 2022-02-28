package ec2_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

func testAccTransitGatewayConnect_basic(t *testing.T) {
	var transitGatewayConnect1 ec2.TransitGatewayConnect
	resourceName := "aws_ec2_transit_gateway_connect.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	transitGatewayVpcAttachmentResourceName := "aws_ec2_transit_gateway_vpc_attachment.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckTransitGatewayVpcAttachment(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTransitGatewayConnectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConnectConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayConnectExists(resourceName, &transitGatewayConnect1),
					resource.TestCheckResourceAttr(resourceName, "protocol", "gre"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_association", "true"),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_propagation", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", transitGatewayResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "transport_attachment_id", transitGatewayVpcAttachmentResourceName, "id"),
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

func testAccTransitGatewayConnect_disappears(t *testing.T) {
	var transitGatewayConnect1 ec2.TransitGatewayConnect
	resourceName := "aws_ec2_transit_gateway_connect.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckTransitGatewayVpcAttachment(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTransitGatewayConnectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConnectConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayConnectExists(resourceName, &transitGatewayConnect1),
					testAccCheckTransitGatewayConnectDisappears(&transitGatewayConnect1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccTransitGatewayConnect_Tags(t *testing.T) {
	var transitGatewayConnect1, transitGatewayConnect2, transitGatewayConnect3 ec2.TransitGatewayConnect
	resourceName := "aws_ec2_transit_gateway_connect.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckTransitGatewayVpcAttachment(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTransitGatewayConnectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConnectTags1Config("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayConnectExists(resourceName, &transitGatewayConnect1),
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
				Config: testAccTransitGatewayConnectTags2Config("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayConnectExists(resourceName, &transitGatewayConnect2),
					testAccCheckTransitGatewayConnectNotRecreated(&transitGatewayConnect1, &transitGatewayConnect2),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccTransitGatewayConnectTags1Config("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayConnectExists(resourceName, &transitGatewayConnect3),
					testAccCheckTransitGatewayConnectNotRecreated(&transitGatewayConnect2, &transitGatewayConnect3),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccTransitGatewayConnect_TransitGatewayDefaultRouteTableAssociationAndPropagationDisabled(t *testing.T) {
	var transitGateway1 ec2.TransitGateway
	var transitGatewayConnect1 ec2.TransitGatewayConnect
	resourceName := "aws_ec2_transit_gateway_connect.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckTransitGatewayVpcAttachment(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTransitGatewayConnectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConnectTransitGatewayDefaultRouteTableAssociationAndPropagationDisabledConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(transitGatewayResourceName, &transitGateway1),
					testAccCheckTransitGatewayConnectExists(resourceName, &transitGatewayConnect1),
					testAccCheckTransitGatewayAssociationDefaultRouteTableAttachmentNotAssociated(&transitGateway1, &transitGatewayConnect1),
					testAccCheckTransitGatewayPropagationDefaultRouteTableAttachmentNotPropagated(&transitGateway1, &transitGatewayConnect1),
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

func testAccTransitGatewayConnect_TransitGatewayDefaultRouteTableAssociation(t *testing.T) {
	var transitGateway1, transitGateway2, transitGateway3 ec2.TransitGateway
	var transitGatewayConnect1, transitGatewayConnect2, transitGatewayConnect3 ec2.TransitGatewayConnect
	resourceName := "aws_ec2_transit_gateway_connect.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckTransitGatewayVpcAttachment(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTransitGatewayConnectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConnectTransitGatewayDefaultRouteTableAssociationConfig(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(transitGatewayResourceName, &transitGateway1),
					testAccCheckTransitGatewayConnectExists(resourceName, &transitGatewayConnect1),
					testAccCheckTransitGatewayAssociationDefaultRouteTableAttachmentNotAssociated(&transitGateway1, &transitGatewayConnect1),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_association", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayConnectTransitGatewayDefaultRouteTableAssociationConfig(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(transitGatewayResourceName, &transitGateway2),
					testAccCheckTransitGatewayConnectExists(resourceName, &transitGatewayConnect2),
					testAccCheckTransitGatewayConnectNotRecreated(&transitGatewayConnect1, &transitGatewayConnect2),
					testAccCheckTransitGatewayAssociationDefaultRouteTableAttachmentAssociated(&transitGateway2, &transitGatewayConnect2),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_association", "true"),
				),
			},
			{
				Config: testAccTransitGatewayConnectTransitGatewayDefaultRouteTableAssociationConfig(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(transitGatewayResourceName, &transitGateway3),
					testAccCheckTransitGatewayConnectExists(resourceName, &transitGatewayConnect3),
					testAccCheckTransitGatewayConnectNotRecreated(&transitGatewayConnect2, &transitGatewayConnect3),
					testAccCheckTransitGatewayAssociationDefaultRouteTableAttachmentNotAssociated(&transitGateway3, &transitGatewayConnect3),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_association", "false"),
				),
			},
		},
	})
}

func testAccTransitGatewayConnect_TransitGatewayDefaultRouteTablePropagation(t *testing.T) {
	var transitGateway1, transitGateway2, transitGateway3 ec2.TransitGateway
	var transitGatewayConnect1, transitGatewayConnect2, transitGatewayConnect3 ec2.TransitGatewayConnect
	resourceName := "aws_ec2_transit_gateway_connect.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckTransitGatewayVpcAttachment(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckTransitGatewayConnectDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConnectTransitGatewayDefaultRouteTablePropagationConfig(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(transitGatewayResourceName, &transitGateway1),
					testAccCheckTransitGatewayConnectExists(resourceName, &transitGatewayConnect1),
					testAccCheckTransitGatewayPropagationDefaultRouteTableAttachmentNotPropagated(&transitGateway1, &transitGatewayConnect1),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_propagation", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTransitGatewayConnectTransitGatewayDefaultRouteTablePropagationConfig(true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(transitGatewayResourceName, &transitGateway2),
					testAccCheckTransitGatewayConnectExists(resourceName, &transitGatewayConnect2),
					testAccCheckTransitGatewayConnectNotRecreated(&transitGatewayConnect1, &transitGatewayConnect2),
					testAccCheckTransitGatewayPropagationDefaultRouteTableAttachmentPropagated(&transitGateway2, &transitGatewayConnect2),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_propagation", "true"),
				),
			},
			{
				Config: testAccTransitGatewayConnectTransitGatewayDefaultRouteTablePropagationConfig(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayExists(transitGatewayResourceName, &transitGateway3),
					testAccCheckTransitGatewayConnectExists(resourceName, &transitGatewayConnect3),
					testAccCheckTransitGatewayConnectNotRecreated(&transitGatewayConnect2, &transitGatewayConnect3),
					testAccCheckTransitGatewayPropagationDefaultRouteTableAttachmentNotPropagated(&transitGateway3, &transitGatewayConnect3),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_default_route_table_propagation", "false"),
				),
			},
		},
	})
}

func testAccCheckTransitGatewayConnectExists(resourceName string, transitGatewayConnect *ec2.TransitGatewayConnect) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Transit Gateway Connect ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		attachment, err := tfec2.DescribeTransitGatewayConnect(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if attachment == nil {
			return fmt.Errorf("EC2 Transit Gateway Connect not found")
		}

		if aws.StringValue(attachment.State) != ec2.TransitGatewayAttachmentStateAvailable && aws.StringValue(attachment.State) != ec2.TransitGatewayAttachmentStatePendingAcceptance {
			return fmt.Errorf("EC2 Transit Gateway Connect (%s) exists in non-available/pending acceptance (%s) state", rs.Primary.ID, aws.StringValue(attachment.State))
		}

		*transitGatewayConnect = *attachment

		return nil
	}
}

func testAccCheckTransitGatewayConnectDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_transit_gateway_route_table" {
			continue
		}

		vpcAttachment, err := tfec2.DescribeTransitGatewayConnect(conn, rs.Primary.ID)

		if tfawserr.ErrMessageContains(err, "InvalidTransitGatewayAttachmentID.NotFound", "") {
			continue
		}

		if err != nil {
			return err
		}

		if vpcAttachment == nil {
			continue
		}

		if aws.StringValue(vpcAttachment.State) != ec2.TransitGatewayAttachmentStateDeleted {
			return fmt.Errorf("EC2 Transit Gateway Connect (%s) still exists in non-deleted (%s) state", rs.Primary.ID, aws.StringValue(vpcAttachment.State))
		}
	}

	return nil
}

func testAccCheckTransitGatewayConnectDisappears(transitGatewayConnect *ec2.TransitGatewayConnect) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		input := &ec2.DeleteTransitGatewayConnectInput{
			TransitGatewayAttachmentId: transitGatewayConnect.TransitGatewayAttachmentId,
		}

		if _, err := conn.DeleteTransitGatewayConnect(input); err != nil {
			return err
		}

		return tfec2.WaitForTransitGatewayAttachmentDeletion(conn, aws.StringValue(transitGatewayConnect.TransitGatewayAttachmentId))
	}
}

func testAccCheckTransitGatewayConnectNotRecreated(i, j *ec2.TransitGatewayConnect) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.TransitGatewayAttachmentId) != aws.StringValue(j.TransitGatewayAttachmentId) {
			return errors.New("EC2 Transit Gateway Connect was recreated")
		}

		return nil
	}
}

func testAccPreCheckTransitGatewayConnect(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	input := &ec2.DescribeTransitGatewayConnectsInput{
		MaxResults: aws.Int64(5),
	}

	_, err := conn.DescribeTransitGatewayConnects(input)

	if acctest.PreCheckSkipError(err) || tfawserr.ErrMessageContains(err, "InvalidAction", "") {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccTransitGatewayConnectConfig() string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(), `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-connect"
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-connect"
  }
}

resource "aws_ec2_transit_gateway" "test" {}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id
}

resource "aws_ec2_transit_gateway_connect" "test" {
  transit_gateway_id      = aws_ec2_transit_gateway.test.id
  transport_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id
}
`)
}

func testAccTransitGatewayConnectTags1Config(tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-connect"
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-connect"
  }
}

resource "aws_ec2_transit_gateway" "test" {}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id
}

resource "aws_ec2_transit_gateway_connect" "test" {
  transit_gateway_id      = aws_ec2_transit_gateway.test.id
  transport_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccTransitGatewayConnectTags2Config(tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-connect"
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-connect"
  }
}

resource "aws_ec2_transit_gateway" "test" {}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id
}

resource "aws_ec2_transit_gateway_connect" "test" {
  transit_gateway_id      = aws_ec2_transit_gateway.test.id
  transport_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccTransitGatewayConnectTransitGatewayDefaultRouteTableAssociationAndPropagationDisabledConfig() string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(), `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-connect"
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-connect"
  }
}

resource "aws_ec2_transit_gateway" "test" {
  default_route_table_association = "disable"
  default_route_table_propagation = "disable"
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids                                      = [aws_subnet.test.id]
  transit_gateway_id                              = aws_ec2_transit_gateway.test.id
  vpc_id                                          = aws_vpc.test.id
  transit_gateway_default_route_table_association = false
  transit_gateway_default_route_table_propagation = false
}

resource "aws_ec2_transit_gateway_connect" "test" {
  transit_gateway_id                              = aws_ec2_transit_gateway.test.id
  transport_attachment_id                         = aws_ec2_transit_gateway_vpc_attachment.test.id
  transit_gateway_default_route_table_association = false
  transit_gateway_default_route_table_propagation = false
}
`)
}

func testAccTransitGatewayConnectTransitGatewayDefaultRouteTableAssociationConfig(transitGatewayDefaultRouteTableAssociation bool) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-connect"
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-connect"
  }
}

resource "aws_ec2_transit_gateway" "test" {
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids                                      = [aws_subnet.test.id]
  transit_gateway_id                              = aws_ec2_transit_gateway.test.id
  vpc_id                                          = aws_vpc.test.id
  transit_gateway_default_route_table_association = %[1]t
}

resource "aws_ec2_transit_gateway_connect" "test" {
  transit_gateway_id                              = aws_ec2_transit_gateway.test.id
  transport_attachment_id                         = aws_ec2_transit_gateway_vpc_attachment.test.id
  transit_gateway_default_route_table_association = %[2]t
}
`, transitGatewayDefaultRouteTableAssociation, transitGatewayDefaultRouteTableAssociation))
}

func testAccTransitGatewayConnectTransitGatewayDefaultRouteTablePropagationConfig(transitGatewayDefaultRouteTablePropagation bool) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-connect"
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-connect"
  }
}

resource "aws_ec2_transit_gateway" "test" {
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids                                      = [aws_subnet.test.id]
  transit_gateway_id                              = aws_ec2_transit_gateway.test.id
  vpc_id                                          = aws_vpc.test.id
  transit_gateway_default_route_table_propagation = %[1]t
}

resource "aws_ec2_transit_gateway_connect" "test" {
  transit_gateway_id                              = aws_ec2_transit_gateway.test.id
  transport_attachment_id                         = aws_ec2_transit_gateway_vpc_attachment.test.id
  transit_gateway_default_route_table_propagation = %[2]t
}
`, transitGatewayDefaultRouteTablePropagation, transitGatewayDefaultRouteTablePropagation))
}
