// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2VPCIPAMsDataSource_Tiered(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_vpc_ipam.test"
	dataSourceName := "data.aws_vpc_ipams.test"
	dataSourceFree := "data.aws_vpc_ipams.free"
	dataSourceAdvanced := "data.aws_vpc_ipams.advanced"

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
				Config: testAccVPCIPAMsDataSourceConfig_filterWithTiers(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "ipams.0.operating_regions.0.region_name", resourceName, "operating_regions.0.region_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ipams.0.tags.#", resourceName, "tags.#"),
					resource.TestCheckResourceAttr(dataSourceAdvanced, "ipams.#", "1"),
					resource.TestCheckResourceAttr(dataSourceFree, "ipams.#", "0"),
				),
			},
		},
	})
}

func testAccVPCIPAMsDataSourceConfig_filterWithTiers() string {
	return acctest.ConfigCompose(testAccIPAMConfig_tags("Some", "Value"), `
data "aws_vpc_ipams" "test" {
  ipam_ids = [aws_vpc_ipam.test.id]
}

data "aws_vpc_ipams" "advanced" {
  ipam_ids = [aws_vpc_ipam.test.id]
  filter {
    name   = "tier"
    values = ["advanced"]
  }
}

data "aws_vpc_ipams" "free" {
  ipam_ids = [aws_vpc_ipam.test.id]
  filter {
    name   = "tier"
    values = ["free"]
  }
}

`)
}
