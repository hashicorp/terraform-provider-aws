package aws

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccDataSourceAwsCoipPools_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsCoipPoolsConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCoipPoolDataSourceExists("data.aws_coip_pools.all"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsCoipPools_filters(t *testing.T) {
	rPoolId := os.Getenv("AWS_COIP_POOL_ID")
	if rPoolId == "" {
		t.Skip(
			"Environment variable AWS_COIP_POOL_ID is not set. " +
				"This environment variable must be set to the ID of " +
				"a deployed Coip Pool to enable this test.")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsCoipPoolsConfig_filters(rPoolId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCoipPoolDataSourceExists("data.aws_coip_pools.selected"),
					testCheckResourceAttrGreaterThanValue("data.aws_coip_pools.selected", "pool_ids.#", "0"),
				),
			},
		},
	})
}

func testAccCheckAwsCoipPoolDataSourceExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("can't find aws_coip_pools data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("aws_coip_pools data source ID not set")
		}
		return nil
	}
}

const testAccDataSourceAwsCoipPoolsConfig = `data "aws_coip_pools" "all" {}`

func testAccDataSourceAwsCoipPoolsConfig_filters(rPoolId string) string {
	return fmt.Sprintf(`
data "aws_coip_pools" "selected" {
  filter {
    name   = "coip-pool.pool-id"
    values = ["%s"]
  }
}
`, rPoolId)
}
