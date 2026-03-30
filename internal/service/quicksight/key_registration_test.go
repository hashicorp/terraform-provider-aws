// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package quicksight_test

import (
	"context"
	"fmt"
	"testing"

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

func testAccKeyRegistration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_quicksight_key_registration.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyRegistrationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyRegistrationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckKeyRegistrationExists(ctx, t, resourceName),
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
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("key_registration"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"default_key": knownvalue.Bool(false),
							"key_arn":     knownvalue.NotNull(),
						}),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrAWSAccountID),
				ImportStateVerifyIdentifierAttribute: names.AttrAWSAccountID,
			},
			{
				Config: testAccKeyRegistrationConfig_updated(rName),
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
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("key_registration"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"default_key": knownvalue.Bool(false),
							"key_arn":     knownvalue.NotNull(),
						}),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"default_key": knownvalue.Bool(true),
							"key_arn":     knownvalue.NotNull(),
						}),
					})),
				},
			},
		},
	})
}

func testAccKeyRegistration_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_quicksight_key_registration.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.QuickSightEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckKeyRegistrationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccKeyRegistrationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKeyRegistrationExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfquicksight.ResourceKeyRegistration, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckKeyRegistrationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).QuickSightClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_quicksight_key_registration" {
				continue
			}

			_, err := tfquicksight.FindKeyRegistrationByID(ctx, conn, rs.Primary.Attributes[names.AttrAWSAccountID])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("QuickSight Account Subscription (%s) still exists", rs.Primary.Attributes[names.AttrAWSAccountID])
		}

		return nil
	}
}

func testAccCheckKeyRegistrationExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).QuickSightClient(ctx)

		_, err := tfquicksight.FindKeyRegistrationByID(ctx, conn, rs.Primary.Attributes[names.AttrAWSAccountID])

		return err
	}
}

func testAccKeyRegistrationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  count = 1

  deletion_window_in_days = 7
  description             = "%[1]s-${count.index}"
  enable_key_rotation     = true
}

resource "aws_quicksight_key_registration" "test" {
  key_registration {
    key_arn = aws_kms_key.test[0].arn
  }
}
`, rName)
}

func testAccKeyRegistrationConfig_updated(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  count = 2

  deletion_window_in_days = 7
  description             = "%[1]s-${count.index}"
  enable_key_rotation     = true
}

resource "aws_quicksight_key_registration" "test" {
  key_registration {
    key_arn = aws_kms_key.test[0].arn
  }
  key_registration {
    key_arn     = aws_kms_key.test[1].arn
    default_key = true
  }
}
`, rName)
}
