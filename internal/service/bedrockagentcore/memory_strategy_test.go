// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfbedrockagentcore "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagentcore"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgentCoreMemoryStrategy_standard(t *testing.T) {
	ctx := acctest.Context(t)
	var m awstypes.MemoryStrategy
	rName := randomMemoryName(t)
	resourceName := "aws_bedrockagentcore_memory_strategy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckMemories(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMemoryStrategyDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Setup: Create memory with execution role (needed for EPISODIC steps)
			{
				Config: testAccMemoryConfig_memoryExecutionRole(rName),
			},
			// Step 1: Create episodic strategy
			{
				Config: testAccMemoryStrategyConfig_withExecutionRole(rName, "EPISODIC", "Episodic strategy", "/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "EPISODIC"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Episodic strategy"),
					resource.TestCheckResourceAttr(resourceName, "namespaces.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "memory_strategy_id"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			// Step 2: Update episodic description (in-place)
			{
				Config: testAccMemoryStrategyConfig_withExecutionRole(rName, "EPISODIC", "Updated episodic strategy", "/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "EPISODIC"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Updated episodic strategy"),
					resource.TestCheckResourceAttr(resourceName, "namespaces.#", "1"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			// Step 3: Change type episodic→semantic (replacement)
			{
				Config: testAccMemoryStrategyConfig_withExecutionRole(rName, "SEMANTIC", names.AttrDescription, "default"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "SEMANTIC"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, "namespaces.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "namespaces.*", "default"),
					resource.TestCheckResourceAttrSet(resourceName, "memory_strategy_id"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
			// Step 4: Update description + namespace (in-place)
			{
				Config: testAccMemoryStrategyConfig_withExecutionRole(rName, "SEMANTIC", "Updated description", "custom"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Updated description"),
					resource.TestCheckResourceAttr(resourceName, "namespaces.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "namespaces.*", "custom"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			// Step 5: Change type semantic→user_preference (replacement)
			{
				Config: testAccMemoryStrategyConfig_withExecutionRole(rName, "USER_PREFERENCE", "User preference strategy", "preferences"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "USER_PREFERENCE"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "User preference strategy"),
					resource.TestCheckResourceAttr(resourceName, "namespaces.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "namespaces.*", "preferences"),
					resource.TestCheckResourceAttrSet(resourceName, "memory_strategy_id"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
			// Step 6: Try to create ANOTHER user_preference strategy → should ERROR
			{
				Config:      testAccMemoryStrategyConfig_duplicateType(rName, "USER_PREFERENCE"),
				ExpectError: regexache.MustCompile("Found multiple strategies of type"),
			},
			// Step 7: Import test - verify composite ID import works
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccMemoryStrategyImportStateIDFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "memory_strategy_id",
				ImportStateVerifyIgnore:              []string{"memory_execution_role_arn"},
			},
		},
	})
}

func TestAccBedrockAgentCoreMemoryStrategy_custom(t *testing.T) {
	ctx := acctest.Context(t)
	var m awstypes.MemoryStrategy
	rName := randomMemoryName(t)
	resourceName := "aws_bedrockagentcore_memory_strategy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckMemories(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMemoryStrategyDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Setup: Create memory
			{
				Config: testAccMemoryConfig_memoryExecutionRole(rName),
			},
			// Step 1: CUSTOM type with missing configuration block → ValidateConfig error
			{
				Config:      testAccMemoryStrategyConfig_customInvalid(rName),
				ExpectError: regexache.MustCompile(`Attribute "configuration" must be configured`),
			},
			// Step 2: Create CUSTOM strategy with consolidation block
			{
				Config: testAccMemoryStrategyConfig_customConsolidationOnly(rName, "SEMANTIC_OVERRIDE", "Focus on semantic relationships", "us.amazon.nova-2-lite-v1:0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.type", "SEMANTIC_OVERRIDE"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.consolidation.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.consolidation.0.append_to_prompt", "Focus on semantic relationships"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.consolidation.0.model_id", "us.amazon.nova-2-lite-v1:0"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.extraction.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "memory_strategy_id"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			// Step 3: Add extraction block and update consolidation properties (same override type)
			{
				Config: testAccMemoryStrategyConfig_custom(rName, "SEMANTIC_OVERRIDE", "Updated semantic consolidation", "amazon.nova-lite-v1:0", "Extract semantic meaning", "us.amazon.nova-2-lite-v1:0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.type", "SEMANTIC_OVERRIDE"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.consolidation.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.consolidation.0.append_to_prompt", "Updated semantic consolidation"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.consolidation.0.model_id", "amazon.nova-lite-v1:0"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.extraction.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.extraction.0.append_to_prompt", "Extract semantic meaning"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.extraction.0.model_id", "us.amazon.nova-2-lite-v1:0"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			// Step 4: Try to remove consolidation block → should ERROR
			{
				Config:      testAccMemoryStrategyConfig_customExtractionOnly(rName, "SEMANTIC_OVERRIDE", "Extract semantic meaning", "us.amazon.nova-2-lite-v1:0"),
				ExpectError: regexache.MustCompile("Removing the previously configured \"consolidation\" block"),
			},
			//// Step 5: Change override type → should replace resource
			{
				Config: testAccMemoryStrategyConfig_custom(rName, "USER_PREFERENCE_OVERRIDE", "Store user preferences", "amazon.nova-lite-v1:0", "Extract user preferences", "us.amazon.nova-2-lite-v1:0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.type", "USER_PREFERENCE_OVERRIDE"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.consolidation.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.consolidation.0.append_to_prompt", "Store user preferences"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.consolidation.0.model_id", "amazon.nova-lite-v1:0"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.extraction.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.extraction.0.append_to_prompt", "Extract user preferences"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.extraction.0.model_id", "us.amazon.nova-2-lite-v1:0"),
					resource.TestCheckResourceAttrSet(resourceName, "memory_strategy_id"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
			//// Step 6: SUMMARY_OVERRIDE with extraction block → ValidateConfig error
			{
				Config:      testAccMemoryStrategyConfig_custom(rName, "SUMMARY_OVERRIDE", "Summary consolidation", "amazon.nova-lite-v1:0", "Summary extraction", "us.amazon.nova-2-lite-v1:0"),
				ExpectError: regexache.MustCompile(`Attribute "configuration\[0\].extraction" must not be configured`),
			},
			//// Step 7: SUMMARY_OVERRIDE with no extraction block → should succeed
			{
				Config: testAccMemoryStrategyConfig_customConsolidationOnly(rName, "SUMMARY_OVERRIDE", "Summary consolidation only", "amazon.nova-lite-v1:0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.type", "SUMMARY_OVERRIDE"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.consolidation.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.consolidation.0.append_to_prompt", "Summary consolidation only"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.consolidation.0.model_id", "amazon.nova-lite-v1:0"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.extraction.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "memory_strategy_id"),
				),
			},
		},
	})
}

func TestAccBedrockAgentCoreMemoryStrategy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var m awstypes.MemoryStrategy
	rName := randomMemoryName(t)
	resourceName := "aws_bedrockagentcore_memory_strategy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckMemories(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMemoryStrategyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccMemoryStrategyConfig_basic(rName, "SEMANTIC", "Example Description", "default"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfbedrockagentcore.ResourceMemoryStrategy, resourceName),
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

func testAccCheckMemoryStrategyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagentcore_memory_strategy" {
				continue
			}

			_, err := tfbedrockagentcore.FindMemoryStrategyByTwoPartKey(ctx, conn, rs.Primary.Attributes["memory_id"], rs.Primary.Attributes["memory_strategy_id"])
			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Bedrock Agent Core Memory Strategy %s,%s still exists", rs.Primary.Attributes["memory_id"], rs.Primary.Attributes["memory_strategy_id"])
		}

		return nil
	}
}

func testAccCheckMemoryStrategyExists(ctx context.Context, t *testing.T, n string, v *awstypes.MemoryStrategy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		resp, err := tfbedrockagentcore.FindMemoryStrategyByTwoPartKey(ctx, conn, rs.Primary.Attributes["memory_id"], rs.Primary.Attributes["memory_strategy_id"])
		if err != nil {
			return err
		}

		*v = *resp

		return nil
	}
}

func testAccMemoryStrategyImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return acctest.AttrsImportStateIdFunc(resourceName, ",", "memory_id", "memory_strategy_id")
}

func testAccMemoryStrategyConfig_basic(rName, strategyType, description, namespace string) string {
	return acctest.ConfigCompose(testAccMemoryConfig_basic(rName), fmt.Sprintf(`	
resource "aws_bedrockagentcore_memory_strategy" "test" {
  name        = %[1]q
  memory_id   = aws_bedrockagentcore_memory.test.id
  type        = %[2]q
  description = %[3]q
  namespaces  = [%[4]q]
}
`, rName, strategyType, description, namespace))
}

func testAccMemoryStrategyConfig_withExecutionRole(rName, strategyType, description, namespace string) string {
	return acctest.ConfigCompose(testAccMemoryConfig_memoryExecutionRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_memory_strategy" "test" {
  name                      = %[1]q
  memory_id                 = aws_bedrockagentcore_memory.test.id
  memory_execution_role_arn = aws_bedrockagentcore_memory.test.memory_execution_role_arn
  type                      = %[2]q
  description               = %[3]q
  namespaces                = [%[4]q]
}
`, rName, strategyType, description, namespace))
}

func testAccMemoryStrategyConfig_duplicateType(rName string, strategyType string) string {
	namespace := "default"
	duplicateNamespace := "duplicate"
	if strategyType == "EPISODIC" {
		namespace = "/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}"
		duplicateNamespace = "/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}"
	}
	return acctest.ConfigCompose(testAccMemoryStrategyConfig_withExecutionRole(rName, strategyType, "Strategy for duplicate test", namespace), fmt.Sprintf(`	
resource "aws_bedrockagentcore_memory_strategy" "test2" {
  name                      = "%[1]s_duplicate"
  memory_id                 = aws_bedrockagentcore_memory.test.id
  memory_execution_role_arn = aws_bedrockagentcore_memory.test.memory_execution_role_arn
  type                      = %[2]q
  description               = "Duplicate strategy"
  namespaces                = [%[3]q]
}
`, rName, strategyType, duplicateNamespace))
}

func testAccMemoryStrategyConfig_custom(rName, overrideType, consolidationPrompt, consolidationModel, extractionPrompt, extractionModel string) string {
	return acctest.ConfigCompose(testAccMemoryConfig_memoryExecutionRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_memory_strategy" "test" {
  name                      = %[1]q
  memory_id                 = aws_bedrockagentcore_memory.test.id
  memory_execution_role_arn = aws_bedrockagentcore_memory.test.memory_execution_role_arn
  type                      = "CUSTOM"
  description               = "Test custom strategy"
  namespaces                = ["{sessionId}"]

  configuration {
    type = %[2]q
    consolidation {
      append_to_prompt = %[3]q
      model_id         = %[4]q
    }
    extraction {
      append_to_prompt = %[5]q
      model_id         = %[6]q
    }
  }
}
`, rName, overrideType, consolidationPrompt, consolidationModel, extractionPrompt, extractionModel))
}

func testAccMemoryStrategyConfig_customConsolidationOnly(rName, overrideType, consolidationPrompt, consolidationModel string) string {
	return acctest.ConfigCompose(testAccMemoryConfig_memoryExecutionRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_memory_strategy" "test" {
  name                      = %[1]q
  memory_id                 = aws_bedrockagentcore_memory.test.id
  memory_execution_role_arn = aws_bedrockagentcore_memory.test.memory_execution_role_arn
  type                      = "CUSTOM"
  description               = "Test custom strategy"
  namespaces                = ["{sessionId}"]

  configuration {
    type = %[2]q
    consolidation {
      append_to_prompt = %[3]q
      model_id         = %[4]q
    }
  }
}
`, rName, overrideType, consolidationPrompt, consolidationModel))
}

func testAccMemoryStrategyConfig_customExtractionOnly(rName, overrideType, extractionPrompt, extractionModel string) string {
	return acctest.ConfigCompose(testAccMemoryConfig_memoryExecutionRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_memory_strategy" "test" {
  name                      = %[1]q
  memory_id                 = aws_bedrockagentcore_memory.test.id
  memory_execution_role_arn = aws_bedrockagentcore_memory.test.memory_execution_role_arn
  type                      = "CUSTOM"
  description               = "Test custom strategy"
  namespaces                = ["{sessionId}"]

  configuration {
    type = %[2]q
    extraction {
      append_to_prompt = %[3]q
      model_id         = %[4]q
    }
  }
}
`, rName, overrideType, extractionPrompt, extractionModel))
}

func testAccMemoryStrategyConfig_customInvalid(rName string) string {
	return acctest.ConfigCompose(testAccMemoryConfig_memoryExecutionRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_memory_strategy" "test" {
  name                      = %[1]q
  memory_execution_role_arn = aws_bedrockagentcore_memory.test.memory_execution_role_arn
  memory_id                 = aws_bedrockagentcore_memory.test.id
  type                      = "CUSTOM"
  description               = "Test custom strategy"
  namespaces                = ["default"]
}
`, rName))
}

func testAccMemoryStrategyConfig_semanticWithReflectionConfiguration(rName string) string {
	return acctest.ConfigCompose(testAccMemoryConfig_memoryExecutionRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_memory_strategy" "test" {
  name                      = %[1]q
  memory_id                 = aws_bedrockagentcore_memory.test.id
  memory_execution_role_arn = aws_bedrockagentcore_memory.test.memory_execution_role_arn
  type                      = "SEMANTIC"
  description               = "Test semantic strategy"
  namespaces                = ["default"]

  reflection_configuration {
    namespace_templates = ["default"]
  }
}
`, rName))
}

func testAccMemoryStrategyConfig_customSemanticWithReflection(rName, consolidationPrompt, consolidationModel, reflectionPrompt, reflectionModel string) string {
	return acctest.ConfigCompose(testAccMemoryConfig_memoryExecutionRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_memory_strategy" "test" {
  name                      = %[1]q
  memory_id                 = aws_bedrockagentcore_memory.test.id
  memory_execution_role_arn = aws_bedrockagentcore_memory.test.memory_execution_role_arn
  type                      = "CUSTOM"
  description               = "Test custom strategy"
  namespaces                = ["default"]

  configuration {
    type = "SEMANTIC_OVERRIDE"
    consolidation {
      append_to_prompt = %[2]q
      model_id         = %[3]q
    }
    reflection {
      append_to_prompt = %[4]q
      model_id         = %[5]q
    }
  }
}
`, rName, consolidationPrompt, consolidationModel, reflectionPrompt, reflectionModel))
}

func testAccMemoryStrategyConfig_episodicReflection(rName, namespace, reflectionNamespace string) string {
	return acctest.ConfigCompose(testAccMemoryConfig_memoryExecutionRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_memory_strategy" "test" {
  name                      = %[1]q
  memory_id                 = aws_bedrockagentcore_memory.test.id
  memory_execution_role_arn = aws_bedrockagentcore_memory.test.memory_execution_role_arn
  type                      = "EPISODIC"
  description               = "Episodic strategy with reflection"
  namespaces                = [%[2]q]

  reflection_configuration {
    namespace_templates = [%[3]q]
  }
}
`, rName, namespace, reflectionNamespace))
}

func testAccMemoryStrategyConfig_customEpisodicReflection(rName, consolidationPrompt, consolidationModel, extractionPrompt, extractionModel, reflectionPrompt, reflectionModel string) string {
	return acctest.ConfigCompose(testAccMemoryConfig_memoryExecutionRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_memory_strategy" "test" {
  name                      = %[1]q
  memory_id                 = aws_bedrockagentcore_memory.test.id
  memory_execution_role_arn = aws_bedrockagentcore_memory.test.memory_execution_role_arn
  type                      = "CUSTOM"
  description               = "Custom episodic strategy with reflection"
  namespaces                = ["/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}"]

  configuration {
    type = "EPISODIC_OVERRIDE"
    consolidation {
      append_to_prompt = %[2]q
      model_id         = %[3]q
    }
    extraction {
      append_to_prompt = %[4]q
      model_id         = %[5]q
    }
    reflection {
      append_to_prompt = %[6]q
      model_id         = %[7]q
      namespace_templates = ["/strategies/{memoryStrategyId}/actors/{actorId}"]
    }
  }
}
`, rName, consolidationPrompt, consolidationModel, extractionPrompt, extractionModel, reflectionPrompt, reflectionModel))
}

func TestAccBedrockAgentCoreMemoryStrategy_episodicReflection(t *testing.T) {
	ctx := acctest.Context(t)
	var m awstypes.MemoryStrategy
	rName := randomMemoryName(t)
	resourceName := "aws_bedrockagentcore_memory_strategy.test"

	episodicNS := "/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}"
	// Reflection namespace must be the same as or a prefix of the episodic namespace.
	reflectionNS := "/strategies/{memoryStrategyId}/actors/{actorId}"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckMemories(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMemoryStrategyDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Setup: Create memory with execution role
			{
				Config: testAccMemoryConfig_memoryExecutionRole(rName),
			},
			// Step 1: Create EPISODIC strategy with reflection_configuration
			{
				Config: testAccMemoryStrategyConfig_episodicReflection(rName, episodicNS, reflectionNS),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "EPISODIC"),
					resource.TestCheckResourceAttr(resourceName, "reflection_configuration.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "reflection_configuration.0.namespace_templates.*", reflectionNS),
					resource.TestCheckResourceAttrSet(resourceName, "memory_strategy_id"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			// Step 2: Import verification
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccMemoryStrategyImportStateIDFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "memory_strategy_id",
				ImportStateVerifyIgnore:              []string{"memory_execution_role_arn"},
			},
			// Step 3: Remove reflection_configuration block
			{
				Config: testAccMemoryStrategyConfig_withExecutionRole(rName, "EPISODIC", "Episodic strategy with reflection", episodicNS),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
					resource.TestCheckResourceAttr(resourceName, "reflection_configuration.#", "0"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			// Step 4: Refresh should not re-surface the block
			{
				Config: testAccMemoryStrategyConfig_withExecutionRole(rName, "EPISODIC", "Episodic strategy with reflection", episodicNS),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
					resource.TestCheckResourceAttr(resourceName, "reflection_configuration.#", "0"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreMemoryStrategy_customEpisodicReflection(t *testing.T) {
	ctx := acctest.Context(t)
	var m awstypes.MemoryStrategy
	rName := randomMemoryName(t)
	resourceName := "aws_bedrockagentcore_memory_strategy.test"

	model := "anthropic.claude-3-haiku-20240307-v1:0"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckMemories(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMemoryStrategyDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Setup: Create memory with execution role
			{
				Config: testAccMemoryConfig_memoryExecutionRole(rName),
			},
			// Step 1: CUSTOM EPISODIC_OVERRIDE with reflection block
			{
				Config: testAccMemoryStrategyConfig_customEpisodicReflection(
					rName,
					"Consolidate episodic memories", model,
					"Extract key episodes", model,
					"Identify patterns across episodes", model,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.type", "EPISODIC_OVERRIDE"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.consolidation.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.extraction.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.reflection.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.reflection.0.append_to_prompt", "Identify patterns across episodes"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.reflection.0.model_id", model),
					resource.TestCheckResourceAttrSet(resourceName, "memory_strategy_id"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			// Step 2: Update reflection prompt (in-place)
			{
				Config: testAccMemoryStrategyConfig_customEpisodicReflection(
					rName,
					"Consolidate episodic memories", model,
					"Extract key episodes", model,
					"Updated: find failure modes and recovery patterns", model,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.reflection.0.append_to_prompt", "Updated: find failure modes and recovery patterns"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			// Step 3: Import verification
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccMemoryStrategyImportStateIDFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "memory_strategy_id",
				ImportStateVerifyIgnore:              []string{"memory_execution_role_arn"},
			},
		},
	})
}

func TestAccBedrockAgentCoreMemoryStrategy_reflectionValidateConfig(t *testing.T) {
	ctx := acctest.Context(t)
	rName := randomMemoryName(t)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckMemories(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckMemoryStrategyDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Setup: Create memory with execution role
			{
				Config: testAccMemoryConfig_memoryExecutionRole(rName),
			},
			// Step 1: reflection_configuration on a non-EPISODIC type → ValidateConfig error
			{
				Config:      testAccMemoryStrategyConfig_semanticWithReflectionConfiguration(rName),
				ExpectError: regexache.MustCompile("The reflection_configuration block is only valid when type is `EPISODIC`"),
			},
			// Step 2: reflection block on a non-EPISODIC_OVERRIDE configuration type → ValidateConfig error
			{
				Config:      testAccMemoryStrategyConfig_customSemanticWithReflection(rName, "Focus on semantic relationships", "us.amazon.nova-2-lite-v1:0", "Identify patterns", "us.amazon.nova-2-lite-v1:0"),
				ExpectError: regexache.MustCompile("The reflection block inside configuration is only valid when configuration type is `EPISODIC_OVERRIDE`"),
			},
		},
	})
}
