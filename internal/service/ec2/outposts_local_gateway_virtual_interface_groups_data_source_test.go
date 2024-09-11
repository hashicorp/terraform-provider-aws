// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2OutpostsLocalGatewayVirtualInterfaceGroupsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ec2_local_gateway_virtual_interface_groups.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOutpostsLocalGatewayVirtualInterfaceGroupsDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "local_gateway_virtual_interface_ids.#", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccEC2OutpostsLocalGatewayVirtualInterfaceGroupsDataSource_filter(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ec2_local_gateway_virtual_interface_groups.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOutpostsLocalGatewayVirtualInterfaceGroupsDataSourceConfig_filter(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "local_gateway_virtual_interface_ids.#", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccEC2OutpostsLocalGatewayVirtualInterfaceGroupsDataSource_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_ec2_local_gateway_virtual_interface_groups.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOutpostsLocalGatewayVirtualInterfaceGroupsDataSourceConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "local_gateway_virtual_interface_ids.#", acctest.Ct2),
				),
			},
		},
	})
}

func testAccOutpostsLocalGatewayVirtualInterfaceGroupsDataSourceConfig_basic() string {
	return `
data "aws_ec2_local_gateway_virtual_interface_groups" "test" {}
`
}

func testAccOutpostsLocalGatewayVirtualInterfaceGroupsDataSourceConfig_filter() string {
	return `
data "aws_ec2_local_gateways" "test" {}

data "aws_ec2_local_gateway_virtual_interface_groups" "test" {
  filter {
    name   = "local-gateway-id"
    values = [tolist(data.aws_ec2_local_gateways.test.ids)[0]]
  }
}
`
}

func testAccOutpostsLocalGatewayVirtualInterfaceGroupsDataSourceConfig_tags(rName string) string {
	return fmt.Sprintf(`
data "aws_ec2_local_gateways" "test" {}

data "aws_ec2_local_gateway_virtual_interface_groups" "source" {
  filter {
    name   = "local-gateway-id"
    values = [tolist(data.aws_ec2_local_gateways.test.ids)[0]]
  }
}

resource "aws_ec2_tag" "test" {
  key         = "TerraformAccTest-aws_ec2_local_gateway_virtual_interface_groups"
  resource_id = tolist(data.aws_ec2_local_gateway_virtual_interface_groups.source.ids)[0]
  value       = %[1]q
}

data "aws_ec2_local_gateway_virtual_interface_groups" "test" {
  tags = {
    (aws_ec2_tag.test.key) = aws_ec2_tag.test.value
  }
}
`, rName)
}
