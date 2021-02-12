package aws

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSEc2TransitGatewayRouteTablesDataSource_TransitGatewayFilter(t *testing.T) {
	dataSourceName := "data.aws_ec2_transit_gateway_route_tables.test"
	resourceName := "aws_ec2_transit_gateway_route_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayRouteTablesDataSourceTransitGatewayFilter(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", dataSourceName, "transit_gateway_id"),
				),
			},
		},
	})
}

func testAccAWSEc2TransitGatewayRouteTablesDataSourceTransitGatewayFilter() string {
	return `
resource "aws_ec2_transit_gateway" "test" {}

resource "aws_ec2_transit_gateway_route_table" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id
}

data "aws_ec2_transit_gateway_route_tables" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id

  depends_on = [aws_ec2_transit_gateway_route_table.test]
}
`
}
