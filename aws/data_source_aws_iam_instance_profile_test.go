package aws

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSDataSourceIAMInstanceProfile_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceAwsIamInstanceProfileConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_iam_instance_profile.test", "role_id"),
					resource.TestCheckResourceAttr("data.aws_iam_instance_profile.test", "path", "/testpath/"),
					resource.TestMatchResourceAttr("data.aws_iam_instance_profile.test", "arn", regexp.MustCompile("^arn:aws:iam::[0-9]{12}:instance-profile/testpath/test-instance-profile$")),
				),
			},
		},
	})
}

const testAccDatasourceAwsIamInstanceProfileConfig = `
provider "aws" {
	region = "us-east-1"
}

resource "aws_iam_role" "test" {
	name = "test-role"
	assume_role_policy = "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"Service\":[\"ec2.amazonaws.com\"]},\"Action\":[\"sts:AssumeRole\"]}]}"
}

resource "aws_iam_instance_profile" "test" {
	name = "test-instance-profile"
	role = "${aws_iam_role.test.name}"
        path = "/testpath/"
}

data "aws_iam_instance_profile" "test" {
	  name = "${aws_iam_instance_profile.test.name}"
}
`
