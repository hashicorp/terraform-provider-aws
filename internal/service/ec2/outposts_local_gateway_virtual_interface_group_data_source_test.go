// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2OutpostsLocalGatewayVirtualInterfaceGroupDataSource_filter(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ec2_local_gateway_virtual_interface_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOutpostsLocalGatewayVirtualInterfaceGroupDataSourceConfig_filter(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, names.AttrID, regexache.MustCompile(`^lgw-vif-grp-`)),
					resource.TestMatchResourceAttr(dataSourceName, "local_gateway_id", regexache.MustCompile(`^lgw-`)),
					resource.TestCheckResourceAttr(dataSourceName, "local_gateway_virtual_interface_ids.#", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccEC2OutpostsLocalGatewayVirtualInterfaceGroupDataSource_localGatewayID(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ec2_local_gateway_virtual_interface_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOutpostsLocalGatewayVirtualInterfaceGroupDataSourceConfig_id(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, names.AttrID, regexache.MustCompile(`^lgw-vif-grp-`)),
					resource.TestMatchResourceAttr(dataSourceName, "local_gateway_id", regexache.MustCompile(`^lgw-`)),
					resource.TestCheckResourceAttr(dataSourceName, "local_gateway_virtual_interface_ids.#", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccEC2OutpostsLocalGatewayVirtualInterfaceGroupDataSource_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDataSourceName := "data.aws_ec2_local_gateway_virtual_interface_group.source"
	dataSourceName := "data.aws_ec2_local_gateway_virtual_interface_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOutpostsLocalGatewayVirtualInterfaceGroupDataSourceConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, sourceDataSourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, "local_gateway_id", sourceDataSourceName, "local_gateway_id"),
				),
			},
		},
	})
}

func testAccOutpostsLocalGatewayVirtualInterfaceGroupDataSourceConfig_filter() string {
	return `
data "aws_ec2_local_gateways" "test" {}

data "aws_ec2_local_gateway_virtual_interface_group" "test" {
  filter {
    name   = "local-gateway-id"
    values = [tolist(data.aws_ec2_local_gateways.test.ids)[0]]
  }
}
`
}

func testAccOutpostsLocalGatewayVirtualInterfaceGroupDataSourceConfig_id() string {
	return `
data "aws_ec2_local_gateways" "test" {}

data "aws_ec2_local_gateway_virtual_interface_group" "test" {
  local_gateway_id = tolist(data.aws_ec2_local_gateways.test.ids)[0]
}
`
}

func testAccOutpostsLocalGatewayVirtualInterfaceGroupDataSourceConfig_tags(rName string) string {
	return fmt.Sprintf(`
data "aws_ec2_local_gateways" "test" {}

data "aws_ec2_local_gateway_virtual_interface_group" "source" {
  filter {
    name   = "local-gateway-id"
    values = [tolist(data.aws_ec2_local_gateways.test.ids)[0]]
  }
}

resource "aws_ec2_tag" "test" {
  key         = "TerraformAccTest-aws_ec2_local_gateway_virtual_interface_group"
  resource_id = data.aws_ec2_local_gateway_virtual_interface_group.source.id
  value       = %[1]q
}

data "aws_ec2_local_gateway_virtual_interface_group" "test" {
  tags = {
    (aws_ec2_tag.test.key) = aws_ec2_tag.test.value
  }
}
`, rName)
}
