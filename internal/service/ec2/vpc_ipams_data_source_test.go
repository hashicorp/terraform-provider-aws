// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIPAMsDataSource_basic(t *testing.T) { // nosemgrep:ci.vpc-in-test-name
	ctx := acctest.Context(t)
	resourceName := "aws_vpc_ipam.test"
	dataSourceName := "data.aws_vpc_ipams.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},

		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMsDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ipams.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ipams.0.id", resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, "ipams.0.operating_regions.0.region_name", resourceName, "operating_regions.0.region_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ipams.0.tags.#", resourceName, "tags.#"),
				),
			},
		},
	})
}

func TestAccIPAMsDataSource_filter(t *testing.T) { // nosemgrep:ci.vpc-in-test-name
	ctx := acctest.Context(t)
	resourceName := "aws_vpc_ipam.test"
	dataSourceNameAdvanced := "data.aws_vpc_ipams.advanced"
	dataSourceNameFree := "data.aws_vpc_ipams.free"

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
				Config: testAccIPAMsDataSourceConfig_filter(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceNameAdvanced, "ipams.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceNameAdvanced, "ipams.0.id", resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceNameAdvanced, "ipams.0.tier", resourceName, "tier"),
					resource.TestCheckResourceAttr(dataSourceNameFree, "ipams.#", "0"),
				),
			},
		},
	})
}

func testAccIPAMsDataSourceConfig_basic() string {
	return acctest.ConfigCompose(testAccIPAMConfig_tags("Some", "Value"), `
data "aws_vpc_ipams" "test" {
  ipam_ids = [aws_vpc_ipam.test.id]
}
`)
}

func testAccIPAMsDataSourceConfig_filter() string {
	return acctest.ConfigCompose(testAccIPAMConfig_tags("Some", "Value"), `
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
