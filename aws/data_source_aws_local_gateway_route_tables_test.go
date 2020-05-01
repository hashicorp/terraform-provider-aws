package aws

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccDataSourceAwsLocalGatewayRouteTables_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsLocalGatewayRouteTablesConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLocalGatewayRouteTableDataSourceExists("data.aws_local_gateway_route_tables.all"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsLocalGatewayRouteTables_filters(t *testing.T) {
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
				Config: testAccDataSourceAwsLocalGatewayRouteTablesConfig_filters(rLocalGatewayRouteTableId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLocalGatewayRouteTableDataSourceExists("data.aws_local_gateway_route_tables.selected"),
					testCheckResourceAttrGreaterThanValue("data.aws_local_gateway_route_tables.selected", "local_gateway_route_table_ids.#", "0"),
				),
			},
		},
	})
}

func testAccCheckAwsLocalGatewayRouteTableDataSourceExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("can't find aws_local_gateway_route_tables data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("aws_local_gateway_route_tables data source ID not set")
		}
		return nil
	}
}

const testAccDataSourceAwsLocalGatewayRouteTablesConfig = `data "aws_local_gateway_route_tables" "all" {}`

func testAccDataSourceAwsLocalGatewayRouteTablesConfig_filters(rLocalGatewayRouteTableId string) string {
	return fmt.Sprintf(`
data "aws_local_gateway_route_tables" "selected" {
  filter {
    name   = "local-gateway-route-table-id"
    values = ["%s"]
  }
}
`, rLocalGatewayRouteTableId)
}
