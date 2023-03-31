package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccTransitGatewayRouteTableAssociationsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_ec2_transit_gateway_route_table_associations.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckTransitGateway(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRouteTableAssociationsDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "ids.#", "0"),
				),
			},
		},
	})
}

func TestAccTransitGatewayRouteTableAssociationsDataSource_filter(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_ec2_transit_gateway_route_table_associations.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckTransitGateway(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRouteTableAssociationsDataSourceConfig_filter(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", "1"),
				),
			},
		},
	})
}

func testAccTransitGatewayRouteTableAssociationsDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
	default_route_table_association = "disable"

	tags = {
		Name = %[1]q
	}
}
resource "aws_vpc" "test" {
	cidr_block = "10.1.0.0/16"
	tags = {
	  Name = %[1]q
	}
}  
resource "aws_subnet" "test" {
	cidr_block = "10.1.1.0/24"
	vpc_id     = aws_vpc.test.id
	tags = {
	  Name = %[1]q
	}
}  
resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
	subnet_ids         = [aws_subnet.test.id]
	transit_gateway_id = aws_ec2_transit_gateway.test.id
	vpc_id             = aws_vpc.test.id

	transit_gateway_default_route_table_association = false
}
resource "aws_ec2_transit_gateway_route_table" "test" {
	transit_gateway_id = aws_ec2_transit_gateway.test.id
	tags = {
		Name = %[1]q
	}
}
resource "aws_ec2_transit_gateway_route_table_association" "test" {
	transit_gateway_attachment_id  = aws_ec2_transit_gateway_vpc_attachment.test.id
	transit_gateway_route_table_id = aws_ec2_transit_gateway_route_table.test.id
}
data "aws_ec2_transit_gateway_route_table_associations" "test" {
	transit_gateway_route_table_id = aws_ec2_transit_gateway_route_table.test.id
	depends_on = [aws_ec2_transit_gateway_route_table_association.test]
}
`, rName)
}

func testAccTransitGatewayRouteTableAssociationsDataSourceConfig_filter(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
	default_route_table_association = "disable"

	tags = {
		Name = %[1]q
	}
}
resource "aws_vpc" "test" {
	cidr_block = "10.1.0.0/16"
	tags = {
	  Name = %[1]q
	}
}  
resource "aws_subnet" "test" {
	cidr_block = "10.1.1.0/24"
	vpc_id     = aws_vpc.test.id
	tags = {
	  Name = %[1]q
	}
}  
resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
	subnet_ids         = [aws_subnet.test.id]
	transit_gateway_id = aws_ec2_transit_gateway.test.id
	vpc_id             = aws_vpc.test.id

	transit_gateway_default_route_table_association = false
}
resource "aws_ec2_transit_gateway_route_table" "test" {
	transit_gateway_id = aws_ec2_transit_gateway.test.id
	tags = {
		Name = %[1]q
	}
}
resource "aws_ec2_transit_gateway_route_table_association" "test" {
	transit_gateway_attachment_id  = aws_ec2_transit_gateway_vpc_attachment.test.id
	transit_gateway_route_table_id = aws_ec2_transit_gateway_route_table.test.id
}
data "aws_ec2_transit_gateway_route_table_associations" "test" {
	transit_gateway_route_table_id = aws_ec2_transit_gateway_route_table.test.id
	filter {
		name   = "transit-gateway-attachment-id"
		values = [aws_ec2_transit_gateway_vpc_attachment.test.id]
	}
	depends_on = [aws_ec2_transit_gateway_route_table_association.test]
}
`, rName)
}
