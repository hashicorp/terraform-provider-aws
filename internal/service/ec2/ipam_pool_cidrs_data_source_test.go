// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIPAMPoolCIDRsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_vpc_ipam_pool_cidrs.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMPoolCIDRsDataSourceConfig_basicOneCIDRs,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ipam_pool_cidrs.#", acctest.Ct1),
				),
			},
			{
				Config: testAccIPAMPoolCIDRsDataSourceConfig_basicTwoCIDRs,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ipam_pool_cidrs.#", acctest.Ct2),
				),
			},
			{
				Config: testAccIPAMPoolCIDRsDataSourceConfig_basicTwoCIDRsFiltered,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ipam_pool_cidrs.#", acctest.Ct1),
				),
			},
		},
	})
}

var testAccIPAMPoolCIDRsDataSourceConfig_basicOneCIDRs = acctest.ConfigCompose(testAccIPAMPoolConfig_basic, `
resource "aws_vpc_ipam_pool_cidr" "test" {
  ipam_pool_id = aws_vpc_ipam_pool.test.id
  cidr         = "172.2.0.0/16"
}

data "aws_vpc_ipam_pool_cidrs" "test" {
  ipam_pool_id = aws_vpc_ipam_pool.test.id

  depends_on = [
    aws_vpc_ipam_pool_cidr.test
  ]
}
`)

var testAccIPAMPoolCIDRsDataSourceConfig_basicTwoCIDRs = acctest.ConfigCompose(testAccIPAMPoolConfig_basic, `
resource "aws_vpc_ipam_pool_cidr" "test" {
  ipam_pool_id = aws_vpc_ipam_pool.test.id
  cidr         = "172.2.0.0/16"
}

resource "aws_vpc_ipam_pool_cidr" "testtwo" {
  ipam_pool_id = aws_vpc_ipam_pool.test.id
  cidr         = "10.2.0.0/16"
}

data "aws_vpc_ipam_pool_cidrs" "test" {
  ipam_pool_id = aws_vpc_ipam_pool.test.id

  depends_on = [
    aws_vpc_ipam_pool_cidr.test,
    aws_vpc_ipam_pool_cidr.testtwo,
  ]
}
`)

var testAccIPAMPoolCIDRsDataSourceConfig_basicTwoCIDRsFiltered = acctest.ConfigCompose(testAccIPAMPoolConfig_basic, `
resource "aws_vpc_ipam_pool_cidr" "test" {
  ipam_pool_id = aws_vpc_ipam_pool.test.id
  cidr         = "172.2.0.0/16"
}

resource "aws_vpc_ipam_pool_cidr" "testtwo" {
  ipam_pool_id = aws_vpc_ipam_pool.test.id
  cidr         = "10.2.0.0/16"
}

data "aws_vpc_ipam_pool_cidrs" "test" {
  ipam_pool_id = aws_vpc_ipam_pool.test.id

  filter {
    name   = "cidr"
    values = ["10.*"]
  }

  depends_on = [
    aws_vpc_ipam_pool_cidr.test,
    aws_vpc_ipam_pool_cidr.testtwo,
  ]
}
`)
