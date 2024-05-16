// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfsync "github.com/hashicorp/terraform-provider-aws/internal/experimental/sync"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccTransitGatewayRouteTablePropagationsDataSource_basic(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_ec2_transit_gateway_route_table_propagations.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRouteTablePropagationsDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "ids.#", 0),
				),
			},
		},
	})
}

func testAccTransitGatewayRouteTablePropagationsDataSource_filter(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_ec2_transit_gateway_route_table_propagations.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayRouteTablePropagationsDataSourceConfig_filter(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", acctest.Ct1),
				),
			},
		},
	})
}

func testAccTransitGatewayRouteTablePropagationsDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccTransitGatewayRouteTablePropagationConfig_basic(rName), `
data "aws_ec2_transit_gateway_route_table_propagations" "test" {
  transit_gateway_route_table_id = aws_ec2_transit_gateway_route_table.test.id

  depends_on = [aws_ec2_transit_gateway_route_table_propagation.test]
}
`)
}

func testAccTransitGatewayRouteTablePropagationsDataSourceConfig_filter(rName string) string {
	return acctest.ConfigCompose(testAccTransitGatewayRouteTablePropagationConfig_basic(rName), `
data "aws_ec2_transit_gateway_route_table_propagations" "test" {
  transit_gateway_route_table_id = aws_ec2_transit_gateway_route_table.test.id

  filter {
    name   = "transit-gateway-attachment-id"
    values = [aws_ec2_transit_gateway_vpc_attachment.test.id]
  }

  depends_on = [aws_ec2_transit_gateway_route_table_propagation.test]
}
`)
}
