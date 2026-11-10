// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
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

func TestAccBedrockAgentCoreAgentRuntimeEndpoint_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var agentruntimeendpoint bedrockagentcorecontrol.GetAgentRuntimeEndpointOutput
	rName := strings.ReplaceAll(acctest.RandomWithPrefix(t, acctest.ResourcePrefix), "-", "_")
	rBucketName := strings.ReplaceAll(rName, "_", "-")
	resourceName := "aws_bedrockagentcore_agent_runtime_endpoint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccAgentRuntimeEndpointPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentRuntimeEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentRuntimeEndpointConfig_basic(rName, rBucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentRuntimeEndpointExists(ctx, t, resourceName, &agentruntimeendpoint),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("agent_runtime_arn"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("agent_runtime_endpoint_arn"), tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`runtime/.+/runtime-endpoint/.+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("agent_runtime_version"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, ",", "agent_runtime_id", names.AttrName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
		},
	})
}

func TestAccBedrockAgentCoreAgentRuntimeEndpoint_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var agentruntimeendpoint bedrockagentcorecontrol.GetAgentRuntimeEndpointOutput
	rName := strings.ReplaceAll(acctest.RandomWithPrefix(t, acctest.ResourcePrefix), "-", "_")
	rBucketName := strings.ReplaceAll(rName, "_", "-")
	resourceName := "aws_bedrockagentcore_agent_runtime_endpoint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccAgentRuntimeEndpointPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentRuntimeEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentRuntimeEndpointConfig_basic(rName, rBucketName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentRuntimeEndpointExists(ctx, t, resourceName, &agentruntimeendpoint),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfbedrockagentcore.ResourceAgentRuntimeEndpoint, resourceName),
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

func TestAccBedrockAgentCoreAgentRuntimeEndpoint_update(t *testing.T) {
	ctx := acctest.Context(t)
	var agentruntimeendpoint bedrockagentcorecontrol.GetAgentRuntimeEndpointOutput
	rName := strings.ReplaceAll(acctest.RandomWithPrefix(t, acctest.ResourcePrefix), "-", "_")
	rBucketName := strings.ReplaceAll(rName, "_", "-")
	resourceName := "aws_bedrockagentcore_agent_runtime_endpoint.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccAgentRuntimeEndpointPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAgentRuntimeEndpointDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccAgentRuntimeEndpointConfig_description(rName, rBucketName, "test endpoint"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentRuntimeEndpointExists(ctx, t, resourceName, &agentruntimeendpoint),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("agent_runtime_version"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("test endpoint")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, ",", "agent_runtime_id", names.AttrName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
			{
				Config: testAccAgentRuntimeEndpointConfig_description(rName, rBucketName, "updated endpoint"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAgentRuntimeEndpointExists(ctx, t, resourceName, &agentruntimeendpoint),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("agent_runtime_version"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("updated endpoint")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
		},
	})
}

func testAccCheckAgentRuntimeEndpointDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagentcore_agent_runtime_endpoint" {
				continue
			}

			_, err := tfbedrockagentcore.FindAgentRuntimeEndpointByTwoPartKey(ctx, conn, rs.Primary.Attributes["agent_runtime_id"], rs.Primary.Attributes[names.AttrName])
			if retry.NotFound(err) {
				return nil
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Bedrock Agent Core Agent Runtime Endpoint %s still exists", rs.Primary.Attributes[names.AttrName])
		}

		return nil
	}
}

func testAccCheckAgentRuntimeEndpointExists(ctx context.Context, t *testing.T, n string, v *bedrockagentcorecontrol.GetAgentRuntimeEndpointOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		resp, err := tfbedrockagentcore.FindAgentRuntimeEndpointByTwoPartKey(ctx, conn, rs.Primary.Attributes["agent_runtime_id"], rs.Primary.Attributes[names.AttrName])
		if err != nil {
			return err
		}

		*v = *resp

		return nil
	}
}

func testAccAgentRuntimeEndpointPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

	input := bedrockagentcorecontrol.ListAgentRuntimeEndpointsInput{
		AgentRuntimeId: aws.String("non_existent_agent-abcde12345"),
	}

	_, err := conn.ListAgentRuntimeEndpoints(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccAgentRuntimeEndpointConfig_basic(rName, rBucketName string) string {
	return acctest.ConfigCompose(testAccAgentRuntimeConfig_codeConfiguration(rName, rBucketName), fmt.Sprintf(`
resource "aws_bedrockagentcore_agent_runtime_endpoint" "test" {
  name                  = %[1]q
  agent_runtime_id      = aws_bedrockagentcore_agent_runtime.test.agent_runtime_id
  agent_runtime_version = aws_bedrockagentcore_agent_runtime.test.agent_runtime_version
}
`, rName))
}

func testAccAgentRuntimeEndpointConfig_description(rName, rBucketName, description string) string {
	return acctest.ConfigCompose(testAccAgentRuntimeConfig_codeConfiguration(rName, rBucketName), fmt.Sprintf(`
resource "aws_bedrockagentcore_agent_runtime_endpoint" "test" {
  name                  = %[1]q
  description           = %[2]q
  agent_runtime_id      = aws_bedrockagentcore_agent_runtime.test.agent_runtime_id
  agent_runtime_version = aws_bedrockagentcore_agent_runtime.test.agent_runtime_version
}
`, rName, description))
}
