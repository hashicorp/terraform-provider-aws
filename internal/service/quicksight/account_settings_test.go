// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"

	tfquicksight "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight"
)

func TestAccQuickSightAccountSettings_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var accountsettings awstypes.AccountSettings
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_quicksight_account_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountSettings_basic(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAccountSettingsExists(ctx, resourceName, &accountsettings),
				),
			},
			{
				ResourceName: resourceName,
				ImportState:  false,
				RefreshState: true,
			},
		},
	})
}

func TestAccQuickSightAccountSettings_resetOnDelete_false(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_quicksight_account_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountSettingsDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_resetOnDelete(rName, false),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("termination_protection_enabled"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("reset_on_delete"), knownvalue.Bool(false)),
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"reset_on_delete",
				},
			},
		},
	})
}

func testAccCheckAccountSettingsDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_quicksight_account_settings" {
				continue
			}

			settings, err := tfquicksight.FindAccountSettingsByID(ctx, conn, rs.Primary.ID)
			if tfresource.NotFound(err) {
				return nil
			}

			fmt.Println("testAccCheckAccountSettingsReset: ")
			fmt.Println(settings)

			if settings.AccountName == nil {
				// Settings have not been reset
				return nil
			}

			if err != nil {
				return create.Error(names.QuickSight, create.ErrActionCheckingDestroyed, tfquicksight.ResNameAccountSettings, rs.Primary.ID, err)
			}

			return create.Error(names.QuickSight, create.ErrActionCheckingDestroyed, tfquicksight.ResNameAccountSettings, rs.Primary.ID, errors.New("Quicksight account settings were reset"))
		}

		return nil
	}
}

func testAccCheckAccountSettingsExists(ctx context.Context, n string, v *types.AccountSettings) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return create.Error(names.QuickSight, create.ErrActionCheckingExistence, tfquicksight.ResNameAccountSettings, n, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.QuickSight, create.ErrActionCheckingExistence, tfquicksight.ResNameAccountSettings, n, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightClient(ctx)

		resp, err := tfquicksight.FindAccountSettingsByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.QuickSight, create.ErrActionCheckingExistence, tfquicksight.ResNameAccountSettings, rs.Primary.ID, err)
		}

		*v = *resp

		return nil
	}
}

func testAccAccountSettings_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_quicksight_account_subscription" "test" {
  account_name          = %[1]q
  authentication_method = "IAM_AND_QUICKSIGHT"
  edition               = "ENTERPRISE"
  notification_email    = %[2]q
}

`, rName, acctest.DefaultEmailAddress)
}

func testAccAccountSettings_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccAccountSettings_base(rName), `
resource "aws_quicksight_account_settings" "test" {
  termination_protection_enabled = false

  depends_on = [aws_quicksight_account_subscription.test]
}
`)
}

func testAccAccountConfig_resetOnDelete(rName string, reset bool) string {
	return acctest.ConfigCompose(
		testAccAccountSettings_base(rName),
		fmt.Sprintf(`
resource "aws_quicksight_account_settings" "test" {
  termination_protection_enabled = false
  reset_on_delete                = %[1]t

  depends_on = [aws_quicksight_account_subscription.test]
}
`, reset))
}
