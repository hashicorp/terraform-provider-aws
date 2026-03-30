// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package quicksight_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfquicksight "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccRoleCustomPermission_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_quicksight_role_custom_permission.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoleCustomPermissionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRoleCustomPermissionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRoleCustomPermissionExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrAWSAccountID), tfknownvalue.AccountID()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("custom_permissions_name"), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrNamespace), knownvalue.StringExact(tfquicksight.DefaultNamespace)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRole), tfknownvalue.StringExact(awstypes.RoleReader)),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccRoleCustomPermissionImportStateID(resourceName),
				ImportStateVerifyIdentifierAttribute: "custom_permissions_name",
			},
		},
	})
}

func testAccRoleCustomPermission_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_quicksight_role_custom_permission.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoleCustomPermissionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRoleCustomPermissionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRoleCustomPermissionExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfquicksight.ResourceRoleCustomPermission, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccRoleCustomPermission_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_quicksight_role_custom_permission.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoleCustomPermissionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRoleCustomPermissionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRoleCustomPermissionExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrAWSAccountID), tfknownvalue.AccountID()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("custom_permissions_name"), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrNamespace), knownvalue.StringExact(tfquicksight.DefaultNamespace)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRole), tfknownvalue.StringExact(awstypes.RoleReader)),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccRoleCustomPermissionImportStateID(resourceName),
				ImportStateVerifyIdentifierAttribute: "custom_permissions_name",
			},
			{
				Config: testAccRoleCustomPermissionConfig_updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRoleCustomPermissionExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
					PostApplyPreRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrAWSAccountID), tfknownvalue.AccountID()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("custom_permissions_name"), knownvalue.StringExact(rName+"-2")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrNamespace), knownvalue.StringExact(tfquicksight.DefaultNamespace)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrRole), tfknownvalue.StringExact(awstypes.RoleReader)),
				},
			},
		},
	})
}

func testAccCheckRoleCustomPermissionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).QuickSightClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_quicksight_role_custom_permission" {
				continue
			}

			_, err := tfquicksight.FindRoleCustomPermissionByThreePartKey(ctx, conn, rs.Primary.Attributes[names.AttrAWSAccountID], rs.Primary.Attributes[names.AttrNamespace], awstypes.Role(rs.Primary.Attributes[names.AttrRole]))

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("QuickSight Role Custom Permission (%s) still exists", rs.Primary.Attributes[names.AttrRole])
		}

		return nil
	}
}

func testAccCheckRoleCustomPermissionExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).QuickSightClient(ctx)

		_, err := tfquicksight.FindRoleCustomPermissionByThreePartKey(ctx, conn, rs.Primary.Attributes[names.AttrAWSAccountID], rs.Primary.Attributes[names.AttrNamespace], awstypes.Role(rs.Primary.Attributes[names.AttrRole]))

		return err
	}
}

func testAccRoleCustomPermissionImportStateID(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		return acctest.AttrsImportStateIdFunc(n, ",", names.AttrAWSAccountID, names.AttrNamespace, names.AttrRole)(s)
	}
}

func testAccRoleCustomPermissionConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccCustomPermissionsConfig_basic(rName), `
resource "aws_quicksight_role_custom_permission" "test" {
  role                    = "READER"
  custom_permissions_name = aws_quicksight_custom_permissions.test.custom_permissions_name
}
`)
}

func testAccRoleCustomPermissionConfig_updated(rName string) string {
	return acctest.ConfigCompose(testAccCustomPermissionsConfig_basic(rName), fmt.Sprintf(`
resource "aws_quicksight_custom_permissions" "test2" {
  custom_permissions_name = "%[1]s-2"

  capabilities {
    create_and_update_datasets     = "DENY"
    create_and_update_data_sources = "DENY"
    export_to_pdf                  = "DENY"
  }
}

resource "aws_quicksight_role_custom_permission" "test" {
  role                    = "READER"
  custom_permissions_name = aws_quicksight_custom_permissions.test2.custom_permissions_name
}
`, rName))
}
