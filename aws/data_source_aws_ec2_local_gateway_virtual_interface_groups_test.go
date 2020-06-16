package aws

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccDataSourceAwsEc2LocalGatewayVirtualInterfaceGroups_basic(t *testing.T) {
	// Hide Outposts testing behind consistent environment variable
	outpostArn := os.Getenv("AWS_OUTPOST_ARN")
	if outpostArn == "" {
		t.Skip(
			"Environment variable AWS_OUTPOST_ARN is not set. " +
				"This environment variable must be set to the ARN of " +
				"a deployed Outpost to enable this test.")
	}

	dataSourceName := "data.aws_ec2_local_gateway_virtual_interface_groups.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsEc2LocalGatewayVirtualInterfaceGroupsConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "local_gateway_virtual_interface_ids.#", "2"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsEc2LocalGatewayVirtualInterfaceGroups_Filter(t *testing.T) {
	// Hide Outposts testing behind consistent environment variable
	outpostArn := os.Getenv("AWS_OUTPOST_ARN")
	if outpostArn == "" {
		t.Skip(
			"Environment variable AWS_OUTPOST_ARN is not set. " +
				"This environment variable must be set to the ARN of " +
				"a deployed Outpost to enable this test.")
	}

	dataSourceName := "data.aws_ec2_local_gateway_virtual_interface_groups.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsEc2LocalGatewayVirtualInterfaceGroupsConfigFilter(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "local_gateway_virtual_interface_ids.#", "2"),
				),
			},
		},
	})
}

func testAccDataSourceAwsEc2LocalGatewayVirtualInterfaceGroupsConfig() string {
	return `
data "aws_ec2_local_gateway_virtual_interface_groups" "test" {}
`
}

func testAccDataSourceAwsEc2LocalGatewayVirtualInterfaceGroupsConfigFilter() string {
	return `
data "aws_ec2_local_gateways" "test" {}

data "aws_ec2_local_gateway_virtual_interface_groups" "test" {
  filter {
    name   = "local-gateway-id"
    values = [tolist(data.aws_ec2_local_gateways.test.ids)[0]]
  }
}
`
}
