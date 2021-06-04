package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/atest"
)

func TestAccDataSourceAwsEc2LocalGateways_basic(t *testing.T) {
	dataSourceName := "data.aws_ec2_local_gateways.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { atest.PreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		ErrorCheck: atest.ErrorCheck(t, ec2.EndpointsID),
		Providers:  atest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsEc2LocalGatewaysConfig(),
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceAttrGreaterThanValue(dataSourceName, "ids.#", "0"),
				),
			},
		},
	})
}

func testAccDataSourceAwsEc2LocalGatewaysConfig() string {
	return `
data "aws_ec2_local_gateways" "test" {}
`
}
