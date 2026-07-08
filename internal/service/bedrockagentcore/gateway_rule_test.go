// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfbedrockagentcore "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagentcore"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccGatewayRuleImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return acctest.AttrsImportStateIdFunc(resourceName, ",", "gateway_identifier", "rule_id")
}

func TestAccBedrockAgentCoreGatewayRule_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var gatewayRule bedrockagentcorecontrol.GetGatewayRuleOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameRuntime := randomWithPrefixAndUnderscore(t)
	rImageUri := acctest.SkipIfEnvVarNotSet(t, "AWS_BEDROCK_AGENTCORE_RUNTIME_IMAGE_V1_URI")
	resourceName := "aws_bedrockagentcore_gateway_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayRuleConfig_basic(rName, rNameRuntime, rImageUri),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayRuleExists(ctx, t, resourceName, &gatewayRule),
					resource.TestCheckResourceAttrSet(resourceName, "gateway_identifier"),
					resource.TestCheckResourceAttrSet(resourceName, "gateway_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "rule_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPriority, "100"),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction+".#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.route_to_target.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.route_to_target.0.static_route.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "action.0.route_to_target.0.static_route.0.target_name", "aws_bedrockagentcore_gateway_target.test", names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "ACTIVE"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccGatewayRuleImportStateIDFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "rule_id",
			},
		},
	})
}

func TestAccBedrockAgentCoreGatewayRule_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var gatewayRule bedrockagentcorecontrol.GetGatewayRuleOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameRuntime := randomWithPrefixAndUnderscore(t)
	rImageUri := acctest.SkipIfEnvVarNotSet(t, "AWS_BEDROCK_AGENTCORE_RUNTIME_IMAGE_V1_URI")
	resourceName := "aws_bedrockagentcore_gateway_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayRuleConfig_basic(rName, rNameRuntime, rImageUri),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayRuleExists(ctx, t, resourceName, &gatewayRule),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfbedrockagentcore.ResourceGatewayRule, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreGatewayRule_update(t *testing.T) {
	ctx := acctest.Context(t)
	var gatewayRule bedrockagentcorecontrol.GetGatewayRuleOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameRuntime := randomWithPrefixAndUnderscore(t)
	rImageUri := acctest.SkipIfEnvVarNotSet(t, "AWS_BEDROCK_AGENTCORE_RUNTIME_IMAGE_V1_URI")
	resourceName := "aws_bedrockagentcore_gateway_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayRuleConfig_basic(rName, rNameRuntime, rImageUri),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayRuleExists(ctx, t, resourceName, &gatewayRule),
					resource.TestCheckResourceAttr(resourceName, names.AttrPriority, "100"),
					resource.TestCheckNoResourceAttr(resourceName, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, names.AttrCondition+".#", "0"),
				),
			},
			{
				Config: testAccGatewayRuleConfig_conditions(rName, rNameRuntime, rImageUri, 200, "updated rule description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayRuleExists(ctx, t, resourceName, &gatewayRule),
					resource.TestCheckResourceAttr(resourceName, names.AttrPriority, "200"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "updated rule description"),
					resource.TestCheckResourceAttr(resourceName, names.AttrCondition+".#", "1"),
					resource.TestCheckResourceAttr(resourceName, "condition.0.match_principals.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "condition.0.match_principals.0.any_of.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "condition.0.match_principals.0.any_of.0.iam_principal.0.operator", "StringLike"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				// Remove the description argument while changing priority. Because
				// the API never clears description and the attribute is
				// Optional+Computed, the prior value is retained rather than
				// producing "inconsistent result after apply".
				Config: testAccGatewayRuleConfig_conditionsNoDescription(rName, rNameRuntime, rImageUri, 300),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayRuleExists(ctx, t, resourceName, &gatewayRule),
					resource.TestCheckResourceAttr(resourceName, names.AttrPriority, "300"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "updated rule description"),
					resource.TestCheckResourceAttr(resourceName, names.AttrCondition+".#", "1"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				// Remove the condition block entirely. The Update path sends an
				// explicit empty conditions list so the server actually clears it,
				// rather than PATCH-retaining it and producing "inconsistent result
				// after apply".
				Config: testAccGatewayRuleConfig_basic(rName, rNameRuntime, rImageUri),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayRuleExists(ctx, t, resourceName, &gatewayRule),
					resource.TestCheckResourceAttr(resourceName, names.AttrPriority, "100"),
					resource.TestCheckResourceAttr(resourceName, names.AttrCondition+".#", "0"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreGatewayRule_matchPrincipals(t *testing.T) {
	ctx := acctest.Context(t)
	var gatewayRule bedrockagentcorecontrol.GetGatewayRuleOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameRuntime := randomWithPrefixAndUnderscore(t)
	rImageUri := acctest.SkipIfEnvVarNotSet(t, "AWS_BEDROCK_AGENTCORE_RUNTIME_IMAGE_V1_URI")
	resourceName := "aws_bedrockagentcore_gateway_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayRuleConfig_matchPrincipals(rName, rNameRuntime, rImageUri),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayRuleExists(ctx, t, resourceName, &gatewayRule),
					resource.TestCheckResourceAttr(resourceName, names.AttrCondition+".#", "1"),
					resource.TestCheckResourceAttr(resourceName, "condition.0.match_principals.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "condition.0.match_principals.0.any_of.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "condition.0.match_principals.0.any_of.0.iam_principal.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "condition.0.match_principals.0.any_of.0.iam_principal.0.operator", "StringLike"),
					resource.TestCheckResourceAttrSet(resourceName, "condition.0.match_principals.0.any_of.0.iam_principal.0.arn"),
				),
			},
		},
	})
}

func TestAccBedrockAgentCoreGatewayRule_weightedRoute(t *testing.T) {
	ctx := acctest.Context(t)
	var gatewayRule bedrockagentcorecontrol.GetGatewayRuleOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameRuntime := randomWithPrefixAndUnderscore(t)
	rImageUri := acctest.SkipIfEnvVarNotSet(t, "AWS_BEDROCK_AGENTCORE_RUNTIME_IMAGE_V1_URI")
	resourceName := "aws_bedrockagentcore_gateway_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayRuleConfig_weightedRoute(rName, rNameRuntime, rImageUri),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayRuleExists(ctx, t, resourceName, &gatewayRule),
					resource.TestCheckResourceAttr(resourceName, names.AttrAction+".#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.route_to_target.0.weighted_route.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "action.0.route_to_target.0.weighted_route.0.traffic_split.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "action.0.route_to_target.0.weighted_route.0.traffic_split.0.weight", "70"),
					resource.TestCheckResourceAttr(resourceName, "action.0.route_to_target.0.weighted_route.0.traffic_split.1.weight", "30"),
				),
			},
		},
	})
}

func testAccCheckGatewayRuleDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagentcore_gateway_rule" {
				continue
			}

			_, err := tfbedrockagentcore.FindGatewayRuleByTwoPartKey(ctx, conn, rs.Primary.Attributes["gateway_identifier"], rs.Primary.Attributes["rule_id"])
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}
			return fmt.Errorf("Bedrock AgentCore Gateway Rule %s still exists", rs.Primary.Attributes["rule_id"])
		}

		return nil
	}
}

func testAccCheckGatewayRuleExists(ctx context.Context, t *testing.T, n string, v *bedrockagentcorecontrol.GetGatewayRuleOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		resp, err := tfbedrockagentcore.FindGatewayRuleByTwoPartKey(ctx, conn, rs.Primary.Attributes["gateway_identifier"], rs.Primary.Attributes["rule_id"])
		if err != nil {
			return err
		}

		*v = *resp
		return nil
	}
}

// testAccGatewayRuleConfig_base provisions an HTTP-protocol agent_runtime, gateway,
// and a gateway_target pointed at that runtime so route_to_target actions can
// reference it by name. `routeToTarget` requires HTTP-protocol targets, not MCP.
func testAccGatewayRuleConfig_base(rName, rNameRuntime, rImageUri string) string {
	return acctest.ConfigCompose(testAccAgentRuntimeConfig_protocolConfiguration(rNameRuntime, rImageUri, "HTTP"), fmt.Sprintf(`
data "aws_iam_policy_document" "gateway_assume" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["bedrock-agentcore.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "gateway" {
  name               = "%[1]s-gateway"
  assume_role_policy = data.aws_iam_policy_document.gateway_assume.json
}

resource "aws_bedrockagentcore_gateway" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.gateway.arn

  authorizer_type = "CUSTOM_JWT"
  authorizer_configuration {
    custom_jwt_authorizer {
      discovery_url    = "https://accounts.google.com/.well-known/openid-configuration"
      allowed_audience = ["test"]
    }
  }
}

resource "aws_bedrockagentcore_gateway_target" "test" {
  name               = %[1]q
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id

  credential_provider_configuration {
    gateway_iam_role {}
  }

  target_configuration {
    http {
      agentcore_runtime {
        arn = aws_bedrockagentcore_agent_runtime.test.agent_runtime_arn
      }
    }
  }
}
`, rName))
}

func testAccGatewayRuleConfig_basic(rName, rNameRuntime, rImageUri string) string {
	return acctest.ConfigCompose(testAccGatewayRuleConfig_base(rName, rNameRuntime, rImageUri), `
resource "aws_bedrockagentcore_gateway_rule" "test" {
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id
  priority           = 100

  action {
    route_to_target {
      static_route {
        target_name = aws_bedrockagentcore_gateway_target.test.name
      }
    }
  }
}
`)
}

func testAccGatewayRuleConfig_conditions(rName, rNameRuntime, rImageUri string, priority int, description string) string {
	return acctest.ConfigCompose(testAccGatewayRuleConfig_base(rName, rNameRuntime, rImageUri), fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_bedrockagentcore_gateway_rule" "test" {
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id
  priority           = %[1]d
  description        = %[2]q

  action {
    route_to_target {
      static_route {
        target_name = aws_bedrockagentcore_gateway_target.test.name
      }
    }
  }

  condition {
    match_principals {
      any_of {
        iam_principal {
          arn      = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:role/agentcore-caller-*"
          operator = "StringLike"
        }
      }
    }
  }
}
`, priority, description))
}

// testAccGatewayRuleConfig_conditionsNoDescription omits the description argument
// entirely to prove it is retained (not cleared, not an error) on update — the
// UpdateGatewayRule API treats a nil description as "leave unchanged".
func testAccGatewayRuleConfig_conditionsNoDescription(rName, rNameRuntime, rImageUri string, priority int) string {
	return acctest.ConfigCompose(testAccGatewayRuleConfig_base(rName, rNameRuntime, rImageUri), fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_bedrockagentcore_gateway_rule" "test" {
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id
  priority           = %[1]d

  action {
    route_to_target {
      static_route {
        target_name = aws_bedrockagentcore_gateway_target.test.name
      }
    }
  }

  condition {
    match_principals {
      any_of {
        iam_principal {
          arn      = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:role/agentcore-caller-*"
          operator = "StringLike"
        }
      }
    }
  }
}
`, priority))
}

func testAccGatewayRuleConfig_matchPrincipals(rName, rNameRuntime, rImageUri string) string {
	return acctest.ConfigCompose(testAccGatewayRuleConfig_base(rName, rNameRuntime, rImageUri), `
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_bedrockagentcore_gateway_rule" "test" {
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id
  priority           = 100

  action {
    route_to_target {
      static_route {
        target_name = aws_bedrockagentcore_gateway_target.test.name
      }
    }
  }

  condition {
    match_principals {
      any_of {
        iam_principal {
          arn      = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:role/service-role/*"
          operator = "StringLike"
        }
      }
    }
  }
}
`)
}

func testAccGatewayRuleConfig_weightedRoute(rName, rNameRuntime, rImageUri string) string {
	return acctest.ConfigCompose(testAccGatewayRuleConfig_base(rName, rNameRuntime, rImageUri), fmt.Sprintf(`
resource "aws_bedrockagentcore_gateway_target" "test2" {
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id
  name               = "%[1]sb"

  credential_provider_configuration {
    gateway_iam_role {}
  }

  target_configuration {
    http {
      agentcore_runtime {
        arn = aws_bedrockagentcore_agent_runtime.test.agent_runtime_arn
      }
    }
  }
}

resource "aws_bedrockagentcore_gateway_rule" "test" {
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id
  priority           = 100

  action {
    route_to_target {
      weighted_route {
        traffic_split {
          name        = "primary"
          target_name = aws_bedrockagentcore_gateway_target.test.name
          weight      = 70
        }
        traffic_split {
          name        = "canary"
          target_name = aws_bedrockagentcore_gateway_target.test2.name
          weight      = 30
        }
      }
    }
  }
}
`, rName))
}
