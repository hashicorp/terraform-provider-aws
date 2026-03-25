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

func TestAccIAMRolePoliciesExclusive_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var role types.Role
	var rolePolicy string
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_iam_role_policies_exclusive.test"
	roleResourceName := "aws_iam_role.test"
	rolePolicyResourceName := "aws_iam_role_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRolePoliciesExclusiveDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRolePoliciesExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, t, roleResourceName, &role),
					testAccCheckRolePolicyExists(ctx, t, rolePolicyResourceName, &rolePolicy),
					testAccCheckRolePoliciesExclusiveExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "role_name", roleResourceName, names.AttrName),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_names.*", rolePolicyResourceName, names.AttrName),
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

func TestAccIAMRolePoliciesExclusive_disappears_Role(t *testing.T) {
	ctx := acctest.Context(t)

	var role types.Role
	var rolePolicy string
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_iam_role_policies_exclusive.test"
	roleResourceName := "aws_iam_role.test"
	rolePolicyResourceName := "aws_iam_role_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRolePoliciesExclusiveDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRolePoliciesExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, t, roleResourceName, &role),
					testAccCheckRolePolicyExists(ctx, t, rolePolicyResourceName, &rolePolicy),
					testAccCheckRolePoliciesExclusiveExists(ctx, t, resourceName),
					// Inline policy must be deleted before the role can be
					acctest.CheckSDKResourceDisappears(ctx, t, tfiam.ResourceRolePolicy(), rolePolicyResourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfiam.ResourceRole(), roleResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIAMRolePoliciesExclusive_multiple(t *testing.T) {
	ctx := acctest.Context(t)

	var role types.Role
	var rolePolicy string
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_iam_role_policies_exclusive.test"
	roleResourceName := "aws_iam_role.test"
	rolePolicyResourceName := "aws_iam_role_policy.test"
	rolePolicyResourceName2 := "aws_iam_role_policy.test2"
	rolePolicyResourceName3 := "aws_iam_role_policy.test3"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRolePoliciesExclusiveDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRolePoliciesExclusiveConfig_multiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, t, roleResourceName, &role),
					testAccCheckRolePolicyExists(ctx, t, rolePolicyResourceName, &rolePolicy),
					testAccCheckRolePoliciesExclusiveExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "role_name", roleResourceName, names.AttrName),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_names.*", rolePolicyResourceName, names.AttrName),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_names.*", rolePolicyResourceName2, names.AttrName),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_names.*", rolePolicyResourceName3, names.AttrName),
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
				Config: testAccRolePoliciesExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, t, roleResourceName, &role),
					testAccCheckRolePolicyExists(ctx, t, rolePolicyResourceName, &rolePolicy),
					testAccCheckRolePoliciesExclusiveExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "role_name", roleResourceName, names.AttrName),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "policy_names.*", rolePolicyResourceName, names.AttrName),
				),
			},
		},
	})
}

func TestAccIAMRolePoliciesExclusive_empty(t *testing.T) {
	ctx := acctest.Context(t)

	var role types.Role
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_iam_role_policies_exclusive.test"
	roleResourceName := "aws_iam_role.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRolePoliciesExclusiveDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRolePoliciesExclusiveConfig_empty(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, t, roleResourceName, &role),
					testAccCheckRolePoliciesExclusiveExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "role_name", roleResourceName, names.AttrName),
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
func TestAccIAMRolePoliciesExclusive_outOfBandRemoval(t *testing.T) {
	ctx := acctest.Context(t)

	var role types.Role
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_iam_role_policies_exclusive.test"
	roleResourceName := "aws_iam_role.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRolePoliciesExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, t, roleResourceName, &role),
					testAccCheckRolePoliciesExclusiveExists(ctx, t, resourceName),
					testAccCheckRolePolicyRemoveInlinePolicy(ctx, t, &role, rName),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccRolePoliciesExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, t, roleResourceName, &role),
					testAccCheckRolePoliciesExclusiveExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "role_name", roleResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "policy_names.#", "1"),
				),
			},
		},
	})
}

// An inline policy added out of band should be removed
func TestAccIAMRolePoliciesExclusive_outOfBandAddition(t *testing.T) {
	ctx := acctest.Context(t)

	var role types.Role
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	policyName := rName + "-out-of-band"
	resourceName := "aws_iam_role_policies_exclusive.test"
	roleResourceName := "aws_iam_role.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRoleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRolePoliciesExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, t, roleResourceName, &role),
					testAccCheckRolePoliciesExclusiveExists(ctx, t, resourceName),
					testAccCheckRolePolicyAddInlinePolicy(ctx, t, &role, policyName),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccRolePoliciesExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoleExists(ctx, t, roleResourceName, &role),
					testAccCheckRolePoliciesExclusiveExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "role_name", roleResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "policy_names.#", "1"),
				),
			},
		},
	})
}

func testAccCheckRolePoliciesExclusiveDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).IAMClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iam_role_policies_exclusive" {
				continue
			}

			roleName := rs.Primary.Attributes["role_name"]
			_, err := tfiam.FindRolePoliciesByName(ctx, conn, roleName)
			if errs.IsA[*types.NoSuchEntityException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.IAM, create.ErrActionCheckingDestroyed, tfiam.ResNameRolePoliciesExclusive, rs.Primary.ID, err)
			}

			return create.Error(names.IAM, create.ErrActionCheckingDestroyed, tfiam.ResNameRolePoliciesExclusive, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckRolePoliciesExclusiveExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.IAM, create.ErrActionCheckingExistence, tfiam.ResNameRolePoliciesExclusive, name, errors.New("not found"))
		}

		roleName := rs.Primary.Attributes["role_name"]
		if roleName == "" {
			return create.Error(names.IAM, create.ErrActionCheckingExistence, tfiam.ResNameRolePoliciesExclusive, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).IAMClient(ctx)
		out, err := tfiam.FindRolePoliciesByName(ctx, conn, roleName)
		if err != nil {
			return create.Error(names.IAM, create.ErrActionCheckingExistence, tfiam.ResNameRolePoliciesExclusive, roleName, err)
		}

		policyCount := rs.Primary.Attributes["policy_names.#"]
		if policyCount != strconv.Itoa(len(out)) {
			return create.Error(names.IAM, create.ErrActionCheckingExistence, tfiam.ResNameRolePoliciesExclusive, roleName, errors.New("unexpected policy_names count"))
		}

		return nil
	}
}

func testAccRolePoliciesExclusiveConfigBase(rName string) string {
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

data "aws_iam_policy_document" "inline" {
  statement {
    actions   = ["s3:ListBucket"]
    resources = ["*"]
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.trust.json
}

resource "aws_iam_role_policy" "test" {
  name   = %[1]q
  role   = aws_iam_role.test.name
  policy = data.aws_iam_policy_document.inline.json
}
`, rName)
}

func testAccRolePoliciesExclusiveConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccRolePoliciesExclusiveConfigBase(rName),
		`
resource "aws_iam_role_policies_exclusive" "test" {
  role_name    = aws_iam_role.test.name
  policy_names = [aws_iam_role_policy.test.name]
}
`)
}

func testAccRolePoliciesExclusiveConfig_multiple(rName string) string {
	return acctest.ConfigCompose(
		testAccRolePoliciesExclusiveConfigBase(rName),
		fmt.Sprintf(`
resource "aws_iam_role_policy" "test2" {
  name   = "%[1]s-2"
  role   = aws_iam_role.test.name
  policy = data.aws_iam_policy_document.inline.json
}

resource "aws_iam_role_policy" "test3" {
  name   = "%[1]s-3"
  role   = aws_iam_role.test.name
  policy = data.aws_iam_policy_document.inline.json
}

resource "aws_iam_role_policies_exclusive" "test" {
  role_name = aws_iam_role.test.name
  policy_names = [
    aws_iam_role_policy.test.name,
    aws_iam_role_policy.test2.name,
    aws_iam_role_policy.test3.name,
  ]
}
`, rName))
}

func testAccRolePoliciesExclusiveConfig_empty(rName string) string {
	return acctest.ConfigCompose(
		testAccRolePoliciesExclusiveConfigBase(rName),
		`
resource "aws_iam_role_policies_exclusive" "test" {
  # Wait until the inline policy is created, then provision
  # the exclusive lock which will remove it. This creates a diff on
  # on the next plan (to re-create aws_iam_role_policy.test)
  # which the test can check for.
  depends_on = [aws_iam_role_policy.test]

  role_name    = aws_iam_role.test.name
  policy_names = []
}
`)
}
