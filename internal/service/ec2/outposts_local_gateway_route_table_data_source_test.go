// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2OutpostsLocalGatewayRouteTableDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ec2_local_gateway_route_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOutpostsLocalGatewayRouteTableDataSourceConfig_routeTableID(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "local_gateway_id", regexache.MustCompile(`^lgw-`)),
					resource.TestMatchResourceAttr(dataSourceName, "local_gateway_route_table_id", regexache.MustCompile(`^lgw-rtb-`)),
					acctest.MatchResourceAttrRegionalARN(dataSourceName, "outpost_arn", "outposts", regexache.MustCompile(`outpost/op-.+`)),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrState, "available"),
				),
			},
		},
	})
}

func TestAccEC2OutpostsLocalGatewayRouteTableDataSource_filter(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ec2_local_gateway_route_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOutpostsLocalGatewayRouteTableDataSourceConfig_filter(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "local_gateway_id", regexache.MustCompile(`^lgw-`)),
					resource.TestMatchResourceAttr(dataSourceName, "local_gateway_route_table_id", regexache.MustCompile(`^lgw-rtb-`)),
					acctest.MatchResourceAttrRegionalARN(dataSourceName, "outpost_arn", "outposts", regexache.MustCompile(`outpost/op-.+`)),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrState, "available"),
				),
			},
		},
	})
}

func TestAccEC2OutpostsLocalGatewayRouteTableDataSource_localGatewayID(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ec2_local_gateway_route_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOutpostsLocalGatewayRouteTableDataSourceConfig_localGatewayID(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "local_gateway_id", regexache.MustCompile(`^lgw-`)),
					resource.TestMatchResourceAttr(dataSourceName, "local_gateway_route_table_id", regexache.MustCompile(`^lgw-rtb-`)),
					acctest.MatchResourceAttrRegionalARN(dataSourceName, "outpost_arn", "outposts", regexache.MustCompile(`outpost/op-.+`)),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrState, "available"),
				),
			},
		},
	})
}

func TestAccEC2OutpostsLocalGatewayRouteTableDataSource_outpostARN(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ec2_local_gateway_route_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOutpostsLocalGatewayRouteTableDataSourceConfig_outpostARN(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "local_gateway_id", regexache.MustCompile(`^lgw-`)),
					resource.TestMatchResourceAttr(dataSourceName, "local_gateway_route_table_id", regexache.MustCompile(`^lgw-rtb-`)),
					acctest.MatchResourceAttrRegionalARN(dataSourceName, "outpost_arn", "outposts", regexache.MustCompile(`outpost/op-.+`)),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrState, "available"),
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
