package ec2_test

import (
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccEC2CoIPPoolDataSource_filter(t *testing.T) {
	dataSourceName := "data.aws_ec2_coip_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck: acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccCoIPPoolDataSourceFilterDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "local_gateway_route_table_id", regexp.MustCompile(`^lgw-rtb-`)),
					resource.TestMatchResourceAttr(dataSourceName, "pool_id", regexp.MustCompile(`^ipv4pool-coip-`)),
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "pool_cidrs.#", "0"),
				),
			},
		},
	})
}

func TestAccEC2CoIPPoolDataSource_id(t *testing.T) {
	dataSourceName := "data.aws_ec2_coip_pool.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck: acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccCoIPPoolDataSourceIDDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "local_gateway_route_table_id", regexp.MustCompile(`^lgw-rtb-`)),
					resource.TestMatchResourceAttr(dataSourceName, "pool_id", regexp.MustCompile(`^ipv4pool-coip-`)),
					acctest.MatchResourceAttrRegionalARN(dataSourceName, "arn", "ec2", regexp.MustCompile(`coip-pool/ipv4pool-coip-.+$`)),
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "pool_cidrs.#", "0"),
				),
			},
		},
	})
}

func testAccCoIPPoolDataSourceFilterDataSourceConfig() string {
	return `
data "aws_ec2_coip_pools" "test" {}

data "aws_ec2_coip_pool" "test" {
  filter {
    name   = "coip-pool.pool-id"
    values = [tolist(data.aws_ec2_coip_pools.test.pool_ids)[0]]
  }
}
`
}

func testAccCoIPPoolDataSourceIDDataSourceConfig() string {
	return `
data "aws_ec2_coip_pools" "test" {}

data "aws_ec2_coip_pool" "test" {
  pool_id = tolist(data.aws_ec2_coip_pools.test.pool_ids)[0]
}
`
}
