// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIPAMPoolDataSource_basic(t *testing.T) { // nosemgrep:ci.vpc-in-test-name
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

func TestAccIPAMPoolDataSource_sourceResourceVPC(t *testing.T) { // nosemgrep:ci.vpc-in-test-name
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_vpc_ipam_pool.vpc"
	dataSourceName := "data.aws_vpc_ipam_pool.vpc"
	vpcResourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMPoolDataSourceConfig_sourceResourceVPC(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "source_resource.#", resourceName, "source_resource.#"),
					resource.TestCheckResourceAttr(dataSourceName, "source_resource.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "source_resource.0.resource_id", vpcResourceName, names.AttrID),
					resource.TestCheckResourceAttrSet(dataSourceName, "source_resource.0.resource_owner"),
					resource.TestCheckResourceAttrSet(dataSourceName, "source_resource.0.resource_region"),
					resource.TestCheckResourceAttr(dataSourceName, "source_resource.0.resource_type", "vpc"),
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

func testAccIPAMPoolDataSourceConfig_sourceResourceVPC(rName string) string {
	return acctest.ConfigCompose(testAccIPAMPoolConfig_base, fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_vpc_ipam_pool" "test" {
  address_family = "ipv4"
  ipam_scope_id  = aws_vpc_ipam.test.private_default_scope_id
  locale         = data.aws_region.current.name

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_ipam_pool_cidr" "test" {
  ipam_pool_id = aws_vpc_ipam_pool.test.id
  cidr         = "10.0.0.0/16"
}

resource "aws_vpc" "test" {
  ipv4_ipam_pool_id   = aws_vpc_ipam_pool.test.id
  ipv4_netmask_length = 24

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_vpc_ipam_pool_cidr.test]
}

resource "aws_vpc_ipam_pool" "vpc" {
  address_family      = "ipv4"
  ipam_scope_id       = aws_vpc_ipam.test.private_default_scope_id
  locale              = data.aws_region.current.name
  source_ipam_pool_id = aws_vpc_ipam_pool.test.id

  source_resource {
    resource_id     = aws_vpc.test.id
    resource_owner  = data.aws_caller_identity.current.account_id
    resource_region = data.aws_region.current.name
    resource_type   = "vpc"
  }

  tags = {
    Name = %[1]q
  }
}

data "aws_vpc_ipam_pool" "vpc" {
  ipam_pool_id = aws_vpc_ipam_pool.vpc.id
}
`, rName))
}
