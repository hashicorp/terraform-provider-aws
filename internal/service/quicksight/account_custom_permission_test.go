// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package quicksight_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/smarterr"
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

func testAccAccountCustomPermission_basic(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_quicksight_account_custom_permission.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountCustomPermissionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountCustomPermissionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccountCustomPermissionExists(ctx, t, resourceName),
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
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrAWSAccountID),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrAWSAccountID,
			},
		},
	})
}

func testAccAccountCustomPermission_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_quicksight_account_custom_permission.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountCustomPermissionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountCustomPermissionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccountCustomPermissionExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfquicksight.ResourceAccountCustomPermission, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func testAccAccountCustomPermission_update(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_quicksight_account_custom_permission.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountCustomPermissionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountCustomPermissionConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccountCustomPermissionExists(ctx, t, resourceName),
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
				},
			},
			{
				Config: testAccAccountCustomPermissionConfig_updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAccountCustomPermissionExists(ctx, t, resourceName),
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
				},
			},
		},
	})
}

func testAccCheckAccountCustomPermissionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).QuickSightClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_quicksight_account_custom_permission" {
				continue
			}
			_, err := tfquicksight.FindAccountCustomPermissionByID(ctx, conn, rs.Primary.Attributes[names.AttrAWSAccountID])

			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return smarterr.NewError(err)
			}
			return smarterr.NewError(fmt.Errorf("QuickSight Account Custom Permission (%s) still exists", rs.Primary.Attributes[names.AttrAWSAccountID]))
		}

		return nil
	}
}

func testAccCheckAccountCustomPermissionExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return smarterr.NewError(fmt.Errorf("Not found: %s", name))
		}

		if rs.Primary.Attributes[names.AttrAWSAccountID] == "" {
			return smarterr.NewError(fmt.Errorf("resource %s has no AWS account ID", name))
		}

		conn := acctest.ProviderMeta(ctx, t).QuickSightClient(ctx)

		_, err := tfquicksight.FindAccountCustomPermissionByID(ctx, conn, rs.Primary.Attributes[names.AttrAWSAccountID])
		if err != nil {
			return smarterr.NewError(err)
		}
		return nil
	}
}

func testAccAccountCustomPermissionConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccCustomPermissionsConfig_basic(rName), `
resource "aws_quicksight_account_custom_permission" "test" {
  custom_permissions_name = aws_quicksight_custom_permissions.test.custom_permissions_name
}
  `)
}

func testAccAccountCustomPermissionConfig_updated(rName string) string {
	return acctest.ConfigCompose(testAccCustomPermissionsConfig_basic(rName), fmt.Sprintf(`
resource "aws_quicksight_custom_permissions" "test2" {
  custom_permissions_name = "%[1]s-2"

  capabilities {
    create_and_update_datasets     = "DENY"
    create_and_update_data_sources = "DENY"
    export_to_pdf                  = "DENY"
  }
}

resource "aws_quicksight_account_custom_permission" "test" {
  custom_permissions_name = aws_quicksight_custom_permissions.test2.custom_permissions_name
}
`, rName))
}
