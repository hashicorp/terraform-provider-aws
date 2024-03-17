// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tfquicksight "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccQuickSightIAMPolicyAssignment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var assignment types.IAMPolicyAssignment
	resourceName := "aws_quicksight_iam_policy_assignment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIAMPolicyAssignmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIAMPolicyAssignmentConfig_basic(rName, flex.StringValueToFramework(ctx, types.AssignmentStatusEnabled).String()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMPolicyAssignmentExists(ctx, resourceName, &assignment),
					resource.TestCheckResourceAttr(resourceName, "assignment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "assignment_status", flex.StringValueToFramework(ctx, types.AssignmentStatusEnabled).String()),
					resource.TestCheckResourceAttr(resourceName, "namespace", tfquicksight.DefaultIAMPolicyAssignmentNamespace),
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

func TestAccQuickSightIAMPolicyAssignment_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var assignment types.IAMPolicyAssignment
	resourceName := "aws_quicksight_iam_policy_assignment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIAMPolicyAssignmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIAMPolicyAssignmentConfig_basic(rName, flex.StringValueToFramework(ctx, types.AssignmentStatusEnabled).String()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMPolicyAssignmentExists(ctx, resourceName, &assignment),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfquicksight.ResourceIAMPolicyAssignment, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccQuickSightIAMPolicyAssignment_assignmentStatus(t *testing.T) {
	ctx := acctest.Context(t)
	var assignment types.IAMPolicyAssignment
	resourceName := "aws_quicksight_iam_policy_assignment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIAMPolicyAssignmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIAMPolicyAssignmentConfig_basic(rName, flex.StringValueToFramework(ctx, types.AssignmentStatusDraft).String()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMPolicyAssignmentExists(ctx, resourceName, &assignment),
					resource.TestCheckResourceAttr(resourceName, "assignment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "assignment_status", flex.StringValueToFramework(ctx, types.AssignmentStatusDraft).String()),
					resource.TestCheckResourceAttr(resourceName, "namespace", tfquicksight.DefaultIAMPolicyAssignmentNamespace),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccIAMPolicyAssignmentConfig_basic(rName, flex.StringValueToFramework(ctx, types.AssignmentStatusEnabled).String()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMPolicyAssignmentExists(ctx, resourceName, &assignment),
					resource.TestCheckResourceAttr(resourceName, "assignment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "assignment_status", flex.StringValueToFramework(ctx, types.AssignmentStatusEnabled).String()),
					resource.TestCheckResourceAttr(resourceName, "namespace", tfquicksight.DefaultIAMPolicyAssignmentNamespace),
				),
			},
			{
				Config: testAccIAMPolicyAssignmentConfig_basic(rName, flex.StringValueToFramework(ctx, types.AssignmentStatusDisabled).String()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMPolicyAssignmentExists(ctx, resourceName, &assignment),
					resource.TestCheckResourceAttr(resourceName, "assignment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "assignment_status", flex.StringValueToFramework(ctx, types.AssignmentStatusDisabled).String()),
					resource.TestCheckResourceAttr(resourceName, "namespace", tfquicksight.DefaultIAMPolicyAssignmentNamespace),
				),
			},
		},
	})
}

func TestAccQuickSightIAMPolicyAssignment_identities(t *testing.T) {
	ctx := acctest.Context(t)
	var assignment types.IAMPolicyAssignment
	resourceName := "aws_quicksight_iam_policy_assignment.test"
	policyResourceName := "aws_iam_policy.test"
	userResourceName := "aws_quicksight_user.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIAMPolicyAssignmentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIAMPolicyAssignmentConfig_identities(rName, flex.StringValueToFramework(ctx, types.AssignmentStatusEnabled).String()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMPolicyAssignmentExists(ctx, resourceName, &assignment),
					resource.TestCheckResourceAttr(resourceName, "assignment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "assignment_status", flex.StringValueToFramework(ctx, types.AssignmentStatusEnabled).String()),
					resource.TestCheckResourceAttr(resourceName, "namespace", tfquicksight.DefaultIAMPolicyAssignmentNamespace),
					resource.TestCheckResourceAttr(resourceName, "identities.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "identities.0.user.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "identities.0.user.0", userResourceName, "user_name"),
					resource.TestCheckResourceAttrPair(resourceName, "policy_arn", policyResourceName, "arn"),
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

func testAccCheckIAMPolicyAssignmentExists(ctx context.Context, resourceName string, assignment *types.IAMPolicyAssignment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightClient(ctx)
		output, err := tfquicksight.FindIAMPolicyAssignmentByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.QuickSight, create.ErrActionCheckingExistence, tfquicksight.ResNameIAMPolicyAssignment, rs.Primary.ID, err)
		}

		*assignment = *output

		return nil
	}
}

func testAccCheckIAMPolicyAssignmentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).QuickSightClient(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_quicksight_iam_policy_assignment" {
				continue
			}

			output, err := tfquicksight.FindIAMPolicyAssignmentByID(ctx, conn, rs.Primary.ID)
			if err != nil {
				if errs.IsA[*types.ResourceNotFoundException](err) {
					return nil
				}
				return err
			}

			if output != nil {
				return create.Error(names.QuickSight, create.ErrActionCheckingDestroyed, tfquicksight.ResNameIAMPolicyAssignment, rs.Primary.ID, err)
			}
		}

		return nil
	}
}

func testAccIAMPolicyAssignmentConfig_basic(rName, assignmentStatus string) string {
	return fmt.Sprintf(`
resource "aws_quicksight_iam_policy_assignment" "test" {
  assignment_name   = %[1]q
  assignment_status = %[2]q
}
`, rName, assignmentStatus)
}

func testAccIAMPolicyAssignmentConfig_identities(rName, assignmentStatus string) string {
	return fmt.Sprintf(`
resource "aws_quicksight_user" "test" {
  user_name     = %[1]q
  email         = %[3]q
  identity_type = "QUICKSIGHT"
  user_role     = "READER"
}

data "aws_iam_policy_document" "test" {
  statement {
    actions   = ["quicksight:ListUsers"]
    resources = ["*"]
  }
}

resource "aws_iam_policy" "test" {
  policy = data.aws_iam_policy_document.test.json
}

resource "aws_quicksight_iam_policy_assignment" "test" {
  assignment_name   = %[1]q
  assignment_status = %[2]q
  policy_arn        = aws_iam_policy.test.arn
  identities {
    user = [aws_quicksight_user.test.user_name]
  }
}
`, rName, assignmentStatus, acctest.DefaultEmailAddress)
}
