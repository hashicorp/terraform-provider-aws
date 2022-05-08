package ec2_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccEC2OutpostsCoIPPoolsDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_ec2_coip_pools.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCoIPPoolsDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "pool_ids.#", "0"),
				),
			},
		},
	})
}

func TestAccEC2OutpostsCoIPPoolsDataSource_filter(t *testing.T) {
	dataSourceName := "data.aws_ec2_coip_pools.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOutpostsOutposts(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCoIPPoolsFilterDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "pool_ids.#", "1"),
				),
			},
		},
	})
}

func testAccCoIPPoolsDataSourceConfig() string {
	return `
data "aws_ec2_coip_pools" "test" {}
`
}

func testAccCoIPPoolsFilterDataSourceConfig() string {
	return `
data "aws_ec2_coip_pools" "all" {}

data "aws_ec2_coip_pools" "test" {
  filter {
    name   = "coip-pool.pool-id"
    values = [tolist(data.aws_ec2_coip_pools.all.pool_ids)[0]]
  }
}
`
}
