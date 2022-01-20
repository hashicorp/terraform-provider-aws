package ec2_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccTransitGatewayRouteTablesDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_ec2_transit_gateway_route_tables.test"

	resource.Test(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck: acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRouteTablesDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "ids.#", "0"),
				),
			},
		},
	})
}

func testAccTransitGatewayRouteTablesDataSource_Filter(t *testing.T) {
	dataSourceName := "data.aws_ec2_transit_gateway_route_tables.test"

	resource.Test(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck: acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRouteTablesTransitGatewayFilterDataSource,
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "ids.#", "0"),
				),
			},
		},
	})
}

const testAccTransitGatewayRouteTablesDataSourceConfig = `
resource "aws_ec2_transit_gateway" "test" {}

resource "aws_ec2_transit_gateway_route_table" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id
}

data "aws_ec2_transit_gateway_route_tables" "test" {
  depends_on = [aws_ec2_transit_gateway_route_table.test]
}
`

const testAccTransitGatewayRouteTablesTransitGatewayFilterDataSource = `
resource "aws_ec2_transit_gateway" "test" {}

resource "aws_ec2_transit_gateway_route_table" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id
}

data "aws_ec2_transit_gateway_route_tables" "test" {
  filter {
    name   = "transit-gateway-id"
    values = [aws_ec2_transit_gateway.test.id]
  }

  depends_on = [aws_ec2_transit_gateway_route_table.test]
}
`
