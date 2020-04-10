package aws

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
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
					testAccCheckAwsVpcsDataSourceExists("data.aws_local_gateways.all"),
				),
			},
		},
	})
}

const testAccDataSourceAwsLocalGatewaysConfig = `data "aws_local_gateways" "all" {}`
