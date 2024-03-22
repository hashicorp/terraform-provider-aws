// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package costoptimizationhub_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/costoptimizationhub"
	"github.com/aws/aws-sdk-go-v2/service/costoptimizationhub/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfcostoptimizationhub "github.com/hashicorp/terraform-provider-aws/internal/service/costoptimizationhub"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccPreferences_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var out costoptimizationhub.GetPreferencesOutput
	resourceName := "aws_costoptimizationhub_preferences.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CostOptimizationHub),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPreferencesDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPreferencesConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPreferencesExists(ctx, resourceName, &out),
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

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CostOptimizationHub),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPreferencesDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPreferencesConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnrollmentStatusIsActive(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfcostoptimizationhub.ResourcePreferences, resourceName),
				),
			},
			{
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// Since this resource manages preferences of the Enrollment, destruction could mean
// one of the following two scenarios:
// 1. The account is not enrolled in Cost Optimization Hub (in which case the "Preferences" does not exist)
// 2. The account is enrolled in Cost Optimizaiton Hub but the preferences are set to the Default values.
func testAccCheckPreferencesDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CostOptimizationHubClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_costoptimizationhub_preferences" {
				continue
			}

			out, err := conn.GetPreferences(ctx, &costoptimizationhub.GetPreferencesInput{})
			if err != nil {
				if errs.MessageContains(err, "AccessDenied", "AWS account is not enrolled for recommendations") {
					// Scenario 1 - Cost Optimization Hub Enrollment status is inactive
					continue
				}
				return err
			}
			if (out.MemberAccountDiscountVisibility == types.MemberAccountDiscountVisibilityAll) &&
				(out.SavingsEstimationMode == types.SavingsEstimationModeBeforeDiscounts) {
				// Scenario 2 - Cost Optimization Hub Enrollment status is Active but there are default values for the preferences
				continue
			}

			return fmt.Errorf("Cost Optimization Hub Preferences %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccPreferences_memberAccountsDiscountVisibility(t *testing.T) {
	ctx := acctest.Context(t)

	var out costoptimizationhub.GetPreferencesOutput
	resourceName := "aws_costoptimizationhub_preferences.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CostOptimizationHub),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPreferencesDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPreferencesConfig_MemberAccountDiscountVisibility(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPreferencesExists(ctx, resourceName, &out),
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

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CostOptimizationHub),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPreferencesDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPreferencesConfig_SavingsEstimationMode(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPreferencesExists(ctx, resourceName, &out),
					resource.TestCheckResourceAttr(resourceName, "savings_estimation_mode", "AfterDiscounts"),
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

func testAccCheckPreferencesExists(ctx context.Context, name string, preferences *costoptimizationhub.GetPreferencesOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.CostOptimizationHub, create.ErrActionCheckingExistence, tfcostoptimizationhub.ResNamePreferences, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.CostOptimizationHub, create.ErrActionCheckingExistence, tfcostoptimizationhub.ResNamePreferences, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CostOptimizationHubClient(ctx)
		resp, err := conn.GetPreferences(ctx, &costoptimizationhub.GetPreferencesInput{})

		if err != nil {
			return create.Error(names.CostOptimizationHub, create.ErrActionCheckingExistence, tfcostoptimizationhub.ResNamePreferences, rs.Primary.ID, err)
		}

		*preferences = *resp

		return nil
	}
}

func testAccPreferencesBase() string {
	return `
resource "aws_costoptimizationhub_enrollment_status" "test_enrollment_status" {
}
`
}

func testAccPreferencesConfig_basic() string {
	return testAccPreferencesBase() + `
resource "aws_costoptimizationhub_preferences" "test" {
  depends_on = [aws_costoptimizationhub_enrollment_status.test_enrollment_status]
}
`
}

func testAccPreferencesConfig_MemberAccountDiscountVisibility() string {
	return testAccPreferencesBase() + `
resource "aws_costoptimizationhub_preferences" "test" {
  member_account_discount_visibility = "None"
  depends_on                         = [aws_costoptimizationhub_enrollment_status.test_enrollment_status]
}
`
}

func testAccPreferencesConfig_SavingsEstimationMode() string {
	return testAccPreferencesBase() + `
resource "aws_costoptimizationhub_preferences" "test" {
  savings_estimation_mode = "AfterDiscounts"
  depends_on              = [aws_costoptimizationhub_enrollment_status.test_enrollment_status]
}
`
}
