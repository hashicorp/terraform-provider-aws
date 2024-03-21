// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIPAMPoolCIDRAllocationsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_vpc_ipam_pool_cidr_allocations.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: TestAccIPAMPoolCIDRAllocationsDataSourceConfig_basicOneAllocations,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ipam_pool_allocations.#", "1"),
				),
			},
			{
				Config: TestAccIPAMPoolCIDRAllocationsDataSourceConfig_basicTwoAllocations,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ipam_pool_allocations.#", "2"),
				),
			},
			{
				Config: TestAccIPAMPoolCIDRAllocationsDataSourceConfig_basicTwoAllocationsFiltered,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ipam_pool_allocations.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "ipam_pool_allocations.0.description", "testtwo"),
				),
			},
			{
				Config: TestAccIPAMPoolCIDRAllocationsDataSourceConfig_basicTwoAllocationsByID,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ipam_pool_allocations.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "ipam_pool_allocations.0.description", "testtwo"),
				),
			},
		},
	})
}

var TestAccIPAMPoolCIDRAllocationsDataSourceConfig_basicOneAllocations = acctest.ConfigCompose(testAccIPAMPoolConfig_basic, `
resource "aws_vpc_ipam_pool_cidr" "test" {
	ipam_pool_id = aws_vpc_ipam_pool.test.id
	cidr         = "172.2.0.0/16"
}

resource "aws_vpc_ipam_pool_cidr_allocation" "test" {
	  ipam_pool_id      = aws_vpc_ipam_pool.test.id
	  netmask_length    = 28
	  description       = "test"
	  depends_on = [
		aws_vpc_ipam_pool_cidr.test
	  ]
}

data "aws_vpc_ipam_pool_cidr_allocations" "test" {
  ipam_pool_id = aws_vpc_ipam_pool.test.id

  depends_on = [
    aws_vpc_ipam_pool_cidr_allocation.test
  ]
}
`)

var TestAccIPAMPoolCIDRAllocationsDataSourceConfig_basicTwoAllocations = acctest.ConfigCompose(testAccIPAMPoolConfig_basic, `

resource "aws_vpc_ipam_pool_cidr" "test" {
	ipam_pool_id = aws_vpc_ipam_pool.test.id
	cidr         = "172.2.0.0/16"
}

resource "aws_vpc_ipam_pool_cidr_allocation" "test" {
	ipam_pool_id      = aws_vpc_ipam_pool.test.id
	netmask_length    = 28
	description       = "test"
	depends_on = [
		aws_vpc_ipam_pool_cidr.test
	  ]
}

resource "aws_vpc_ipam_pool_cidr_allocation" "testtwo" {
	ipam_pool_id      = aws_vpc_ipam_pool.test.id
	netmask_length    = 28
	description       = "testtwo"
	depends_on = [
		aws_vpc_ipam_pool_cidr.test
	  ]
}

data "aws_vpc_ipam_pool_cidr_allocations" "test" {
  ipam_pool_id = aws_vpc_ipam_pool.test.id

  depends_on = [
    aws_vpc_ipam_pool_cidr_allocation.test,
    aws_vpc_ipam_pool_cidr_allocation.testtwo,
  ]
}
`)

var TestAccIPAMPoolCIDRAllocationsDataSourceConfig_basicTwoAllocationsFiltered = acctest.ConfigCompose(testAccIPAMPoolConfig_basic, `
resource "aws_vpc_ipam_pool_cidr" "test" {
	ipam_pool_id = aws_vpc_ipam_pool.test.id
	cidr         = "172.2.0.0/16"
}

resource "aws_vpc_ipam_pool_cidr" "testtwo" {
	ipam_pool_id = aws_vpc_ipam_pool.test.id
	cidr         = "10.2.0.0/16"
}

resource "aws_vpc_ipam_pool_cidr_allocation" "test" {
	ipam_pool_id = aws_vpc_ipam_pool.test.id
	netmask_length    = 28
	description       = "test"
	depends_on = [
		aws_vpc_ipam_pool_cidr.test
	  ]
}

resource "aws_vpc_ipam_pool_cidr_allocation" "testtwo" {
	ipam_pool_id = aws_vpc_ipam_pool.test.id
	netmask_length    = 28
	description       = "testtwo"
	depends_on = [
		aws_vpc_ipam_pool_cidr.testtwo
	  ]
}

data "aws_vpc_ipam_pool_cidr_allocations" "test" {
  ipam_pool_id = aws_vpc_ipam_pool.test.id

  filter {
    name   = "cidr"
    values = [aws_vpc_ipam_pool_cidr_allocation.testtwo.cidr]
  }

  depends_on = [
    aws_vpc_ipam_pool_cidr_allocation.test,
    aws_vpc_ipam_pool_cidr_allocation.testtwo,
  ]
}
`)

var TestAccIPAMPoolCIDRAllocationsDataSourceConfig_basicTwoAllocationsByID = acctest.ConfigCompose(testAccIPAMPoolConfig_basic, `
resource "aws_vpc_ipam_pool_cidr" "test" {
	ipam_pool_id = aws_vpc_ipam_pool.test.id
	cidr         = "172.2.0.0/16"
}

resource "aws_vpc_ipam_pool_cidr" "testtwo" {
	ipam_pool_id = aws_vpc_ipam_pool.test.id
	cidr         = "10.2.0.0/16"
}

resource "aws_vpc_ipam_pool_cidr_allocation" "test" {
	ipam_pool_id = aws_vpc_ipam_pool.test.id
	netmask_length    = 28
	description       = "test"
	depends_on = [
		aws_vpc_ipam_pool_cidr.test
	  ]
}

resource "aws_vpc_ipam_pool_cidr_allocation" "testtwo" {
	ipam_pool_id = aws_vpc_ipam_pool.test.id
	netmask_length    = 28
	description       = "testtwo"
	depends_on = [
		aws_vpc_ipam_pool_cidr.testtwo
	  ]
}

data "aws_vpc_ipam_pool_cidr_allocations" "test" {
  ipam_pool_id = aws_vpc_ipam_pool.test.id
  ipam_pool_allocation_id = element(split("_", aws_vpc_ipam_pool_cidr_allocation.testtwo.id), 0)

  depends_on = [
    aws_vpc_ipam_pool_cidr_allocation.test,
    aws_vpc_ipam_pool_cidr_allocation.testtwo,
  ]
}
`)
