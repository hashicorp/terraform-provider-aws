// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package quicksight_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
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

func TestAccQuickSightUserCustomPermission_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := "tfacctest" + sdkacctest.RandString(10)
	resourceName := "aws_quicksight_user_custom_permission.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserCustomPermissionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserCustomPermissionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserCustomPermissionExists(ctx, t, resourceName),
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
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("custom_permissions_name"), knownvalue.StringExact(rName+"-perm")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrNamespace), knownvalue.StringExact(tfquicksight.DefaultNamespace)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrUserName), knownvalue.StringExact(rName)),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccUserCustomPermissionImportStateID(resourceName),
				ImportStateVerifyIdentifierAttribute: "custom_permissions_name",
			},
		},
	})
}

func TestAccQuickSightUserCustomPermission_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := "tfacctest" + sdkacctest.RandString(10)
	resourceName := "aws_quicksight_user_custom_permission.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserCustomPermissionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserCustomPermissionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserCustomPermissionExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfquicksight.ResourceUserCustomPermission, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccQuickSightUserCustomPermission_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := "tfacctest" + sdkacctest.RandString(10)
	resourceName := "aws_quicksight_user_custom_permission.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserCustomPermissionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserCustomPermissionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserCustomPermissionExists(ctx, t, resourceName),
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
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("custom_permissions_name"), knownvalue.StringExact(rName+"-perm")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrNamespace), knownvalue.StringExact(tfquicksight.DefaultNamespace)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrUserName), knownvalue.StringExact(rName)),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccUserCustomPermissionImportStateID(resourceName),
				ImportStateVerifyIdentifierAttribute: "custom_permissions_name",
			},
			{
				Config: testAccUserCustomPermissionConfig_updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckUserCustomPermissionExists(ctx, t, resourceName),
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
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("custom_permissions_name"), knownvalue.StringExact(rName+"-perm2")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrNamespace), knownvalue.StringExact(tfquicksight.DefaultNamespace)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrUserName), knownvalue.StringExact(rName)),
				},
			},
		},
	})
}

func testAccCheckUserCustomPermissionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).QuickSightClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_quicksight_user_custom_permission" {
				continue
			}

			_, err := tfquicksight.FindUserCustomPermissionByThreePartKey(ctx, conn, rs.Primary.Attributes[names.AttrAWSAccountID], rs.Primary.Attributes[names.AttrNamespace], rs.Primary.Attributes[names.AttrUserName])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("QuickSight User Custom Permission (%s) still exists", rs.Primary.Attributes[names.AttrUserName])
		}

		return nil
	}
}

func testAccCheckUserCustomPermissionExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).QuickSightClient(ctx)

		_, err := tfquicksight.FindUserCustomPermissionByThreePartKey(ctx, conn, rs.Primary.Attributes[names.AttrAWSAccountID], rs.Primary.Attributes[names.AttrNamespace], rs.Primary.Attributes[names.AttrUserName])

		return err
	}
}

func testAccUserCustomPermissionImportStateID(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		return acctest.AttrsImportStateIdFunc(n, ",", names.AttrAWSAccountID, names.AttrNamespace, names.AttrUserName)(s)
	}
}

func testAccUserCustomPermissionConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccCustomPermissionsConfig_basic(rName+"-perm"), fmt.Sprintf(`
resource "aws_quicksight_user" "test" {
  user_name     = %[1]q
  email         = %[2]q
  identity_type = "QUICKSIGHT"
  user_role     = "READER"
}

resource "aws_quicksight_user_custom_permission" "test" {
  user_name               = aws_quicksight_user.test.user_name
  custom_permissions_name = aws_quicksight_custom_permissions.test.custom_permissions_name
}
`, rName, acctest.DefaultEmailAddress))
}

func testAccUserCustomPermissionConfig_updated(rName string) string {
	return acctest.ConfigCompose(testAccCustomPermissionsConfig_basic(rName+"-perm"), fmt.Sprintf(`
resource "aws_quicksight_custom_permissions" "test2" {
  custom_permissions_name = "%[1]s-perm2"

  capabilities {
    create_and_update_datasets     = "DENY"
    create_and_update_data_sources = "DENY"
    export_to_pdf                  = "DENY"
  }
}

resource "aws_quicksight_user" "test" {
  user_name     = %[1]q
  email         = %[2]q
  identity_type = "QUICKSIGHT"
  user_role     = "READER"
}

resource "aws_quicksight_user_custom_permission" "test" {
  user_name               = aws_quicksight_user.test.user_name
  custom_permissions_name = aws_quicksight_custom_permissions.test2.custom_permissions_name
}
`, rName, acctest.DefaultEmailAddress))
}
