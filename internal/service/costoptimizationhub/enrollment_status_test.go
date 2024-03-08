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
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
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
  status = "Active"
  include_member_accounts = true
}
`
}

func testAccEnrollmentStatusConfig_IncludeMemberAccounts_False() string {
	return `
resource "aws_costoptimizationhub_enrollment_status" "test" {
  status = "Active"
  include_member_accounts = true
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
