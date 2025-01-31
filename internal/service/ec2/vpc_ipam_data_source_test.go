// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2VPCIPAMDataSource_Basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_vpc_ipam.test"
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
				Config: testAccVPCIPAMDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "operating_regions.0.region_name", resourceName, "operating_regions.0.region_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.Test", resourceName, "tags.Test"),
				),
			},
		},
	})
}

func testAccVPCIPAMDataSourceConfig_basic() string {
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
  id = aws_vpc_ipam.test.id
}
`
}
