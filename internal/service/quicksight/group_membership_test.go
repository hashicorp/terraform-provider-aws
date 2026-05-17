// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package quicksight_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfquicksight "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccQuickSightGroupMembership_basic(t *testing.T) {
	ctx := acctest.Context(t)
	groupName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	memberName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_quicksight_group_membership.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		CheckDestroy:             testAccCheckGroupMembershipDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupMembershipConfig_basic(groupName, memberName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupMembershipExists(ctx, t, resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccQuickSightGroupMembership_withNamespace(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_quicksight_group_membership.test"
	groupResourceName := "aws_quicksight_group.test"
	userResourceName := "aws_quicksight_user.test"
	namespaceResourceName := "aws_quicksight_namespace.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		CheckDestroy:             testAccCheckGroupMembershipDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupMembershipConfig_withNamespace(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupMembershipExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrGroupName, groupResourceName, names.AttrGroupName),
					resource.TestCheckResourceAttrPair(resourceName, "member_name", userResourceName, names.AttrUserName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrNamespace, namespaceResourceName, names.AttrNamespace),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
func TestAccQuickSightGroupMembership_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	groupName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	memberName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_quicksight_group_membership.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupMembershipDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupMembershipConfig_basic(groupName, memberName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupMembershipExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfquicksight.ResourceGroupMembership(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckGroupMembershipDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).QuickSightClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_quicksight_group_membership" {
				continue
			}

			_, err := tfquicksight.FindGroupMembershipByFourPartKey(ctx, conn, rs.Primary.Attributes[names.AttrAWSAccountID], rs.Primary.Attributes[names.AttrNamespace], rs.Primary.Attributes[names.AttrGroupName], rs.Primary.Attributes["member_name"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("QuickSight Group Membership (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckGroupMembershipExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).QuickSightClient(ctx)

		_, err := tfquicksight.FindGroupMembershipByFourPartKey(ctx, conn, rs.Primary.Attributes[names.AttrAWSAccountID], rs.Primary.Attributes[names.AttrNamespace], rs.Primary.Attributes[names.AttrGroupName], rs.Primary.Attributes["member_name"])

		return err
	}
}

func testAccGroupMembershipConfig_basic(groupName string, memberName string) string {
	return acctest.ConfigCompose(
		testAccGroupConfig_basic(groupName),
		testAccUserConfig_basic(memberName),
		fmt.Sprintf(`
resource "aws_quicksight_group_membership" "test" {
  group_name  = aws_quicksight_group.default.group_name
  member_name = aws_quicksight_user.%s.user_name
}
`, memberName))
}

func testAccGroupMembershipConfig_withNamespace(rName string) string {
	return acctest.ConfigCompose(
		testAccNamespaceConfig_basic(rName),
		fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_quicksight_group" "test" {
  group_name = %[1]q
  namespace  = aws_quicksight_namespace.test.namespace
}

resource "aws_quicksight_user" "test" {
  aws_account_id = data.aws_caller_identity.current.account_id
  user_name      = %[1]q
  email          = %[2]q
  namespace      = aws_quicksight_namespace.test.namespace
  identity_type  = "QUICKSIGHT"
  user_role      = "READER"
}

resource "aws_quicksight_group_membership" "test" {
  group_name  = aws_quicksight_group.test.group_name
  member_name = aws_quicksight_user.test.user_name
  namespace   = aws_quicksight_namespace.test.namespace
}
`, rName, acctest.DefaultEmailAddress))
}
