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

func TestAccIAMGroupPoliciesExclusive_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var group types.Group
	var groupPolicy string
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_group_policies_exclusive.test"
	groupResourceName := "aws_iam_group.test"
	groupPolicyResourceName := "aws_iam_group_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupPoliciesExclusiveDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupPoliciesExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, groupResourceName, &group),
					testAccCheckGroupPolicyExists(ctx, groupPolicyResourceName, &groupPolicy),
					testAccCheckGroupPoliciesExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrGroupName, groupResourceName, names.AttrName),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_names.*", groupPolicyResourceName, names.AttrName),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGroupPoliciesExclusiveImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrGroupName,
			},
		},
	})
}

func TestAccIAMGroupPoliciesExclusive_disappears_Group(t *testing.T) {
	ctx := acctest.Context(t)

	var group types.Group
	var groupPolicy string
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_group_policies_exclusive.test"
	groupResourceName := "aws_iam_group.test"
	groupPolicyResourceName := "aws_iam_group_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupPoliciesExclusiveDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupPoliciesExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, groupResourceName, &group),
					testAccCheckGroupPolicyExists(ctx, groupPolicyResourceName, &groupPolicy),
					testAccCheckGroupPoliciesExclusiveExists(ctx, resourceName),
					// Inline policy must be deleted before the group can be
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfiam.ResourceGroupPolicy(), groupPolicyResourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfiam.ResourceGroup(), groupResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIAMGroupPoliciesExclusive_multiple(t *testing.T) {
	ctx := acctest.Context(t)

	var group types.Group
	var groupPolicy string
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_group_policies_exclusive.test"
	groupResourceName := "aws_iam_group.test"
	groupPolicyResourceName := "aws_iam_group_policy.test"
	groupPolicyResourceName2 := "aws_iam_group_policy.test2"
	groupPolicyResourceName3 := "aws_iam_group_policy.test3"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupPoliciesExclusiveDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupPoliciesExclusiveConfig_multiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, groupResourceName, &group),
					testAccCheckGroupPolicyExists(ctx, groupPolicyResourceName, &groupPolicy),
					testAccCheckGroupPoliciesExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrGroupName, groupResourceName, names.AttrName),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_names.*", groupPolicyResourceName, names.AttrName),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_names.*", groupPolicyResourceName2, names.AttrName),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_names.*", groupPolicyResourceName3, names.AttrName),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGroupPoliciesExclusiveImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrGroupName,
			},
			{
				Config: testAccGroupPoliciesExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, groupResourceName, &group),
					testAccCheckGroupPolicyExists(ctx, groupPolicyResourceName, &groupPolicy),
					testAccCheckGroupPoliciesExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrGroupName, groupResourceName, names.AttrName),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_names.*", groupPolicyResourceName, names.AttrName),
				),
			},
		},
	})
}

func TestAccIAMGroupPoliciesExclusive_empty(t *testing.T) {
	ctx := acctest.Context(t)

	var group types.Group
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_group_policies_exclusive.test"
	groupResourceName := "aws_iam_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupPoliciesExclusiveDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupPoliciesExclusiveConfig_empty(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, groupResourceName, &group),
					testAccCheckGroupPoliciesExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrGroupName, groupResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "policy_names.#", acctest.Ct0),
				),
				// The empty `policy_names` argument in the exclusive lock will remove the
				// inline policy defined in this configuration, so a diff is expected
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// An inline policy removed out of band should be recreated
func TestAccIAMGroupPoliciesExclusive_outOfBandRemoval(t *testing.T) {
	ctx := acctest.Context(t)

	var group types.Group
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_group_policies_exclusive.test"
	groupResourceName := "aws_iam_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupPoliciesExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, groupResourceName, &group),
					testAccCheckGroupPoliciesExclusiveExists(ctx, resourceName),
					testAccCheckGroupPolicyRemoveInlinePolicy(ctx, &group, rName),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccGroupPoliciesExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, groupResourceName, &group),
					testAccCheckGroupPoliciesExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrGroupName, groupResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "policy_names.#", acctest.Ct1),
				),
			},
		},
	})
}

// An inline policy added out of band should be removed
func TestAccIAMGroupPoliciesExclusive_outOfBandAddition(t *testing.T) {
	ctx := acctest.Context(t)

	var group types.Group
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	policyName := rName + "-out-of-band"
	resourceName := "aws_iam_group_policies_exclusive.test"
	groupResourceName := "aws_iam_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGroupPoliciesExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, groupResourceName, &group),
					testAccCheckGroupPoliciesExclusiveExists(ctx, resourceName),
					testAccCheckGroupPolicyAddInlinePolicy(ctx, &group, policyName),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccGroupPoliciesExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroupExists(ctx, groupResourceName, &group),
					testAccCheckGroupPoliciesExclusiveExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrGroupName, groupResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "policy_names.#", acctest.Ct1),
				),
			},
		},
	})
}

func testAccCheckGroupPoliciesExclusiveDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iam_group_policies_exclusive" {
				continue
			}

			groupName := rs.Primary.Attributes[names.AttrGroupName]
			_, err := tfiam.FindGroupPoliciesByName(ctx, conn, groupName)
			if errs.IsA[*types.NoSuchEntityException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.IAM, create.ErrActionCheckingDestroyed, tfiam.ResNameGroupPoliciesExclusive, rs.Primary.ID, err)
			}

			return create.Error(names.IAM, create.ErrActionCheckingDestroyed, tfiam.ResNameGroupPoliciesExclusive, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckGroupPoliciesExclusiveExists(ctx context.Context, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.IAM, create.ErrActionCheckingExistence, tfiam.ResNameGroupPoliciesExclusive, name, errors.New("not found"))
		}

		groupName := rs.Primary.Attributes[names.AttrGroupName]
		if groupName == "" {
			return create.Error(names.IAM, create.ErrActionCheckingExistence, tfiam.ResNameGroupPoliciesExclusive, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)
		out, err := tfiam.FindGroupPoliciesByName(ctx, conn, groupName)
		if err != nil {
			return create.Error(names.IAM, create.ErrActionCheckingExistence, tfiam.ResNameGroupPoliciesExclusive, groupName, err)
		}

		policyCount := rs.Primary.Attributes["policy_names.#"]
		if policyCount != fmt.Sprint(len(out)) {
			return create.Error(names.IAM, create.ErrActionCheckingExistence, tfiam.ResNameGroupPoliciesExclusive, groupName, errors.New("unexpected policy_names count"))
		}

		return nil
	}
}

func testAccGroupPoliciesExclusiveImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes[names.AttrGroupName], nil
	}
}

func testAccCheckGroupPolicyAddInlinePolicy(ctx context.Context, group *types.Group, inlinePolicy string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

		_, err := conn.PutGroupPolicy(ctx, &iam.PutGroupPolicyInput{
			PolicyDocument: aws.String(testAccGroupPolicyExtraInlineConfig()),
			PolicyName:     aws.String(inlinePolicy),
			GroupName:      group.GroupName,
		})

		return err
	}
}

func testAccCheckGroupPolicyRemoveInlinePolicy(ctx context.Context, group *types.Group, inlinePolicy string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMClient(ctx)

		_, err := conn.DeleteGroupPolicy(ctx, &iam.DeleteGroupPolicyInput{
			PolicyName: aws.String(inlinePolicy),
			GroupName:  group.GroupName,
		})

		return err
	}
}

func testAccGroupPolicyExtraInlineConfig() string {
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

func testAccGroupPoliciesExclusiveConfigBase(rName string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "inline" {
  statement {
    actions   = ["s3:ListBucket"]
    resources = ["*"]
  }
}

resource "aws_iam_group" "test" {
  name = %[1]q
}

resource "aws_iam_group_policy" "test" {
  name   = %[1]q
  group  = aws_iam_group.test.name
  policy = data.aws_iam_policy_document.inline.json
}
`, rName)
}

func testAccGroupPoliciesExclusiveConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccGroupPoliciesExclusiveConfigBase(rName),
		`
resource "aws_iam_group_policies_exclusive" "test" {
  group_name   = aws_iam_group.test.name
  policy_names = [aws_iam_group_policy.test.name]
}
`)
}

func testAccGroupPoliciesExclusiveConfig_multiple(rName string) string {
	return acctest.ConfigCompose(
		testAccGroupPoliciesExclusiveConfigBase(rName),
		fmt.Sprintf(`
resource "aws_iam_group_policy" "test2" {
  name   = "%[1]s-2"
  group  = aws_iam_group.test.name
  policy = data.aws_iam_policy_document.inline.json
}

resource "aws_iam_group_policy" "test3" {
  name   = "%[1]s-3"
  group  = aws_iam_group.test.name
  policy = data.aws_iam_policy_document.inline.json
}

resource "aws_iam_group_policies_exclusive" "test" {
  group_name = aws_iam_group.test.name
  policy_names = [
    aws_iam_group_policy.test.name,
    aws_iam_group_policy.test2.name,
    aws_iam_group_policy.test3.name,
  ]
}
`, rName))
}

func testAccGroupPoliciesExclusiveConfig_empty(rName string) string {
	return acctest.ConfigCompose(
		testAccGroupPoliciesExclusiveConfigBase(rName),
		`
resource "aws_iam_group_policies_exclusive" "test" {
  # Wait until the inline policy is created, then provision
  # the exclusive lock which will remove it. This creates a diff on
  # on the next plan (to re-create aws_iam_group_policy.test)
  # which the test can check for.
  depends_on = [aws_iam_group_policy.test]

  group_name   = aws_iam_group.test.name
  policy_names = []
}
`)
}
