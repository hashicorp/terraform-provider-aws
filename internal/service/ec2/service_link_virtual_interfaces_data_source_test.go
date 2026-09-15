// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2ServiceLinkVirtualInterfacesDataSource_filter(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ec2_service_link_virtual_interfaces.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceLinkVirtualInterfacesDataSourceConfig_filter(),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "ids.#", 0),
				),
			},
		},
	})
}

func testAccServiceLinkVirtualInterfacesDataSourceConfig_filter() string {
	return `
data "aws_outposts_outposts" "test" {}

data "aws_ec2_service_link_virtual_interfaces" "test" {
  filter {
    name   = "outpost-arn"
    values = [tolist(data.aws_outposts_outposts.test.arns)[0]]
  }
}
`
}
