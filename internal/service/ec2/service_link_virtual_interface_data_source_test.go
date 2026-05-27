// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2ServiceLinkVirtualInterfaceDataSource_filter(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ec2_service_link_virtual_interface.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceLinkVirtualInterfaceDataSourceConfig_filter(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, names.AttrID, regexache.MustCompile(`^slvif-`)),
					acctest.MatchResourceAttrRegionalARN(ctx, dataSourceName, names.AttrARN, "ec2", regexache.MustCompile(`service-link-virtual-interface/slvif-.+`)),
					resource.TestCheckResourceAttrSet(dataSourceName, "outpost_arn"),
					resource.TestCheckResourceAttrSet(dataSourceName, "outpost_lag_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "vlan"),
				),
			},
		},
	})
}

func TestAccEC2ServiceLinkVirtualInterfaceDataSource_id(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ec2_service_link_virtual_interface.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceLinkVirtualInterfaceDataSourceConfig_id(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, names.AttrID, regexache.MustCompile(`^slvif-`)),
					acctest.MatchResourceAttrRegionalARN(ctx, dataSourceName, names.AttrARN, "ec2", regexache.MustCompile(`service-link-virtual-interface/slvif-.+`)),
					resource.TestCheckResourceAttrSet(dataSourceName, "outpost_arn"),
					resource.TestCheckResourceAttrSet(dataSourceName, "outpost_lag_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "vlan"),
				),
			},
		},
	})
}

func testAccServiceLinkVirtualInterfaceDataSourceConfig_base() string {
	return `
data "aws_outposts_outposts" "test" {}

data "aws_ec2_service_link_virtual_interfaces" "base" {
  filter {
    name   = "outpost-arn"
    values = [tolist(data.aws_outposts_outposts.test.arns)[0]]
  }
}
`
}

func testAccServiceLinkVirtualInterfaceDataSourceConfig_filter() string {
	return acctest.ConfigCompose(testAccServiceLinkVirtualInterfaceDataSourceConfig_base(), `
data "aws_ec2_service_link_virtual_interface" "test" {
  filter {
    name   = "service-link-virtual-interface-id"
    values = [tolist(data.aws_ec2_service_link_virtual_interfaces.base.ids)[0]]
  }
}
`)
}

func testAccServiceLinkVirtualInterfaceDataSourceConfig_id() string {
	return acctest.ConfigCompose(testAccServiceLinkVirtualInterfaceDataSourceConfig_base(), `
data "aws_ec2_service_link_virtual_interface" "test" {
  id = tolist(data.aws_ec2_service_link_virtual_interfaces.base.ids)[0]
}
`)
}
