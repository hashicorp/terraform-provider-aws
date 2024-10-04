// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIPAMPoolDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_vpc_ipam_pool.test"
	dataSourceName := "data.aws_vpc_ipam_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMPoolDataSourceConfig_optionsBasic,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "address_family", resourceName, "address_family"),
					resource.TestCheckResourceAttrPair(dataSourceName, "allocation_default_netmask_length", resourceName, "allocation_default_netmask_length"),
					resource.TestCheckResourceAttrPair(dataSourceName, "allocation_max_netmask_length", resourceName, "allocation_max_netmask_length"),
					resource.TestCheckResourceAttrPair(dataSourceName, "allocation_min_netmask_length", resourceName, "allocation_min_netmask_length"),
					resource.TestCheckResourceAttrPair(dataSourceName, "allocation_resource_tags.%", resourceName, "allocation_resource_tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "auto_import", resourceName, "auto_import"),
					resource.TestCheckResourceAttrPair(dataSourceName, "aws_service", resourceName, "aws_service"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, "ipam_scope_id", resourceName, "ipam_scope_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ipam_scope_type", resourceName, "ipam_scope_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "locale", resourceName, "locale"),
					resource.TestCheckResourceAttrPair(dataSourceName, "pool_depth", resourceName, "pool_depth"),
					resource.TestCheckResourceAttrPair(dataSourceName, "publicly_advertisable", resourceName, "publicly_advertisable"),
					resource.TestCheckResourceAttrPair(dataSourceName, "source_ipam_pool_id", resourceName, "source_ipam_pool_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrState, resourceName, names.AttrState),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

var testAccIPAMPoolDataSourceConfig_optionsBasic = acctest.ConfigCompose(testAccIPAMPoolConfig_base, `
resource "aws_vpc_ipam_pool" "test" {
  address_family                    = "ipv4"
  ipam_scope_id                     = aws_vpc_ipam.test.private_default_scope_id
  auto_import                       = true
  allocation_default_netmask_length = 32
  allocation_max_netmask_length     = 32
  allocation_min_netmask_length     = 32
  allocation_resource_tags = {
    test = "1"
  }
  description = "test"
}

data "aws_vpc_ipam_pool" "test" {
  ipam_pool_id = aws_vpc_ipam_pool.test.id
}
`)
