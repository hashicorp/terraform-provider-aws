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
				ExpectError: regexache.MustCompile("When type is `CUSTOM`, the configuration block is required"),
			},
			// Step 2: Create CUSTOM strategy with consolidation block
			{
				Config: testAccMemoryStrategyConfig_customConsolidationOnly(rName, "SEMANTIC_OVERRIDE", "Focus on semantic relationships", "anthropic.claude-3-haiku-20240307-v1:0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.type", "SEMANTIC_OVERRIDE"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.consolidation.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.consolidation.0.append_to_prompt", "Focus on semantic relationships"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.consolidation.0.model_id", "anthropic.claude-3-haiku-20240307-v1:0"),
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
				Config: testAccMemoryStrategyConfig_custom(rName, "SEMANTIC_OVERRIDE", "Updated semantic consolidation", "anthropic.claude-3-sonnet-20240229-v1:0", "Extract semantic meaning", "anthropic.claude-3-haiku-20240307-v1:0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.type", "SEMANTIC_OVERRIDE"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.consolidation.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.consolidation.0.append_to_prompt", "Updated semantic consolidation"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.consolidation.0.model_id", "anthropic.claude-3-sonnet-20240229-v1:0"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.extraction.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.extraction.0.append_to_prompt", "Extract semantic meaning"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.extraction.0.model_id", "anthropic.claude-3-haiku-20240307-v1:0"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			// Step 4: Try to remove consolidation block → should ERROR
			{
				Config:      testAccMemoryStrategyConfig_customExtractionOnly(rName, "SEMANTIC_OVERRIDE", "Extract semantic meaning", "anthropic.claude-3-haiku-20240307-v1:0"),
				ExpectError: regexache.MustCompile("Removing the previously configured \"consolidation\" block is not allowed"),
			},
			//// Step 5: Change override type → should replace resource
			{
				Config: testAccMemoryStrategyConfig_custom(rName, "USER_PREFERENCE_OVERRIDE", "Store user preferences", "anthropic.claude-3-sonnet-20240229-v1:0", "Extract user preferences", "anthropic.claude-3-haiku-20240307-v1:0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.type", "USER_PREFERENCE_OVERRIDE"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.consolidation.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.consolidation.0.append_to_prompt", "Store user preferences"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.consolidation.0.model_id", "anthropic.claude-3-sonnet-20240229-v1:0"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.extraction.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.extraction.0.append_to_prompt", "Extract user preferences"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.extraction.0.model_id", "anthropic.claude-3-haiku-20240307-v1:0"),
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
				Config:      testAccMemoryStrategyConfig_custom(rName, "SUMMARY_OVERRIDE", "Summary consolidation", "anthropic.claude-3-sonnet-20240229-v1:0", "Summary extraction", "anthropic.claude-3-haiku-20240307-v1:0"),
				ExpectError: regexache.MustCompile("(?s)When\\s+configuration\\s+type\\s+is\\s+`SUMMARY_OVERRIDE`,\\s+the\\s+extraction\\s+block\\s+cannot\\s+be\\s+defined"),
			},
			//// Step 7: SUMMARY_OVERRIDE with no extraction block → should succeed
			{
				Config: testAccMemoryStrategyConfig_customConsolidationOnly(rName, "SUMMARY_OVERRIDE", "Summary consolidation only", "anthropic.claude-3-sonnet-20240229-v1:0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.type", "SUMMARY_OVERRIDE"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.consolidation.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.consolidation.0.append_to_prompt", "Summary consolidation only"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.consolidation.0.model_id", "anthropic.claude-3-sonnet-20240229-v1:0"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.extraction.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "memory_strategy_id"),
				),
			},
			//Step 8: Add 6 more CUSTOM strategies → should ERROR on too many
			{
				Config:      testAccMemoryStrategyConfig_customTooMany(rName, "USER_PREFERENCE_OVERRIDE", "Store user preferences", "anthropic.claude-3-sonnet-20240229-v1:0", "Extract user preferences", "anthropic.claude-3-haiku-20240307-v1:0"),
				ExpectError: regexache.MustCompile("Resource limit exceeded for memory strategies for memory"),
			},
			// Step 9: Import test
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

func TestAccBedrockAgentCoreMemoryStrategy_memoryRecordSchema(t *testing.T) {
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
			// Setup: Create memory with execution role
			{
				Config: testAccMemoryConfig_memoryExecutionRole(rName),
			},
			// Step 1: Create SEMANTIC strategy with a memory_record_schema entry
			{
				Config: testAccMemoryStrategyConfig_memoryRecordSchema(rName, "priority", "LLM_INFERRED", "STRING", "The priority of the record"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "SEMANTIC"),
					resource.TestCheckResourceAttr(resourceName, "memory_record_schema.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "memory_record_schema.0.metadata_schema.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "memory_record_schema.0.metadata_schema.0.key", "priority"),
					resource.TestCheckResourceAttr(resourceName, "memory_record_schema.0.metadata_schema.0.extraction_type", "LLM_INFERRED"),
					resource.TestCheckResourceAttr(resourceName, "memory_record_schema.0.metadata_schema.0.type", "STRING"),
					resource.TestCheckResourceAttr(resourceName, "memory_record_schema.0.metadata_schema.0.extraction_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "memory_record_schema.0.metadata_schema.0.extraction_config.0.llm_extraction_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "memory_record_schema.0.metadata_schema.0.extraction_config.0.llm_extraction_config.0.definition", "The priority of the record"),
					resource.TestCheckResourceAttr(resourceName, "memory_record_schema.0.metadata_schema.0.extraction_config.0.llm_extraction_config.0.validation.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "memory_record_schema.0.metadata_schema.0.extraction_config.0.llm_extraction_config.0.validation.0.string_validation.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "memory_record_schema.0.metadata_schema.0.extraction_config.0.llm_extraction_config.0.validation.0.string_validation.0.allowed_values.#", "3"),
					resource.TestCheckResourceAttrSet(resourceName, "memory_strategy_id"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			// Step 2: Update the metadata entry definition (in-place)
			{
				Config: testAccMemoryStrategyConfig_memoryRecordSchema(rName, "priority", "LLM_INFERRED", "STRING", "Updated priority definition"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
					resource.TestCheckResourceAttr(resourceName, "memory_record_schema.0.metadata_schema.0.extraction_config.0.llm_extraction_config.0.definition", "Updated priority definition"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			// Step 3: Import test
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

func testAccMemoryStrategyConfig_customDummy(rName string) string {
	return fmt.Sprintf(`
resource "aws_bedrockagentcore_memory_strategy" "%[1]s" {
  name                      = %[1]q
  memory_id                 = aws_bedrockagentcore_memory.test.id
  memory_execution_role_arn = aws_bedrockagentcore_memory.test.memory_execution_role_arn
  type                      = "CUSTOM"
  description               = "Test custom strategy"
  namespaces                = ["{sessionId}"]

  configuration {
    type = "SEMANTIC_OVERRIDE"
    extraction {
      append_to_prompt = "test"
      model_id         = "test_model"
    }
  }
  # Preventing the main strategy from being lost due to "too many resources" error
  depends_on = [aws_bedrockagentcore_memory_strategy.test]
}
`, rName)
}

func testAccMemoryStrategyConfig_customTooMany(rName, overrideType, consolidationPrompt, consolidationModel, extractionPrompt, extractionModel string) string {
	compose := acctest.ConfigCompose(testAccMemoryStrategyConfig_custom(rName, overrideType, consolidationPrompt, consolidationModel, extractionPrompt, extractionModel),
		testAccMemoryStrategyConfig_customDummy(rName+"_2"),
		testAccMemoryStrategyConfig_customDummy(rName+"_3"),
		testAccMemoryStrategyConfig_customDummy(rName+"_4"),
		testAccMemoryStrategyConfig_customDummy(rName+"_5"),
		testAccMemoryStrategyConfig_customDummy(rName+"_6"),
		testAccMemoryStrategyConfig_customDummy(rName+"_7"))
	return compose
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

func testAccMemoryStrategyConfig_memoryRecordSchema(rName, key, extractionType, valueType, definition string) string {
	return acctest.ConfigCompose(testAccMemoryConfig_memoryExecutionRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_memory_strategy" "test" {
  name                      = %[1]q
  memory_id                 = aws_bedrockagentcore_memory.test.id
  memory_execution_role_arn = aws_bedrockagentcore_memory.test.memory_execution_role_arn
  type                      = "SEMANTIC"
  description               = "Test memory record schema"
  namespaces                = ["default"]

  memory_record_schema {
    metadata_schema {
      key             = %[2]q
      extraction_type = %[3]q
      type            = %[4]q

      extraction_config {
        llm_extraction_config {
          definition = %[5]q

          validation {
            string_validation {
              allowed_values = ["high", "medium", "low"]
            }
          }
        }
      }
    }
  }
}
`, rName, key, extractionType, valueType, definition))
}

// TestAccBedrockAgentCoreMemoryStrategy_validationAtLeastOneOf confirms an empty
// number_validation {} / string_list_validation {} is rejected at plan time. Such an
// empty union member previously passed validation but was discarded by the API,
// producing "inconsistent result after apply: validation block count 1 -> 0".
func TestAccBedrockAgentCoreMemoryStrategy_validationAtLeastOneOf(t *testing.T) {
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
			{
				Config:      testAccMemoryStrategyConfig_emptyNumberValidation(rName),
				ExpectError: regexache.MustCompile("Invalid Attribute Combination"),
			},
		},
	})
}

// TestAccBedrockAgentCoreMemoryStrategy_nameForceNew confirms that changing name
// forces replacement. ModifyMemoryStrategyInput has no Name field, so an in-place
// rename was silently ignored by the API and produced "inconsistent result after
// apply: .name".
func TestAccBedrockAgentCoreMemoryStrategy_nameForceNew(t *testing.T) {
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
				Config: testAccMemoryStrategyConfig_semanticNamed(rName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
				),
			},
			{
				Config: testAccMemoryStrategyConfig_semanticNamed(rName, rName+"_renamed"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
		},
	})
}

// TestAccBedrockAgentCoreMemoryStrategy_recordSchemaRemovalForceNew confirms that
// removing memory_record_schema forces replacement. The API ignores a nil
// MemoryRecordSchema on update and retains the prior value, so an in-place removal
// produced "inconsistent result after apply: memory_record_schema block count 0 -> 1".
func TestAccBedrockAgentCoreMemoryStrategy_recordSchemaRemovalForceNew(t *testing.T) {
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
				Config: testAccMemoryStrategyConfig_recordSchemaToggle(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
					resource.TestCheckResourceAttr(resourceName, "memory_record_schema.#", "1"),
				),
			},
			{
				Config: testAccMemoryStrategyConfig_recordSchemaToggle(rName, false),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
		},
	})
}

// TestAccBedrockAgentCoreMemoryStrategy_typeSwitchNonCustomToCustom confirms that
// switching a strategy from a non-CUSTOM type to CUSTOM (which forces replacement)
// plans without crashing. The errorIfSingleBlockRemoved plan modifier previously
// read configuration[0].type into a non-nullable enum and panicked on the null a
// non-CUSTOM prior state produces, aborting plan generation.
func TestAccBedrockAgentCoreMemoryStrategy_typeSwitchNonCustomToCustom(t *testing.T) {
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
				Config: testAccMemoryStrategyConfig_withExecutionRole(rName, "SEMANTIC", "Semantic strategy", "default"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "SEMANTIC"),
				),
			},
			{
				// Plan-only: the bug was a crash during plan generation (the
				// errorIfSingleBlockRemoved modifier read a null configuration type from
				// the non-CUSTOM prior state). A successful non-empty plan proves the
				// modifier no longer panics; applying is unnecessary (and avoids relying
				// on a specific Bedrock model being enabled in the test region).
				Config:             testAccMemoryStrategyConfig_custom(rName, "SEMANTIC_OVERRIDE", "consolidate", "anthropic.claude-3-sonnet-20240229-v1:0", "extract", "anthropic.claude-3-haiku-20240307-v1:0"),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccMemoryStrategyConfig_semanticNamed(rName, strategyName string) string {
	return acctest.ConfigCompose(testAccMemoryConfig_memoryExecutionRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_memory_strategy" "test" {
  name                      = %[2]q
  memory_id                 = aws_bedrockagentcore_memory.test.id
  memory_execution_role_arn = aws_bedrockagentcore_memory.test.memory_execution_role_arn
  type                      = "SEMANTIC"
  description               = "Rename test"
  namespaces                = ["default"]
}
`, rName, strategyName))
}

func testAccMemoryStrategyConfig_recordSchemaToggle(rName string, withSchema bool) string {
	schemaBlock := ""
	if withSchema {
		schemaBlock = `
  memory_record_schema {
    metadata_schema {
      key             = "priority"
      extraction_type = "LLM_INFERRED"
      type            = "STRING"

      extraction_config {
        llm_extraction_config {
          definition = "The priority of the record"

          validation {
            string_validation {
              allowed_values = ["high", "medium", "low"]
            }
          }
        }
      }
    }
  }`
	}
	return acctest.ConfigCompose(testAccMemoryConfig_memoryExecutionRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_memory_strategy" "test" {
  name                      = %[1]q
  memory_id                 = aws_bedrockagentcore_memory.test.id
  memory_execution_role_arn = aws_bedrockagentcore_memory.test.memory_execution_role_arn
  type                      = "SEMANTIC"
  description               = "Record schema removal test"
  namespaces                = ["default"]
%[2]s
}
`, rName, schemaBlock))
}

func testAccMemoryStrategyConfig_emptyNumberValidation(rName string) string {
	return acctest.ConfigCompose(testAccMemoryConfig_memoryExecutionRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_memory_strategy" "test" {
  name                      = %[1]q
  memory_id                 = aws_bedrockagentcore_memory.test.id
  memory_execution_role_arn = aws_bedrockagentcore_memory.test.memory_execution_role_arn
  type                      = "SEMANTIC"
  description               = "Empty number_validation test"
  namespaces                = ["default"]

  memory_record_schema {
    metadata_schema {
      key = "priority"

      extraction_config {
        llm_extraction_config {
          definition = "The priority of the record"

          validation {
            number_validation {}
          }
        }
      }
    }
  }
}
`, rName))
}
