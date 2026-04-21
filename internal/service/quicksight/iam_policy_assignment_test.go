// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package quicksight_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfquicksight "github.com/hashicorp/terraform-provider-aws/internal/service/quicksight"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccQuickSightIAMPolicyAssignment_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var assignment awstypes.IAMPolicyAssignment
	resourceName := "aws_quicksight_iam_policy_assignment.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIAMPolicyAssignmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIAMPolicyAssignmentConfig_basic(rName, string(awstypes.AssignmentStatusEnabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMPolicyAssignmentExists(ctx, t, resourceName, &assignment),
					resource.TestCheckResourceAttr(resourceName, "assignment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "assignment_status", string(awstypes.AssignmentStatusEnabled)),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamespace, tfquicksight.DefaultNamespace),
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
	var assignment awstypes.IAMPolicyAssignment
	resourceName := "aws_quicksight_iam_policy_assignment.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIAMPolicyAssignmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIAMPolicyAssignmentConfig_basic(rName, string(awstypes.AssignmentStatusEnabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMPolicyAssignmentExists(ctx, t, resourceName, &assignment),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfquicksight.ResourceIAMPolicyAssignment, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccQuickSightIAMPolicyAssignment_assignmentStatus(t *testing.T) {
	ctx := acctest.Context(t)
	var assignment awstypes.IAMPolicyAssignment
	resourceName := "aws_quicksight_iam_policy_assignment.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIAMPolicyAssignmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIAMPolicyAssignmentConfig_basic(rName, string(awstypes.AssignmentStatusDraft)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMPolicyAssignmentExists(ctx, t, resourceName, &assignment),
					resource.TestCheckResourceAttr(resourceName, "assignment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "assignment_status", string(awstypes.AssignmentStatusDraft)),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamespace, tfquicksight.DefaultNamespace),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccIAMPolicyAssignmentConfig_basic(rName, string(awstypes.AssignmentStatusEnabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMPolicyAssignmentExists(ctx, t, resourceName, &assignment),
					resource.TestCheckResourceAttr(resourceName, "assignment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "assignment_status", string(awstypes.AssignmentStatusEnabled)),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamespace, tfquicksight.DefaultNamespace),
				),
			},
			{
				Config: testAccIAMPolicyAssignmentConfig_basic(rName, string(awstypes.AssignmentStatusDisabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMPolicyAssignmentExists(ctx, t, resourceName, &assignment),
					resource.TestCheckResourceAttr(resourceName, "assignment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "assignment_status", string(awstypes.AssignmentStatusDisabled)),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamespace, tfquicksight.DefaultNamespace),
				),
			},
		},
	})
}

func TestAccQuickSightIAMPolicyAssignment_identities(t *testing.T) {
	ctx := acctest.Context(t)
	var assignment awstypes.IAMPolicyAssignment
	resourceName := "aws_quicksight_iam_policy_assignment.test"
	policyResourceName := "aws_iam_policy.test"
	userResourceName := "aws_quicksight_user.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.QuickSightServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIAMPolicyAssignmentDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccIAMPolicyAssignmentConfig_identities(rName, string(awstypes.AssignmentStatusEnabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIAMPolicyAssignmentExists(ctx, t, resourceName, &assignment),
					resource.TestCheckResourceAttr(resourceName, "assignment_name", rName),
					resource.TestCheckResourceAttr(resourceName, "assignment_status", string(awstypes.AssignmentStatusEnabled)),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamespace, tfquicksight.DefaultNamespace),
					resource.TestCheckResourceAttr(resourceName, "identities.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "identities.0.user.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "identities.0.user.0", userResourceName, names.AttrUserName),
					resource.TestCheckResourceAttrPair(resourceName, "policy_arn", policyResourceName, names.AttrARN),
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

func testAccCheckIAMPolicyAssignmentExists(ctx context.Context, t *testing.T, n string, v *awstypes.IAMPolicyAssignment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).QuickSightClient(ctx)

		output, err := tfquicksight.FindIAMPolicyAssignmentByThreePartKey(ctx, conn, rs.Primary.Attributes[names.AttrAWSAccountID], rs.Primary.Attributes[names.AttrNamespace], rs.Primary.Attributes["assignment_name"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckIAMPolicyAssignmentDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).QuickSightClient(ctx)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_quicksight_iam_policy_assignment" {
				continue
			}

			_, err := tfquicksight.FindIAMPolicyAssignmentByThreePartKey(ctx, conn, rs.Primary.Attributes[names.AttrAWSAccountID], rs.Primary.Attributes[names.AttrNamespace], rs.Primary.Attributes["assignment_name"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("QuickSight IAM Policy Assignment (%s) still exists", rs.Primary.ID)
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
