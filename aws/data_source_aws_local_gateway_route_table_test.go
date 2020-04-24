package aws

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccDataSourceAwsLocalGatewayRouteTable_basic(t *testing.T) {
	rLocalGatewayRouteTableId := os.Getenv("AWS_LOCAL_GATEWAY_ROUTE_TABLE_ID")
	if rLocalGatewayRouteTableId == "" {
		t.Skip(
			"Environment variable AWS_LOCAL_GATEWAY_ROUTE_TABLE_ID is not set. " +
				"This environment variable must be set to the ID of " +
				"a deployed Local Gateway Route Table to enable this test.")
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsLocalGatewayRouteTableConfig(rLocalGatewayRouteTableId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_local_gateway_route_table.by_id", "local_gateway_route_table_id", rLocalGatewayRouteTableId),
					resource.TestCheckResourceAttrSet("data.aws_local_gateway_route_table.by_id", "local_gateway_id"),
					resource.TestCheckResourceAttrSet("data.aws_local_gateway_route_table.by_id", "outpost_arn"),
					resource.TestCheckResourceAttrSet("data.aws_local_gateway_route_table.by_id", "state"),
				),
			},
		},
	})
}

func testAccDataSourceAwsLocalGatewayRouteTableConfig(rLocalGatewayRouteTableId string) string {
	return fmt.Sprintf(`
data "aws_local_gateway_route_table" "by_id" {
  local_gateway_route_table_id = "%s"
}
`, rLocalGatewayRouteTableId)
}
