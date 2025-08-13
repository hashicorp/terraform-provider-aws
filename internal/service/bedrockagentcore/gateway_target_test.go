// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	// TIP: ==== IMPORTS ====
	// This is a common set of imports but not customized to your code since
	// your code hasn't been written yet. Make sure you, your IDE, or
	// goimports -w <file> fixes these imports.
	//
	// The provider linter wants your imports to be in two groups: first,
	// standard library (i.e., "fmt" or "strings"), second, everything else.
	//
	// Also, AWS Go SDK v2 may handle nested structures differently than v1,
	// using the services/bedrockagentcorecontrol/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// types.<Type Name>.
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfbedrockagentcore "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagentcore"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgentCoreGatewayTarget_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var gatewayTarget bedrockagentcorecontrol.GetGatewayTargetOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_gateway_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayTargetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayTargetConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, resourceName, &gatewayTarget),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, "gateway_identifier"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.gateway_iam_role.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.0.lambda.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "target_configuration.0.mcp.0.lambda.0.lambda_arn"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.0.lambda.0.tool_schema.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.0.lambda.0.tool_schema.0.inline_payload.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccGatewayTargetImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccBedrockAgentCoreGatewayTarget_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var gatewayTarget bedrockagentcorecontrol.GetGatewayTargetOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_gateway_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayTargetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayTargetConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, resourceName, &gatewayTarget),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfbedrockagentcore.ResourceGatewayTarget, resourceName),
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

func testAccCheckGatewayTargetDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentCoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagentcore_gateway_target" {
				continue
			}

			_, err := tfbedrockagentcore.FindGatewayTargetByID(ctx, conn, rs.Primary.Attributes["gateway_identifier"], rs.Primary.ID)
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.BedrockAgentCore, create.ErrActionCheckingDestroyed, tfbedrockagentcore.ResNameGatewayTarget, rs.Primary.ID, err)
			}

			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingDestroyed, tfbedrockagentcore.ResNameGatewayTarget, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckGatewayTargetExists(ctx context.Context, name string, gatewayTarget *bedrockagentcorecontrol.GetGatewayTargetOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingExistence, tfbedrockagentcore.ResNameGatewayTarget, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingExistence, tfbedrockagentcore.ResNameGatewayTarget, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BedrockAgentCoreClient(ctx)

		resp, err := tfbedrockagentcore.FindGatewayTargetByID(ctx, conn, rs.Primary.Attributes["gateway_identifier"], rs.Primary.ID)
		if err != nil {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingExistence, tfbedrockagentcore.ResNameGatewayTarget, rs.Primary.ID, err)
		}

		*gatewayTarget = *resp

		return nil
	}
}

func testAccCheckGatewayTargetNotRecreated(before, after *bedrockagentcorecontrol.GetGatewayTargetOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.TargetId), aws.ToString(after.TargetId); before != after {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingNotRecreated, tfbedrockagentcore.ResNameGatewayTarget, before, errors.New("recreated"))
		}

		return nil
	}
}

func testAccGatewayTargetConfig_infra(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_iam_policy_document" "test_assume" {
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
  assume_role_policy = data.aws_iam_policy_document.test_assume.json
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

  authorizer_configuration {
    custom_jwt_authorizer {
      discovery_url    = "https://accounts.google.com/.well-known/openid-configuration"
      allowed_audience = ["test"]
    }
  }
}
`, rName)
}

func testAccGatewayTargetImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}

		gatewayIdentifier := rs.Primary.Attributes["gateway_identifier"]
		targetId := rs.Primary.ID

		return fmt.Sprintf("%s,%s", gatewayIdentifier, targetId), nil
	}
}

func testAccGatewayTargetConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccGatewayTargetConfig_infra(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_gateway_target" "test" {
  name               = %[1]q
  gateway_identifier = aws_bedrockagentcore_gateway.test.id

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

func TestAccBedrockAgentCoreGatewayTarget_targetConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var gatewayTarget, gatewayTargetPrev bedrockagentcorecontrol.GetGatewayTargetOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_gateway_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayTargetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayTargetConfig_targetConfiguration(rName, testAccSchema_primitive()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, resourceName, &gatewayTarget),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.0.lambda.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.0.lambda.0.tool_schema.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.0.lambda.0.tool_schema.0.inline_payload.#", "1"),
				),
			},
			// Example 2: Object with properties + required
			{
				Config: testAccGatewayTargetConfig_targetConfiguration(rName, testAccSchema_objectWithProperties()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, resourceName, &gatewayTargetPrev),
					testAccCheckGatewayTargetNotRecreated(&gatewayTarget, &gatewayTargetPrev),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			// Example 3: Array of primitives
			{
				Config: testAccGatewayTargetConfig_targetConfiguration(rName, testAccSchema_arrayOfPrimitives()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, resourceName, &gatewayTarget),
					testAccCheckGatewayTargetNotRecreated(&gatewayTargetPrev, &gatewayTarget),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			// Example 4: Array of objects
			{
				Config: testAccGatewayTargetConfig_targetConfiguration(rName, testAccSchema_arrayOfObjects()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, resourceName, &gatewayTarget),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			// Example 5: Array of arrays
			{
				Config: testAccGatewayTargetConfig_targetConfiguration(rName, testAccSchema_arrayOfArrays()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, resourceName, &gatewayTarget),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			//Example 6: Mixed nested object/array
			{
				Config: testAccGatewayTargetConfig_targetConfiguration(rName, testAccSchema_mixedNested()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, resourceName, &gatewayTargetPrev),
					testAccCheckGatewayTargetNotRecreated(&gatewayTarget, &gatewayTargetPrev),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			// Example 7: Array with ignored keywords
			{
				Config: testAccGatewayTargetConfig_targetConfiguration(rName, testAccSchema_arrayWithIgnoredKeywords()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, resourceName, &gatewayTarget),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
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

func TestAccBedrockAgentCoreGatewayTarget_credentialProvider(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var gatewayTarget, gatewayTargetPrev bedrockagentcorecontrol.GetGatewayTargetOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_gateway_target.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayTargetDestroy(ctx),
		Steps: []resource.TestStep{
			// Step 1: Gateway IAM Role provider with Lambda target
			{
				Config: testAccGatewayTargetConfig_credentialProvider(rName, testAccCredentialProvider_gatewayIAMRole()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, resourceName, &gatewayTarget),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.gateway_iam_role.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.api_key.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.oauth.#", "0"),
				),
			},
			// Step 2: API Key provider with OpenAPI Schema target (creates new resource)
			{
				Config: testAccGatewayTargetConfig_credentialProviderNonLambda(rName, testAccCredentialProvider_apiKey()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, resourceName, &gatewayTargetPrev),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.api_key.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.gateway_iam_role.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.oauth.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "credential_provider_configuration.0.api_key.0.provider_arn"),
				),
			},
			// Step 3: OAuth provider with OpenAPI Schema target (updates credential provider only)
			{
				Config: testAccGatewayTargetConfig_credentialProviderNonLambda(rName, testAccCredentialProvider_oauth()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, resourceName, &gatewayTarget),
					testAccCheckGatewayTargetNotRecreated(&gatewayTargetPrev, &gatewayTarget),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.oauth.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.api_key.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.gateway_iam_role.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "credential_provider_configuration.0.oauth.0.provider_arn"),
				),
			},
			// Step 4: Gateway IAM Role provider with Smithy Model target (creates new resource due to both changes)
			{
				Config: testAccGatewayTargetConfig_credentialProviderSmithy(rName, testAccCredentialProvider_gatewayIAMRole()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, resourceName, &gatewayTargetPrev),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.gateway_iam_role.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.api_key.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.oauth.#", "0"),
				),
			},
			// Step 5: Back to Gateway IAM Role with Lambda target (creates new resource again)
			{
				Config: testAccGatewayTargetConfig_credentialProvider(rName, testAccCredentialProvider_gatewayIAMRole()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, resourceName, &gatewayTargetPrev),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.gateway_iam_role.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.api_key.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "credential_provider_configuration.0.oauth.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccGatewayTargetImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccBedrockAgentCoreGatewayTarget_credentialProvider_invalid(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayTargetDestroy(ctx),
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

func testAccGatewayTargetConfig_credentialProvider(rName, credentialProviderContent string) string {
	return acctest.ConfigCompose(testAccGatewayTargetConfig_infra(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_gateway_target" "test" {
  name               = %[1]q
  gateway_identifier = aws_bedrockagentcore_gateway.test.id

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
  gateway_identifier = aws_bedrockagentcore_gateway.test.id

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
  gateway_identifier = aws_bedrockagentcore_gateway.test.id

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
  gateway_identifier = aws_bedrockagentcore_gateway.test.id

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

// Example 1: Primitive schema - { "type": "string", "description": "A token" }
func testAccSchema_primitive() string {
	return `
			type        = "string"
			description = "A token"
 		   `
}

// Example 2: Object with properties + required
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

// Example 3: Array of primitives
func testAccSchema_arrayOfPrimitives() string {
	return `
			type        = "array"
			description = "Tags"

			items {
				type = "string"
			}
		 `
}

// Example 4: Array of objects (element has flat props)
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

// Example 5: Array of arrays (consecutive arrays)
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

// Example 6: Mixed nested object/array
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

// Example 7: Array with extra (ignored) JSON-Schema keywords
func testAccSchema_arrayWithIgnoredKeywords() string {
	return `
			type = "array"

			items {
				type = "string"
			}
		 `
}

// Invalid Example A: items present but type != "array"
//func testAccSchema_invalidItemsOnObject() string {
//	return `
//			type = "object"
//
//			items {
//				type = "string"
//			}
//		 `
//}
//
//// Invalid Example B: properties present but type != "object"
//func testAccSchema_invalidPropertiesOnArray() string {
//	return `
//			type = "array"
//
//			property {
//				name = "x"
//				type = "string"
//			}
//		 `
//}
//
//// Invalid Example C: required outside of properties / not a subset
//func testAccSchema_invalidRequiredNotSubset() string {
//	return `
//			type = "object"
//
//			property {
//				name     = "a"
//				type     = "string"
//				required = true
//			}
//
//			property {
//				name     = "b"
//				type     = "string"
//				required = true
//			}
//		 `
//}

// Invalid Example D: Both items and properties at the same node
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

// Invalid Example E: Missing type
func testAccSchema_invalidMissingType() string {
	return `
			description = "No type here"
		 `
}

// Invalid Example F: Unsupported type
func testAccSchema_invalidUnsupportedType() string {
	return `
			type = "date"
		 `
}

// Credential Provider Helper Functions

// Gateway IAM Role provider (no configuration needed)
func testAccCredentialProvider_gatewayIAMRole() string {
	return `    gateway_iam_role {}`
}

// API Key provider with all optional parameters
func testAccCredentialProvider_apiKey() string {
	return `    api_key {
      provider_arn              = "arn:${data.aws_partition.current.partition}:iam::123456789012:oidc-provider/example.com"
      credential_location       = "HEADER"
      credential_parameter_name = "X-API-Key"
      credential_prefix         = "Bearer"
    }`
}

// OAuth provider with required and optional parameters
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

// Invalid: Multiple providers configured simultaneously
func testAccCredentialProvider_multipleProviders() string {
	return `    gateway_iam_role {}
    api_key {
      provider_arn = "arn:${data.aws_partition.current.partition}:iam::123456789012:oidc-provider/example.com"
    }`
}

// Invalid: Empty credential provider configuration
func testAccCredentialProvider_empty() string {
	return `    # No providers configured`
}
