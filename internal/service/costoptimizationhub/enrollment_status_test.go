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
	tfcostoptimizationhub "github.com/hashicorp/terraform-provider-aws/internal/service/costoptimizationhub"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccEnrollmentStatus_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var les_out costoptimizationhub.ListEnrollmentStatusesOutput
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

func testAccEnrollmentStatus_disappears(t *testing.T) {
	ctx := acctest.Context(t)

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
				Config: testAccEnrollmentStatusConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEnrollmentStatusIsActive(ctx, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfcostoptimizationhub.ResourceEnrollmentStatus, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				RefreshState:       true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// Since this resource manages enrollment/unenrollment destroying the resource
// only means that account enrollment is inactive.
func testAccCheckEnrollmentStatusDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).CostOptimizationHubClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_costoptimizationhub_enrollment_status" {
				continue
			}

			out, err := conn.ListEnrollmentStatuses(ctx, &costoptimizationhub.ListEnrollmentStatusesInput{})
			if err != nil {
				return err
			}
			if out.Items[0].Status == types.EnrollmentStatusInactive {
				continue
			}

			return fmt.Errorf("Cost Optimization Hub Enrollment Status %s still exists", rs.Primary.ID)
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
		if out == nil || out.Items[0].Status != "Active" {
			return create.Error(names.CostOptimizationHub, create.ErrActionCheckingExistence, tfcostoptimizationhub.ResNameEnrollmentStatus, rs.Primary.ID, errors.New("Cost Optimization Hub not active"))
		}

		return nil
	}
}

func testAccEnrollmentStatus_includeMemberAccounts(t *testing.T) {
	ctx := acctest.Context(t)

	var les_out costoptimizationhub.ListEnrollmentStatusesOutput
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
				Config: testAccEnrollmentStatusConfig_IncludeMemberAccounts(),
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
}
`
}

func testAccEnrollmentStatusConfig_IncludeMemberAccounts() string {
	return `
resource "aws_costoptimizationhub_enrollment_status" "test" {
  include_member_accounts = true
}
`
}
