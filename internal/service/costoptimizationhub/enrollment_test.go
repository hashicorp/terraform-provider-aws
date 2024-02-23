// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package costoptimizationhub_test

// **PLEASE DELETE THIS AND ALL TIP COMMENTS BEFORE SUBMITTING A PR FOR REVIEW!**
//
// TIP: ==== INTRODUCTION ====
// Thank you for trying the skaff tool!
//
// You have opted to include these helpful comments. They all include "TIP:"
// to help you find and remove them when you're done with them.
//
// While some aspects of this file are customized to your input, the
// scaffold tool does *not* look at the AWS API and ensure it has correct
// function, structure, and variable names. It makes guesses based on
// commonalities. You will need to make significant adjustments.
//
// In other words, as generated, this is a rough outline of the work you will
// need to do. If something doesn't make sense for your situation, get rid of
// it.

import (
	// TIP: ==== IMPORTS ====
	// This is a common set of imports but not customized to your code since
	// your code hasn't been written yet. Make sure you, your IDE, or
	// goimports -w <file> fixes these imports.
	//
	// The provider linter wants your imports to be in two groups: first,
	// standard library (i.e., "fmt" or "strings"), second, everything else.
	//
	// Also, AWS Go SDK v2 may handle nested structures differently than v1,
	// using the services/costoptimizationhub/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// types.<Type Name>.
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
	"github.com/hashicorp/terraform-provider-aws/names"

	// TIP: You will often need to import the package that this test file lives
	// in. Since it is in the "test" context, it must import the package to use
	// any normal context constants, variables, or functions.
	tfcostoptimizationhub "github.com/hashicorp/terraform-provider-aws/internal/service/costoptimizationhub"
)

// TIP: File Structure. The basic outline for all test files should be as
// follows. Improve this resource's maintainability by following this
// outline.
//
// 1. Package declaration (add "_test" since this is a test file)
// 2. Imports
// 3. Unit tests
// 4. Basic test
// 5. Disappears test
// 6. All the other tests
// 7. Helper functions (exists, destroy, check, etc.)
// 8. Functions that return Terraform configurations

// TIP: ==== ACCEPTANCE TESTS ====
// This is an example of a basic acceptance test. This should test as much of
// standard functionality of the resource as possible, and test importing, if
// applicable. We prefix its name with "TestAcc", the service, and the
// resource name.
//
// Acceptance test access AWS and cost money to run.
func TestAccCostOptimizationHubEnrollment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	// TIP: This is a long-running test guard for tests that run longer than
	// 300s (5 min) generally.
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var les_out costoptimizationhub.ListEnrollmentStatusesOutput
	//var es_in costoptimizationhub.ListEnrollmentStatusesInput
	//var up_out costoptimizationhub.UpdateEnrollmentStatusOutput
	//var up_in costoptimizationhub.UpdatePreferencesInput

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
					// TIP: The Plugin-Framework disappears helper is similar to the Plugin-SDK version,
					// but expects a new resource factory function as the third argument. To expose this
					// private function to the testing package, you may need to add a line like the following
					// to exports_test.go:
					//
					//   var ResourceEnrollment = newResourceEnrollment
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

// func testAccPreCheck(ctx context.Context, t *testing.T) {
// 	conn := acctest.Provider.Meta().(*conns.AWSClient).CostOptimizationHubClient(ctx)

// 	input := &costoptimizationhub.ListEnrollmentsInput{}
// 	_, err := conn.ListEnrollments(ctx, input)

// 	if acctest.PreCheckSkipError(err) {
// 		t.Skipf("skipping acceptance testing: %s", err)
// 	}
// 	if err != nil {
// 		t.Fatalf("unexpected PreCheck error: %s", err)
// 	}
// }

// func testAccCheckEnrollmentNotRecreated(before, after *costoptimizationhub.DescribeEnrollmentResponse) resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		if before, after := aws.ToString(before.EnrollmentId), aws.ToString(after.EnrollmentId); before != after {
// 			return create.Error(names.CostOptimizationHub, create.ErrActionCheckingNotRecreated, tfcostoptimizationhub.ResNameEnrollment, aws.ToString(before.EnrollmentId), errors.New("recreated"))
// 		}

// 		return nil
// 	}
// }

func testAccEnrollmentConfig_basic() string {
	return `
	resource "aws_costoptimizationhub_enrollment" "test" {
	}	
`
}
