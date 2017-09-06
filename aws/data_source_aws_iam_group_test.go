package aws

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSDataSourceIAMGroup_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsIAMGroupConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_iam_group.test", "group_id"),
					resource.TestCheckResourceAttr("data.aws_iam_group.test", "path", "/"),
					resource.TestCheckResourceAttr("data.aws_iam_group.test", "group_name", "test-datasource-group"),
					resource.TestMatchResourceAttr("data.aws_iam_group.test", "arn", regexp.MustCompile("^arn:aws:iam::[0-9]{12}:group/test-datasource-group")),
				),
			},
		},
	})
}

const testAccAwsIAMGroupConfig = `
resource "aws_iam_group" "group" {
	name = "test-datasource-group"
	path = "/"
}

data "aws_iam_group" "test" {
	  group_name = "${aws_iam_group.group.name}"
}
`
