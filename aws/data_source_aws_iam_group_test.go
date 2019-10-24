package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccAWSDataSourceIAMGroup_basic(t *testing.T) {
	groupName := fmt.Sprintf("test-datasource-user-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsIAMGroupConfig(groupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_iam_group.test", "group_id"),
					resource.TestCheckResourceAttr("data.aws_iam_group.test", "path", "/"),
					resource.TestCheckResourceAttr("data.aws_iam_group.test", "group_name", groupName),
					resource.TestMatchResourceAttr("data.aws_iam_group.test", "arn", regexp.MustCompile("^arn:aws:iam::[0-9]{12}:group/"+groupName)),
				),
			},
		},
	})
}

func TestAccAWSDataSourceIAMGroup_users(t *testing.T) {
	groupName := fmt.Sprintf("test-datasource-user-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsIAMGroupUsersConfig(groupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_iam_group.test", "group_id"),
					resource.TestCheckResourceAttr("data.aws_iam_group.test", "group_name", groupName),
					resource.TestCheckResourceAttr("data.aws_iam_group.test", "users.#", "2"),
				),
			},
		},
	})
}

func testAccAwsIAMGroupConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_iam_group" "group" {
  name = "%s"
  path = "/"
}

data "aws_iam_group" "test" {
  group_name = "${aws_iam_group.group.name}"
}
`, name)
}

func testAccAwsIAMGroupUsersConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "user_one" {
  name = "%s_user_one"
  path = "/"
}

resource "aws_iam_user" "user_two" {
  name = "%s_user_two"
  path = "/"
}

resource "aws_iam_group" "group" {
  name = "%s"
  path = "/"
}

resource "aws_iam_group_membership" "group" {
  name = "%s"
  group = "${aws_iam_group.group.name}"

  users = [
    "${aws_iam_user.user_one.name}",
    "${aws_iam_user.user_two.name}",
  ]
}

data "aws_iam_group" "test" {
  group_name = "${aws_iam_group_membership.group.name}"
}
`, name, name, name, name)
}
