// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMUserPolicyAttachmentsExclusive_basic(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_user_policy_attachments_exclusive.test"
	userResourceName := "aws_iam_user.test"
	attachmentResourceName := "aws_iam_user_policy_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPolicyAttachmentsExclusiveDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPolicyAttachmentsExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserPolicyAttachmentExists(ctx, attachmentResourceName),
					testAccCheckUserPolicyAttachmentCount(ctx, rName, 1),
					testAccCheckUserPolicyAttachmentsExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUserName, userResourceName, names.AttrName),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_arns.*", attachmentResourceName, "policy_arn"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccUserPolicyAttachmentsExclusiveImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrUserName,
			},
		},
	})
}

func TestAccIAMUserPolicyAttachmentsExclusive_disappears_User(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_user_policy_attachments_exclusive.test"
	userResourceName := "aws_iam_user.test"
	attachmentResourceName := "aws_iam_user_policy_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPolicyAttachmentsExclusiveDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPolicyAttachmentsExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserPolicyAttachmentExists(ctx, attachmentResourceName),
					testAccCheckUserPolicyAttachmentCount(ctx, rName, 1),
					testAccCheckUserPolicyAttachmentsExclusiveExists(ctx, resourceName),
					// Managed policies must be detached before user can be deleted
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfiam.ResourceUserPolicyAttachment(), attachmentResourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfiam.ResourceUser(), userResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIAMUserPolicyAttachmentsExclusive_disappears_Policy(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_user_policy_attachments_exclusive.test"
	policyResourceName := "aws_iam_policy.test"
	attachmentResourceName := "aws_iam_user_policy_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPolicyAttachmentsExclusiveDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPolicyAttachmentsExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserPolicyAttachmentExists(ctx, attachmentResourceName),
					testAccCheckUserPolicyAttachmentCount(ctx, rName, 1),
					testAccCheckUserPolicyAttachmentsExclusiveExists(ctx, resourceName),
					// Managed policy must be detached before it can be deleted
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfiam.ResourceUserPolicyAttachment(), attachmentResourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfiam.ResourcePolicy(), policyResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIAMUserPolicyAttachmentsExclusive_multiple(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_user_policy_attachments_exclusive.test"
	userResourceName := "aws_iam_user.test"
	attachmentResourceName := "aws_iam_user_policy_attachment.test"
	attachmentResourceName2 := "aws_iam_user_policy_attachment.test2"
	attachmentResourceName3 := "aws_iam_user_policy_attachment.test3"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPolicyAttachmentsExclusiveDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPolicyAttachmentsExclusiveConfig_multiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserPolicyAttachmentExists(ctx, attachmentResourceName),
					testAccCheckUserPolicyAttachmentExists(ctx, attachmentResourceName2),
					testAccCheckUserPolicyAttachmentExists(ctx, attachmentResourceName3),
					testAccCheckUserPolicyAttachmentCount(ctx, rName, 3),
					testAccCheckUserPolicyAttachmentsExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUserName, userResourceName, names.AttrName),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_arns.*", attachmentResourceName, "policy_arn"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_arns.*", attachmentResourceName2, "policy_arn"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_arns.*", attachmentResourceName3, "policy_arn"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccUserPolicyAttachmentsExclusiveImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrUserName,
			},
			{
				Config: testAccUserPolicyAttachmentsExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserPolicyAttachmentExists(ctx, attachmentResourceName),
					testAccCheckUserPolicyAttachmentCount(ctx, rName, 1),
					testAccCheckUserPolicyAttachmentsExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUserName, userResourceName, names.AttrName),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_arns.*", attachmentResourceName, "policy_arn"),
				),
			},
		},
	})
}

func TestAccIAMUserPolicyAttachmentsExclusive_empty(t *testing.T) {
	ctx := acctest.Context(t)

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_user_policy_attachments_exclusive.test"
	userResourceName := "aws_iam_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPolicyAttachmentsExclusiveDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPolicyAttachmentsExclusiveConfig_empty(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserPolicyAttachmentsExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUserName, userResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "policy_arns.#", acctest.Ct0),
				),
				// The empty `policy_arns` argument in the exclusive lock will remove the
				// managed policy defined in this configuration, so a diff is expected
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// A managed policy removed out of band should be recreated
func TestAccIAMUserPolicyAttachmentsExclusive_outOfBandRemoval(t *testing.T) {
	ctx := acctest.Context(t)

	var user types.User
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_user_policy_attachments_exclusive.test"
	userResourceName := "aws_iam_user.test"
	attachmentResourceName := "aws_iam_user_policy_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPolicyAttachmentsExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, userResourceName, &user),
					testAccCheckUserPolicyAttachmentExists(ctx, attachmentResourceName),
					testAccCheckUserPolicyAttachmentCount(ctx, rName, 1),
					testAccCheckUserPolicyAttachmentsExclusiveExists(ctx, resourceName),
					testAccCheckUserPolicyDetachManagedPolicy(ctx, &user, rName),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccUserPolicyAttachmentsExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, userResourceName, &user),
					testAccCheckUserPolicyAttachmentExists(ctx, attachmentResourceName),
					testAccCheckUserPolicyAttachmentCount(ctx, rName, 1),
					testAccCheckUserPolicyAttachmentsExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUserName, userResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "policy_arns.#", acctest.Ct1),
				),
			},
		},
	})
}

// A managed policy added out of band should be removed
func TestAccIAMUserPolicyAttachmentsExclusive_outOfBandAddition(t *testing.T) {
	ctx := acctest.Context(t)

	var user types.User
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	oobPolicyName := rName + "-out-of-band"
	resourceName := "aws_iam_user_policy_attachments_exclusive.test"
	userResourceName := "aws_iam_user.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPolicyAttachmentsExclusiveConfig_outOfBandAddition(rName, oobPolicyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, userResourceName, &user),
					testAccCheckUserPolicyAttachmentsExclusiveExists(ctx, resourceName),
					testAccCheckUserPolicyAttachManagedPolicy(ctx, &user, oobPolicyName),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccUserPolicyAttachmentsExclusiveConfig_outOfBandAddition(rName, oobPolicyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, userResourceName, &user),
					testAccCheckUserPolicyAttachmentsExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUserName, userResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "policy_arns.#", acctest.Ct1),
				),
			},
		},
	})
}

func testAccCheckUserPolicyAttachmentsExclusiveDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iam_user_policy_attachments_exclusive" {
				continue
			}

			userName := rs.Primary.Attributes[names.AttrUserName]
			_, err := tfiam.FindUserPolicyAttachmentsByName(ctx, conn, userName)
			if errs.IsA[*types.NoSuchEntityException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.IAM, create.ErrActionCheckingDestroyed, tfiam.ResNameUserPolicyAttachmentsExclusive, userName, err)
			}

			return create.Error(names.IAM, create.ErrActionCheckingDestroyed, tfiam.ResNameUserPolicyAttachmentsExclusive, userName, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckUserPolicyAttachmentsExclusiveExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.IAM, create.ErrActionCheckingExistence, tfiam.ResNameUserPolicyAttachmentsExclusive, name, errors.New("not found"))
		}

		userName := rs.Primary.Attributes[names.AttrUserName]
		if userName == "" {
			return create.Error(names.IAM, create.ErrActionCheckingExistence, tfiam.ResNameUserPolicyAttachmentsExclusive, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)
		out, err := tfiam.FindUserPolicyAttachmentsByName(ctx, conn, userName)
		if err != nil {
			return create.Error(names.IAM, create.ErrActionCheckingExistence, tfiam.ResNameUserPolicyAttachmentsExclusive, userName, err)
		}

		policyCount := rs.Primary.Attributes["policy_arns.#"]
		if policyCount != fmt.Sprint(len(out)) {
			return create.Error(names.IAM, create.ErrActionCheckingExistence, tfiam.ResNameUserPolicyAttachmentsExclusive, userName, errors.New("unexpected policy_arns count"))
		}

		return nil
	}
}

func testAccUserPolicyAttachmentsExclusiveImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes[names.AttrUserName], nil
	}
}

func testAccCheckUserPolicyDetachManagedPolicy(ctx context.Context, user *types.User, policyName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

		var managedARN string
		input := &iam.ListAttachedUserPoliciesInput{
			UserName: user.UserName,
		}

		pages := iam.NewListAttachedUserPoliciesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)

			if err != nil && !errs.IsA[*types.NoSuchEntityException](err) {
				return fmt.Errorf("finding managed policy (%s): %w", policyName, err)
			}

			if err != nil {
				return err
			}

			for _, v := range page.AttachedPolicies {
				if *v.PolicyName == policyName {
					managedARN = *v.PolicyArn
					break
				}
			}
		}

		if managedARN == "" {
			return fmt.Errorf("managed policy (%s) not found", policyName)
		}

		_, err := conn.DetachUserPolicy(ctx, &iam.DetachUserPolicyInput{
			PolicyArn: aws.String(managedARN),
			UserName:  user.UserName,
		})

		return err
	}
}

func testAccCheckUserPolicyAttachManagedPolicy(ctx context.Context, user *types.User, policyName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

		var managedARN string
		input := &iam.ListPoliciesInput{
			PathPrefix:        aws.String("/tf-testing/"),
			PolicyUsageFilter: types.PolicyUsageType("PermissionsPolicy"),
			Scope:             types.PolicyScopeType("Local"),
		}

		pages := iam.NewListPoliciesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)

			if err != nil && !errs.IsA[*types.NoSuchEntityException](err) {
				return fmt.Errorf("finding managed policy (%s): %w", policyName, err)
			}

			if err != nil {
				return err
			}

			for _, v := range page.Policies {
				if *v.PolicyName == policyName {
					managedARN = *v.Arn
					break
				}
			}
		}

		if managedARN == "" {
			return fmt.Errorf("managed policy (%s) not found", policyName)
		}

		_, err := conn.AttachUserPolicy(ctx, &iam.AttachUserPolicyInput{
			PolicyArn: aws.String(managedARN),
			UserName:  user.UserName,
		})

		return err
	}
}

func testAccUserPolicyAttachmentsExclusiveConfigBase(rName string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "managed" {
  statement {
    actions   = ["sts:GetCallerIdentity"]
    resources = ["*"]
  }
}

resource "aws_iam_user" "test" {
  name = %[1]q
}

resource "aws_iam_policy" "test" {
  name   = %[1]q
  policy = data.aws_iam_policy_document.managed.json
}

resource "aws_iam_user_policy_attachment" "test" {
  user       = aws_iam_user.test.name
  policy_arn = aws_iam_policy.test.arn
}
`, rName)
}

func testAccUserPolicyAttachmentsExclusiveConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccUserPolicyAttachmentsExclusiveConfigBase(rName),
		`
resource "aws_iam_user_policy_attachments_exclusive" "test" {
  user_name   = aws_iam_user.test.name
  policy_arns = [aws_iam_user_policy_attachment.test.policy_arn]
}
`)
}

func testAccUserPolicyAttachmentsExclusiveConfig_multiple(rName string) string {
	return acctest.ConfigCompose(
		testAccUserPolicyAttachmentsExclusiveConfigBase(rName),
		fmt.Sprintf(`
resource "aws_iam_policy" "test2" {
  name   = "%[1]s-2"
  policy = data.aws_iam_policy_document.managed.json
}

resource "aws_iam_user_policy_attachment" "test2" {
  user       = aws_iam_user.test.name
  policy_arn = aws_iam_policy.test2.arn
}

resource "aws_iam_policy" "test3" {
  name   = "%[1]s-3"
  policy = data.aws_iam_policy_document.managed.json
}

resource "aws_iam_user_policy_attachment" "test3" {
  user       = aws_iam_user.test.name
  policy_arn = aws_iam_policy.test3.arn
}

resource "aws_iam_user_policy_attachments_exclusive" "test" {
  user_name = aws_iam_user.test.name
  policy_arns = [
    aws_iam_user_policy_attachment.test.policy_arn,
    aws_iam_user_policy_attachment.test2.policy_arn,
    aws_iam_user_policy_attachment.test3.policy_arn,
  ]
}
`, rName))
}

func testAccUserPolicyAttachmentsExclusiveConfig_empty(rName string) string {
	return acctest.ConfigCompose(
		testAccUserPolicyAttachmentsExclusiveConfigBase(rName),
		`
resource "aws_iam_user_policy_attachments_exclusive" "test" {
  # Wait until the managed policy is attached, then provision
  # the exclusive lock which will remove it. This creates a diff on
  # on the next plan (to re-create aws_iam_user_policy_attachment.test)
  # which the test can check for.
  depends_on = [aws_iam_user_policy_attachment.test]

  user_name   = aws_iam_user.test.name
  policy_arns = []
}
`)
}

func testAccUserPolicyAttachmentsExclusiveConfig_outOfBandAddition(rName, oobPolicyName string) string {
	return acctest.ConfigCompose(
		testAccUserPolicyAttachmentsExclusiveConfigBase(rName),
		fmt.Sprintf(`
# This will be attached out-of-band via a test check helper
resource "aws_iam_policy" "test2" {
  name   = %[1]q
  path   = "/tf-testing/"
  policy = data.aws_iam_policy_document.managed.json
}

resource "aws_iam_user_policy_attachments_exclusive" "test" {
  user_name   = aws_iam_user.test.name
  policy_arns = [aws_iam_user_policy_attachment.test.policy_arn]
}
`, oobPolicyName))
}
