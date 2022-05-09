package ec2_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccTransitGatewayRouteTableDataSource_Filter(t *testing.T) {
	dataSourceName := "data.aws_ec2_transit_gateway_route_table.test"
	resourceName := "aws_ec2_transit_gateway_route_table.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRouteTableFilterDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "default_association_route_table", dataSourceName, "default_association_route_table"),
					resource.TestCheckResourceAttrPair(resourceName, "default_propagation_route_table", dataSourceName, "default_propagation_route_table"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", dataSourceName, "transit_gateway_id"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
				),
			},
		},
	})
}

func testAccTransitGatewayRouteTableDataSource_ID(t *testing.T) {
	dataSourceName := "data.aws_ec2_transit_gateway_route_table.test"
	resourceName := "aws_ec2_transit_gateway_route_table.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRouteTableIDDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "default_association_route_table", dataSourceName, "default_association_route_table"),
					resource.TestCheckResourceAttrPair(resourceName, "default_propagation_route_table", dataSourceName, "default_propagation_route_table"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", dataSourceName, "transit_gateway_id"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
				),
			},
		},
	})
}

func testAccTransitGatewayRouteTableFilterDataSourceConfig() string {
	return `
resource "aws_ec2_transit_gateway" "test" {}

resource "aws_ec2_transit_gateway_route_table" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id
}

data "aws_ec2_transit_gateway_route_table" "test" {
  filter {
    name   = "transit-gateway-route-table-id"
    values = [aws_ec2_transit_gateway_route_table.test.id]
  }
}
`
}

func testAccTransitGatewayRouteTableIDDataSourceConfig() string {
	return `
resource "aws_ec2_transit_gateway" "test" {}

resource "aws_ec2_transit_gateway_route_table" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id
}

data "aws_ec2_transit_gateway_route_table" "test" {
  id = aws_ec2_transit_gateway_route_table.test.id
}
`
}
