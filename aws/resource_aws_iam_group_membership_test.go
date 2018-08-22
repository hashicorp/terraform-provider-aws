package aws

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSGroupMembership_basic(t *testing.T) {
	var group iam.GetGroupOutput

	rString := acctest.RandString(8)
	groupName := fmt.Sprintf("tf-acc-group-gm-basic-%s", rString)
	userName := fmt.Sprintf("tf-acc-user-gm-basic-%s", rString)
	userName2 := fmt.Sprintf("tf-acc-user-gm-basic-two-%s", rString)
	userName3 := fmt.Sprintf("tf-acc-user-gm-basic-three-%s", rString)
	membershipName := fmt.Sprintf("tf-acc-membership-gm-basic-%s", rString)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGroupMembershipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGroupMemberConfig(groupName, userName, membershipName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGroupMembershipExists("aws_iam_group_membership.team", &group),
					testAccCheckAWSGroupMembershipAttributes(&group, groupName, []string{userName}),
				),
			},

			{
				Config: testAccAWSGroupMemberConfigUpdate(groupName, userName, userName2, userName3, membershipName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGroupMembershipExists("aws_iam_group_membership.team", &group),
					testAccCheckAWSGroupMembershipAttributes(&group, groupName, []string{userName2, userName3}),
				),
			},

			{
				Config: testAccAWSGroupMemberConfigUpdateDown(groupName, userName3, membershipName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGroupMembershipExists("aws_iam_group_membership.team", &group),
					testAccCheckAWSGroupMembershipAttributes(&group, groupName, []string{userName3}),
				),
			},
		},
	})
}

func TestAccAWSGroupMembership_paginatedUserList(t *testing.T) {
	var group iam.GetGroupOutput

	rString := acctest.RandString(8)
	groupName := fmt.Sprintf("tf-acc-group-gm-pul-%s", rString)
	membershipName := fmt.Sprintf("tf-acc-membership-gm-pul-%s", rString)
	userNamePrefix := fmt.Sprintf("tf-acc-user-gm-pul-%s-", rString)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGroupMembershipDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGroupMemberConfigPaginatedUserList(groupName, membershipName, userNamePrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSGroupMembershipExists("aws_iam_group_membership.team", &group),
					resource.TestCheckResourceAttr(
						"aws_iam_group_membership.team", "users.#", "101"),
				),
			},
		},
	})
}

func testAccCheckAWSGroupMembershipDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).iamconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iam_group_membership" {
			continue
		}

		group := rs.Primary.Attributes["group"]

		_, err := conn.GetGroup(&iam.GetGroupInput{
			GroupName: aws.String(group),
		})
		if err != nil {
			// Verify the error is what we want
			if ae, ok := err.(awserr.Error); ok && ae.Code() == "NoSuchEntity" {
				continue
			}
			return err
		}

		return fmt.Errorf("still exists")
	}

	return nil
}

func testAccCheckAWSGroupMembershipExists(n string, g *iam.GetGroupOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No User name is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).iamconn
		gn := rs.Primary.Attributes["group"]

		resp, err := conn.GetGroup(&iam.GetGroupInput{
			GroupName: aws.String(gn),
		})

		if err != nil {
			return fmt.Errorf("Error: Group (%s) not found", gn)
		}

		*g = *resp

		return nil
	}
}

func testAccCheckAWSGroupMembershipAttributes(group *iam.GetGroupOutput, groupName string, users []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !strings.Contains(*group.Group.GroupName, groupName) {
			return fmt.Errorf("Bad group membership: expected %s, got %s", groupName, *group.Group.GroupName)
		}

		uc := len(users)
		for _, u := range users {
			for _, gu := range group.Users {
				if u == *gu.UserName {
					uc--
				}
			}
		}

		if uc > 0 {
			return fmt.Errorf("Bad group membership count, expected (%d), but only (%d) found", len(users), uc)
		}
		return nil
	}
}

func testAccAWSGroupMemberConfig(groupName, userName, membershipName string) string {
	return fmt.Sprintf(`
resource "aws_iam_group" "group" {
	name = "%s"
}

resource "aws_iam_user" "user" {
	name = "%s"
}

resource "aws_iam_group_membership" "team" {
	name = "%s"
	users = ["${aws_iam_user.user.name}"]
	group = "${aws_iam_group.group.name}"
}
`, groupName, userName, membershipName)
}

func testAccAWSGroupMemberConfigUpdate(groupName, userName, userName2, userName3, membershipName string) string {
	return fmt.Sprintf(`
resource "aws_iam_group" "group" {
	name = "%s"
}

resource "aws_iam_user" "user" {
	name = "%s"
}

resource "aws_iam_user" "user_two" {
	name = "%s"
}

resource "aws_iam_user" "user_three" {
	name = "%s"
}

resource "aws_iam_group_membership" "team" {
	name = "%s"
	users = [
		"${aws_iam_user.user_two.name}",
		"${aws_iam_user.user_three.name}",
	]
	group = "${aws_iam_group.group.name}"
}
`, groupName, userName, userName2, userName3, membershipName)
}

func testAccAWSGroupMemberConfigUpdateDown(groupName, userName3, membershipName string) string {
	return fmt.Sprintf(`
resource "aws_iam_group" "group" {
	name = "%s"
}

resource "aws_iam_user" "user_three" {
	name = "%s"
}

resource "aws_iam_group_membership" "team" {
	name = "%s"
	users = [
		"${aws_iam_user.user_three.name}",
	]
	group = "${aws_iam_group.group.name}"
}
`, groupName, userName3, membershipName)
}

func testAccAWSGroupMemberConfigPaginatedUserList(groupName, membershipName, userNamePrefix string) string {
	return fmt.Sprintf(`
resource "aws_iam_group" "group" {
	name = "%s"
}

resource "aws_iam_group_membership" "team" {
	name = "%s"
	users = ["${aws_iam_user.user.*.name}"]
	group = "${aws_iam_group.group.name}"
}

resource "aws_iam_user" "user" {
	count = 101
	name = "${format("%s%%d", count.index + 1)}"
}
`, groupName, membershipName, userNamePrefix)
}
