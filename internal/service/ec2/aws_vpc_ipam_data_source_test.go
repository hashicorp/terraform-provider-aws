// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2AwsVpcIpamDataSourceBasic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_vpc_ipam.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsVpcIpamDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "tags.Test", "Test"),
				),
			},
		},
	})
}

func testAccAwsVpcIpamDataSourceConfig_basic() string {
	return `
data "aws_region" "current" {}

resource "aws_vpc_ipam" "test" {
  description = "My IPAM"
  operating_regions {
    region_name = data.aws_region.current.name
  }

  tags = {
    Test = "Test"
  }
}

data "aws_vpc_ipam" "test" {
	ipam_id = aws_vpc_ipam.test.id
}
`
}
