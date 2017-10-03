package aws

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSDataSourceIAMUser_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsDataSourceIAMUserConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_iam_user.test", "user_id"),
					resource.TestCheckResourceAttr("data.aws_iam_user.test", "path", "/"),
					resource.TestCheckResourceAttr("data.aws_iam_user.test", "user_name", "test-datasource-user"),
					resource.TestMatchResourceAttr("data.aws_iam_user.test", "arn", regexp.MustCompile("^arn:aws:iam::[0-9]{12}:user/test-datasource-user")),
				),
			}, {
				Config: testAccAwsDataSourceIAMUserUnknownConfig,
				// Should fail with an error that looks like this:
				//   Error refreshing: 1 error(s) occurred:
				//   * data.aws_iam_user.test: 1 error(s) occurred:
				//   * data.aws_iam_user.test: data.aws_iam_user.test: error getting user: NoSuchEntity: The user with name test-datasource-user-unknown cannot be found.
				//           status code: 404, request id: 16760dec-a88d-11e7-afb8-bbe194fb2bc8
				ExpectError: regexp.MustCompile("(?s)\\berror getting user\\b.*\\btest-datasource-user-unknown\\b.*\\b404\\b"),
			},
		},
	})
}

const testAccAwsDataSourceIAMUserConfig = `
resource "aws_iam_user" "user" {
	name = "test-datasource-user"
	path = "/"
}

data "aws_iam_user" "test" {
	  user_name = "${aws_iam_user.user.name}"
}
`

const testAccAwsDataSourceIAMUserUnknownConfig = `
data "aws_iam_user" "test" {
	  user_name = "test-datasource-user-unknown"
}
`
