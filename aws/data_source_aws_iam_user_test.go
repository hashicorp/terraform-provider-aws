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
