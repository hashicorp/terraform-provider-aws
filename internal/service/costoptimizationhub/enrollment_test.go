// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package costoptimizationhub_test

import (
	"context"
	"errors"
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

func TestAccCostOptimizationHubEnrollment_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var les_out costoptimizationhub.ListEnrollmentStatusesOutput

	resourceName := "aws_costoptimizationhub_enrollment.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CostOptimizationHub),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnrollmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnrollmentConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnrollmentExists(ctx, resourceName, &les_out),
					resource.TestCheckResourceAttr(resourceName, "include_member_accounts", "false"),
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

func TestAccCostOptimizationHubEnrollment_IncludeMemberAccounts(t *testing.T) {
	ctx := acctest.Context(t)

	var les_out costoptimizationhub.ListEnrollmentStatusesOutput

	resourceName := "aws_costoptimizationhub_enrollment.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CostOptimizationHub),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnrollmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnrollmentConfig_IncludeMemberAccounts(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnrollmentExists(ctx, resourceName, &les_out),
					resource.TestCheckResourceAttr(resourceName, "include_member_accounts", "true"),
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

func TestAccCostOptimizationHubEnrollment_MemberAccountDiscountVisibility(t *testing.T) {
	ctx := acctest.Context(t)

	var les_out costoptimizationhub.ListEnrollmentStatusesOutput

	resourceName := "aws_costoptimizationhub_enrollment.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CostOptimizationHub),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnrollmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnrollmentConfig_MemberAccountDiscountVisibility(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnrollmentExists(ctx, resourceName, &les_out),
					resource.TestCheckResourceAttr(resourceName, "include_member_accounts", "false"),
					resource.TestCheckResourceAttr(resourceName, "member_account_discount_visibility", "None"),
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
func TestAccCostOptimizationHubEnrollment_SavingsEstimationMode(t *testing.T) {
	ctx := acctest.Context(t)

	var les_out costoptimizationhub.ListEnrollmentStatusesOutput

	resourceName := "aws_costoptimizationhub_enrollment.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CostOptimizationHub),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnrollmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnrollmentConfig_SavingsEstimationMode(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnrollmentExists(ctx, resourceName, &les_out),
					resource.TestCheckResourceAttr(resourceName, "include_member_accounts", "false"),
					resource.TestCheckResourceAttr(resourceName, "member_account_discount_visibility", "All"),
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

func TestAccCostOptimizationHubEnrollment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var enrollment costoptimizationhub.ListEnrollmentStatusesOutput
	resourceName := "aws_costoptimizationhub_enrollment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CostOptimizationHub),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnrollmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnrollmentConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnrollmentExists(ctx, resourceName, &enrollment),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfcostoptimizationhub.ResourceEnrollment, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckEnrollmentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CostOptimizationHubClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_costoptimizationhub_enrollment" {
				continue
			}

			les_out, err := conn.ListEnrollmentStatuses(ctx, &costoptimizationhub.ListEnrollmentStatusesInput{})
			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.CostOptimizationHub, create.ErrActionCheckingDestroyed, tfcostoptimizationhub.ResNameEnrollment, rs.Primary.ID, err)
			}
			if les_out.Items[0].Status == "Active" {
				return create.Error(names.CostOptimizationHub, create.ErrActionCheckingDestroyed, tfcostoptimizationhub.ResNameEnrollment, rs.Primary.ID, errors.New("not destroyed"))
			}

		}
		return nil
	}
}

func testAccCheckEnrollmentExists(ctx context.Context, name string, enrollment *costoptimizationhub.ListEnrollmentStatusesOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.CostOptimizationHub, create.ErrActionCheckingExistence, tfcostoptimizationhub.ResNameEnrollment, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.CostOptimizationHub, create.ErrActionCheckingExistence, tfcostoptimizationhub.ResNameEnrollment, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CostOptimizationHubClient(ctx)
		les_out, err := conn.ListEnrollmentStatuses(ctx, &costoptimizationhub.ListEnrollmentStatusesInput{})

		if err != nil || les_out.Items[0].Status != "Active" {
			return create.Error(names.CostOptimizationHub, create.ErrActionCheckingExistence, tfcostoptimizationhub.ResNameEnrollment, rs.Primary.ID, err)
		}

		*enrollment = *les_out

		return nil
	}
}

func testAccEnrollmentConfig_basic() string {
	return `
	resource "aws_costoptimizationhub_enrollment" "test" {
	}	
`
}

func testAccEnrollmentConfig_IncludeMemberAccounts() string {
	return `
	resource "aws_costoptimizationhub_enrollment" "test" {
		include_member_accounts = true
	}	
`
}
func testAccEnrollmentConfig_MemberAccountDiscountVisibility() string {
	return `
	resource "aws_costoptimizationhub_enrollment" "test" {
		member_account_discount_visibility = "None"
	}	
`
}
func testAccEnrollmentConfig_SavingsEstimationMode() string {
	return `
	resource "aws_costoptimizationhub_enrollment" "test" {
		savings_estimation_mode = "AfterDiscounts"
	}	
`
}
