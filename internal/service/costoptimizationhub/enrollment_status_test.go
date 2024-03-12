// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package costoptimizationhub_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/costoptimizationhub"
	"github.com/aws/aws-sdk-go-v2/service/costoptimizationhub/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfcostoptimizationhub "github.com/hashicorp/terraform-provider-aws/internal/service/costoptimizationhub"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCostOptimizationHubEnrollmentStatus_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var les_out costoptimizationhub.ListEnrollmentStatusesOutput
	resourceName := "aws_costoptimizationhub_enrollment_status.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CostOptimizationHub),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		//CheckDestroy:             testAccCheckEnrollmentStatusDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnrollmentStatusConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnrollmentStatusExists(ctx, resourceName, &les_out),
					resource.TestCheckResourceAttr(resourceName, "status", "Active"),
					resource.TestCheckResourceAttr(resourceName, "include_member_accounts", "false"),
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

func TestAccCostOptimizationHubEnrollmentStatus_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if os.Getenv("COSTOPTIMIZATIONHUB_UNENROLL_ACCOUNT_ON_DESTROY") == "" {
		t.Skip("Environment variable COSTOPTIMIZATIONHUB_UNENROLL_ACCOUNT_ON_DESTROY is not set")
	}

	resourceName := "aws_costoptimizationhub_enrollment_status.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CostOptimizationHub),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckEnrollmentStatusDestroy(ctx),
		Steps: []resource.TestStep{
			{
				// unenroll_on_destroy must be enabled for the disappears helper to disable
				// Cost Optimization Hub on destroy and trigger the non-empty plan after state refresh
				Config: testAccEnrollmentStatusConfig_unenrollOnDestroy(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnrollmentStatusIsActive(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfcostoptimizationhub.ResourceEnrollmentStatus, resourceName),
				),
			},
			{
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// testAccCheckEnrollmentStatusDestroy verfies ListEnrollmentStatuses does not return an error
//
// Since this resource manages enrollment/unenrollment to Cost Optimization Hub, there is nothing
// to destroy. Additionally, because enrollment may remain active depending on whether
// the unenroll_on_destroy attribute was set, this function does not check that account
// enrollment is inactive, simply that the status check returns a valid response.
func testAccCheckEnrollmentStatusDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CostOptimizationHubClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_costoptimizationhub_enrollment_status" {
				continue
			}

			_, err := conn.ListEnrollmentStatuses(ctx, &costoptimizationhub.ListEnrollmentStatusesInput{})
			if err != nil {
				return err
			}
		}

		return nil
	}
}

// testAccCheckEnrollmentStatusIsActive verifies Cost Optimization Hub is active in the current account/region combination
func testAccCheckEnrollmentStatusIsActive(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.CostOptimizationHub, create.ErrActionCheckingExistence, tfcostoptimizationhub.ResNameEnrollmentStatus, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.CostOptimizationHub, create.ErrActionCheckingExistence, tfcostoptimizationhub.ResNameEnrollmentStatus, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CostOptimizationHubClient(ctx)
		out, err := conn.ListEnrollmentStatuses(ctx, &costoptimizationhub.ListEnrollmentStatusesInput{})
		if err != nil {
			return create.Error(names.CostOptimizationHub, create.ErrActionCheckingExistence, tfcostoptimizationhub.ResNameEnrollmentStatus, rs.Primary.ID, err)
		}
		if out == nil || out.Items[0].Status != types.EnrollmentStatusActive {
			return create.Error(names.CostOptimizationHub, create.ErrActionCheckingExistence, tfcostoptimizationhub.ResNameEnrollmentStatus, rs.Primary.ID, errors.New("Cost Optimization Hub not active"))
		}

		return nil
	}
}

func TestAccCostOptimizationHubEnrollmentStatus_IncludeMemberAccounts_True(t *testing.T) {
	ctx := acctest.Context(t)

	var les_out costoptimizationhub.ListEnrollmentStatusesOutput
	resourceName := "aws_costoptimizationhub_enrollment_status.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CostOptimizationHub),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		//CheckDestroy:             testAccCheckEnrollmentStatusDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnrollmentStatusConfig_IncludeMemberAccounts_True(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnrollmentStatusExists(ctx, resourceName, &les_out),
					resource.TestCheckResourceAttr(resourceName, "status", "Active"),
					resource.TestCheckResourceAttr(resourceName, "include_member_accounts", "true"),
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

func TestAccCostOptimizationHubEnrollmentStatus_IncludeMemberAccounts_False(t *testing.T) {
	ctx := acctest.Context(t)

	var les_out costoptimizationhub.ListEnrollmentStatusesOutput
	resourceName := "aws_costoptimizationhub_enrollment_status.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CostOptimizationHub),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		//CheckDestroy:             testAccCheckEnrollmentStatusDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnrollmentStatusConfig_IncludeMemberAccounts_False(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnrollmentStatusExists(ctx, resourceName, &les_out),
					resource.TestCheckResourceAttr(resourceName, "status", "Active"),
					resource.TestCheckResourceAttr(resourceName, "include_member_accounts", "false"),
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

func TestAccCostOptimizationHubEnrollmentStatus_Status_Inactive(t *testing.T) {
	ctx := acctest.Context(t)

	var les_out costoptimizationhub.ListEnrollmentStatusesOutput
	resourceName := "aws_costoptimizationhub_enrollment_status.test"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.CostOptimizationHub),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		//CheckDestroy:             testAccCheckEnrollmentStatusDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccEnrollmentStatusConfig_Status_Inactive(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnrollmentStatusExists(ctx, resourceName, &les_out),
					resource.TestCheckResourceAttr(resourceName, "status", "Inactive"),
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

func testAccCheckEnrollmentStatusExists(ctx context.Context, name string, enrollmentstatus *costoptimizationhub.ListEnrollmentStatusesOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.CostOptimizationHub, create.ErrActionCheckingExistence, tfcostoptimizationhub.ResNameEnrollmentStatus, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.CostOptimizationHub, create.ErrActionCheckingExistence, tfcostoptimizationhub.ResNameEnrollmentStatus, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CostOptimizationHubClient(ctx)
		resp, err := conn.ListEnrollmentStatuses(ctx, &costoptimizationhub.ListEnrollmentStatusesInput{
			IncludeOrganizationInfo: false,
		})

		if err != nil {
			return create.Error(names.CostOptimizationHub, create.ErrActionCheckingExistence, tfcostoptimizationhub.ResNameEnrollmentStatus, rs.Primary.ID, err)
		}

		*enrollmentstatus = *resp

		return nil
	}
}

func testAccEnrollmentStatusConfig_basic() string {
	return `
resource "aws_costoptimizationhub_enrollment_status" "test" {
  status = "Active"
}
`
}

func testAccEnrollmentStatusConfig_IncludeMemberAccounts_True() string {
	return `
resource "aws_costoptimizationhub_enrollment_status" "test" {
  status                  = "Active"
  include_member_accounts = true
}
`
}

func testAccEnrollmentStatusConfig_IncludeMemberAccounts_False() string {
	return `
resource "aws_costoptimizationhub_enrollment_status" "test" {
  status                  = "Active"
  include_member_accounts = false
}
`
}

func testAccEnrollmentStatusConfig_Status_Inactive() string {
	return `
resource "aws_costoptimizationhub_enrollment_status" "test" {
  status = "Inactive"
}
`
}

func testAccEnrollmentStatusConfig_unenrollOnDestroy() string {
	return `
resource "aws_costoptimizationhub_enrollment_status" "test" {
  status              = "Active"
  unenroll_on_destroy = true
}
`
}
