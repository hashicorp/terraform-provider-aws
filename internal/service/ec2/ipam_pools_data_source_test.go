package ec2_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccIPAMPoolsDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_vpc_ipam_pools.test"
	dataSourceNameTwo := "data.aws_vpc_ipam_pools.testtwo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccIPAMPreCheck(t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMPoolsDataSourceConfig_Basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ipam_pools.#", "1"),
				),
			},
			{
				Config: testAccIPAMPoolsDataSourceConfig_BasicTwoPools,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ipam_pools.#", "3"),
					resource.TestCheckResourceAttr(dataSourceNameTwo, "ipam_pools.#", "1"),
					resource.TestCheckResourceAttr(dataSourceNameTwo, "ipam_pools.0.allocation_resource_tags.test", "3"),
					resource.TestCheckResourceAttr(dataSourceNameTwo, "ipam_pools.0.allocation_default_netmask_length", "32"),
					resource.TestCheckResourceAttr(dataSourceNameTwo, "ipam_pools.0.allocation_max_netmask_length", "32"),
					resource.TestCheckResourceAttr(dataSourceNameTwo, "ipam_pools.0.allocation_min_netmask_length", "32"),
					resource.TestCheckResourceAttr(dataSourceNameTwo, "ipam_pools.0.description", "testthree"),

					resource.TestCheckResourceAttr(dataSourceNameTwo, "ipam_pools.0.tags.tagtest", "3"),
				),
			},
		},
	})
}

var testAccIPAMPoolsDataSourceConfig_Basic = acctest.ConfigCompose(testAccIPAMPoolConfig_base, `
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

data "aws_vpc_ipam_pools" "test" {
  depends_on = [
    aws_vpc_ipam_pool.test
  ]
}
`)

var testAccIPAMPoolsDataSourceConfig_BasicTwoPools = acctest.ConfigCompose(testAccIPAMPoolConfig_base, `
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

resource "aws_vpc_ipam_pool" "testtwo" {
  address_family = "ipv4"
  ipam_scope_id  = aws_vpc_ipam.test.private_default_scope_id
  allocation_resource_tags = {
    test = "2"
  }
  description = "testtwo"
}

resource "aws_vpc_ipam_pool" "testthree" {
  address_family                    = "ipv4"
  ipam_scope_id                     = aws_vpc_ipam.test.private_default_scope_id
  allocation_default_netmask_length = 32
  allocation_max_netmask_length     = 32
  allocation_min_netmask_length     = 32
  allocation_resource_tags = {
    test = "3"
  }
  description = "testthree"
  tags = {
    tagtest = 3
  }
}

data "aws_vpc_ipam_pools" "test" {
  depends_on = [
    aws_vpc_ipam_pool.test,
    aws_vpc_ipam_pool.testtwo,
    aws_vpc_ipam_pool.testthree
  ]
}

data "aws_vpc_ipam_pools" "testtwo" {
  filter {
    name   = "description"
    values = ["*three*"]
  }

  depends_on = [
    aws_vpc_ipam_pool.test,
    aws_vpc_ipam_pool.testtwo,
    aws_vpc_ipam_pool.testthree
  ]
}
`)
