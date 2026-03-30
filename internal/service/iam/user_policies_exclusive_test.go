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

func TestAccIAMUserPoliciesExclusive_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var user types.User
	var userPolicy string
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_iam_user_policies_exclusive.test"
	userResourceName := "aws_iam_user.test"
	userPolicyResourceName := "aws_iam_user_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoliciesExclusiveDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoliciesExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, t, userResourceName, &user),
					testAccCheckUserPolicyExists(ctx, t, userPolicyResourceName, &userPolicy),
					testAccCheckUserPoliciesExclusiveExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUserName, userResourceName, names.AttrName),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_names.*", userPolicyResourceName, names.AttrName),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrUserName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrUserName,
			},
		},
	})
}

func TestAccIAMUserPoliciesExclusive_disappears_User(t *testing.T) {
	ctx := acctest.Context(t)

	var user types.User
	var userPolicy string
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_iam_user_policies_exclusive.test"
	userResourceName := "aws_iam_user.test"
	userPolicyResourceName := "aws_iam_user_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoliciesExclusiveDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoliciesExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, t, userResourceName, &user),
					testAccCheckUserPolicyExists(ctx, t, userPolicyResourceName, &userPolicy),
					testAccCheckUserPoliciesExclusiveExists(ctx, t, resourceName),
					// Inline policy must be deleted before the user can be
					acctest.CheckSDKResourceDisappears(ctx, t, tfiam.ResourceUserPolicy(), userPolicyResourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfiam.ResourceUser(), userResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIAMUserPoliciesExclusive_multiple(t *testing.T) {
	ctx := acctest.Context(t)

	var user types.User
	var userPolicy string
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_iam_user_policies_exclusive.test"
	userResourceName := "aws_iam_user.test"
	userPolicyResourceName := "aws_iam_user_policy.test"
	userPolicyResourceName2 := "aws_iam_user_policy.test2"
	userPolicyResourceName3 := "aws_iam_user_policy.test3"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoliciesExclusiveDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoliciesExclusiveConfig_multiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, t, userResourceName, &user),
					testAccCheckUserPolicyExists(ctx, t, userPolicyResourceName, &userPolicy),
					testAccCheckUserPoliciesExclusiveExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUserName, userResourceName, names.AttrName),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_names.*", userPolicyResourceName, names.AttrName),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_names.*", userPolicyResourceName2, names.AttrName),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_names.*", userPolicyResourceName3, names.AttrName),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrUserName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrUserName,
			},
			{
				Config: testAccUserPoliciesExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, t, userResourceName, &user),
					testAccCheckUserPolicyExists(ctx, t, userPolicyResourceName, &userPolicy),
					testAccCheckUserPoliciesExclusiveExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUserName, userResourceName, names.AttrName),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_names.*", userPolicyResourceName, names.AttrName),
				),
			},
		},
	})
}

func TestAccIAMUserPoliciesExclusive_empty(t *testing.T) {
	ctx := acctest.Context(t)

	var user types.User
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_iam_user_policies_exclusive.test"
	userResourceName := "aws_iam_user.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserPoliciesExclusiveDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoliciesExclusiveConfig_empty(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, t, userResourceName, &user),
					testAccCheckUserPoliciesExclusiveExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUserName, userResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "policy_names.#", "0"),
				),
				// The empty `policy_names` argument in the exclusive lock will remove the
				// inline policy defined in this configuration, so a diff is expected
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// An inline policy removed out of band should be recreated
func TestAccIAMUserPoliciesExclusive_outOfBandRemoval(t *testing.T) {
	ctx := acctest.Context(t)

	var user types.User
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_iam_user_policies_exclusive.test"
	userResourceName := "aws_iam_user.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoliciesExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, t, userResourceName, &user),
					testAccCheckUserPoliciesExclusiveExists(ctx, t, resourceName),
					testAccCheckUserPolicyRemoveInlinePolicy(ctx, t, &user, rName),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccUserPoliciesExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, t, userResourceName, &user),
					testAccCheckUserPoliciesExclusiveExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUserName, userResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "policy_names.#", "1"),
				),
			},
		},
	})
}

// An inline policy added out of band should be removed
func TestAccIAMUserPoliciesExclusive_outOfBandAddition(t *testing.T) {
	ctx := acctest.Context(t)

	var user types.User
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	policyName := rName + "-out-of-band"
	resourceName := "aws_iam_user_policies_exclusive.test"
	userResourceName := "aws_iam_user.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckUserDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccUserPoliciesExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, t, userResourceName, &user),
					testAccCheckUserPoliciesExclusiveExists(ctx, t, resourceName),
					testAccCheckUserPolicyAddInlinePolicy(ctx, t, &user, policyName),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccUserPoliciesExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserExists(ctx, t, userResourceName, &user),
					testAccCheckUserPoliciesExclusiveExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrUserName, userResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "policy_names.#", "1"),
				),
			},
		},
	})
}

func testAccCheckUserPoliciesExclusiveDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).IAMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iam_user_policies_exclusive" {
				continue
			}

			userName := rs.Primary.Attributes[names.AttrUserName]
			_, err := tfiam.FindUserPoliciesByName(ctx, conn, userName)
			if errs.IsA[*types.NoSuchEntityException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.IAM, create.ErrActionCheckingDestroyed, tfiam.ResNameUserPoliciesExclusive, rs.Primary.ID, err)
			}

			return create.Error(names.IAM, create.ErrActionCheckingDestroyed, tfiam.ResNameUserPoliciesExclusive, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckUserPoliciesExclusiveExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.IAM, create.ErrActionCheckingExistence, tfiam.ResNameUserPoliciesExclusive, name, errors.New("not found"))
		}

		userName := rs.Primary.Attributes[names.AttrUserName]
		if userName == "" {
			return create.Error(names.IAM, create.ErrActionCheckingExistence, tfiam.ResNameUserPoliciesExclusive, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).IAMClient(ctx)
		out, err := tfiam.FindUserPoliciesByName(ctx, conn, userName)
		if err != nil {
			return create.Error(names.IAM, create.ErrActionCheckingExistence, tfiam.ResNameUserPoliciesExclusive, userName, err)
		}

		policyCount := rs.Primary.Attributes["policy_names.#"]
		if policyCount != strconv.Itoa(len(out)) {
			return create.Error(names.IAM, create.ErrActionCheckingExistence, tfiam.ResNameUserPoliciesExclusive, userName, errors.New("unexpected policy_names count"))
		}

		return nil
	}
}

func testAccCheckUserPolicyAddInlinePolicy(ctx context.Context, t *testing.T, user *types.User, inlinePolicy string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).IAMClient(ctx)

		_, err := conn.PutUserPolicy(ctx, &iam.PutUserPolicyInput{
			PolicyDocument: aws.String(testAccUserPolicyExtraInlineConfig()),
			PolicyName:     aws.String(inlinePolicy),
			UserName:       user.UserName,
		})

		return err
	}
}

func testAccCheckUserPolicyRemoveInlinePolicy(ctx context.Context, t *testing.T, user *types.User, inlinePolicy string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).IAMClient(ctx)

		_, err := conn.DeleteUserPolicy(ctx, &iam.DeleteUserPolicyInput{
			PolicyName: aws.String(inlinePolicy),
			UserName:   user.UserName,
		})

		return err
	}
}

func testAccUserPolicyExtraInlineConfig() string {
	return `{
	"Version": "2012-10-17",
	"Statement": [
		{
		"Action": [
			"ec2:Describe*"
		],
		"Effect": "Allow",
		"Resource": "*"
		}
	]
}`
}

func testAccUserPoliciesExclusiveConfigBase(rName string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "inline" {
  statement {
    actions   = ["s3:ListBucket"]
    resources = ["*"]
  }
}

resource "aws_iam_user" "test" {
  name = %[1]q
}

resource "aws_iam_user_policy" "test" {
  name   = %[1]q
  user   = aws_iam_user.test.name
  policy = data.aws_iam_policy_document.inline.json
}
`, rName)
}

func testAccUserPoliciesExclusiveConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccUserPoliciesExclusiveConfigBase(rName),
		`
resource "aws_iam_user_policies_exclusive" "test" {
  user_name    = aws_iam_user.test.name
  policy_names = [aws_iam_user_policy.test.name]
}
`)
}

func testAccUserPoliciesExclusiveConfig_multiple(rName string) string {
	return acctest.ConfigCompose(
		testAccUserPoliciesExclusiveConfigBase(rName),
		fmt.Sprintf(`
resource "aws_iam_user_policy" "test2" {
  name   = "%[1]s-2"
  user   = aws_iam_user.test.name
  policy = data.aws_iam_policy_document.inline.json
}

resource "aws_iam_user_policy" "test3" {
  name   = "%[1]s-3"
  user   = aws_iam_user.test.name
  policy = data.aws_iam_policy_document.inline.json
}

resource "aws_iam_user_policies_exclusive" "test" {
  user_name = aws_iam_user.test.name
  policy_names = [
    aws_iam_user_policy.test.name,
    aws_iam_user_policy.test2.name,
    aws_iam_user_policy.test3.name,
  ]
}
`, rName))
}

func testAccUserPoliciesExclusiveConfig_empty(rName string) string {
	return acctest.ConfigCompose(
		testAccUserPoliciesExclusiveConfigBase(rName),
		`
resource "aws_iam_user_policies_exclusive" "test" {
  # Wait until the inline policy is created, then provision
  # the exclusive lock which will remove it. This creates a diff on
  # on the next plan (to re-create aws_iam_user_policy.test)
  # which the test can check for.
  depends_on = [aws_iam_user_policy.test]

  user_name    = aws_iam_user.test.name
  policy_names = []
}
`)
}
