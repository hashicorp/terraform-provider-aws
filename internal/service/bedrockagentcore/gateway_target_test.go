// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
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

func TestAccBedrockAgentCoreGatewayTarget_targetConfigurationAPIGateway(t *testing.T) {
	ctx := acctest.Context(t)
	var gatewayTarget bedrockagentcorecontrol.GetGatewayTargetOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_gateway_target.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayTargetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayTargetConfig_targetConfigurationAPIGateway(rName, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, t, resourceName, &gatewayTarget),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.0.api_gateway.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "target_configuration.0.mcp.0.api_gateway.0.rest_api_id", "aws_api_gateway_rest_api.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "target_configuration.0.mcp.0.api_gateway.0.stage", "aws_api_gateway_stage.test", "stage_name"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.0.api_gateway.0.api_gateway_tool_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.0.api_gateway.0.api_gateway_tool_configuration.0.tool_filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.0.api_gateway.0.api_gateway_tool_configuration.0.tool_filter.0.filter_path", "/pets"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.0.api_gateway.0.api_gateway_tool_configuration.0.tool_filter.0.methods.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "target_configuration.0.mcp.0.api_gateway.0.api_gateway_tool_configuration.0.tool_filter.0.methods.*", "GET"),
					resource.TestCheckTypeSetElemAttr(resourceName, "target_configuration.0.mcp.0.api_gateway.0.api_gateway_tool_configuration.0.tool_filter.0.methods.*", "POST"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target_configuration.0.mcp.0.api_gateway.0.api_gateway_tool_configuration.0.tool_override.*", map[string]string{
						names.AttrName:        "ListPets",
						names.AttrPath:        "/pets",
						"method":              "GET",
						names.AttrDescription: "Retrieves all available pets",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target_configuration.0.mcp.0.api_gateway.0.api_gateway_tool_configuration.0.tool_override.*", map[string]string{
						names.AttrName:        "RegisterPets",
						names.AttrPath:        "/pets",
						"method":              "POST",
						names.AttrDescription: "Register pets",
					}),
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
			{
				Config: testAccGatewayTargetConfig_targetConfigurationAPIGateway(rName, "2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, t, resourceName, &gatewayTarget),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.0.api_gateway.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "target_configuration.0.mcp.0.api_gateway.0.rest_api_id", "aws_api_gateway_rest_api.test", names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "target_configuration.0.mcp.0.api_gateway.0.stage", "aws_api_gateway_stage.test", "stage_name"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.0.api_gateway.0.api_gateway_tool_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.0.api_gateway.0.api_gateway_tool_configuration.0.tool_filter.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.0.api_gateway.0.api_gateway_tool_configuration.0.tool_filter.0.filter_path", "/pets"),
					resource.TestCheckResourceAttr(resourceName, "target_configuration.0.mcp.0.api_gateway.0.api_gateway_tool_configuration.0.tool_filter.0.methods.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "target_configuration.0.mcp.0.api_gateway.0.api_gateway_tool_configuration.0.tool_filter.0.methods.*", "GET"),
					resource.TestCheckTypeSetElemAttr(resourceName, "target_configuration.0.mcp.0.api_gateway.0.api_gateway_tool_configuration.0.tool_filter.0.methods.*", "POST"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target_configuration.0.mcp.0.api_gateway.0.api_gateway_tool_configuration.0.tool_override.*", map[string]string{
						names.AttrName:        "ListPets2",
						names.AttrPath:        "/pets",
						"method":              "GET",
						names.AttrDescription: "Retrieves all available pets2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "target_configuration.0.mcp.0.api_gateway.0.api_gateway_tool_configuration.0.tool_override.*", map[string]string{
						names.AttrName:        "RegisterPets2",
						names.AttrPath:        "/pets",
						"method":              "POST",
						names.AttrDescription: "Register pets2",
					}),
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

func TestAccBedrockAgentCoreGatewayTarget_metadataConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var gatewayTarget bedrockagentcorecontrol.GetGatewayTargetOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_gateway_target.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayTargetDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayTargetConfig_metadataConfiguration(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, t, resourceName, &gatewayTarget),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "metadata_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metadata_configuration.0.allowed_request_headers.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "metadata_configuration.0.allowed_response_headers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metadata_configuration.0.allowed_query_parameters.#", "1"),
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
			// Update metadata configuration
			{
				Config: testAccGatewayTargetConfig_metadataConfigurationUpdated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, t, resourceName, &gatewayTarget),
					resource.TestCheckResourceAttr(resourceName, "metadata_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metadata_configuration.0.allowed_request_headers.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "metadata_configuration.0.allowed_response_headers.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "metadata_configuration.0.allowed_query_parameters.#", "2"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			// Remove metadata configuration
			{
				Config: testAccGatewayTargetConfig_metadataConfigurationRemoved(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayTargetExists(ctx, t, resourceName, &gatewayTarget),
					resource.TestCheckResourceAttr(resourceName, "metadata_configuration.#", "0"),
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

func TestAccBedrockAgentCoreGatewayTarget_metadataConfiguration_invalidHeaders(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayTargetDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Invalid: restricted header Authorization
			{
				Config:      testAccGatewayTargetConfig_metadataConfigurationInvalidHeader(rName, "Authorization"),
				ExpectError: regexache.MustCompile(`none of \(case-insensitive\)`),
			},
			// Invalid: restricted header Content-Type
			{
				Config:      testAccGatewayTargetConfig_metadataConfigurationInvalidHeader(rName, "Content-Type"),
				ExpectError: regexache.MustCompile(`none of \(case-insensitive\)`),
			},
			// Invalid: restricted header Host
			{
				Config:      testAccGatewayTargetConfig_metadataConfigurationInvalidHeader(rName, "Host"),
				ExpectError: regexache.MustCompile(`none of \(case-insensitive\)`),
			},
			// Invalid: X-Amzn- prefix
			{
				Config:      testAccGatewayTargetConfig_metadataConfigurationInvalidHeader(rName, "X-Amzn-Custom"),
				ExpectError: regexache.MustCompile(`must not begin with \(case-insensitive\)`),
			},
			// Invalid: header with special characters
			{
				Config:      testAccGatewayTargetConfig_metadataConfigurationInvalidHeader(rName, "Invalid Header!"),
				ExpectError: regexache.MustCompile(`header names must contain only alphanumeric characters`),
			},
			// Valid: X-Amzn-Bedrock-AgentCore-Runtime-Custom- prefix is allowed
			{
				Config: testAccGatewayTargetConfig_metadataConfigurationInvalidHeader(rName, "X-Amzn-Bedrock-AgentCore-Runtime-Custom-MyHeader"),
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

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "lambda:*",
    "Resource": "*"
  }
}
  EOF
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

  protocol_configuration {
    mcp {
      instructions       = "Do something"
      search_type        = "SEMANTIC"
      supported_versions = ["2025-11-25"]
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
      provider_arn       = "arn:${data.aws_partition.current.partition}:iam::123456789012:oidc-provider/oauth.example.com"
      scopes             = ["read", "write"]
	  grant_type         = "AUTHORIZATION_CODE"
	  default_return_url = "https://example.com/callback"

      custom_parameters = {
        "client_type" = "confidential"
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

func testAccGatewayTargetConfig_metadataConfiguration(rName string) string {
	return acctest.ConfigCompose(testAccGatewayTargetConfig_infra(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_gateway_target" "test" {
  name               = %[1]q
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id

  target_configuration {
    mcp {
      mcp_server {
        endpoint = "https://docs.mcp.cloudflare.com/mcp"
      }
    }
  }

  metadata_configuration {
    allowed_request_headers  = ["x-correlation-id", "x-tenant-id"]
    allowed_response_headers = ["x-rate-limit-remaining"]
    allowed_query_parameters = ["version"]
  }
}
`, rName))
}

func testAccGatewayTargetConfig_metadataConfigurationUpdated(rName string) string {
	return acctest.ConfigCompose(testAccGatewayTargetConfig_infra(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_gateway_target" "test" {
  name               = %[1]q
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id

  target_configuration {
    mcp {
      mcp_server {
        endpoint = "https://docs.mcp.cloudflare.com/mcp"
      }
    }
  }

  metadata_configuration {
    allowed_request_headers  = ["x-correlation-id", "x-tenant-id", "x-request-id"]
    allowed_response_headers = ["x-rate-limit-remaining", "x-request-id"]
    allowed_query_parameters = ["version", "format"]
  }
}
`, rName))
}

func testAccGatewayTargetConfig_metadataConfigurationRemoved(rName string) string {
	return acctest.ConfigCompose(testAccGatewayTargetConfig_infra(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_gateway_target" "test" {
  name               = %[1]q
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id

  target_configuration {
    mcp {
      mcp_server {
        endpoint = "https://docs.mcp.cloudflare.com/mcp"
      }
    }
  }
}
`, rName))
}

func testAccGatewayTargetConfig_metadataConfigurationInvalidHeader(rName, headerName string) string {
	return acctest.ConfigCompose(testAccGatewayTargetConfig_infra(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_gateway_target" "test" {
  name               = %[1]q
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id

  target_configuration {
    mcp {
      mcp_server {
        endpoint = "https://docs.mcp.cloudflare.com/mcp"
      }
    }
  }

  metadata_configuration {
    allowed_request_headers = [%[2]q]
  }
}
`, rName, headerName))
}

func testAccGatewayTargetConfig_targetConfigurationAPIGateway(rName, toolOverrideSuffix string) string {
	return acctest.ConfigCompose(testAccGatewayTargetConfig_infra(rName), fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "test" {
  name = %[1]q
}

resource "aws_api_gateway_resource" "pets" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  parent_id   = aws_api_gateway_rest_api.test.root_resource_id
  path_part   = "pets"
}

resource "aws_api_gateway_method" "get_pets" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  resource_id   = aws_api_gateway_resource.pets.id
  http_method   = "GET"
  authorization = "NONE"
}

resource "aws_api_gateway_method" "post_pets" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  resource_id   = aws_api_gateway_resource.pets.id
  http_method   = "POST"
  authorization = "NONE"
}

resource "aws_api_gateway_method_response" "get_pets_200" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.pets.id
  http_method = aws_api_gateway_method.get_pets.http_method
  status_code = "200"
}

resource "aws_api_gateway_method_response" "post_pets_200" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.pets.id
  http_method = aws_api_gateway_method.post_pets.http_method
  status_code = "200"
}

resource "aws_api_gateway_integration" "get_pets" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.pets.id
  http_method = aws_api_gateway_method.get_pets.http_method
  type        = "MOCK"
}

resource "aws_api_gateway_integration" "post_pets" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  resource_id = aws_api_gateway_resource.pets.id
  http_method = aws_api_gateway_method.post_pets.http_method
  type        = "MOCK"
}

resource "aws_api_gateway_deployment" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id
  depends_on = [
    aws_api_gateway_integration.get_pets,
    aws_api_gateway_integration.post_pets,
  ]
}

resource "aws_api_gateway_stage" "test" {
  stage_name    = "prod"
  rest_api_id   = aws_api_gateway_rest_api.test.id
  deployment_id = aws_api_gateway_deployment.test.id
}

resource "aws_bedrockagentcore_gateway_target" "test" {
  name               = %[1]q
  gateway_identifier = aws_bedrockagentcore_gateway.test.gateway_id

  target_configuration {
    mcp {
      api_gateway {
        rest_api_id = aws_api_gateway_rest_api.test.id
        stage       = aws_api_gateway_stage.test.stage_name

        api_gateway_tool_configuration {
          tool_filter {
            filter_path = "/pets"
            methods     = ["GET", "POST"]
          }

          tool_override {
            name        = "ListPets%[2]s"
            path        = "/pets"
            method      = "GET"
            description = "Retrieves all available pets%[2]s"
          }

          tool_override {
            name        = "RegisterPets%[2]s"
            path        = "/pets"
            method      = "POST"
            description = "Register pets%[2]s"
          }
        }
      }
    }
  }
}
`, rName, toolOverrideSuffix))
}
