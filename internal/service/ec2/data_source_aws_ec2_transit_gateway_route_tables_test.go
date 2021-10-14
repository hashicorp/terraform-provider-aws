package ec2_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func testAccDataSourceAwsEc2TransitGatewayRouteTables_basic(t *testing.T) {
	dataSourceName := "data.aws_ec2_transit_gateway_route_tables.test"

	resource.Test(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		ErrorCheck: acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsEc2TransitGatewayRouteTablesConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceAttrGreaterThanValue(dataSourceName, "ids.#", "0"),
				),
			},
		},
	})
}

func testAccDataSourceAwsEc2TransitGatewayRouteTables_Filter(t *testing.T) {
	dataSourceName := "data.aws_ec2_transit_gateway_route_tables.test"

	resource.Test(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		ErrorCheck: acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsEc2TransitGatewayRouteTablesTransitGatewayFilter,
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceAttrGreaterThanValue(dataSourceName, "ids.#", "0"),
				),
			},
		},
	})
}

const testAccDataSourceAwsEc2TransitGatewayRouteTablesConfig = `
resource "aws_ec2_transit_gateway" "test" {}

resource "aws_ec2_transit_gateway_route_table" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id
}

data "aws_ec2_transit_gateway_route_tables" "test" {
  depends_on = [aws_ec2_transit_gateway_route_table.test]
}
`

const testAccDataSourceAwsEc2TransitGatewayRouteTablesTransitGatewayFilter = `
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
