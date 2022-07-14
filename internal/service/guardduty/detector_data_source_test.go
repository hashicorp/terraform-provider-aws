package guardduty_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccDetectorDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(t) },
		ErrorCheck:                acctest.ErrorCheck(t, guardduty.EndpointsID),
		ProviderFactories:         acctest.ProviderFactories,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccDetectorDataSourceConfig_basicResource(),
				Check:  resource.ComposeTestCheckFunc(),
			},
			{
				Config: testAccDetectorDataSourceConfig_basicResource2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.aws_guardduty_detector.test", "id", "aws_guardduty_detector.test", "id"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "status", "ENABLED"),
					acctest.CheckResourceAttrGlobalARN("data.aws_guardduty_detector.test", "service_role_arn", "iam", "role/aws-service-role/guardduty.amazonaws.com/AWSServiceRoleForAmazonGuardDuty"),
					resource.TestCheckResourceAttrPair("data.aws_guardduty_detector.test", "finding_publishing_frequency", "aws_guardduty_detector.test", "finding_publishing_frequency"),
				),
			},
		},
	})
}

func testAccDetectorDataSource_ID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, guardduty.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDetectorDataSourceConfig_explicit(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.aws_guardduty_detector.test", "id", "aws_guardduty_detector.test", "id"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "status", "ENABLED"),
					acctest.CheckResourceAttrGlobalARN("data.aws_guardduty_detector.test", "service_role_arn", "iam", "role/aws-service-role/guardduty.amazonaws.com/AWSServiceRoleForAmazonGuardDuty"),
					resource.TestCheckResourceAttrPair("data.aws_guardduty_detector.test", "finding_publishing_frequency", "aws_guardduty_detector.test", "finding_publishing_frequency"),
				),
			},
		},
	})
}

func testAccDetectorDataSourceConfig_basicResource() string {
	return `
resource "aws_guardduty_detector" "test" {}
`
}

func testAccDetectorDataSourceConfig_basicResource2() string {
	return `
resource "aws_guardduty_detector" "test" {}

data "aws_guardduty_detector" "test" {}
`
}

func testAccDetectorDataSourceConfig_explicit() string {
	return `
resource "aws_guardduty_detector" "test" {}

data "aws_guardduty_detector" "test" {
  id = aws_guardduty_detector.test.id
}
`
}
