// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"context"
	"fmt"
	"os"
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

const (
	policyEngineIDEnvVar = "BEDROCK_AGENTCORE_POLICY_ENGINE_ID"
)

func TestAccBedrockAgentCorePolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v bedrockagentcorecontrol.GetPolicyOutput
	rName := strings.ReplaceAll(acctest.RandomWithPrefix(t, "tfacc"), "-", "_")
	resourceName := "aws_bedrockagentcore_policy.test"
	policyEngineID := os.Getenv(policyEngineIDEnvVar)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckPolicy(ctx, t, policyEngineID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_basic(rName, policyEngineID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_arn"), tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`policy-engine/.+/policy/.+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccPolicyImportStateIDFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: "policy_id",
				ImportStateVerifyIgnore:              []string{"validation_mode"},
			},
		},
	})
}

func TestAccBedrockAgentCorePolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v bedrockagentcorecontrol.GetPolicyOutput
	rName := strings.ReplaceAll(acctest.RandomWithPrefix(t, "tfacc"), "-", "_")
	resourceName := "aws_bedrockagentcore_policy.test"
	policyEngineID := os.Getenv(policyEngineIDEnvVar)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckPolicy(ctx, t, policyEngineID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_basic(rName, policyEngineID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfbedrockagentcore.ResourcePolicy, resourceName),
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

func TestAccBedrockAgentCorePolicy_description(t *testing.T) {
	ctx := acctest.Context(t)
	var v bedrockagentcorecontrol.GetPolicyOutput
	rName := strings.ReplaceAll(acctest.RandomWithPrefix(t, "tfacc"), "-", "_")
	resourceName := "aws_bedrockagentcore_policy.test"
	policyEngineID := os.Getenv(policyEngineIDEnvVar)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckPolicy(ctx, t, policyEngineID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_description(rName, policyEngineID, "initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("initial description")),
				},
			},
			{
				Config: testAccPolicyConfig_description(rName, policyEngineID, "updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &v),
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

func TestAccBedrockAgentCorePolicy_cedarStatementUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var v bedrockagentcorecontrol.GetPolicyOutput
	rName := strings.ReplaceAll(acctest.RandomWithPrefix(t, "tfacc"), "-", "_")
	resourceName := "aws_bedrockagentcore_policy.test"
	policyEngineID := os.Getenv(policyEngineIDEnvVar)

	stmt1 := "permit(principal, action, resource is AgentCore::Gateway);"
	stmt2 := "forbid(principal, action, resource is AgentCore::Gateway);"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckPolicy(ctx, t, policyEngineID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_cedarStatement(rName, policyEngineID, stmt1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("definition").AtSliceIndex(0).AtMapKey("cedar").AtSliceIndex(0).AtMapKey("statement"), knownvalue.StringExact(stmt1)),
				},
			},
			{
				Config: testAccPolicyConfig_cedarStatement(rName, policyEngineID, stmt2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("definition").AtSliceIndex(0).AtMapKey("cedar").AtSliceIndex(0).AtMapKey("statement"), knownvalue.StringExact(stmt2)),
				},
			},
		},
	})
}

func testAccPreCheckPolicy(ctx context.Context, t *testing.T, policyEngineID string) {
	if policyEngineID == "" {
		t.Skipf("skipping acceptance testing: %s must be set", policyEngineIDEnvVar)
	}

	conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

	input := bedrockagentcorecontrol.ListPoliciesInput{
		PolicyEngineId: aws.String(policyEngineID),
	}

	_, err := conn.ListPolicies(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckPolicyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagentcore_policy" {
				continue
			}

			_, err := tfbedrockagentcore.FindPolicyByTwoPartKey(ctx, conn, rs.Primary.Attributes["policy_engine_id"], rs.Primary.Attributes["policy_id"])
			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Bedrock Agent Core Policy %s still exists", rs.Primary.Attributes["policy_id"])
		}

		return nil
	}
}

func testAccCheckPolicyExists(ctx context.Context, t *testing.T, n string, v *bedrockagentcorecontrol.GetPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		resp, err := tfbedrockagentcore.FindPolicyByTwoPartKey(ctx, conn, rs.Primary.Attributes["policy_engine_id"], rs.Primary.Attributes["policy_id"])
		if err != nil {
			return err
		}

		*v = *resp

		return nil
	}
}

func testAccPolicyImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}
		return fmt.Sprintf("%s,%s", rs.Primary.Attributes["policy_engine_id"], rs.Primary.Attributes["policy_id"]), nil
	}
}

func testAccPolicyConfig_basic(rName, policyEngineID string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_policy" "test" {
  name             = %[1]q
  policy_engine_id = %[2]q
  validation_mode  = "IGNORE_ALL_FINDINGS"

  definition {
    cedar {
      statement = "permit(principal, action, resource is AgentCore::Gateway);"
    }
  }
}
`, rName, policyEngineID)
}

func testAccPolicyConfig_cedarStatement(rName, policyEngineID, statement string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_policy" "test" {
  name             = %[1]q
  policy_engine_id = %[2]q
  validation_mode  = "IGNORE_ALL_FINDINGS"

  definition {
    cedar {
      statement = %[3]q
    }
  }
}
`, rName, policyEngineID, statement)
}

func testAccPolicyConfig_description(rName, policyEngineID, description string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_policy" "test" {
  name             = %[1]q
  policy_engine_id = %[2]q
  description      = %[3]q
  validation_mode  = "IGNORE_ALL_FINDINGS"

  definition {
    cedar {
      statement = "permit(principal, action, resource is AgentCore::Gateway);"
    }
  }
}
`, rName, policyEngineID, description)
}
