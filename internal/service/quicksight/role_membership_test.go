// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfquicksight "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccRoleMembership_basic(t *testing.T) {
	ctx := acctest.Context(t)
	role := string(types.RoleReader)
	resourceName := "aws_quicksight_role_membership.test"

	memberName := acctest.SkipIfEnvVarNotSet(t, "TF_AWS_QUICKSIGHT_IDC_GROUP")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.QuickSightEndpointID)
			// Role Membership APIs are only available when QuickSight is configured with IAM Identity Center
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoleMembershipDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRoleMembershipConfig_basic(role, memberName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRoleMembershipExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "member_name", memberName),
					resource.TestCheckResourceAttr(resourceName, names.AttrRole, role),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccRoleMembershipImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "member_name",
			},
		},
	})
}

func testAccRoleMembership_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	role := string(types.RoleReader)
	resourceName := "aws_quicksight_role_membership.test"

	memberName := acctest.SkipIfEnvVarNotSet(t, "TF_AWS_QUICKSIGHT_IDC_GROUP")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.QuickSightEndpointID)
			// Role Membership APIs are only available when QuickSight is configured with IAM Identity Center
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoleMembershipDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRoleMembershipConfig_basic(role, memberName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRoleMembershipExists(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfquicksight.ResourceRoleMembership, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccRoleMembership_role(t *testing.T) {
	ctx := acctest.Context(t)
	role := string(types.RoleReader)
	roleUpdated := string(types.RoleAuthor)
	resourceName := "aws_quicksight_role_membership.test"

	memberName := acctest.SkipIfEnvVarNotSet(t, "TF_AWS_QUICKSIGHT_IDC_GROUP")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.QuickSightEndpointID)
			// Role Membership APIs are only available when QuickSight is configured with IAM Identity Center
			acctest.PreCheckSSOAdminInstances(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoleMembershipDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRoleMembershipConfig_basic(role, memberName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRoleMembershipExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "member_name", memberName),
					resource.TestCheckResourceAttr(resourceName, names.AttrRole, role),
				),
			},
			{
				Config: testAccRoleMembershipConfig_basic(roleUpdated, memberName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRoleMembershipExists(ctx, resourceName),
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

func testAccCheckRoleMembershipDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_quicksight_role_membership" {
				continue
			}

			accountID := rs.Primary.Attributes[names.AttrAWSAccountID]
			namespace := rs.Primary.Attributes[names.AttrNamespace]
			role := rs.Primary.Attributes[names.AttrRole]
			memberName := rs.Primary.Attributes["member_name"]

			err := tfquicksight.FindRoleMembershipByMultiPartKey(ctx, conn, accountID, namespace, types.Role(role), memberName)
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.QuickSight, create.ErrActionCheckingDestroyed, tfquicksight.ResNameRoleMembership, rs.Primary.ID, err)
			}

			return create.Error(names.QuickSight, create.ErrActionCheckingDestroyed, tfquicksight.ResNameRoleMembership, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckRoleMembershipExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.QuickSight, create.ErrActionCheckingExistence, tfquicksight.ResNameRoleMembership, name, errors.New("not found"))
		}

		accountID := rs.Primary.Attributes[names.AttrAWSAccountID]
		namespace := rs.Primary.Attributes[names.AttrNamespace]
		role := rs.Primary.Attributes[names.AttrRole]
		memberName := rs.Primary.Attributes["member_name"]
		if accountID == "" || namespace == "" || role == "" || memberName == "" {
			return create.Error(names.QuickSight, create.ErrActionCheckingExistence, tfquicksight.ResNameRoleMembership, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightClient(ctx)

		err := tfquicksight.FindRoleMembershipByMultiPartKey(ctx, conn, accountID, namespace, types.Role(role), memberName)
		if err != nil {
			return create.Error(names.QuickSight, create.ErrActionCheckingExistence, tfquicksight.ResNameRoleMembership, rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccRoleMembershipImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s,%s,%s,%s",
			rs.Primary.Attributes[names.AttrAWSAccountID],
			rs.Primary.Attributes[names.AttrNamespace],
			rs.Primary.Attributes[names.AttrRole],
			rs.Primary.Attributes["member_name"],
		), nil
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
