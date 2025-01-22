// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMGroupMembership_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var group iam.GetGroupOutput

	rString := sdkacctest.RandString(8)
	groupName := fmt.Sprintf("tf-acc-group-gm-basic-%s", rString)
	userName := fmt.Sprintf("tf-acc-user-gm-basic-%s", rString)
	userName2 := fmt.Sprintf("tf-acc-user-gm-basic-two-%s", rString)
	userName3 := fmt.Sprintf("tf-acc-user-gm-basic-three-%s", rString)
	membershipName := fmt.Sprintf("tf-acc-membership-gm-basic-%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupMembershipDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupMembershipConfig_member(groupName, userName, membershipName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupMembershipExists(ctx, "aws_iam_group_membership.team", &group),
					testAccCheckGroupMembershipAttributes(&group, groupName, []string{userName}),
				),
			},

			{
				Config: testAccGroupMembershipConfig_memberUpdate(groupName, userName, userName2, userName3, membershipName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupMembershipExists(ctx, "aws_iam_group_membership.team", &group),
					testAccCheckGroupMembershipAttributes(&group, groupName, []string{userName2, userName3}),
				),
			},

			{
				Config: testAccGroupMembershipConfig_memberUpdateDown(groupName, userName3, membershipName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupMembershipExists(ctx, "aws_iam_group_membership.team", &group),
					testAccCheckGroupMembershipAttributes(&group, groupName, []string{userName3}),
				),
			},
		},
	})
}

func TestAccIAMGroupMembership_paginatedUserList(t *testing.T) {
	ctx := acctest.Context(t)
	var group iam.GetGroupOutput

	rString := sdkacctest.RandString(8)
	groupName := fmt.Sprintf("tf-acc-group-gm-pul-%s", rString)
	membershipName := fmt.Sprintf("tf-acc-membership-gm-pul-%s", rString)
	userNamePrefix := fmt.Sprintf("tf-acc-user-gm-pul-%s-", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupMembershipDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupMembershipConfig_memberPaginatedUserList(groupName, membershipName, userNamePrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupMembershipExists(ctx, "aws_iam_group_membership.team", &group),
					resource.TestCheckResourceAttr(
						"aws_iam_group_membership.team", "users.#", "101"),
				),
			},
		},
	})
}

func testAccCheckGroupMembershipDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iam_group_membership" {
				continue
			}

			group := rs.Primary.Attributes["group"]

			_, err := conn.GetGroup(ctx, &iam.GetGroupInput{
				GroupName: aws.String(group),
			})

			if errs.IsA[*awstypes.NoSuchEntityException](err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("still exists")
		}

		return nil
	}
}

func testAccCheckGroupMembershipExists(ctx context.Context, n string, g *iam.GetGroupOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No User name is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)
		gn := rs.Primary.Attributes["group"]

		resp, err := conn.GetGroup(ctx, &iam.GetGroupInput{
			GroupName: aws.String(gn),
		})

		if err != nil {
			return fmt.Errorf("Error: Group (%s) not found", gn)
		}

		*g = *resp

		return nil
	}
}

func testAccCheckGroupMembershipAttributes(group *iam.GetGroupOutput, groupName string, users []string) resource.TestCheckFunc {
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

func testAccGroupMembershipConfig_member(groupName, userName, membershipName string) string {
	return fmt.Sprintf(`
resource "aws_iam_group" "group" {
  name = "%s"
}

resource "aws_iam_user" "user" {
  name = "%s"
}

resource "aws_iam_group_membership" "team" {
  name  = "%s"
  users = [aws_iam_user.user.name]
  group = aws_iam_group.group.name
}
`, groupName, userName, membershipName)
}

func testAccGroupMembershipConfig_memberUpdate(groupName, userName, userName2, userName3, membershipName string) string {
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
    aws_iam_user.user_two.name,
    aws_iam_user.user_three.name,
  ]

  group = aws_iam_group.group.name
}
`, groupName, userName, userName2, userName3, membershipName)
}

func testAccGroupMembershipConfig_memberUpdateDown(groupName, userName3, membershipName string) string {
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
    aws_iam_user.user_three.name,
  ]

  group = aws_iam_group.group.name
}
`, groupName, userName3, membershipName)
}

func testAccGroupMembershipConfig_memberPaginatedUserList(groupName, membershipName, userNamePrefix string) string {
	return fmt.Sprintf(`
resource "aws_iam_group" "group" {
  name = "%s"
}

resource "aws_iam_group_membership" "team" {
  name  = "%s"
  group = aws_iam_group.group.name

  # TODO: Switch back to simple list reference when test configurations are upgraded to 0.12 syntax
  users = [
    aws_iam_user.user[0].name,
    aws_iam_user.user[1].name,
    aws_iam_user.user[2].name,
    aws_iam_user.user[3].name,
    aws_iam_user.user[4].name,
    aws_iam_user.user[5].name,
    aws_iam_user.user[6].name,
    aws_iam_user.user[7].name,
    aws_iam_user.user[8].name,
    aws_iam_user.user[9].name,
    aws_iam_user.user[10].name,
    aws_iam_user.user[11].name,
    aws_iam_user.user[12].name,
    aws_iam_user.user[13].name,
    aws_iam_user.user[14].name,
    aws_iam_user.user[15].name,
    aws_iam_user.user[16].name,
    aws_iam_user.user[17].name,
    aws_iam_user.user[18].name,
    aws_iam_user.user[19].name,
    aws_iam_user.user[20].name,
    aws_iam_user.user[21].name,
    aws_iam_user.user[22].name,
    aws_iam_user.user[23].name,
    aws_iam_user.user[24].name,
    aws_iam_user.user[25].name,
    aws_iam_user.user[26].name,
    aws_iam_user.user[27].name,
    aws_iam_user.user[28].name,
    aws_iam_user.user[29].name,
    aws_iam_user.user[30].name,
    aws_iam_user.user[31].name,
    aws_iam_user.user[32].name,
    aws_iam_user.user[33].name,
    aws_iam_user.user[34].name,
    aws_iam_user.user[35].name,
    aws_iam_user.user[36].name,
    aws_iam_user.user[37].name,
    aws_iam_user.user[38].name,
    aws_iam_user.user[39].name,
    aws_iam_user.user[40].name,
    aws_iam_user.user[41].name,
    aws_iam_user.user[42].name,
    aws_iam_user.user[43].name,
    aws_iam_user.user[44].name,
    aws_iam_user.user[45].name,
    aws_iam_user.user[46].name,
    aws_iam_user.user[47].name,
    aws_iam_user.user[48].name,
    aws_iam_user.user[49].name,
    aws_iam_user.user[50].name,
    aws_iam_user.user[51].name,
    aws_iam_user.user[52].name,
    aws_iam_user.user[53].name,
    aws_iam_user.user[54].name,
    aws_iam_user.user[55].name,
    aws_iam_user.user[56].name,
    aws_iam_user.user[57].name,
    aws_iam_user.user[58].name,
    aws_iam_user.user[59].name,
    aws_iam_user.user[60].name,
    aws_iam_user.user[61].name,
    aws_iam_user.user[62].name,
    aws_iam_user.user[63].name,
    aws_iam_user.user[64].name,
    aws_iam_user.user[65].name,
    aws_iam_user.user[66].name,
    aws_iam_user.user[67].name,
    aws_iam_user.user[68].name,
    aws_iam_user.user[69].name,
    aws_iam_user.user[70].name,
    aws_iam_user.user[71].name,
    aws_iam_user.user[72].name,
    aws_iam_user.user[73].name,
    aws_iam_user.user[74].name,
    aws_iam_user.user[75].name,
    aws_iam_user.user[76].name,
    aws_iam_user.user[77].name,
    aws_iam_user.user[78].name,
    aws_iam_user.user[79].name,
    aws_iam_user.user[80].name,
    aws_iam_user.user[81].name,
    aws_iam_user.user[82].name,
    aws_iam_user.user[83].name,
    aws_iam_user.user[84].name,
    aws_iam_user.user[85].name,
    aws_iam_user.user[86].name,
    aws_iam_user.user[87].name,
    aws_iam_user.user[88].name,
    aws_iam_user.user[89].name,
    aws_iam_user.user[90].name,
    aws_iam_user.user[91].name,
    aws_iam_user.user[92].name,
    aws_iam_user.user[93].name,
    aws_iam_user.user[94].name,
    aws_iam_user.user[95].name,
    aws_iam_user.user[96].name,
    aws_iam_user.user[97].name,
    aws_iam_user.user[98].name,
    aws_iam_user.user[99].name,
    aws_iam_user.user[100].name,
  ]
}

resource "aws_iam_user" "user" {
  count = 101
  name  = format("%s%%d", count.index + 1)
}
`, groupName, membershipName, userNamePrefix)
}
