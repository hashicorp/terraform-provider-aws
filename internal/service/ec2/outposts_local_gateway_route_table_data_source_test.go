package ec2_test

import (
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccEC2OutpostsLocalGatewayRouteTableDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_ec2_local_gateway_route_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOutpostsLocalGatewayRouteTableDataSourceConfig_routeTableID(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "local_gateway_id", regexp.MustCompile(`^lgw-`)),
					resource.TestMatchResourceAttr(dataSourceName, "local_gateway_route_table_id", regexp.MustCompile(`^lgw-rtb-`)),
					acctest.MatchResourceAttrRegionalARN(dataSourceName, "outpost_arn", "outposts", regexp.MustCompile(`outpost/op-.+`)),
					resource.TestCheckResourceAttr(dataSourceName, "state", "available"),
				),
			},
		},
	})
}

func TestAccEC2OutpostsLocalGatewayRouteTableDataSource_filter(t *testing.T) {
	dataSourceName := "data.aws_ec2_local_gateway_route_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOutpostsLocalGatewayRouteTableDataSourceConfig_filter(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "local_gateway_id", regexp.MustCompile(`^lgw-`)),
					resource.TestMatchResourceAttr(dataSourceName, "local_gateway_route_table_id", regexp.MustCompile(`^lgw-rtb-`)),
					acctest.MatchResourceAttrRegionalARN(dataSourceName, "outpost_arn", "outposts", regexp.MustCompile(`outpost/op-.+`)),
					resource.TestCheckResourceAttr(dataSourceName, "state", "available"),
				),
			},
		},
	})
}

func TestAccEC2OutpostsLocalGatewayRouteTableDataSource_localGatewayID(t *testing.T) {
	dataSourceName := "data.aws_ec2_local_gateway_route_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOutpostsLocalGatewayRouteTableDataSourceConfig_localGatewayID(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "local_gateway_id", regexp.MustCompile(`^lgw-`)),
					resource.TestMatchResourceAttr(dataSourceName, "local_gateway_route_table_id", regexp.MustCompile(`^lgw-rtb-`)),
					acctest.MatchResourceAttrRegionalARN(dataSourceName, "outpost_arn", "outposts", regexp.MustCompile(`outpost/op-.+`)),
					resource.TestCheckResourceAttr(dataSourceName, "state", "available"),
				),
			},
		},
	})
}

func TestAccEC2OutpostsLocalGatewayRouteTableDataSource_outpostARN(t *testing.T) {
	dataSourceName := "data.aws_ec2_local_gateway_route_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOutpostsLocalGatewayRouteTableDataSourceConfig_outpostARN(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "local_gateway_id", regexp.MustCompile(`^lgw-`)),
					resource.TestMatchResourceAttr(dataSourceName, "local_gateway_route_table_id", regexp.MustCompile(`^lgw-rtb-`)),
					acctest.MatchResourceAttrRegionalARN(dataSourceName, "outpost_arn", "outposts", regexp.MustCompile(`outpost/op-.+`)),
					resource.TestCheckResourceAttr(dataSourceName, "state", "available"),
				),
			},
		},
	})
}

func testAccOutpostsLocalGatewayRouteTableDataSourceConfig_filter() string {
	return `
data "aws_outposts_outposts" "test" {}

data "aws_ec2_local_gateway_route_table" "test" {
  filter {
    name   = "outpost-arn"
    values = [tolist(data.aws_outposts_outposts.test.arns)[0]]
  }
}
`
}

func testAccOutpostsLocalGatewayRouteTableDataSourceConfig_localGatewayID() string {
	return `
data "aws_ec2_local_gateways" "test" {}

data "aws_ec2_local_gateway_route_table" "test" {
  local_gateway_id = tolist(data.aws_ec2_local_gateways.test.ids)[0]
}
`
}

func testAccOutpostsLocalGatewayRouteTableDataSourceConfig_routeTableID() string {
	return `
data "aws_ec2_local_gateway_route_tables" "test" {}

data "aws_ec2_local_gateway_route_table" "test" {
  local_gateway_route_table_id = tolist(data.aws_ec2_local_gateway_route_tables.test.ids)[0]
}
`
}

func testAccOutpostsLocalGatewayRouteTableDataSourceConfig_outpostARN() string {
	return `
data "aws_outposts_outposts" "test" {}

data "aws_ec2_local_gateway_route_table" "test" {
  outpost_arn = tolist(data.aws_outposts_outposts.test.arns)[0]
}
`
}
