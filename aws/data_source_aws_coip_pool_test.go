package aws

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccDataSourceAwsCoipPool_basic(t *testing.T) {
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
				Config: testAccDataSourceAwsLocalGatewayConfig(rPoolId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_coip_pool.by_id", "pool_id", rPoolId),
					resource.TestCheckResourceAttrSet("data.aws_coip_pool.by_id", "local_gateway_route_table_id"),
					testCheckResourceAttrGreaterThanValue("data.aws_coip_pool.by_id", "pool_cidrs.#", "0"),
				),
			},
		},
	})
}

func testAccDataSourceAwsLocalGatewayConfig(rPoolId string) string {
	return fmt.Sprintf(`
data "aws_coip_pool" "by_id" {
  pool_id = "%s"
}
`, rPoolId)
}
