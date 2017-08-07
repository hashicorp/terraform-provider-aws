package aws

import (
	"regexp"
	"testing"

	"fmt"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSDataSourceIAMRole_basic(t *testing.T) {
	rName := acctest.RandString(10)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsIAMRoleConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_iam_role.test", "unique_id"),
					resource.TestCheckResourceAttrSet("data.aws_iam_role.test", "assume_role_policy"),
					resource.TestCheckResourceAttr("data.aws_iam_role.test", "path", "/testpath/"),
					resource.TestCheckResourceAttr("data.aws_iam_role.test", "name", fmt.Sprintf("test-role-%s", rName)),
					resource.TestCheckResourceAttrSet("data.aws_iam_role.test", "create_date"),
					resource.TestMatchResourceAttr("data.aws_iam_role.test", "arn", regexp.MustCompile(`^arn:[\w-]+:([a-zA-Z0-9\-])+:([a-z]{2}-(gov-)?[a-z]+-\d{1})?:(\d{12})?:(.*)$`)),
				),
			},
		},
	})
}

func testAccAwsIAMRoleConfig(rName string) string {
	return fmt.Sprintf(`
provider "aws" {
  region = "us-east-1"
}

resource "aws_iam_role" "test_role" {
  name = "test-role-%s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF

  path = "/testpath/"
}

data "aws_iam_role" "test" {
  name = "${aws_iam_role.test_role.name}"
}
`, rName)
}
