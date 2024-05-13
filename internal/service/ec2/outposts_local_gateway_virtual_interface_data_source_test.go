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

func TestAccEC2OutpostsLocalGatewayVirtualInterfaceDataSource_filter(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ec2_local_gateway_virtual_interface.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOutpostsLocalGatewayVirtualInterfaceDataSourceConfig_filter(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, names.AttrID, regexache.MustCompile(`^lgw-vif-`)),
					resource.TestMatchResourceAttr(dataSourceName, "local_address", regexache.MustCompile(`^\d+\.\d+\.\d+\.\d+/\d+$`)),
					resource.TestMatchResourceAttr(dataSourceName, "local_bgp_asn", regexache.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr(dataSourceName, "local_gateway_id", regexache.MustCompile(`^lgw-`)),
					resource.TestMatchResourceAttr(dataSourceName, "peer_address", regexache.MustCompile(`^\d+\.\d+\.\d+\.\d+/\d+$`)),
					resource.TestMatchResourceAttr(dataSourceName, "peer_bgp_asn", regexache.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr(dataSourceName, "vlan", regexache.MustCompile(`^\d+$`)),
				),
			},
		},
	})
}

func TestAccEC2OutpostsLocalGatewayVirtualInterfaceDataSource_id(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ec2_local_gateway_virtual_interface.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOutpostsLocalGatewayVirtualInterfaceDataSourceConfig_id(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, names.AttrID, regexache.MustCompile(`^lgw-vif-`)),
					resource.TestMatchResourceAttr(dataSourceName, "local_address", regexache.MustCompile(`^\d+\.\d+\.\d+\.\d+/\d+$`)),
					resource.TestMatchResourceAttr(dataSourceName, "local_bgp_asn", regexache.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr(dataSourceName, "local_gateway_id", regexache.MustCompile(`^lgw-`)),
					resource.TestMatchResourceAttr(dataSourceName, "peer_address", regexache.MustCompile(`^\d+\.\d+\.\d+\.\d+/\d+$`)),
					resource.TestMatchResourceAttr(dataSourceName, "peer_bgp_asn", regexache.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr(dataSourceName, "vlan", regexache.MustCompile(`^\d+$`)),
				),
			},
		},
	})
}

func TestAccEC2OutpostsLocalGatewayVirtualInterfaceDataSource_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	sourceDataSourceName := "data.aws_ec2_local_gateway_virtual_interface.source"
	dataSourceName := "data.aws_ec2_local_gateway_virtual_interface.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOutpostsLocalGatewayVirtualInterfaceDataSourceConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, sourceDataSourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, "local_address", sourceDataSourceName, "local_address"),
					resource.TestCheckResourceAttrPair(dataSourceName, "local_bgp_asn", sourceDataSourceName, "local_bgp_asn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "local_gateway_id", sourceDataSourceName, "local_gateway_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "peer_address", sourceDataSourceName, "peer_address"),
					resource.TestCheckResourceAttrPair(dataSourceName, "peer_bgp_asn", sourceDataSourceName, "peer_bgp_asn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vlan", sourceDataSourceName, "vlan"),
				),
			},
		},
	})
}

func testAccOutpostsLocalGatewayVirtualInterfaceDataSourceConfig_filter() string {
	return `
data "aws_ec2_local_gateways" "test" {}

data "aws_ec2_local_gateway_virtual_interface_group" "test" {
  local_gateway_id = tolist(data.aws_ec2_local_gateways.test.ids)[0]
}

data "aws_ec2_local_gateway_virtual_interface" "test" {
  filter {
    name   = "local-gateway-virtual-interface-id"
    values = [tolist(data.aws_ec2_local_gateway_virtual_interface_group.test.local_gateway_virtual_interface_ids)[0]]
  }
}
`
}

func testAccOutpostsLocalGatewayVirtualInterfaceDataSourceConfig_id() string {
	return `
data "aws_ec2_local_gateways" "test" {}

data "aws_ec2_local_gateway_virtual_interface_group" "test" {
  local_gateway_id = tolist(data.aws_ec2_local_gateways.test.ids)[0]
}

data "aws_ec2_local_gateway_virtual_interface" "test" {
  id = tolist(data.aws_ec2_local_gateway_virtual_interface_group.test.local_gateway_virtual_interface_ids)[0]
}
`
}

func testAccOutpostsLocalGatewayVirtualInterfaceDataSourceConfig_tags(rName string) string {
	return fmt.Sprintf(`
data "aws_ec2_local_gateways" "test" {}

data "aws_ec2_local_gateway_virtual_interface_group" "test" {
  local_gateway_id = tolist(data.aws_ec2_local_gateways.test.ids)[0]
}

data "aws_ec2_local_gateway_virtual_interface" "source" {
  id = tolist(data.aws_ec2_local_gateway_virtual_interface_group.test.local_gateway_virtual_interface_ids)[0]
}

resource "aws_ec2_tag" "test" {
  key         = "TerraformAccTest-aws_ec2_local_gateway_virtual_interface"
  resource_id = data.aws_ec2_local_gateway_virtual_interface.source.id
  value       = %[1]q
}

data "aws_ec2_local_gateway_virtual_interface" "test" {
  tags = {
    (aws_ec2_tag.test.key) = aws_ec2_tag.test.value
  }
}
`, rName)
}
