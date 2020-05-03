package aws

import (
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccDataSourceAwsEc2LocalGateway_basic(t *testing.T) {
	// Hide Outposts testing behind consistent environment variable
	outpostArn := os.Getenv("AWS_OUTPOST_ARN")
	if outpostArn == "" {
		t.Skip(
			"Environment variable AWS_OUTPOST_ARN is not set. " +
				"This environment variable must be set to the ARN of " +
				"a deployed Outpost to enable this test.")
	}

	dataSourceName := "data.aws_ec2_local_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsEc2LocalGatewayConfigId(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "id", regexp.MustCompile(`^lgw-`)),
					resource.TestCheckResourceAttr(dataSourceName, "outpost_arn", outpostArn),
					testAccCheckResourceAttrAccountID(dataSourceName, "owner_id"),
					resource.TestCheckResourceAttr(dataSourceName, "state", "available"),
				),
			},
		},
	})
}

func testAccDataSourceAwsEc2LocalGatewayConfigId() string {
	return `
data "aws_ec2_local_gateways" "test" {}

data "aws_ec2_local_gateway" "test" {
  id = tolist(data.aws_ec2_local_gateways.test.ids)[0]
}
`
}
