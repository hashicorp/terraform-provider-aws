// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIPAMDataSource_basic(t *testing.T) { // nosemgrep:ci.vpc-in-test-name
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
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, "default_resource_discovery_id", resourceName, "default_resource_discovery_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "default_resource_discovery_association_id", resourceName, "default_resource_discovery_association_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "enable_private_gua", resourceName, "enable_private_gua"),
					resource.TestCheckResourceAttrPair(dataSourceName, "operating_regions.0.region_name", resourceName, "operating_regions.0.region_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "private_default_scope_id", resourceName, "private_default_scope_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "public_default_scope_id", resourceName, "public_default_scope_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "scope_count", resourceName, "scope_count"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tier", resourceName, "tier"),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.Test", resourceName, "tags.Test"),
				),
			},
		},
	})
}

func testAccIPAMDataSourceConfig_basic() string {
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
