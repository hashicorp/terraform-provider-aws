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

func TestAccBedrockAgentCorePolicyEngine_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := randomWithPrefixAndUnderscore(t)
	resourceName := "aws_bedrockagentcore_policy_engine.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckPolicyEngines(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyEngineDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyEngineConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyEngineExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_engine_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_engine_arn"), tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`policy-engine/.+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "policy_engine_id"),
				ImportStateVerifyIdentifierAttribute: "policy_engine_id",
			},
		},
	})
}

func TestAccBedrockAgentCorePolicyEngine_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := randomWithPrefixAndUnderscore(t)
	resourceName := "aws_bedrockagentcore_policy_engine.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckPolicyEngines(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyEngineDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyEngineConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyEngineExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfbedrockagentcore.ResourcePolicyEngine, resourceName),
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

func TestAccBedrockAgentCorePolicyEngine_description(t *testing.T) {
	ctx := acctest.Context(t)
	rName := randomWithPrefixAndUnderscore(t)
	resourceName := "aws_bedrockagentcore_policy_engine.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckPolicyEngines(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyEngineDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyEngineConfig_description(rName, "initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyEngineExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("initial description")),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "policy_engine_id"),
				ImportStateVerifyIdentifierAttribute: "policy_engine_id",
			},
			{
				Config: testAccPolicyEngineConfig_description(rName, "updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyEngineExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("updated description")),
				},
			},
		},
	})
}

func testAccCheckPolicyEngineDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagentcore_policy_engine" {
				continue
			}

			_, err := tfbedrockagentcore.FindPolicyEngineByID(ctx, conn, rs.Primary.Attributes["policy_engine_id"])
			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Bedrock Agent Core Policy Engine %s still exists", rs.Primary.Attributes["policy_engine_id"])
		}

		return nil
	}
}

func testAccCheckPolicyEngineExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		_, err := tfbedrockagentcore.FindPolicyEngineByID(ctx, conn, rs.Primary.Attributes["policy_engine_id"])
		return err
	}
}

func testAccPreCheckPolicyEngines(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)
	input := bedrockagentcorecontrol.ListPolicyEnginesInput{}

	_, err := conn.ListPolicyEngines(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccPolicyEngineConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_policy_engine" "test" {
  name = %[1]q
}
`, rName)
}

func testAccPolicyEngineConfig_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_policy_engine" "test" {
  name        = %[1]q
  description = %[2]q
}
`, rName, description)
}
