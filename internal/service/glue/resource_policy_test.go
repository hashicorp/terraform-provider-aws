// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package glue_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awspolicy "github.com/hashicorp/awspolicyequivalence"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfglue "github.com/hashicorp/terraform-provider-aws/internal/service/glue"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccResourcePolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_glue_resource_policy.test"
	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_required("glue:CreateTable"),
				Check: resource.ComposeTestCheckFunc(
					testAccResourcePolicy(ctx, t, resourceName, "glue:CreateTable"),
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
	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx, t),
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
	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_required("glue:CreateTable"),
				Check: resource.ComposeTestCheckFunc(
					testAccResourcePolicy(ctx, t, resourceName, "glue:CreateTable"),
					acctest.CheckSDKResourceDisappears(ctx, t, tfglue.ResourceResourcePolicy(), resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfglue.ResourceResourcePolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccResourcePolicy_update(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_glue_resource_policy.test"
	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_required("glue:CreateTable"),
				Check: resource.ComposeTestCheckFunc(
					testAccResourcePolicy(ctx, t, resourceName, "glue:CreateTable"),
				),
			},
			{
				Config: testAccResourcePolicyConfig_required("glue:DeleteTable"),
				Check: resource.ComposeTestCheckFunc(
					testAccResourcePolicy(ctx, t, resourceName, "glue:DeleteTable"),
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

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccResourcePolicyConfig_equivalent(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourcePolicy(ctx, t, resourceName, "glue:CreateTable"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccResourcePolicyConfig_equivalent2(),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func testAccResourcePolicy(ctx context.Context, t *testing.T, n string, action string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).GlueClient(ctx)

		output, err := tfglue.FindResourcePolicy(ctx, conn)

		if err != nil {
			return err
		}

		actualPolicyText, expectedPolicy := aws.ToString(output.PolicyInJson), testAccNewResourcePolicy(ctx, action)
		equivalent, err := awspolicy.PoliciesAreEquivalent(actualPolicyText, expectedPolicy)
		if err != nil {
			return fmt.Errorf("Error testing policy equivalence: %w", err)
		}
		if !equivalent {
			return fmt.Errorf("Non-equivalent policy error:\n\nexpected: %s\n\n     got: %s\n",
				expectedPolicy, actualPolicyText)
		}

		return nil
	}
}

func testAccCheckResourcePolicyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).GlueClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_glue_resource_policy" {
				continue
			}

			_, err := tfglue.FindResourcePolicy(ctx, conn)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Glue Resource Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccNewResourcePolicy(ctx context.Context, action string) string {
	return fmt.Sprintf(`{
  "Version" : "2012-10-17",
  "Statement" : [
    {
      "Effect" : "Allow",
      "Action" : [
        %[1]q
      ],
      "Principal" : {
         "AWS": "*"
       },
      "Resource" : "arn:%[2]s:glue:%[3]s:%[4]s:*"
    }
  ]
}`, action, acctest.Partition(), acctest.Region(), acctest.AccountID(ctx))
}

func testAccResourcePolicyConfig_required(action string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

data "aws_region" "current" {}

data "aws_iam_policy_document" "glue-example-policy" {
  statement {
    actions   = [%[1]q]
    resources = ["arn:${data.aws_partition.current.partition}:glue:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:*"]
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
    resources = ["arn:${data.aws_partition.current.partition}:glue:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:*"]
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
        "arn:${data.aws_partition.current.partition}:glue:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:*"
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
      Resource = "arn:${data.aws_partition.current.partition}:glue:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:*"
      Principal = {
        AWS = ["*"]
      }
    }
  })
}
`
}
