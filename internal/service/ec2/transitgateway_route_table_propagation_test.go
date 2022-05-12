package ec2_test

import (
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

func testAccTransitGatewayRouteTablePropagation_basic(t *testing.T) {
	var transitGatewayRouteTablePropagtion1 ec2.TransitGatewayRouteTablePropagation
	resourceName := "aws_ec2_transit_gateway_route_table_propagation.test"
	transitGatewayRouteTableResourceName := "aws_ec2_transit_gateway_route_table.test"
	transitGatewayVpcAttachmentResourceName := "aws_ec2_transit_gateway_vpc_attachment.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTransitGatewayRouteTablePropagationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRouteTablePropagationConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayRouteTablePropagationExists(resourceName, &transitGatewayRouteTablePropagtion1),
					resource.TestCheckResourceAttrSet(resourceName, "resource_id"),
					resource.TestCheckResourceAttrSet(resourceName, "resource_type"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_attachment_id", transitGatewayVpcAttachmentResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_route_table_id", transitGatewayRouteTableResourceName, "id"),
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

func testAccCheckTransitGatewayRouteTablePropagationExists(resourceName string, transitGatewayRouteTablePropagation *ec2.TransitGatewayRouteTablePropagation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Transit Gateway Route ID is set")
		}

		transitGatewayRouteTableID, transitGatewayAttachmentID, err := tfec2.DecodeTransitGatewayRouteTablePropagationID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		propagation, err := tfec2.FindTransitGatewayRouteTablePropagation(conn, transitGatewayRouteTableID, transitGatewayAttachmentID)

		if err != nil {
			return err
		}

		if propagation == nil {
			return fmt.Errorf("EC2 Transit Gateway Route Table Propagation not found")
		}

		if aws.StringValue(propagation.State) != ec2.TransitGatewayPropagationStateEnabled {
			return fmt.Errorf("EC2 Transit Gateway Route Table Propagation not in enabled state: %s", aws.StringValue(propagation.State))
		}

		*transitGatewayRouteTablePropagation = *propagation

		return nil
	}
}

func testAccCheckTransitGatewayRouteTablePropagationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_transit_gateway_route_table_propagation" {
			continue
		}

		transitGatewayRouteTableID, transitGatewayAttachmentID, err := tfec2.DecodeTransitGatewayRouteTablePropagationID(rs.Primary.ID)

		if err != nil {
			return err
		}

		propagation, err := tfec2.FindTransitGatewayRouteTablePropagation(conn, transitGatewayRouteTableID, transitGatewayAttachmentID)

		if tfawserr.ErrCodeEquals(err, "InvalidRouteTableID.NotFound") {
			continue
		}

		if err != nil {
			return err
		}

		if propagation == nil {
			continue
		}

		return fmt.Errorf("EC2 Transit Gateway Route Table (%s) Propagation (%s) still exists", transitGatewayRouteTableID, transitGatewayAttachmentID)
	}

	return nil
}

func testAccTransitGatewayRouteTablePropagationConfig() string {
	return `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-route"
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.0.0.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-route"
  }
}

resource "aws_ec2_transit_gateway" "test" {}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id
}

resource "aws_ec2_transit_gateway_route_table" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id
}

resource "aws_ec2_transit_gateway_route_table_propagation" "test" {
  transit_gateway_attachment_id  = aws_ec2_transit_gateway_vpc_attachment.test.id
  transit_gateway_route_table_id = aws_ec2_transit_gateway_route_table.test.id
}
`
}
