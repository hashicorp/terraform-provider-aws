// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMRolePolicyAttachmentsExclusive_basic(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_iam_role_policy_attachments_exclusive.test"
	roleResourceName := "aws_iam_role.test"
	attachmentResourceName := "aws_iam_role_policy_attachment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRolePolicyAttachmentsExclusiveDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRolePolicyAttachmentsExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRolePolicyAttachmentExists(ctx, t, attachmentResourceName),
					testAccCheckRolePolicyAttachmentCount(ctx, t, rName, 1),
					testAccCheckRolePolicyAttachmentsExclusiveExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "role_name", roleResourceName, names.AttrName),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_arns.*", attachmentResourceName, "policy_arn"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "role_name"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "role_name",
			},
		},
	})
}

func TestAccIAMRolePolicyAttachmentsExclusive_disappears_Role(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_iam_role_policy_attachments_exclusive.test"
	roleResourceName := "aws_iam_role.test"
	attachmentResourceName := "aws_iam_role_policy_attachment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRolePolicyAttachmentsExclusiveDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRolePolicyAttachmentsExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRolePolicyAttachmentExists(ctx, t, attachmentResourceName),
					testAccCheckRolePolicyAttachmentCount(ctx, t, rName, 1),
					testAccCheckRolePolicyAttachmentsExclusiveExists(ctx, t, resourceName),
					// Managed policies must be detached before role can be deleted
					acctest.CheckSDKResourceDisappears(ctx, t, tfiam.ResourceRolePolicyAttachment(), attachmentResourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfiam.ResourceRole(), roleResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIAMRolePolicyAttachmentsExclusive_disappears_Policy(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_iam_role_policy_attachments_exclusive.test"
	policyResourceName := "aws_iam_policy.test"
	attachmentResourceName := "aws_iam_role_policy_attachment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRolePolicyAttachmentsExclusiveDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRolePolicyAttachmentsExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRolePolicyAttachmentExists(ctx, t, attachmentResourceName),
					testAccCheckRolePolicyAttachmentCount(ctx, t, rName, 1),
					testAccCheckRolePolicyAttachmentsExclusiveExists(ctx, t, resourceName),
					// Managed policy must be detached before it can be deleted
					acctest.CheckSDKResourceDisappears(ctx, t, tfiam.ResourceRolePolicyAttachment(), attachmentResourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfiam.ResourcePolicy(), policyResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIAMRolePolicyAttachmentsExclusive_multiple(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_iam_role_policy_attachments_exclusive.test"
	roleResourceName := "aws_iam_role.test"
	attachmentResourceName := "aws_iam_role_policy_attachment.test"
	attachmentResourceName2 := "aws_iam_role_policy_attachment.test2"
	attachmentResourceName3 := "aws_iam_role_policy_attachment.test3"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRolePolicyAttachmentsExclusiveDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRolePolicyAttachmentsExclusiveConfig_multiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRolePolicyAttachmentExists(ctx, t, attachmentResourceName),
					testAccCheckRolePolicyAttachmentExists(ctx, t, attachmentResourceName2),
					testAccCheckRolePolicyAttachmentExists(ctx, t, attachmentResourceName3),
					testAccCheckRolePolicyAttachmentCount(ctx, t, rName, 3),
					testAccCheckRolePolicyAttachmentsExclusiveExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "role_name", roleResourceName, names.AttrName),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_arns.*", attachmentResourceName, "policy_arn"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_arns.*", attachmentResourceName2, "policy_arn"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_arns.*", attachmentResourceName3, "policy_arn"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "role_name"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "role_name",
			},
			{
				Config: testAccRolePolicyAttachmentsExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRolePolicyAttachmentExists(ctx, t, attachmentResourceName),
					testAccCheckRolePolicyAttachmentCount(ctx, t, rName, 1),
					testAccCheckRolePolicyAttachmentsExclusiveExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "role_name", roleResourceName, names.AttrName),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_arns.*", attachmentResourceName, "policy_arn"),
				),
			},
		},
	})
}

func TestAccIAMRolePolicyAttachmentsExclusive_empty(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_iam_role_policy_attachments_exclusive.test"
	roleResourceName := "aws_iam_role.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRolePolicyAttachmentsExclusiveDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRolePolicyAttachmentsExclusiveConfig_empty(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRolePolicyAttachmentsExclusiveExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "role_name", roleResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "policy_arns.#", "0"),
				),
				// The empty `policy_arns` argument in the exclusive lock will remove the
				// managed policy defined in this configuration, so a diff is expected
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// A managed policy removed out of band should be recreated
func TestAccIAMRolePolicyAttachmentsExclusive_outOfBandRemoval(t *testing.T) {
	ctx := acctest.Context(t)

	var role types.Role
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_iam_role_policy_attachments_exclusive.test"
	roleResourceName := "aws_iam_role.test"
	attachmentResourceName := "aws_iam_role_policy_attachment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRolePolicyAttachmentsExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, t, roleResourceName, &role),
					testAccCheckRolePolicyAttachmentExists(ctx, t, attachmentResourceName),
					testAccCheckRolePolicyAttachmentCount(ctx, t, rName, 1),
					testAccCheckRolePolicyAttachmentsExclusiveExists(ctx, t, resourceName),
					testAccCheckRolePolicyDetachManagedPolicy(ctx, t, &role, rName),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccRolePolicyAttachmentsExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, t, roleResourceName, &role),
					testAccCheckRolePolicyAttachmentExists(ctx, t, attachmentResourceName),
					testAccCheckRolePolicyAttachmentCount(ctx, t, rName, 1),
					testAccCheckRolePolicyAttachmentsExclusiveExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "role_name", roleResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "policy_arns.#", "1"),
				),
			},
		},
	})
}

// A managed policy added out of band should be removed
func TestAccIAMRolePolicyAttachmentsExclusive_outOfBandAddition(t *testing.T) {
	ctx := acctest.Context(t)

	var role types.Role
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	oobPolicyName := rName + "-out-of-band"
	resourceName := "aws_iam_role_policy_attachments_exclusive.test"
	roleResourceName := "aws_iam_role.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRolePolicyAttachmentsExclusiveConfig_outOfBandAddition(rName, oobPolicyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, t, roleResourceName, &role),
					testAccCheckRolePolicyAttachmentsExclusiveExists(ctx, t, resourceName),
					testAccCheckRolePolicyAttachManagedPolicy(ctx, t, &role, oobPolicyName),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccRolePolicyAttachmentsExclusiveConfig_outOfBandAddition(rName, oobPolicyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, t, roleResourceName, &role),
					testAccCheckRolePolicyAttachmentsExclusiveExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "role_name", roleResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "policy_arns.#", "1"),
				),
			},
		},
	})
}

func testAccCheckRolePolicyAttachmentsExclusiveDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).IAMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iam_role_policy_attachments_exclusive" {
				continue
			}

			roleName := rs.Primary.Attributes["role_name"]
			_, err := tfiam.FindRolePolicyAttachmentsByName(ctx, conn, roleName)
			if errs.IsA[*types.NoSuchEntityException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.IAM, create.ErrActionCheckingDestroyed, tfiam.ResNameRolePolicyAttachmentsExclusive, roleName, err)
			}

			return create.Error(names.IAM, create.ErrActionCheckingDestroyed, tfiam.ResNameRolePolicyAttachmentsExclusive, roleName, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckRolePolicyAttachmentsExclusiveExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.IAM, create.ErrActionCheckingExistence, tfiam.ResNameRolePolicyAttachmentsExclusive, name, errors.New("not found"))
		}

		roleName := rs.Primary.Attributes["role_name"]
		if roleName == "" {
			return create.Error(names.IAM, create.ErrActionCheckingExistence, tfiam.ResNameRolePolicyAttachmentsExclusive, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).IAMClient(ctx)
		out, err := tfiam.FindRolePolicyAttachmentsByName(ctx, conn, roleName)
		if err != nil {
			return create.Error(names.IAM, create.ErrActionCheckingExistence, tfiam.ResNameRolePolicyAttachmentsExclusive, roleName, err)
		}

		policyCount := rs.Primary.Attributes["policy_arns.#"]
		if policyCount != strconv.Itoa(len(out)) {
			return create.Error(names.IAM, create.ErrActionCheckingExistence, tfiam.ResNameRolePolicyAttachmentsExclusive, roleName, errors.New("unexpected policy_arns count"))
		}

		return nil
	}
}

func testAccRolePolicyAttachmentsExclusiveConfigBase(rName string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "trust" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["ec2.amazonaws.com"]
    }
  }
}

data "aws_iam_policy_document" "managed" {
  statement {
    actions   = ["s3:ListBucket"]
    resources = ["*"]
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.trust.json
}

resource "aws_iam_policy" "test" {
  name   = %[1]q
  policy = data.aws_iam_policy_document.managed.json
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = aws_iam_policy.test.arn
}
`, rName)
}

func testAccRolePolicyAttachmentsExclusiveConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccRolePolicyAttachmentsExclusiveConfigBase(rName),
		`
resource "aws_iam_role_policy_attachments_exclusive" "test" {
  role_name   = aws_iam_role.test.name
  policy_arns = [aws_iam_role_policy_attachment.test.policy_arn]
}
`)
}

func testAccRolePolicyAttachmentsExclusiveConfig_multiple(rName string) string {
	return acctest.ConfigCompose(
		testAccRolePolicyAttachmentsExclusiveConfigBase(rName),
		fmt.Sprintf(`
resource "aws_iam_policy" "test2" {
  name   = "%[1]s-2"
  policy = data.aws_iam_policy_document.managed.json
}

resource "aws_iam_role_policy_attachment" "test2" {
  role       = aws_iam_role.test.name
  policy_arn = aws_iam_policy.test2.arn
}

resource "aws_iam_policy" "test3" {
  name   = "%[1]s-3"
  policy = data.aws_iam_policy_document.managed.json
}

resource "aws_iam_role_policy_attachment" "test3" {
  role       = aws_iam_role.test.name
  policy_arn = aws_iam_policy.test3.arn
}

resource "aws_iam_role_policy_attachments_exclusive" "test" {
  role_name = aws_iam_role.test.name
  policy_arns = [
    aws_iam_role_policy_attachment.test.policy_arn,
    aws_iam_role_policy_attachment.test2.policy_arn,
    aws_iam_role_policy_attachment.test3.policy_arn,
  ]
}
`, rName))
}

func testAccRolePolicyAttachmentsExclusiveConfig_empty(rName string) string {
	return acctest.ConfigCompose(
		testAccRolePolicyAttachmentsExclusiveConfigBase(rName),
		`
resource "aws_iam_role_policy_attachments_exclusive" "test" {
  # Wait until the managed policy is attached, then provision
  # the exclusive lock which will remove it. This creates a diff on
  # on the next plan (to re-create aws_iam_role_policy_attachment.test)
  # which the test can check for.
  depends_on = [aws_iam_role_policy_attachment.test]

  role_name   = aws_iam_role.test.name
  policy_arns = []
}
`)
}

func testAccRolePolicyAttachmentsExclusiveConfig_outOfBandAddition(rName, oobPolicyName string) string {
	return acctest.ConfigCompose(
		testAccRolePolicyAttachmentsExclusiveConfigBase(rName),
		fmt.Sprintf(`
# This will be attached out-of-band via a test check helper
resource "aws_iam_policy" "test2" {
  name   = %[1]q
  path   = "/tf-testing/"
  policy = data.aws_iam_policy_document.managed.json
}

resource "aws_iam_role_policy_attachments_exclusive" "test" {
  role_name   = aws_iam_role.test.name
  policy_arns = [aws_iam_role_policy_attachment.test.policy_arn]
}
`, oobPolicyName))
}
