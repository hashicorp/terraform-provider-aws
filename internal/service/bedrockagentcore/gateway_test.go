// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfbedrockagentcore "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagentcore"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgentCoreGateway_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var gateway bedrockagentcorecontrol.GetGatewayOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_gateway.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckGateways(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, t, resourceName, &gateway),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("gateway_arn"), tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`gateway/.+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("gateway_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("gateway_url"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("workload_identity_details"), knownvalue.ListSizeExact(1)),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "gateway_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "gateway_id",
			},
		},
	})
}

func TestAccBedrockAgentCoreGateway_xrayDelivery(t *testing.T) {
	ctx := acctest.Context(t)
	var gateway bedrockagentcorecontrol.GetGatewayOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_gateway.test"
	deliveryResourceName := "aws_cloudwatch_log_delivery.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckGateways(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_xrayDelivery(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, t, resourceName, &gateway),
					resource.TestCheckResourceAttrSet(deliveryResourceName, names.AttrID),
					resource.TestCheckResourceAttr(deliveryResourceName, "delivery_source_name", rName+"-source"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
						plancheck.ExpectResourceAction(deliveryResourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreGateway_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var gateway bedrockagentcorecontrol.GetGatewayOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_gateway.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckGateways(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, t, resourceName, &gateway),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfbedrockagentcore.ResourceGateway, resourceName),
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

func TestAccBedrockAgentCoreGateway_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var gateway bedrockagentcorecontrol.GetGatewayOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_gateway.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckGateways(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(ctx, t, resourceName, &gateway),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "gateway_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "gateway_id",
			},
			{
				Config: testAccGatewayConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(ctx, t, resourceName, &gateway),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
			{
				Config: testAccGatewayConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGatewayExists(ctx, t, resourceName, &gateway),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreGateway_interceptorConfigurations(t *testing.T) {
	ctx := acctest.Context(t)
	var gateway bedrockagentcorecontrol.GetGatewayOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_gateway.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckGateways(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_interceptorConfigurations(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, t, resourceName, &gateway),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("interceptor_configuration"), knownvalue.ListSizeExact(1)),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "gateway_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "gateway_id",
			},
		},
	})
}

func TestAccBedrockAgentCoreGateway_description(t *testing.T) {
	ctx := acctest.Context(t)
	var gateway bedrockagentcorecontrol.GetGatewayOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_gateway.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckGateways(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_description(rName, "Initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, t, resourceName, &gateway),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("Initial description")),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "gateway_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "gateway_id",
			},
			{
				Config: testAccGatewayConfig_description(rName, "Updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, t, resourceName, &gateway),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("Updated description")),
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreGateway_IAMAuthorizer(t *testing.T) {
	ctx := acctest.Context(t)
	var gateway bedrockagentcorecontrol.GetGatewayOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_gateway.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckGateways(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_IAMAuthorizer(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, t, resourceName, &gateway),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("authorizer_type"), knownvalue.StringExact(string(awstypes.AuthorizerTypeAwsIam))),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "gateway_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "gateway_id",
			},
			{
				Config: testAccGatewayConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, t, resourceName, &gateway),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("authorizer_type"), knownvalue.StringExact(string(awstypes.AuthorizerTypeCustomJwt))),
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreGateway_kmsKey(t *testing.T) {
	acctest.Skip(t, "KMS key returns HTTP 500")
	ctx := acctest.Context(t)
	var gateway bedrockagentcorecontrol.GetGatewayOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_gateway.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckGateways(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_kmsKey(rName, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, t, resourceName, &gateway),
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
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "gateway_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "gateway_id",
			},
			{
				Config: testAccGatewayConfig_kmsKey(rName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, t, resourceName, &gateway),
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

func TestAccBedrockAgentCoreGateway_protocolConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var gateway bedrockagentcorecontrol.GetGatewayOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_bedrockagentcore_gateway.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckGateways(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGatewayDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGatewayConfig_protocolConfiguration(rName, "First set of instructions"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, t, resourceName, &gateway),
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
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "gateway_id"),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "gateway_id",
			},
			{
				Config: testAccGatewayConfig_protocolConfiguration(rName, "Second set of instructions"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGatewayExists(ctx, t, resourceName, &gateway),
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

func testAccCheckGatewayDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagentcore_gateway" {
				continue
			}

			_, err := tfbedrockagentcore.FindGatewayByID(ctx, conn, rs.Primary.Attributes["gateway_id"])
			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Bedrock Agent Core Gateway %s still exists", rs.Primary.Attributes["gateway_id"])
		}

		return nil
	}
}

func testAccCheckGatewayExists(ctx context.Context, t *testing.T, n string, v *bedrockagentcorecontrol.GetGatewayOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		resp, err := tfbedrockagentcore.FindGatewayByID(ctx, conn, rs.Primary.Attributes["gateway_id"])
		if err != nil {
			return err
		}

		*v = *resp

		return nil
	}
}

func testAccPreCheckGateways(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

	var input bedrockagentcorecontrol.ListGatewaysInput

	_, err := conn.ListGateways(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccGatewayConfig_iamRole(rName string) string {
	return fmt.Sprintf(`
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
`, rName)
}

func testAccGatewayConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccGatewayConfig_iamRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_gateway" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  authorizer_type = "CUSTOM_JWT"
  authorizer_configuration {
    custom_jwt_authorizer {
      discovery_url    = "https://accounts.google.com/.well-known/openid-configuration"
      allowed_audience = ["test1", "test2"]
    }
  }

  protocol_type = "MCP"
}
`, rName))
}

func testAccGatewayConfig_xrayDelivery(rName string) string {
	return acctest.ConfigCompose(testAccGatewayConfig_iamRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_gateway" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  authorizer_type = "CUSTOM_JWT"
  authorizer_configuration {
    custom_jwt_authorizer {
      discovery_url    = "https://accounts.google.com/.well-known/openid-configuration"
      allowed_audience = ["test1", "test2"]
    }
  }

  protocol_type = "MCP"
}

resource "aws_cloudwatch_log_delivery_source" "test" {
  name         = "%[1]s-source"
  log_type     = "TRACES"
  resource_arn = aws_bedrockagentcore_gateway.test.gateway_arn
}

resource "aws_cloudwatch_log_delivery_destination" "test" {
  name                      = "%[1]s-destination"
  delivery_destination_type = "XRAY"
}

resource "aws_cloudwatch_log_delivery" "test" {
  delivery_source_name     = aws_cloudwatch_log_delivery_source.test.name
  delivery_destination_arn = aws_cloudwatch_log_delivery_destination.test.arn
}
`, rName))
}

func testAccGatewayConfig_protocolConfiguration(rName, instructions string) string {
	return acctest.ConfigCompose(testAccGatewayConfig_iamRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_gateway" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  authorizer_type = "CUSTOM_JWT"
  authorizer_configuration {
    custom_jwt_authorizer {
      discovery_url    = "https://accounts.google.com/.well-known/openid-configuration"
      allowed_audience = ["test1", "test2"]
    }
  }

  protocol_configuration {
    mcp {
      instructions       = %[2]q
      search_type        = "SEMANTIC"
      supported_versions = ["2025-03-26"]
    }
  }

  protocol_type = "MCP"
}
`, rName, instructions))
}

func testAccGatewayConfig_description(rName, description string) string {
	return acctest.ConfigCompose(testAccGatewayConfig_iamRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_gateway" "test" {
  name        = %[1]q
  description = %[2]q
  role_arn    = aws_iam_role.test.arn

  authorizer_type = "CUSTOM_JWT"
  authorizer_configuration {
    custom_jwt_authorizer {
      discovery_url    = "https://accounts.google.com/.well-known/openid-configuration"
      allowed_audience = ["test1", "test2"]
    }
  }

  protocol_type = "MCP"
}
`, rName, description))
}

func testAccGatewayConfig_kmsKey(rName string, idx int) string {
	return acctest.ConfigCompose(testAccGatewayConfig_iamRole(rName), fmt.Sprintf(`
resource "aws_kms_key" "test" {
  count = 2

  description             = "Test key for %[1]s ${count.index}"
  deletion_window_in_days = 7
}

resource "aws_bedrockagentcore_gateway" "test" {
  name            = %[1]q
  role_arn        = aws_iam_role.test.arn
  exception_level = "DEBUG"
  kms_key_arn     = aws_kms_key.test[%[2]d].arn

  authorizer_type = "CUSTOM_JWT"
  authorizer_configuration {
    custom_jwt_authorizer {
      discovery_url    = "https://accounts.google.com/.well-known/openid-configuration"
      allowed_audience = ["test1", "test2"]
    }
  }

  protocol_type = "MCP"
}
`, rName, idx))
}

func testAccGatewayConfig_IAMAuthorizer(rName string) string {
	return acctest.ConfigCompose(testAccGatewayConfig_iamRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_gateway" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  authorizer_type = "AWS_IAM"

  protocol_type = "MCP"
}
`, rName))
}

func testAccGatewayConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccGatewayConfig_iamRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_gateway" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  authorizer_type = "CUSTOM_JWT"
  authorizer_configuration {
    custom_jwt_authorizer {
      discovery_url    = "https://accounts.google.com/.well-known/openid-configuration"
      allowed_audience = ["test1", "test2"]
    }
  }

  protocol_type = "MCP"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccGatewayConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccGatewayConfig_iamRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_gateway" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  authorizer_type = "CUSTOM_JWT"
  authorizer_configuration {
    custom_jwt_authorizer {
      discovery_url    = "https://accounts.google.com/.well-known/openid-configuration"
      allowed_audience = ["test1", "test2"]
    }
  }

  protocol_type = "MCP"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccGatewayConfig_interceptorConfigurations(rName string) string {
	return acctest.ConfigCompose(testAccGatewayConfig_iamRole(rName), fmt.Sprintf(`
resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.lambda.arn
  handler       = "index.handler"
  runtime       = "python3.12"
}

resource "aws_iam_role" "lambda" {
  name = "%[1]s-lambda"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "lambda.amazonaws.com"
      }
    }]
  })
}

resource "aws_bedrockagentcore_gateway" "test" {
  name     = %[1]q
  role_arn = aws_iam_role.test.arn

  authorizer_type = "AWS_IAM"
  protocol_type   = "MCP"

  interceptor_configuration {
    interception_points = ["REQUEST", "RESPONSE"]

    interceptor {
      lambda {
        arn = aws_lambda_function.test.arn
      }
    }

    input_configuration {
      pass_request_headers = true
    }
  }
}
`, rName))
}
