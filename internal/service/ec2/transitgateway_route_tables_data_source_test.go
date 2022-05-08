package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccTransitGatewayRouteTablesDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_ec2_transit_gateway_route_tables.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRouteTablesDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "ids.#", "0"),
				),
			},
		},
	})
}

func testAccTransitGatewayRouteTablesDataSource_filter(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_ec2_transit_gateway_route_tables.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRouteTablesTransitGatewayFilterDataSource(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", "2"),
				),
			},
		},
	})
}

func testAccTransitGatewayRouteTablesDataSource_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_ec2_transit_gateway_route_tables.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRouteTablesTransitGatewayTagsDataSource(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", "1"),
				),
			},
		},
	})
}

func testAccTransitGatewayRouteTablesDataSource_empty(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_ec2_transit_gateway_route_tables.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRouteTablesTransitGatewayEmptyDataSource(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", "0"),
				),
			},
		},
	})
}

func testAccTransitGatewayRouteTablesDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_route_table" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id

  tags = {
    Name = %[1]q
  }
}

data "aws_ec2_transit_gateway_route_tables" "test" {
  depends_on = [aws_ec2_transit_gateway_route_table.test]
}
`, rName)
}

func testAccTransitGatewayRouteTablesTransitGatewayFilterDataSource(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_route_table" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id

  tags = {
    Name = %[1]q
  }
}

data "aws_ec2_transit_gateway_route_tables" "test" {
  filter {
    name   = "transit-gateway-id"
    values = [aws_ec2_transit_gateway.test.id]
  }

  depends_on = [aws_ec2_transit_gateway_route_table.test]
}
`, rName)
}

func testAccTransitGatewayRouteTablesTransitGatewayTagsDataSource(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_route_table" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id

  tags = {
    Name = %[1]q
  }
}

data "aws_ec2_transit_gateway_route_tables" "test" {
  tags = {
    Name = %[1]q
  }

  depends_on = [aws_ec2_transit_gateway_route_table.test]
}
`, rName)
}

func testAccTransitGatewayRouteTablesTransitGatewayEmptyDataSource(rName string) string {
	return fmt.Sprintf(`
data "aws_ec2_transit_gateway_route_tables" "test" {
  tags = {
    Name = %[1]q
  }
}
`, rName)
}
