package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccAWSGuarddutyDetectorDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsGuarddutyDetectorConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.aws_guardduty_detector.test", "id", "aws_guardduty_detector.test", "id"),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "enabled", "true"),
					resource.TestMatchResourceAttr("data.aws_guardduty_detector.test", "service_role_arn", regexp.MustCompile("^arn:aws:iam::[0-9]{12}:role/aws-service-role/guardduty.amazonaws.com/AWSServiceRoleForAmazonGuardDuty")),
					resource.TestCheckResourceAttr("data.aws_guardduty_detector.test", "finding_publishing_frequency", "SIX_HOURS"),
				),
			},
		},
	})
}

func testAccAwsGuarddutyDetectorConfig() string {
	return fmt.Sprintf(`
resource "aws_guardduty_detector" "test" {
}

data "aws_guardduty_detector" "test" {
	id = "${aws_guardduty_detector.test.id}"
}
`)
}
