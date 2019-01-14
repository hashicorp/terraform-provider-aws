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

func TestAccAWSDataSourceIAMGroup_member(t *testing.T) {
	groupName := fmt.Sprintf("test-datasource-group-%d", acctest.RandInt())
	userName := fmt.Sprintf("test-datasource-user-%d", acctest.RandInt())
	groupMemberShipName := fmt.Sprintf("test-datasource-group-membership-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsIAMGroupConfigMember(groupName, userName, groupMemberShipName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_iam_group.test", "group_id"),
					resource.TestCheckResourceAttr("data.aws_iam_group.test", "path", "/"),
					resource.TestCheckResourceAttr("data.aws_iam_group.test", "group_name", groupName),
					resource.TestMatchResourceAttr("data.aws_iam_group.test", "arn", regexp.MustCompile("^arn:aws:iam::[0-9]{12}:group/"+groupName)),
					resource.TestCheckResourceAttr("data.aws_iam_group.test", "members.#", "1"),
					resource.TestCheckResourceAttrPair("data.aws_iam_group.test", "members.0.arn", "aws_iam_user.user", "arn"),
					resource.TestCheckResourceAttrSet("data.aws_iam_group.test", "members.0.user_id"),
					resource.TestCheckResourceAttrPair("data.aws_iam_group.test", "members.0.user_name", "aws_iam_user.user", "name"),
					resource.TestCheckResourceAttrPair("data.aws_iam_group.test", "members.0.path", "aws_iam_user.user", "path"),
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

func testAccAwsIAMGroupConfigMember(groupName, userName, membershipName string) string {
	return fmt.Sprintf(`
resource "aws_iam_group" "group" {
	name = "%s"
	path = "/"
}

resource "aws_iam_user" "user" {
	name = "%s"
}

resource "aws_iam_group_membership" "team" {
	name = "%s"
	users = ["${aws_iam_user.user.name}"]
	group = "${aws_iam_group.group.name}"
}

data "aws_iam_group" "test" {
	group_name = "${aws_iam_group_membership.team.group}"	
}
`, groupName, userName, membershipName)
}
