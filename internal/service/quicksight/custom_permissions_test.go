// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package quicksight_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
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

func TestAccQuickSightCustomPermissions_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.CustomPermissions
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_quicksight_custom_permissions.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomPermissionsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomPermissionsConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomPermissionsExists(ctx, t, resourceName, &v),
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
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("quicksight", regexache.MustCompile(`custompermissions/.+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrAWSAccountID), tfknownvalue.AccountID()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("capabilities"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"add_or_run_anomaly_detection_for_analyses":  knownvalue.Null(),
							"create_and_update_dashboard_email_reports":  knownvalue.Null(),
							"create_and_update_datasets":                 knownvalue.Null(),
							"create_and_update_data_sources":             knownvalue.Null(),
							"create_and_update_themes":                   knownvalue.Null(),
							"create_and_update_threshold_alerts":         knownvalue.Null(),
							"create_shared_folders":                      knownvalue.Null(),
							"create_spice_dataset":                       knownvalue.Null(),
							"export_to_csv":                              knownvalue.Null(),
							"export_to_csv_in_scheduled_reports":         knownvalue.Null(),
							"export_to_excel":                            knownvalue.Null(),
							"export_to_excel_in_scheduled_reports":       knownvalue.Null(),
							"export_to_pdf":                              knownvalue.Null(),
							"export_to_pdf_in_scheduled_reports":         knownvalue.Null(),
							"include_content_in_scheduled_reports_email": knownvalue.Null(),
							"print_reports":                              tfknownvalue.StringExact(awstypes.CapabilityStateDeny),
							"rename_shared_folders":                      knownvalue.Null(),
							"share_analyses":                             knownvalue.Null(),
							"share_dashboards":                           tfknownvalue.StringExact(awstypes.CapabilityStateDeny),
							"share_datasets":                             knownvalue.Null(),
							"share_data_sources":                         knownvalue.Null(),
							"subscribe_dashboard_email_reports":          knownvalue.Null(),
							"view_account_spice_capacity":                knownvalue.Null(),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("custom_permissions_name"), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccCustomPermissionsImportStateID(resourceName),
				ImportStateVerifyIdentifierAttribute: "custom_permissions_name",
			},
		},
	})
}

func TestAccQuickSightCustomPermissions_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.CustomPermissions
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_quicksight_custom_permissions.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomPermissionsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomPermissionsConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomPermissionsExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfquicksight.ResourceCustomPermissions, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccQuickSightCustomPermissions_update(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.CustomPermissions
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_quicksight_custom_permissions.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCustomPermissionsDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCustomPermissionsConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomPermissionsExists(ctx, t, resourceName, &v),
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
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("quicksight", regexache.MustCompile(`custompermissions/.+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrAWSAccountID), tfknownvalue.AccountID()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("capabilities"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"add_or_run_anomaly_detection_for_analyses":  knownvalue.Null(),
							"create_and_update_dashboard_email_reports":  knownvalue.Null(),
							"create_and_update_datasets":                 knownvalue.Null(),
							"create_and_update_data_sources":             knownvalue.Null(),
							"create_and_update_themes":                   knownvalue.Null(),
							"create_and_update_threshold_alerts":         knownvalue.Null(),
							"create_shared_folders":                      knownvalue.Null(),
							"create_spice_dataset":                       knownvalue.Null(),
							"export_to_csv":                              knownvalue.Null(),
							"export_to_csv_in_scheduled_reports":         knownvalue.Null(),
							"export_to_excel":                            knownvalue.Null(),
							"export_to_excel_in_scheduled_reports":       knownvalue.Null(),
							"export_to_pdf":                              knownvalue.Null(),
							"export_to_pdf_in_scheduled_reports":         knownvalue.Null(),
							"include_content_in_scheduled_reports_email": knownvalue.Null(),
							"print_reports":                              tfknownvalue.StringExact(awstypes.CapabilityStateDeny),
							"rename_shared_folders":                      knownvalue.Null(),
							"share_analyses":                             knownvalue.Null(),
							"share_dashboards":                           tfknownvalue.StringExact(awstypes.CapabilityStateDeny),
							"share_datasets":                             knownvalue.Null(),
							"share_data_sources":                         knownvalue.Null(),
							"subscribe_dashboard_email_reports":          knownvalue.Null(),
							"view_account_spice_capacity":                knownvalue.Null(),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("custom_permissions_name"), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccCustomPermissionsImportStateID(resourceName),
				ImportStateVerifyIdentifierAttribute: "custom_permissions_name",
			},
			{
				Config: testAccCustomPermissionsConfig_updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCustomPermissionsExists(ctx, t, resourceName, &v),
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
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("quicksight", regexache.MustCompile(`custompermissions/.+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrAWSAccountID), tfknownvalue.AccountID()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("capabilities"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectPartial(map[string]knownvalue.Check{
							"create_and_update_datasets":     tfknownvalue.StringExact(awstypes.CapabilityStateDeny),
							"create_and_update_data_sources": tfknownvalue.StringExact(awstypes.CapabilityStateDeny),
							"export_to_pdf":                  tfknownvalue.StringExact(awstypes.CapabilityStateDeny),
							"print_reports":                  knownvalue.Null(),
							"share_dashboards":               knownvalue.Null(),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("custom_permissions_name"), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
		},
	})
}

func testAccCheckCustomPermissionsDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).QuickSightClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_quicksight_ip_restriction" {
				continue
			}

			_, err := tfquicksight.FindCustomPermissionsByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrAWSAccountID], rs.Primary.Attributes["custom_permissions_name"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("QuickSight Custom Permissions (%s) still exists", rs.Primary.Attributes["custom_permissions_name"])
		}

		return nil
	}
}

func testAccCheckCustomPermissionsExists(ctx context.Context, t *testing.T, n string, v *awstypes.CustomPermissions) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).QuickSightClient(ctx)

		output, err := tfquicksight.FindCustomPermissionsByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrAWSAccountID], rs.Primary.Attributes["custom_permissions_name"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCustomPermissionsImportStateID(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		return acctest.AttrsImportStateIdFunc(n, ",", names.AttrAWSAccountID, "custom_permissions_name")(s)
	}
}

func testAccCustomPermissionsConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_quicksight_custom_permissions" "test" {
  custom_permissions_name = %[1]q

  capabilities {
    print_reports    = "DENY"
    share_dashboards = "DENY"
  }
}
`, rName)
}

func testAccCustomPermissionsConfig_updated(rName string) string {
	return fmt.Sprintf(`
resource "aws_quicksight_custom_permissions" "test" {
  custom_permissions_name = %[1]q

  capabilities {
    create_and_update_datasets     = "DENY"
    create_and_update_data_sources = "DENY"
    export_to_pdf                  = "DENY"
  }
}
`, rName)
}
