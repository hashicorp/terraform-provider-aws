package aws

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"testing"
)

func TestAccDataSourceAwsLocalGateways_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsLocalGatewaysConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLocalGatewaysDataSourceExists("data.aws_local_gateways.all"),
				),
			},
		},
	})
}

func testAccCheckAwsLocalGatewaysDataSourceExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find aws_local_gateways data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("aws_local_gateways data source ID not set")
		}
		return nil
	}
}

const testAccDataSourceAwsLocalGatewaysConfig = `data "aws_local_gateways" "all" {}`
