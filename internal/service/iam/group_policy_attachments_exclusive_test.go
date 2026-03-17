// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfiam "github.com/hashicorp/terraform-provider-aws/internal/service/iam"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMGroupPolicyAttachmentsExclusive_basic(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_iam_group_policy_attachments_exclusive.test"
	groupResourceName := "aws_iam_group.test"
	attachmentResourceName := "aws_iam_group_policy_attachment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupPolicyAttachmentsExclusiveDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupPolicyAttachmentsExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupPolicyAttachmentExists(ctx, t, attachmentResourceName),
					testAccCheckGroupPolicyAttachmentCount(ctx, t, rName, 1),
					testAccCheckGroupPolicyAttachmentsExclusiveExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrGroupName, groupResourceName, names.AttrName),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_arns.*", attachmentResourceName, "policy_arn"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrGroupName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrGroupName,
			},
		},
	})
}

func TestAccIAMGroupPolicyAttachmentsExclusive_disappears_Group(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_iam_group_policy_attachments_exclusive.test"
	groupResourceName := "aws_iam_group.test"
	attachmentResourceName := "aws_iam_group_policy_attachment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupPolicyAttachmentsExclusiveDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupPolicyAttachmentsExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupPolicyAttachmentExists(ctx, t, attachmentResourceName),
					testAccCheckGroupPolicyAttachmentCount(ctx, t, rName, 1),
					testAccCheckGroupPolicyAttachmentsExclusiveExists(ctx, t, resourceName),
					// Managed policies must be detached before group can be deleted
					acctest.CheckSDKResourceDisappears(ctx, t, tfiam.ResourceGroupPolicyAttachment(), attachmentResourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfiam.ResourceGroup(), groupResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIAMGroupPolicyAttachmentsExclusive_disappears_Policy(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_iam_group_policy_attachments_exclusive.test"
	policyResourceName := "aws_iam_policy.test"
	attachmentResourceName := "aws_iam_group_policy_attachment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupPolicyAttachmentsExclusiveDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupPolicyAttachmentsExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupPolicyAttachmentExists(ctx, t, attachmentResourceName),
					testAccCheckGroupPolicyAttachmentCount(ctx, t, rName, 1),
					testAccCheckGroupPolicyAttachmentsExclusiveExists(ctx, t, resourceName),
					// Managed policy must be detached before it can be deleted
					acctest.CheckSDKResourceDisappears(ctx, t, tfiam.ResourceGroupPolicyAttachment(), attachmentResourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfiam.ResourcePolicy(), policyResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIAMGroupPolicyAttachmentsExclusive_multiple(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_iam_group_policy_attachments_exclusive.test"
	groupResourceName := "aws_iam_group.test"
	attachmentResourceName := "aws_iam_group_policy_attachment.test"
	attachmentResourceName2 := "aws_iam_group_policy_attachment.test2"
	attachmentResourceName3 := "aws_iam_group_policy_attachment.test3"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupPolicyAttachmentsExclusiveDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupPolicyAttachmentsExclusiveConfig_multiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupPolicyAttachmentExists(ctx, t, attachmentResourceName),
					testAccCheckGroupPolicyAttachmentExists(ctx, t, attachmentResourceName2),
					testAccCheckGroupPolicyAttachmentExists(ctx, t, attachmentResourceName3),
					testAccCheckGroupPolicyAttachmentCount(ctx, t, rName, 3),
					testAccCheckGroupPolicyAttachmentsExclusiveExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrGroupName, groupResourceName, names.AttrName),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_arns.*", attachmentResourceName, "policy_arn"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_arns.*", attachmentResourceName2, "policy_arn"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_arns.*", attachmentResourceName3, "policy_arn"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrGroupName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrGroupName,
			},
			{
				Config: testAccGroupPolicyAttachmentsExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupPolicyAttachmentExists(ctx, t, attachmentResourceName),
					testAccCheckGroupPolicyAttachmentCount(ctx, t, rName, 1),
					testAccCheckGroupPolicyAttachmentsExclusiveExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrGroupName, groupResourceName, names.AttrName),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_arns.*", attachmentResourceName, "policy_arn"),
				),
			},
		},
	})
}

func TestAccIAMGroupPolicyAttachmentsExclusive_empty(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_iam_group_policy_attachments_exclusive.test"
	groupResourceName := "aws_iam_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupPolicyAttachmentsExclusiveDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupPolicyAttachmentsExclusiveConfig_empty(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupPolicyAttachmentsExclusiveExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrGroupName, groupResourceName, names.AttrName),
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
func TestAccIAMGroupPolicyAttachmentsExclusive_outOfBandRemoval(t *testing.T) {
	ctx := acctest.Context(t)

	var group types.Group
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_iam_group_policy_attachments_exclusive.test"
	groupResourceName := "aws_iam_group.test"
	attachmentResourceName := "aws_iam_group_policy_attachment.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupPolicyAttachmentsExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, t, groupResourceName, &group),
					testAccCheckGroupPolicyAttachmentExists(ctx, t, attachmentResourceName),
					testAccCheckGroupPolicyAttachmentCount(ctx, t, rName, 1),
					testAccCheckGroupPolicyAttachmentsExclusiveExists(ctx, t, resourceName),
					testAccCheckGroupPolicyDetachManagedPolicy(ctx, t, &group, rName),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccGroupPolicyAttachmentsExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, t, groupResourceName, &group),
					testAccCheckGroupPolicyAttachmentExists(ctx, t, attachmentResourceName),
					testAccCheckGroupPolicyAttachmentCount(ctx, t, rName, 1),
					testAccCheckGroupPolicyAttachmentsExclusiveExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrGroupName, groupResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "policy_arns.#", "1"),
				),
			},
		},
	})
}

// A managed policy added out of band should be removed
func TestAccIAMGroupPolicyAttachmentsExclusive_outOfBandAddition(t *testing.T) {
	ctx := acctest.Context(t)

	var group types.Group
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	oobPolicyName := rName + "-out-of-band"
	resourceName := "aws_iam_group_policy_attachments_exclusive.test"
	groupResourceName := "aws_iam_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupPolicyAttachmentsExclusiveConfig_outOfBandAddition(rName, oobPolicyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, t, groupResourceName, &group),
					testAccCheckGroupPolicyAttachmentsExclusiveExists(ctx, t, resourceName),
					testAccCheckGroupPolicyAttachManagedPolicy(ctx, t, &group, oobPolicyName),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccGroupPolicyAttachmentsExclusiveConfig_outOfBandAddition(rName, oobPolicyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, t, groupResourceName, &group),
					testAccCheckGroupPolicyAttachmentsExclusiveExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrGroupName, groupResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "policy_arns.#", "1"),
				),
			},
		},
	})
}

func testAccCheckGroupPolicyAttachmentsExclusiveDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).IAMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iam_group_policy_attachments_exclusive" {
				continue
			}

			groupName := rs.Primary.Attributes[names.AttrGroupName]
			_, err := tfiam.FindGroupPolicyAttachmentsByName(ctx, conn, groupName)
			if errs.IsA[*types.NoSuchEntityException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.IAM, create.ErrActionCheckingDestroyed, tfiam.ResNameGroupPolicyAttachmentsExclusive, groupName, err)
			}

			return create.Error(names.IAM, create.ErrActionCheckingDestroyed, tfiam.ResNameGroupPolicyAttachmentsExclusive, groupName, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckGroupPolicyAttachmentsExclusiveExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.IAM, create.ErrActionCheckingExistence, tfiam.ResNameGroupPolicyAttachmentsExclusive, name, errors.New("not found"))
		}

		groupName := rs.Primary.Attributes[names.AttrGroupName]
		if groupName == "" {
			return create.Error(names.IAM, create.ErrActionCheckingExistence, tfiam.ResNameGroupPolicyAttachmentsExclusive, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).IAMClient(ctx)
		out, err := tfiam.FindGroupPolicyAttachmentsByName(ctx, conn, groupName)
		if err != nil {
			return create.Error(names.IAM, create.ErrActionCheckingExistence, tfiam.ResNameGroupPolicyAttachmentsExclusive, groupName, err)
		}

		policyCount := rs.Primary.Attributes["policy_arns.#"]
		if policyCount != strconv.Itoa(len(out)) {
			return create.Error(names.IAM, create.ErrActionCheckingExistence, tfiam.ResNameGroupPolicyAttachmentsExclusive, groupName, errors.New("unexpected policy_arns count"))
		}

		return nil
	}
}

func testAccCheckGroupPolicyDetachManagedPolicy(ctx context.Context, t *testing.T, group *types.Group, policyName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).IAMClient(ctx)

		var managedARN string
		input := &iam.ListAttachedGroupPoliciesInput{
			GroupName: group.GroupName,
		}

		pages := iam.NewListAttachedGroupPoliciesPaginator(conn, input)
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

		_, err := conn.DetachGroupPolicy(ctx, &iam.DetachGroupPolicyInput{
			PolicyArn: aws.String(managedARN),
			GroupName: group.GroupName,
		})

		return err
	}
}

func testAccCheckGroupPolicyAttachManagedPolicy(ctx context.Context, t *testing.T, group *types.Group, policyName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).IAMClient(ctx)

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

		_, err := conn.AttachGroupPolicy(ctx, &iam.AttachGroupPolicyInput{
			PolicyArn: aws.String(managedARN),
			GroupName: group.GroupName,
		})

		return err
	}
}

func testAccGroupPolicyAttachmentsExclusiveConfigBase(rName string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "managed" {
  statement {
    actions   = ["sts:GetCallerIdentity"]
    resources = ["*"]
  }
}

resource "aws_iam_group" "test" {
  name = %[1]q
}

resource "aws_iam_policy" "test" {
  name   = %[1]q
  policy = data.aws_iam_policy_document.managed.json
}

resource "aws_iam_group_policy_attachment" "test" {
  group      = aws_iam_group.test.name
  policy_arn = aws_iam_policy.test.arn
}
`, rName)
}

func testAccGroupPolicyAttachmentsExclusiveConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccGroupPolicyAttachmentsExclusiveConfigBase(rName), `
resource "aws_iam_group_policy_attachments_exclusive" "test" {
  group_name  = aws_iam_group.test.name
  policy_arns = [aws_iam_group_policy_attachment.test.policy_arn]
}
`,
	)
}

func testAccGroupPolicyAttachmentsExclusiveConfig_multiple(rName string) string {
	return acctest.ConfigCompose(
		testAccGroupPolicyAttachmentsExclusiveConfigBase(rName),
		fmt.Sprintf(`
resource "aws_iam_policy" "test2" {
  name   = "%[1]s-2"
  policy = data.aws_iam_policy_document.managed.json
}

resource "aws_iam_group_policy_attachment" "test2" {
  group      = aws_iam_group.test.name
  policy_arn = aws_iam_policy.test2.arn
}

resource "aws_iam_policy" "test3" {
  name   = "%[1]s-3"
  policy = data.aws_iam_policy_document.managed.json
}

resource "aws_iam_group_policy_attachment" "test3" {
  group      = aws_iam_group.test.name
  policy_arn = aws_iam_policy.test3.arn
}

resource "aws_iam_group_policy_attachments_exclusive" "test" {
  group_name = aws_iam_group.test.name
  policy_arns = [
    aws_iam_group_policy_attachment.test.policy_arn,
    aws_iam_group_policy_attachment.test2.policy_arn,
    aws_iam_group_policy_attachment.test3.policy_arn,
  ]
}
`, rName))
}

func testAccGroupPolicyAttachmentsExclusiveConfig_empty(rName string) string {
	return acctest.ConfigCompose(
		testAccGroupPolicyAttachmentsExclusiveConfigBase(rName), `
resource "aws_iam_group_policy_attachments_exclusive" "test" {
  # Wait until the managed policy is attached, then provision
  # the exclusive lock which will remove it. This creates a diff on
  # on the next plan (to re-create aws_iam_group_policy_attachment.test)
  # which the test can check for.
  depends_on = [aws_iam_group_policy_attachment.test]

  group_name  = aws_iam_group.test.name
  policy_arns = []
}
`,
	)
}

func testAccGroupPolicyAttachmentsExclusiveConfig_outOfBandAddition(rName, oobPolicyName string) string {
	return acctest.ConfigCompose(
		testAccGroupPolicyAttachmentsExclusiveConfigBase(rName),
		fmt.Sprintf(`
# This will be attached out-of-band via a test check helper
resource "aws_iam_policy" "test2" {
  name   = %[1]q
  path   = "/tf-testing/"
  policy = data.aws_iam_policy_document.managed.json
}

resource "aws_iam_group_policy_attachments_exclusive" "test" {
  group_name  = aws_iam_group.test.name
  policy_arns = [aws_iam_group_policy_attachment.test.policy_arn]
}
`, oobPolicyName))
}
