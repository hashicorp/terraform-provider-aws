package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

func testAccTransitGatewayRouteTableAssociation_basic(t *testing.T) {
	var transitGatewayRouteTablePropagtion1 ec2.TransitGatewayRouteTableAssociation
	resourceName := "aws_ec2_transit_gateway_route_table_association.test"
	transitGatewayRouteTableResourceName := "aws_ec2_transit_gateway_route_table.test"
	transitGatewayVpcAttachmentResourceName := "aws_ec2_transit_gateway_vpc_attachment.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTransitGatewayRouteTableAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRouteTableAssociationConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTransitGatewayRouteTableAssociationExists(resourceName, &transitGatewayRouteTablePropagtion1),
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

func testAccCheckTransitGatewayRouteTableAssociationExists(resourceName string, transitGatewayRouteTableAssociation *ec2.TransitGatewayRouteTableAssociation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Transit Gateway Route ID is set")
		}

		transitGatewayRouteTableID, transitGatewayAttachmentID, err := tfec2.DecodeTransitGatewayRouteTableAssociationID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		association, err := tfec2.DescribeTransitGatewayRouteTableAssociation(conn, transitGatewayRouteTableID, transitGatewayAttachmentID)

		if err != nil {
			return err
		}

		if association == nil {
			return fmt.Errorf("EC2 Transit Gateway Route Table Association not found")
		}

		*transitGatewayRouteTableAssociation = *association

		return nil
	}
}

func testAccCheckTransitGatewayRouteTableAssociationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_transit_gateway_route_table_association" {
			continue
		}

		transitGatewayRouteTableID, transitGatewayAttachmentID, err := tfec2.DecodeTransitGatewayRouteTableAssociationID(rs.Primary.ID)

		if err != nil {
			return err
		}

		association, err := tfec2.DescribeTransitGatewayRouteTableAssociation(conn, transitGatewayRouteTableID, transitGatewayAttachmentID)

		if tfawserr.ErrCodeEquals(err, "InvalidRouteTableID.NotFound") {
			continue
		}

		if err != nil {
			return err
		}

		if association == nil {
			continue
		}

		return fmt.Errorf("EC2 Transit Gateway Route Table (%s) Association (%s) still exists", transitGatewayRouteTableID, transitGatewayAttachmentID)
	}

	return nil
}

func testAccTransitGatewayRouteTableAssociationConfig_basic() string {
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
  subnet_ids                                      = [aws_subnet.test.id]
  transit_gateway_default_route_table_association = false
  transit_gateway_id                              = aws_ec2_transit_gateway.test.id
  vpc_id                                          = aws_vpc.test.id
}

resource "aws_ec2_transit_gateway_route_table" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id
}

resource "aws_ec2_transit_gateway_route_table_association" "test" {
  transit_gateway_attachment_id  = aws_ec2_transit_gateway_vpc_attachment.test.id
  transit_gateway_route_table_id = aws_ec2_transit_gateway_route_table.test.id
}
`
}
