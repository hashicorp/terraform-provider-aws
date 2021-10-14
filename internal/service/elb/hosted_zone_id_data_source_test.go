package elb_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccAWSElbHostedZoneId_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, elb.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsElbHostedZoneIdConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_elb_hosted_zone_id.main", "id", elbHostedZoneIdPerRegionMap[acctest.Region()]),
				),
			},
			{
				Config: testAccCheckAwsElbHostedZoneIdExplicitRegionConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_elb_hosted_zone_id.regional", "id", "Z32O12XQLNTSW2"),
				),
			},
		},
	})
}

const testAccCheckAwsElbHostedZoneIdConfig = `
data "aws_elb_hosted_zone_id" "main" {}
`

//lintignore:AWSAT003
const testAccCheckAwsElbHostedZoneIdExplicitRegionConfig = `
data "aws_elb_hosted_zone_id" "regional" {
  region = "eu-west-1"
}
`
