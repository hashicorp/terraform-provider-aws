// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfsync "github.com/hashicorp/terraform-provider-aws/internal/experimental/sync"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccTransitGatewayRouteTableAssociationsDataSource_basic(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_ec2_transit_gateway_route_table_associations.test"

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
				Config: testAccTransitGatewayRouteTableAssociationsDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "ids.#", 0),
				),
			},
		},
	})
}

func testAccTransitGatewayRouteTableAssociationsDataSource_filter(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_ec2_transit_gateway_route_table_associations.test"

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
				Config: testAccTransitGatewayRouteTableAssociationsDataSourceConfig_filter(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", acctest.Ct1),
				),
			},
		},
	})
}

func testAccTransitGatewayRouteTableAssociationsDataSourceConfig_base(rName string) string {
	return acctest.ConfigCompose(testAccTransitGatewayRouteTableAssociationConfig_base(rName), fmt.Sprintf(`
resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = aws_subnet.test[*].id
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id

  transit_gateway_default_route_table_association = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_route_table_association" "test" {
  transit_gateway_attachment_id  = aws_ec2_transit_gateway_vpc_attachment.test.id
  transit_gateway_route_table_id = aws_ec2_transit_gateway_route_table.test.id
}
`, rName))
}

func testAccTransitGatewayRouteTableAssociationsDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccTransitGatewayRouteTableAssociationsDataSourceConfig_base(rName), `
data "aws_ec2_transit_gateway_route_table_associations" "test" {
  transit_gateway_route_table_id = aws_ec2_transit_gateway_route_table.test.id

  depends_on = [aws_ec2_transit_gateway_route_table_association.test]
}
`)
}

func testAccTransitGatewayRouteTableAssociationsDataSourceConfig_filter(rName string) string {
	return acctest.ConfigCompose(testAccTransitGatewayRouteTableAssociationsDataSourceConfig_base(rName), `
data "aws_ec2_transit_gateway_route_table_associations" "test" {
  transit_gateway_route_table_id = aws_ec2_transit_gateway_route_table.test.id

  filter {
    name   = "transit-gateway-attachment-id"
    values = [aws_ec2_transit_gateway_vpc_attachment.test.id]
  }

  depends_on = [aws_ec2_transit_gateway_route_table_association.test]
}
`)
}
