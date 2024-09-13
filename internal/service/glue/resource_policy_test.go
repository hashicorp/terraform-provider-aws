// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	awspolicy "github.com/hashicorp/awspolicyequivalence"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfglue "github.com/hashicorp/terraform-provider-aws/internal/service/glue"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccResourcePolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_glue_resource_policy.test"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_required("glue:CreateTable"),
				Check: resource.ComposeTestCheckFunc(
					testAccResourcePolicy(ctx, resourceName, "glue:CreateTable"),
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

func testAccResourcePolicy_hybrid(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_glue_resource_policy.test"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_hybrid("glue:CreateTable", acctest.CtTrueCaps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enable_hybrid", acctest.CtTrueCaps),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"enable_hybrid"},
			},
			{
				Config: testAccResourcePolicyConfig_hybrid("glue:CreateTable", acctest.CtFalseCaps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enable_hybrid", acctest.CtFalseCaps),
				),
			},
			{
				Config: testAccResourcePolicyConfig_hybrid("glue:CreateTable", acctest.CtTrueCaps),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enable_hybrid", acctest.CtTrueCaps),
				),
			},
		},
	})
}

func testAccResourcePolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_glue_resource_policy.test"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_required("glue:CreateTable"),
				Check: resource.ComposeTestCheckFunc(
					testAccResourcePolicy(ctx, resourceName, "glue:CreateTable"),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfglue.ResourceResourcePolicy(), resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfglue.ResourceResourcePolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccResourcePolicy_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_glue_resource_policy.test"
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_required("glue:CreateTable"),
				Check: resource.ComposeTestCheckFunc(
					testAccResourcePolicy(ctx, resourceName, "glue:CreateTable"),
				),
			},
			{
				Config: testAccResourcePolicyConfig_required("glue:DeleteTable"),
				Check: resource.ComposeTestCheckFunc(
					testAccResourcePolicy(ctx, resourceName, "glue:DeleteTable"),
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

func testAccResourcePolicy_ignoreEquivalent(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_glue_resource_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_equivalent(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourcePolicy(ctx, resourceName, "glue:CreateTable"),
				),
			},
			{
				Config:   testAccResourcePolicyConfig_equivalent2(),
				PlanOnly: true,
			},
		},
	})
}

func testAccResourcePolicy(ctx context.Context, n string, action string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No policy id set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueClient(ctx)

		policy, err := conn.GetResourcePolicy(ctx, &glue.GetResourcePolicyInput{})
		if err != nil {
			return fmt.Errorf("Get resource policy error: %v", err)
		}

		actualPolicyText := aws.ToString(policy.PolicyInJson)

		expectedPolicy := CreateTablePolicy(action)
		equivalent, err := awspolicy.PoliciesAreEquivalent(actualPolicyText, expectedPolicy)
		if err != nil {
			return fmt.Errorf("Error testing policy equivalence: %s", err)
		}
		if !equivalent {
			return fmt.Errorf("Non-equivalent policy error:\n\nexpected: %s\n\n     got: %s\n",
				expectedPolicy, actualPolicyText)
		}

		return nil
	}
}

func testAccCheckResourcePolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueClient(ctx)

		policy, err := conn.GetResourcePolicy(ctx, &glue.GetResourcePolicyInput{})

		if err != nil {
			if errs.IsAErrorMessageContains[*awstypes.EntityNotFoundException](err, "Policy not found") {
				return nil
			}
			return err
		}

		if *policy.PolicyInJson != "" {
			return fmt.Errorf("Aws glue resource policy still exists: %s", *policy.PolicyInJson)
		}
		return nil
	}
}

func CreateTablePolicy(action string) string {
	return fmt.Sprintf(`{
  "Version" : "2012-10-17",
  "Statement" : [
    {
      "Effect" : "Allow",
      "Action" : [
        "%s"
      ],
      "Principal" : {
         "AWS": "*"
       },
      "Resource" : "arn:%s:glue:%s:%s:*"
    }
  ]
}`, action, acctest.Partition(), acctest.Region(), acctest.AccountID())
}

func testAccResourcePolicyConfig_required(action string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

data "aws_region" "current" {}

data "aws_iam_policy_document" "glue-example-policy" {
  statement {
    actions   = [%[1]q]
    resources = ["arn:${data.aws_partition.current.partition}:glue:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:*"]
    principals {
      identifiers = ["*"]
      type        = "AWS"
    }
  }
}

resource "aws_glue_resource_policy" "test" {
  policy = data.aws_iam_policy_document.glue-example-policy.json
}
`, action)
}

func testAccResourcePolicyConfig_hybrid(action, hybrid string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

data "aws_region" "current" {}

data "aws_iam_policy_document" "glue-example-policy" {
  statement {
    actions   = [%[1]q]
    resources = ["arn:${data.aws_partition.current.partition}:glue:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:*"]
    principals {
      identifiers = ["*"]
      type        = "AWS"
    }
  }
}

resource "aws_glue_resource_policy" "test" {
  policy        = data.aws_iam_policy_document.glue-example-policy.json
  enable_hybrid = %[2]q
}
`, action, hybrid)
}

func testAccResourcePolicyConfig_equivalent() string {
	return `
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_glue_resource_policy" "test" {
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = {
      Action = "glue:CreateTable"
      Effect = "Allow"
      Resource = [
        "arn:${data.aws_partition.current.partition}:glue:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:*"
      ]
      Principal = {
        AWS = "*"
      }
    }
  })
}
`
}

func testAccResourcePolicyConfig_equivalent2() string {
	return `
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_glue_resource_policy" "test" {
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = {
      Effect = "Allow"
      Action = [
        "glue:CreateTable",
      ]
      Resource = "arn:${data.aws_partition.current.partition}:glue:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:*"
      Principal = {
        AWS = ["*"]
      }
    }
  })
}
`
}
