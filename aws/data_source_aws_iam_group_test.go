package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccAWSDataSourceIAMGroup_basic(t *testing.T) {
	groupName := fmt.Sprintf("test-datasource-group-%d", acctest.RandInt())
	userName := fmt.Sprintf("test-datasource-user-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsIAMGroupConfig(groupName, userName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_iam_group.test", "group_id"),
					resource.TestCheckResourceAttr("data.aws_iam_group.test", "path", "/"),
					resource.TestCheckResourceAttr("data.aws_iam_group.test", "group_name", groupName),
					resource.TestMatchResourceAttr("data.aws_iam_group.test", "arn", regexp.MustCompile("^arn:aws:iam::[0-9]{12}:group/"+groupName)),
					resource.TestCheckResourceAttr("data.aws_iam_group.test", "users.#", "1"),
				),
			},
		},
	})
}

func testAccAwsIAMGroupConfig(groupname string, username string) string {
	return fmt.Sprintf(`
resource "aws_iam_group" "group" {
  name = "%s"
  path = "/"
}

resource "aws_iam_user" "user" {
	name = "%s"
}

resource "aws_iam_user_group_membership" "user_membership" {
	user = "${aws_iam_user.user.name}"

	groups = [
		"${aws_iam_group.group.name}",
	]
}

data "aws_iam_group" "test" {
  /*
  Getting the group_name from the aws_iam_user_group_membership
  enforce an implicit dependency which is needed for the test
  */
  group_name = "${element(tolist(aws_iam_user_group_membership.user_membership.groups), 0)}"
}
`, groupname, username)
}
