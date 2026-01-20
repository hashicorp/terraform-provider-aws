// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package costoptimizationhub_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/costoptimizationhub"
	"github.com/aws/aws-sdk-go-v2/service/costoptimizationhub/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfcostoptimizationhub "github.com/hashicorp/terraform-provider-aws/internal/service/costoptimizationhub"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccPreferences_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var out costoptimizationhub.GetPreferencesOutput
	resourceName := "aws_costoptimizationhub_preferences.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CostOptimizationHub),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPreferencesDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPreferencesConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPreferencesExists(ctx, t, resourceName, &out),
					resource.TestCheckResourceAttr(resourceName, "member_account_discount_visibility", "All"),
					resource.TestCheckResourceAttr(resourceName, "savings_estimation_mode", "BeforeDiscounts"),
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

func testAccPreferences_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_costoptimizationhub_preferences.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CostOptimizationHub),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPreferencesDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPreferencesConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnrollmentStatusIsActive(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfcostoptimizationhub.ResourcePreferences, resourceName),
				),
			},
			{
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccPreferences_memberAccountsDiscountVisibility(t *testing.T) {
	ctx := acctest.Context(t)

	var out costoptimizationhub.GetPreferencesOutput
	resourceName := "aws_costoptimizationhub_preferences.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CostOptimizationHub),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPreferencesDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPreferencesConfig_MemberAccountDiscountVisibility(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPreferencesExists(ctx, t, resourceName, &out),
					resource.TestCheckResourceAttr(resourceName, "member_account_discount_visibility", "None"),
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

func testAccPreferences_savingsEstimationMode(t *testing.T) {
	ctx := acctest.Context(t)

	var out costoptimizationhub.GetPreferencesOutput
	resourceName := "aws_costoptimizationhub_preferences.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
			acctest.PreCheckOrganizationManagementAccount(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CostOptimizationHub),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPreferencesDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPreferencesConfig_SavingsEstimationMode(string(types.SavingsEstimationModeAfterDiscounts)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPreferencesExists(ctx, t, resourceName, &out),
					resource.TestCheckResourceAttr(resourceName, "savings_estimation_mode", string(types.SavingsEstimationModeAfterDiscounts)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccPreferencesConfig_SavingsEstimationMode(string(types.SavingsEstimationModeBeforeDiscounts)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPreferencesExists(ctx, t, resourceName, &out),
					resource.TestCheckResourceAttr(resourceName, "savings_estimation_mode", string(types.SavingsEstimationModeBeforeDiscounts)),
				),
			},
		},
	})
}

// Since this resource manages preferences of the Enrollment, destruction could mean
// one of the following two scenarios:
// 1. The account is not enrolled in Cost Optimization Hub (in which case the "Preferences" does not exist)
// 2. The account is enrolled in Cost Optimizaiton Hub but the preferences are set to the Default values.
func testAccCheckPreferencesDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).CostOptimizationHubClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_costoptimizationhub_preferences" {
				continue
			}

			output, err := tfcostoptimizationhub.FindPreferences(ctx, conn)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			if (output.MemberAccountDiscountVisibility == types.MemberAccountDiscountVisibilityAll) &&
				(output.SavingsEstimationMode == types.SavingsEstimationModeBeforeDiscounts) {
				// Cost Optimization Hub Enrollment status is Active but there are default values for the preferences.
				continue
			}

			return fmt.Errorf("Cost Optimization Hub Preferences %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckPreferencesExists(ctx context.Context, t *testing.T, n string, v *costoptimizationhub.GetPreferencesOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).CostOptimizationHubClient(ctx)

		output, err := tfcostoptimizationhub.FindPreferences(ctx, conn)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreferencesConfig_base() string {
	return `
resource "aws_costoptimizationhub_enrollment_status" "test_enrollment_status" {
}
`
}

func testAccPreferencesConfig_basic() string {
	return acctest.ConfigCompose(testAccPreferencesConfig_base(), `
resource "aws_costoptimizationhub_preferences" "test" {
  depends_on = [aws_costoptimizationhub_enrollment_status.test_enrollment_status]
}
`)
}

func testAccPreferencesConfig_MemberAccountDiscountVisibility() string {
	return acctest.ConfigCompose(testAccPreferencesConfig_base(), `
resource "aws_costoptimizationhub_preferences" "test" {
  member_account_discount_visibility = "None"
  depends_on                         = [aws_costoptimizationhub_enrollment_status.test_enrollment_status]
}
`)
}

func testAccPreferencesConfig_SavingsEstimationMode(mode string) string {
	return acctest.ConfigCompose(
		testAccPreferencesConfig_base(),
		fmt.Sprintf(`
resource "aws_costoptimizationhub_preferences" "test" {
  savings_estimation_mode = %[1]q
  depends_on              = [aws_costoptimizationhub_enrollment_status.test_enrollment_status]
}
`, mode))
}
