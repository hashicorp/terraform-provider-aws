package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsCloudwatchLogResourcePolicy_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsCloudwatchLogResourcePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudwatchLogResourcePolicyConfig(acctest.RandString(5)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudwatchLogResourcePolicyExists("aws_cloudwatch_log_resource_policy.test"),
				),
			},
		},
	})
}

func testAccCheckAwsCloudwatchLogResourcePolicyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloudwatchlogsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudwatch_log_resource_policy" {
			continue
		}
		_, exists, err := lookupCloudWatchLogResourcePolicy(conn, rs.Primary.ID, nil)
		if err != nil {
			return nil
		}

		if exists {
			return fmt.Errorf("Cloudwatch Logs Resource Policy still exists: %s", rs.Primary.ID)
		}
	}

	return nil

}

func testAccCheckAwsCloudwatchLogResourcePolicyExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccAwsCloudwatchLogResourcePolicyConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = "tf-cwl-%s"
}

resource "aws_cloudwatch_log_resource_policy" "test" {
  policy_name = "tf-cwl-rp-%s"
  policy_document = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
  {
    "Sid": "Route53LogsToCloudWatchLogs",
    "Effect": "Allow",
    "Principal": {
      "Service": [
      "route53.amazonaws.com"
      ]
    },
    "Action": "logs:PutLogEvents",
    "Resource": "${aws_cloudwatch_log_group.test.arn}"
  }
  ]
}
POLICY
}
`, rName, rName)
}
