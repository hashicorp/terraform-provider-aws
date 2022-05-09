package elb_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfelb "github.com/hashicorp/terraform-provider-aws/internal/service/elb"
)

func TestAccELBHostedZoneIDDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, elb.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAWSElbHostedZoneIdConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_elb_hosted_zone_id.main", "id", tfelb.HostedZoneIdPerRegionMap[acctest.Region()]),
				),
			},
			{
				Config: testAccCheckAWSElbHostedZoneIdExplicitRegionConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_elb_hosted_zone_id.regional", "id", "Z32O12XQLNTSW2"),
				),
			},
		},
	})
}

const testAccCheckAWSElbHostedZoneIdConfig = `
data "aws_elb_hosted_zone_id" "main" {}
`

//lintignore:AWSAT003
const testAccCheckAWSElbHostedZoneIdExplicitRegionConfig = `
data "aws_elb_hosted_zone_id" "regional" {
  region = "eu-west-1"
}
`
