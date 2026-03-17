// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package quicksight_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfquicksight "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccRoleMembership_basic(t *testing.T) {
	ctx := acctest.Context(t)
	memberName := acctest.SkipIfEnvVarNotSet(t, "TF_AWS_QUICKSIGHT_IDC_GROUP")
	role := string(awstypes.RoleReader)
	resourceName := "aws_quicksight_role_membership.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.QuickSightEndpointID)
			// Role Membership APIs are only available when QuickSight is configured with IAM Identity Center
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoleMembershipDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRoleMembershipConfig_basic(role, memberName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRoleMembershipExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "member_name", memberName),
					resource.TestCheckResourceAttr(resourceName, names.AttrRole, role),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccRoleMembershipImportStateIDFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "member_name",
			},
		},
	})
}

func testAccRoleMembership_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	memberName := acctest.SkipIfEnvVarNotSet(t, "TF_AWS_QUICKSIGHT_IDC_GROUP")
	role := string(awstypes.RoleReader)
	resourceName := "aws_quicksight_role_membership.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.QuickSightEndpointID)
			// Role Membership APIs are only available when QuickSight is configured with IAM Identity Center
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoleMembershipDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRoleMembershipConfig_basic(role, memberName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRoleMembershipExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfquicksight.ResourceRoleMembership, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccRoleMembership_role(t *testing.T) {
	ctx := acctest.Context(t)
	memberName := acctest.SkipIfEnvVarNotSet(t, "TF_AWS_QUICKSIGHT_IDC_GROUP")
	role := string(awstypes.RoleReader)
	roleUpdated := string(awstypes.RoleAuthor)
	resourceName := "aws_quicksight_role_membership.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.QuickSightEndpointID)
			// Role Membership APIs are only available when QuickSight is configured with IAM Identity Center
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoleMembershipDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRoleMembershipConfig_basic(role, memberName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRoleMembershipExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "member_name", memberName),
					resource.TestCheckResourceAttr(resourceName, names.AttrRole, role),
				),
			},
			{
				Config: testAccRoleMembershipConfig_basic(roleUpdated, memberName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRoleMembershipExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "member_name", memberName),
					resource.TestCheckResourceAttr(resourceName, names.AttrRole, roleUpdated),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
			},
		},
	})
}

func testAccCheckRoleMembershipDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).QuickSightClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_quicksight_role_membership" {
				continue
			}

			err := tfquicksight.FindRoleMembershipByFourPartKey(ctx, conn, rs.Primary.Attributes[names.AttrAWSAccountID], rs.Primary.Attributes[names.AttrNamespace], awstypes.Role(rs.Primary.Attributes[names.AttrRole]), rs.Primary.Attributes["member_name"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("QuickSight Role Membership (%s) still exists", rs.Primary.Attributes[names.AttrRole])
		}

		return nil
	}
}

func testAccCheckRoleMembershipExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).QuickSightClient(ctx)

		err := tfquicksight.FindRoleMembershipByFourPartKey(ctx, conn, rs.Primary.Attributes[names.AttrAWSAccountID], rs.Primary.Attributes[names.AttrNamespace], awstypes.Role(rs.Primary.Attributes[names.AttrRole]), rs.Primary.Attributes["member_name"])

		return err
	}
}

func testAccRoleMembershipImportStateIDFunc(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		return acctest.AttrsImportStateIdFunc(n, ",", names.AttrAWSAccountID, names.AttrNamespace, names.AttrRole, "member_name")(s)
	}
}

func testAccRoleMembershipConfig_basic(role, memberName string) string {
	return fmt.Sprintf(`
resource "aws_quicksight_role_membership" "test" {
  role        = %[1]q
  member_name = %[2]q
}
`, role, memberName)
}
