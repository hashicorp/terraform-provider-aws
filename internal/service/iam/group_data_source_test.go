// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMGroupDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	groupName := fmt.Sprintf("test-datasource-user-%d", sdkacctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupDataSourceConfig_basic(groupName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_iam_group.test", "group_id"),
					resource.TestCheckResourceAttr("data.aws_iam_group.test", names.AttrPath, "/"),
					resource.TestCheckResourceAttr("data.aws_iam_group.test", names.AttrGroupName, groupName),
					acctest.CheckResourceAttrGlobalARN("data.aws_iam_group.test", names.AttrARN, "iam", fmt.Sprintf("group/%s", groupName)),
				),
			},
		},
	})
}

func TestAccIAMGroupDataSource_users(t *testing.T) {
	ctx := acctest.Context(t)
	groupName := fmt.Sprintf("test-datasource-group-%d", sdkacctest.RandInt())
	userName := fmt.Sprintf("test-datasource-user-%d", sdkacctest.RandInt())
	groupMemberShipName := fmt.Sprintf("test-datasource-group-membership-%d", sdkacctest.RandInt())
	userCount := 101

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupDataSourceConfig_user(groupName, userName, groupMemberShipName, userCount),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_iam_group.test", "group_id"),
					resource.TestCheckResourceAttr("data.aws_iam_group.test", names.AttrPath, "/"),
					resource.TestCheckResourceAttr("data.aws_iam_group.test", names.AttrGroupName, groupName),
					acctest.CheckResourceAttrGlobalARN("data.aws_iam_group.test", names.AttrARN, "iam", fmt.Sprintf("group/%s", groupName)),
					resource.TestCheckResourceAttr("data.aws_iam_group.test", "users.#", fmt.Sprint(userCount)),
					resource.TestCheckResourceAttrSet("data.aws_iam_group.test", "users.0.arn"),
					resource.TestCheckResourceAttrSet("data.aws_iam_group.test", "users.0.user_id"),
					resource.TestCheckResourceAttrSet("data.aws_iam_group.test", "users.0.user_name"),
					resource.TestCheckResourceAttrSet("data.aws_iam_group.test", "users.0.path"),
				),
			},
		},
	})
}

func testAccGroupDataSourceConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_iam_group" "group" {
  name = "%s"
  path = "/"
}

data "aws_iam_group" "test" {
  group_name = aws_iam_group.group.name
}
`, name)
}

func testAccGroupDataSourceConfig_user(groupName, userName, membershipName string, userCount int) string {
	return fmt.Sprintf(`
resource "aws_iam_group" "group" {
  name = "%s"
  path = "/"
}

resource "aws_iam_user" "user" {
  name  = "%s-${count.index}"
  count = %d
}

resource "aws_iam_group_membership" "team" {
  name  = "%s"
  users = aws_iam_user.user[*].name
  group = aws_iam_group.group.name
}

data "aws_iam_group" "test" {
  group_name = aws_iam_group_membership.team.group
}
`, groupName, userName, userCount, membershipName)
}
