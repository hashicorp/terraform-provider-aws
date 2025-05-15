// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccAccountSettings_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_quicksight_account_settings.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountSettingsConfig_basic(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
					},
				},
			},
			{
				ResourceName: resourceName,
				ImportState:  false,
				RefreshState: true,
			},
		},
	})
}

func testAccAccountSettingsConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_quicksight_account_subscription" "test" {
  account_name          = %[1]q
  authentication_method = "IAM_AND_QUICKSIGHT"
  edition               = "ENTERPRISE"
  notification_email    = %[2]q
}
`, rName, acctest.DefaultEmailAddress)
}

func testAccAccountSettingsConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccAccountSettingsConfig_base(rName), `
resource "aws_quicksight_account_settings" "test" {
  termination_protection_enabled = false

  depends_on = [aws_quicksight_account_subscription.test]
}
`)
}
