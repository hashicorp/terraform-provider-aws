// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrock_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
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
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGuardrailResourcePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGuardrailResourcePolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGuardrailResourcePolicyExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrResourceARN, "aws_bedrock_guardrail.test", "guardrail_arn"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrPolicy),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrResourceARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrResourceARN,
				// JSON is not normalized during attribute comparison.
				ImportStateVerifyIgnore: []string{names.AttrPolicy},
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
					PostApplyPostRefresh: []plancheck.PlanCheck{
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
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGuardrailResourcePolicyExists(ctx, t, resourceName),
				),
			},
		},
	})
}

// Guardrail profiles are system-defined and shared across the account and Region, so this test is serialized.
func testAccGuardrailResourcePolicy_guardrailProfile(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_bedrock_guardrail_resource_policy.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			// The "us" guardrail profile is only available in US Regions.
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID, endpoints.UsEast2RegionID, endpoints.UsWest2RegionID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGuardrailResourcePolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGuardrailResourcePolicyConfig_guardrailProfile("test1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGuardrailResourcePolicyExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrResourceARN), tfknownvalue.RegionalARNExact("bedrock", "guardrail-profile/us.guardrail.v1:0")),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrResourceARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrResourceARN,
				// JSON is not normalized during attribute comparison.
				ImportStateVerifyIgnore: []string{names.AttrPolicy},
			},
			{
				Config: testAccGuardrailResourcePolicyConfig_guardrailProfile("test2"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGuardrailResourcePolicyExists(ctx, t, resourceName),
				),
			},
		},
	})
}

func testAccCheckGuardrailResourcePolicyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrock_guardrail_resource_policy" {
				continue
			}

			ctx := acctest.NewTestResourceContext(ctx, rs.Type, rs.Primary.Attributes[names.AttrRegion])
			conn := acctest.ProviderMeta(ctx, t).BedrockClient(ctx)

			arn := rs.Primary.Attributes[names.AttrResourceARN]

			_, err := tfbedrock.FindGuardrailResourcePolicyByARN(ctx, conn, arn)
			if retry.NotFound(err) {
				continue
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

		arn := rs.Primary.Attributes[names.AttrResourceARN]
		if arn == "" {
			return create.Error(names.Bedrock, create.ErrActionCheckingExistence, tfbedrock.ResNameGuardrailResourcePolicy, name, errors.New("resource_arn not set"))
		}

		ctx := acctest.NewTestResourceContext(ctx, rs.Type, rs.Primary.Attributes[names.AttrRegion])
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

data "aws_partition" "current" {}

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
resource "aws_bedrock_guardrail_resource_policy" "test" {
  resource_arn = aws_bedrock_guardrail.test.guardrail_arn

  policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    effect    = "Allow"
    actions   = ["bedrock:GetGuardrail", "bedrock:ApplyGuardrail"]
    resources = [aws_bedrock_guardrail.test.guardrail_arn]

    principals {
      type        = "AWS"
      identifiers = ["arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"]
    }
  }
}
`)
}

func testAccGuardrailResourcePolicyConfig_updated(rName string) string {
	return acctest.ConfigCompose(testAccGuardrailResourcePolicyConfig_base(rName), `
resource "aws_bedrock_guardrail_resource_policy" "test" {
  resource_arn = aws_bedrock_guardrail.test.guardrail_arn

  policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    effect    = "Allow"
    actions   = ["bedrock:ApplyGuardrail"]
    resources = [aws_bedrock_guardrail.test.guardrail_arn]

    principals {
      type        = "AWS"
      identifiers = ["arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"]
    }
  }
}
`)
}

func testAccGuardrailResourcePolicyConfig_guardrailProfile(sid string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

data "aws_region" "current" {}

locals {
  guardrail_profile_arn = "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:guardrail-profile/us.guardrail.v1:0"
}

resource "aws_bedrock_guardrail_resource_policy" "test" {
  resource_arn = local.guardrail_profile_arn

  policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "test" {
  statement {
    sid       = %[1]q
    effect    = "Allow"
    actions   = ["bedrock:ApplyGuardrail"]
    resources = [local.guardrail_profile_arn]

    principals {
      type        = "AWS"
      identifiers = ["arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"]
    }
  }
}
`, sid)
}
