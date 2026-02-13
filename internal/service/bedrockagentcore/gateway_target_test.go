// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfbedrockagentcore "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagentcore"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgentCoreGatewayTarget_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var gatewayTarget bedrockagentcorecontrol.GetGatewayTargetOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_gateway_target.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckGatewayTargets(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayTargetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayTargetConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, t, resourceName, &gatewayTarget),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, "gateway_identifier"),
					resource.TestCheckResourceAttrSet(resourceName, "target_id"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.gateway_iam_role.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.0.lambda.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "target_configuration.0.mcp.0.lambda.0.lambda_arn"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.0.lambda.0.tool_schema.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.0.lambda.0.tool_schema.0.inline_payload.#", "1"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, ",", "gateway_identifier", "target_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "target_id",
			},
		},
	})
}

func TestAccBedrockAgentCoreGatewayTarget_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var gatewayTarget bedrockagentcorecontrol.GetGatewayTargetOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_gateway_target.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckGatewayTargets(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayTargetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayTargetConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, t, resourceName, &gatewayTarget),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfbedrockagentcore.ResourceGatewayTarget, resourceName),
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

func TestAccBedrockAgentCoreGatewayTarget_targetConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var gatewayTarget, gatewayTargetPrev bedrockagentcorecontrol.GetGatewayTargetOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_gateway_target.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckGatewayTargets(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayTargetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayTargetConfig_targetConfiguration(rName, testAccSchema_primitive()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, t, resourceName, &gatewayTarget),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.0.lambda.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.0.lambda.0.tool_schema.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.0.lambda.0.tool_schema.0.inline_payload.#", "1"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			// Example 2: Object with properties + required
			{
				Config: testAccGatewayTargetConfig_targetConfiguration(rName, testAccSchema_objectWithProperties()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, t, resourceName, &gatewayTargetPrev),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			// Example 3: Array of primitives
			{
				Config: testAccGatewayTargetConfig_targetConfiguration(rName, testAccSchema_arrayOfPrimitives()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, t, resourceName, &gatewayTarget),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			// Example 4: Array of objects
			{
				Config: testAccGatewayTargetConfig_targetConfiguration(rName, testAccSchema_arrayOfObjects()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, t, resourceName, &gatewayTarget),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			// Example 5: Array of arrays
			{
				Config: testAccGatewayTargetConfig_targetConfiguration(rName, testAccSchema_arrayOfArrays()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, t, resourceName, &gatewayTarget),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			//Example 6: Mixed nested object/array
			{
				Config: testAccGatewayTargetConfig_targetConfiguration(rName, testAccSchema_mixedNested()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, t, resourceName, &gatewayTargetPrev),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			// Example 7: Array with ignored keywords
			{
				Config: testAccGatewayTargetConfig_targetConfiguration(rName, testAccSchema_arrayWithIgnoredKeywords()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, t, resourceName, &gatewayTarget),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			// Invalid Example 8: Both items and properties at the same node
			{
				Config:      testAccGatewayTargetConfig_targetConfiguration(rName, testAccSchema_invalidBothItemsAndProperties()),
				ExpectError: regexache.MustCompile("Invalid Attribute Combination"),
			},
			// Invalid Example 9: Missing type
			{
				Config:      testAccGatewayTargetConfig_targetConfiguration(rName, testAccSchema_invalidMissingType()),
				ExpectError: regexache.MustCompile("Missing required argument"),
			},
			// Invalid Example 10: Unsupported type
			{
				Config:      testAccGatewayTargetConfig_targetConfiguration(rName, testAccSchema_invalidUnsupportedType()),
				ExpectError: regexache.MustCompile("Invalid String Enum Value"),
			},
			// Return to valid configuration to proceed with post-test destroy
			{
				Config: testAccGatewayTargetConfig_targetConfiguration(rName, testAccSchema_objectWithProperties()),
			},
		},
	})
}

func TestAccBedrockAgentCoreGatewayTarget_targetConfigurationMCPServer(t *testing.T) {
	ctx := acctest.Context(t)
	var gatewayTarget bedrockagentcorecontrol.GetGatewayTargetOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_gateway_target.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckGatewayTargets(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayTargetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayTargetConfig_targetConfigurationMCPServer(rName, "https://knowledge-mcp.global.api.aws"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, t, resourceName, &gatewayTarget),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.0.mcp_server.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.0.mcp_server.0.endpoint", "https://knowledge-mcp.global.api.aws"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccGatewayTargetConfig_targetConfigurationMCPServer(rName, "https://docs.mcp.cloudflare.com/mcp"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, t, resourceName, &gatewayTarget),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.0.mcp_server.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.0.mcp_server.0.endpoint", "https://docs.mcp.cloudflare.com/mcp"),
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

func TestAccBedrockAgentCoreGatewayTarget_credentialProvider(t *testing.T) {
	ctx := acctest.Context(t)
	var gatewayTarget, gatewayTargetPrev bedrockagentcorecontrol.GetGatewayTargetOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_gateway_target.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckGatewayTargets(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayTargetDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Gateway IAM Role provider with Lambda target
			{
				Config: testAccGatewayTargetConfig_credentialProvider(rName, testAccCredentialProvider_gatewayIAMRole()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, t, resourceName, &gatewayTarget),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.gateway_iam_role.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.api_key.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.oauth.#", "0"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			// Step 2: API Key provider with OpenAPI Schema target (creates new resource)
			{
				Config: testAccGatewayTargetConfig_credentialProviderNonLambda(rName, testAccCredentialProvider_apiKey()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, t, resourceName, &gatewayTargetPrev),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.api_key.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.gateway_iam_role.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.oauth.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "credential_provider_configuration.0.api_key.0.provider_arn"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
			// Step 3: OAuth provider with OpenAPI Schema target (updates credential provider only)
			{
				Config: testAccGatewayTargetConfig_credentialProviderNonLambda(rName, testAccCredentialProvider_oauth()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, t, resourceName, &gatewayTarget),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.oauth.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.api_key.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.gateway_iam_role.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "credential_provider_configuration.0.oauth.0.provider_arn"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			// Step 4: Gateway IAM Role provider with Smithy Model target (creates new resource due to both changes)
			{
				Config: testAccGatewayTargetConfig_credentialProviderSmithy(rName, testAccCredentialProvider_gatewayIAMRole()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, t, resourceName, &gatewayTargetPrev),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.gateway_iam_role.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.api_key.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.oauth.#", "0"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
			// Step 5: Back to Gateway IAM Role with Lambda target (creates new resource again)
			{
				Config: testAccGatewayTargetConfig_credentialProvider(rName, testAccCredentialProvider_gatewayIAMRole()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, t, resourceName, &gatewayTargetPrev),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.gateway_iam_role.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.api_key.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.oauth.#", "0"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, ",", "gateway_identifier", "target_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "target_id",
			},
		},
	})
}

func TestAccBedrockAgentCoreGatewayTarget_credentialProvider_invalid(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckGatewayTargets(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayTargetDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Invalid: Multiple credential providers
			{
				Config:      testAccGatewayTargetConfig_credentialProvider(rName, testAccCredentialProvider_multipleProviders()),
				ExpectError: regexache.MustCompile(`Invalid Attribute Combination|cannot be specified`),
			},
			{
				Config:      testAccGatewayTargetConfig_credentialProvider(rName, testAccCredentialProvider_empty()),
				ExpectError: regexache.MustCompile("Invalid Credential Provider Configuration|At least one credential provider must be configured"),
			},
		},
	})
}

func testAccCheckGatewayTargetDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagentcore_gateway_target" {
				continue
			}

			_, err := tfbedrockagentcore.FindGatewayTargetByTwoPartKey(ctx, conn, rs.Primary.Attributes["gateway_identifier"], rs.Primary.Attributes["target_id"])
			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Bedrock Agent Core Gateway Target %s still exists", rs.Primary.Attributes["target_id"])
		}

		return nil
	}
}

func testAccCheckGatewayTargetExists(ctx context.Context, t *testing.T, n string, v *bedrockagentcorecontrol.GetGatewayTargetOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		resp, err := tfbedrockagentcore.FindGatewayTargetByTwoPartKey(ctx, conn, rs.Primary.Attributes["gateway_identifier"], rs.Primary.Attributes["target_id"])
		if err != nil {
			return err
		}

		*v = *resp

		return nil
	}
}

func testAccPreCheckGatewayTargets(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

	input := bedrockagentcorecontrol.ListGatewayTargetsInput{
		GatewayIdentifier: aws.String("test-guthipm3lw"), // Using a dummy ID for the precheck
	}

	_, err := conn.ListGatewayTargets(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccGatewayTargetConfig_infra(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_iam_policy_document" "test" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["bedrock-agentcore.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test.json
}

data "aws_iam_policy_document" "lambda_assume" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "lambda" {
  name = "%[1]s-lambda"

  assume_role_policy = data.aws_iam_policy_document.lambda_assume.json
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.lambda.arn
  handler       = "lambdatest.handler"
  runtime       = "nodejs20.x"
}

resource "aws_bedrockagentcore_gateway" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  authorizer_type = "CUSTOM_JWT"
  authorizer_configuration {
    custom_jwt_authorizer {
      discovery_url    = "https://accounts.google.com/.well-known/openid-configuration"
      allowed_audience = ["test"]
    }
  }

  protocol_type = "MCP"
}
`, rName)
}

func testAccGatewayTargetConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccGatewayTargetConfig_infra(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_gateway_target" "test" {
  name               = %[1]q
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id

  credential_provider_configuration {
    gateway_iam_role {}
  }

  target_configuration {
    mcp {
      lambda {
        lambda_arn = aws_lambda_function.test.arn

        tool_schema {
          inline_payload {
            name        = "test_tool"
            description = "A test tool"

            input_schema {
              type = "object"

              property {
                name        = "input"
                description = "some input"
                type        = "string"
                required    = true
              }
            }
          }
        }
      }
    }
  }
}

`, rName))
}

func testAccGatewayTargetConfig_credentialProvider(rName, credentialProviderContent string) string {
	return acctest.ConfigCompose(testAccGatewayTargetConfig_infra(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_gateway_target" "test" {
  name               = %[1]q
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id

  credential_provider_configuration {
%[2]s
  }

  target_configuration {
    mcp {
      lambda {
        lambda_arn = aws_lambda_function.test.arn

        tool_schema {
          inline_payload {
            name        = "test_tool"
            description = "A test tool"

            input_schema {
              type        = "string"
              description = "Basic schema for credential provider test"
            }
          }
        }
      }
    }
  }
}
`, rName, credentialProviderContent))
}

func testAccGatewayTargetConfig_credentialProviderNonLambda(rName, credentialProviderContent string) string {
	return acctest.ConfigCompose(testAccGatewayTargetConfig_infra(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_gateway_target" "test" {
  name               = %[1]q
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id

  credential_provider_configuration {
%[2]s
  }

  target_configuration {
    mcp {
      open_api_schema {
        inline_payload {
          payload = jsonencode({
            openapi = "3.0.0"
            info = {
              title   = "Test API"
              version = "1.0.0"
            }
            servers = [
              {
                url = "https://api.example.com"
              }
            ]
            paths = {
              "/test" = {
                get = {
                  operationId = "getTest"
                  summary     = "Test endpoint"
                  responses = {
                    "200" = {
                      description = "Success"
                    }
                  }
                }
              }
            }
          })
        }
      }
    }
  }
}
`, rName, credentialProviderContent))
}

func testAccGatewayTargetConfig_credentialProviderSmithy(rName, credentialProviderContent string) string {
	return acctest.ConfigCompose(testAccGatewayTargetConfig_infra(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_gateway_target" "test" {
  name               = %[1]q
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id

  credential_provider_configuration {
%[2]s
  }

  target_configuration {
    mcp {
      smithy_model {
        inline_payload {
          payload = jsonencode({
            "smithy" = "2.0"
            "shapes" = {
              "com.example#TestService" = {
                "type"    = "service"
                "version" = "1.0"
                "operations" = [
                  {
                    "target" = "com.example#TestOperation"
                  }
                ]
                "traits" = {
                  "aws.auth#sigv4" = {
                    "name" = "testservice"
                  }
                  "aws.protocols#restJson1" = {}
                }
              }
              "com.example#TestOperation" = {
                "type" = "operation"
                "input" = {
                  "target" = "com.example#TestInput"
                }
                "output" = {
                  "target" = "com.example#TestOutput"
                }
                "traits" = {
                  "smithy.api#http" = {
                    "method" = "POST"
                    "uri"    = "/test"
                  }
                }
              }
              "com.example#TestInput" = {
                "type" = "structure"
                "members" = {
                  "message" = {
                    "target" = "smithy.api#String"
                    "traits" = {
                      "smithy.api#required" = {}
                    }
                  }
                }
              }
              "com.example#TestOutput" = {
                "type" = "structure"
                "members" = {
                  "result" = {
                    "target" = "smithy.api#String"
                  }
                }
              }
            }
          })
        }
      }
    }
  }
}
`, rName, credentialProviderContent))
}

func testAccGatewayTargetConfig_targetConfiguration(rName, schemaContent string) string {
	return acctest.ConfigCompose(testAccGatewayTargetConfig_infra(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_gateway_target" "test" {
  name               = %[1]q
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id

  credential_provider_configuration {
    gateway_iam_role {}
  }

  target_configuration {
    mcp {
      lambda {
        lambda_arn = aws_lambda_function.test.arn

        tool_schema {
          inline_payload {
            name        = "test_tool"
            description = "A test tool"

            input_schema {
              %[2]s
            }
          }
        }
      }
    }
  }
}
`, rName, schemaContent))
}

func testAccGatewayTargetConfig_targetConfigurationMCPServer(rName, endpoint string) string {
	return acctest.ConfigCompose(testAccGatewayTargetConfig_infra(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_gateway_target" "test" {
  name               = %[1]q
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id

  target_configuration {
    mcp {
      mcp_server {
        endpoint = %[2]q
      }
    }
  }
}
`, rName, endpoint))
}

func testAccSchema_primitive() string {
	return `
			type        = "string"
			description = "A token"
 		   `
}

func testAccSchema_objectWithProperties() string {
	return `
			type        = "object"
			description = "User"

			property {
				name     = "id"
				type     = "string"
				required = true
			}

			property {
				name = "age"
				type = "integer"
			}

			property {
				name = "paid"
				type = "boolean"
			}
		 `
}

func testAccSchema_arrayOfPrimitives() string {
	return `
			type        = "array"
			description = "Tags"

			items {
				type = "string"
			}
		 `
}

func testAccSchema_arrayOfObjects() string {
	return `
			type = "array"

			items {
				type = "object"

				property {
					name     = "id"
					type     = "string"
					required = true
				}

				property {
					name = "email"
					type = "string"
				}

				property {
					name = "age"
					type = "integer"
				}
			}
		 `
}

func testAccSchema_arrayOfArrays() string {
	return `
			type = "array"

			items {
				type = "array"

				items {
					type = "number"
				}
			}
		 `
}

func testAccSchema_mixedNested() string {
	return `
			type = "object"

			property {
				name = "profile"
				type = "object"

				property {
					name       = "nested_tags"
					type       = "array"
					items_json = jsonencode({
						type = "string"
					})
				}
			}
		 `
}

func testAccSchema_arrayWithIgnoredKeywords() string {
	return `
			type = "array"

			items {
				type = "string"
			}
		 `
}

func testAccSchema_invalidBothItemsAndProperties() string {
	return `
			type = "object"

			items {
				type = "string"
			}

			property {
				name = "a"
				type = "string"
			}
		 `
}

func testAccSchema_invalidMissingType() string {
	return `
			description = "No type here"
		 `
}

func testAccSchema_invalidUnsupportedType() string {
	return `
			type = "date"
		 `
}

func testAccCredentialProvider_gatewayIAMRole() string {
	return `    gateway_iam_role {}`
}

func testAccCredentialProvider_apiKey() string {
	return `    api_key {
      provider_arn              = "arn:${data.aws_partition.current.partition}:iam::123456789012:oidc-provider/example.com"
      credential_location       = "HEADER"
      credential_parameter_name = "X-API-Key"
      credential_prefix         = "Bearer"
    }`
}

func testAccCredentialProvider_oauth() string {
	return `    oauth {
      provider_arn = "arn:${data.aws_partition.current.partition}:iam::123456789012:oidc-provider/oauth.example.com"
      scopes       = ["read", "write"]
      custom_parameters = {
        "client_type" = "confidential"
        "grant_type"  = "authorization_code"
      }
    }`
}

func testAccCredentialProvider_multipleProviders() string {
	return `    gateway_iam_role {}
    api_key {
      provider_arn = "arn:${data.aws_partition.current.partition}:iam::123456789012:oidc-provider/example.com"
    }`
}

func testAccCredentialProvider_empty() string {
	return `    # No providers configured`
}
