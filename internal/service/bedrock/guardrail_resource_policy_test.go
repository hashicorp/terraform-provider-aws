// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrock_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfbedrock "github.com/hashicorp/terraform-provider-aws/internal/service/bedrock"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockGuardrailResourcePolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrock_guardrail_resource_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGuardrailResourcePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGuardrailResourcePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGuardrailResourcePolicyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", "aws_bedrock_guardrail.test", "guardrail_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "policy"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGuardrailResourcePolicyImportStateIDFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrResourceARN,
			},
		},
	})
}

func TestAccBedrockGuardrailResourcePolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrock_guardrail_resource_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGuardrailResourcePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGuardrailResourcePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGuardrailResourcePolicyExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfbedrock.ResourceGuardrailResourcePolicy, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccBedrockGuardrailResourcePolicy_update(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrock_guardrail_resource_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			acctest.PreCheckOrganizationsEnabled(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGuardrailResourcePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGuardrailResourcePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGuardrailResourcePolicyExists(ctx, t, resourceName),
				),
			},
			{
				Config: testAccGuardrailResourcePolicyConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGuardrailResourcePolicyExists(ctx, t, resourceName),
				),
			},
		},
	})
}

func testAccGuardrailResourcePolicyImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes["resource_arn"], nil
	}
}

func testAccCheckGuardrailResourcePolicyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrock_guardrail_resource_policy" {
				continue
			}

			arn := rs.Primary.Attributes["resource_arn"]

			_, err := tfbedrock.FindGuardrailResourcePolicyByARN(ctx, conn, arn)
			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.Bedrock, create.ErrActionCheckingDestroyed, tfbedrock.ResNameGuardrailResourcePolicy, arn, err)
			}

			return create.Error(names.Bedrock, create.ErrActionCheckingDestroyed, tfbedrock.ResNameGuardrailResourcePolicy, arn, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckGuardrailResourcePolicyExists(ctx context.Context, t *testing.T, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Bedrock, create.ErrActionCheckingExistence, tfbedrock.ResNameGuardrailResourcePolicy, name, errors.New("not found"))
		}

		arn := rs.Primary.Attributes["resource_arn"]
		if arn == "" {
			return create.Error(names.Bedrock, create.ErrActionCheckingExistence, tfbedrock.ResNameGuardrailResourcePolicy, name, errors.New("resource_arn not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).BedrockClient(ctx)

		_, err := tfbedrock.FindGuardrailResourcePolicyByARN(ctx, conn, arn)
		if err != nil {
			return create.Error(names.Bedrock, create.ErrActionCheckingExistence, tfbedrock.ResNameGuardrailResourcePolicy, arn, err)
		}

		return nil
	}
}

func testAccGuardrailResourcePolicyConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_bedrock_guardrail" "test" {
  name                      = %[1]q
  blocked_input_messaging   = "test"
  blocked_outputs_messaging = "test"
  description               = "test"

  content_policy_config {
    filters_config {
      input_strength  = "MEDIUM"
      output_strength = "MEDIUM"
      type            = "HATE"
    }
  }
}
`, rName)
}

func testAccGuardrailResourcePolicyConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccGuardrailResourcePolicyConfig_base(rName), `
data "aws_organizations_organization" "current" {}

resource "aws_bedrock_guardrail_resource_policy" "test" {
  resource_arn = aws_bedrock_guardrail.test.guardrail_arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = "*"
      Action    = ["bedrock:GetGuardrail", "bedrock:ApplyGuardrail"]
      Resource  = aws_bedrock_guardrail.test.guardrail_arn
      Condition = {
        StringEquals = { "aws:PrincipalOrgID" = data.aws_organizations_organization.current.id }
      }
    }]
  })
}
`)
}

func testAccGuardrailResourcePolicyConfig_updated(rName string) string {
	return acctest.ConfigCompose(testAccGuardrailResourcePolicyConfig_base(rName), `
data "aws_organizations_organization" "current" {}

resource "aws_bedrock_guardrail_resource_policy" "test" {
  resource_arn = aws_bedrock_guardrail.test.guardrail_arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = "*"
      Action    = ["bedrock:ApplyGuardrail"]
      Resource  = aws_bedrock_guardrail.test.guardrail_arn
      Condition = {
        StringEquals = { "aws:PrincipalOrgID" = data.aws_organizations_organization.current.id }
      }
    }]
  })
}
`)
}
