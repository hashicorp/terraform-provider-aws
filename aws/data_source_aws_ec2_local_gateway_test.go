package aws

import (
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/atest"
)

func TestAccDataSourceAwsEc2LocalGateway_basic(t *testing.T) {
	dataSourceName := "data.aws_ec2_local_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { atest.PreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		ErrorCheck: atest.ErrorCheck(t, ec2.EndpointsID),
		Providers:  atest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsEc2LocalGatewayConfigId(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "id", regexp.MustCompile(`^lgw-`)),
					atest.MatchAttrRegionalARN(dataSourceName, "outpost_arn", "outposts", regexp.MustCompile(`outpost/op-.+`)),
					atest.CheckAttrAccountID(dataSourceName, "owner_id"),
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
