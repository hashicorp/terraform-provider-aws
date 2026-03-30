// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfbedrockagentcore "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagentcore"
	"github.com/hashicorp/terraform-provider-aws/names"
)

var memoryConfig = func(rName string) string {
	return testAccMemoryConfig_basic(rName)
}

func TestAccBedrockAgentCoreMemoryStrategy_standard(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var memoryStrategy1, memoryStrategy2, memoryStrategy3, memoryStrategy4 awstypes.MemoryStrategy
	rName := strings.ReplaceAll(acctest.RandomWithPrefix(t, acctest.ResourcePrefix), "-", "_")
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
				Config: memoryConfig(rName),
			},
			// Step 1: Create semantic strategy
			{
				Config: testAccMemoryStrategyConfig(rName, "SEMANTIC", names.AttrDescription, "default"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &memoryStrategy1),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "SEMANTIC"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, names.AttrDescription),
					resource.TestCheckResourceAttr(resourceName, "namespaces.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "namespaces.*", "default"),
					resource.TestCheckResourceAttrSet(resourceName, "memory_strategy_id"),
				),
			},
			// Step 2: Update description + namespace (in-place)
			{
				Config: testAccMemoryStrategyConfig(rName, "SEMANTIC", "Updated description", "custom"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &memoryStrategy2),
					testAccCheckMemoryStrategyNotRecreated(&memoryStrategy1, &memoryStrategy2),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Updated description"),
					resource.TestCheckResourceAttr(resourceName, "namespaces.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "namespaces.*", "custom"),
				),
			},
			// Step 3: Change type semantic→summary (replacement)
			{
				Config: testAccMemoryStrategyConfig(rName, "SUMMARIZATION", "Summary strategy", "{sessionId}"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &memoryStrategy3),
					testAccCheckMemoryStrategyRecreated(&memoryStrategy2, &memoryStrategy3),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "SUMMARIZATION"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Summary strategy"),
					resource.TestCheckResourceAttr(resourceName, "namespaces.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "namespaces.*", "{sessionId}"),
					resource.TestCheckResourceAttrSet(resourceName, "memory_strategy_id"),
				),
			},
			// Step 4: Change type summary→user_preference (replacement)
			{
				Config: testAccMemoryStrategyConfig(rName, "USER_PREFERENCE", "User preference strategy", "preferences"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &memoryStrategy4),
					testAccCheckMemoryStrategyRecreated(&memoryStrategy3, &memoryStrategy4),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, "USER_PREFERENCE"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "User preference strategy"),
					resource.TestCheckResourceAttr(resourceName, "namespaces.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "namespaces.*", "preferences"),
					resource.TestCheckResourceAttrSet(resourceName, "memory_strategy_id"),
				),
			},
			// Step 5: Try to create ANOTHER user_preference strategy → should ERROR
			{
				Config:      testAccMemoryStrategyConfig_duplicateType(rName),
				ExpectError: regexache.MustCompile("Found multiple strategies of type"),
			},
			// Step 6: Import test - verify composite ID import works
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccMemoryStrategyImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "memory_strategy_id",
			},
		},
	})
}

func TestAccBedrockAgentCoreMemoryStrategy_custom(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var r1, r2, r5 awstypes.MemoryStrategy
	rName := strings.ReplaceAll(acctest.RandomWithPrefix(t, acctest.ResourcePrefix), "-", "_")
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
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &r1),
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
			},
			// Step 3: Add extraction block and update consolidation properties (same override type)
			{
				Config: testAccMemoryStrategyConfig_custom(rName, "SEMANTIC_OVERRIDE", "Updated semantic consolidation", "anthropic.claude-3-sonnet-20240229-v1:0", "Extract semantic meaning", "anthropic.claude-3-haiku-20240307-v1:0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &r2),
					testAccCheckMemoryStrategyNotRecreated(&r1, &r2),
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
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &r5),
					testAccCheckMemoryStrategyRecreated(&r2, &r5),
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
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &r5),
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
				ImportStateIdFunc:                    testAccMemoryStrategyImportStateIdFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "memory_strategy_id",
				ImportStateVerifyIgnore:              []string{"memory_execution_role_arn"},
			},
		},
	})
}

func TestAccBedrockAgentCoreMemoryStrategy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var memorystrategy awstypes.MemoryStrategy
	rName := strings.ReplaceAll(acctest.RandomWithPrefix(t, acctest.ResourcePrefix), "-", "_")
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
				Config: testAccMemoryStrategyConfig(rName, "SEMANTIC", "Example Description", "default"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &memorystrategy),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfbedrockagentcore.ResourceMemoryStrategy, resourceName),
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

func testAccCheckMemoryStrategyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagentcore_memory_strategy" {
				continue
			}

			memoryStrategyId := rs.Primary.Attributes["memory_strategy_id"]
			_, err := tfbedrockagentcore.FindMemoryStrategyByID(ctx, conn, rs.Primary.Attributes["memory_id"], memoryStrategyId)
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.BedrockAgentCore, create.ErrActionCheckingDestroyed, tfbedrockagentcore.ResNameMemoryStrategy, memoryStrategyId, err)
			}

			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingDestroyed, tfbedrockagentcore.ResNameMemoryStrategy, memoryStrategyId, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckMemoryStrategyExists(ctx context.Context, t *testing.T, name string, memorystrategy *awstypes.MemoryStrategy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingExistence, tfbedrockagentcore.ResNameMemoryStrategy, name, errors.New("not found"))
		}

		memoryStrategyId := rs.Primary.Attributes["memory_strategy_id"]
		if memoryStrategyId == "" {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingExistence, tfbedrockagentcore.ResNameMemoryStrategy, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		resp, err := tfbedrockagentcore.FindMemoryStrategyByID(ctx, conn, rs.Primary.Attributes["memory_id"], memoryStrategyId)
		if err != nil {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingExistence, tfbedrockagentcore.ResNameMemoryStrategy, memoryStrategyId, err)
		}

		*memorystrategy = *resp

		return nil
	}
}

func testAccCheckMemoryStrategyNotRecreated(before, after *awstypes.MemoryStrategy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if beforeID, afterID := aws.ToString(before.StrategyId), aws.ToString(after.StrategyId); beforeID != afterID {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingNotRecreated, tfbedrockagentcore.ResNameMemoryStrategy, beforeID, errors.New("recreated"))
		}

		return nil
	}
}

func testAccCheckMemoryStrategyRecreated(before, after *awstypes.MemoryStrategy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if beforeID, afterID := aws.ToString(before.StrategyId), aws.ToString(after.StrategyId); beforeID == afterID {
			return create.Error(names.BedrockAgentCore, create.ErrActionCheckingRecreated, tfbedrockagentcore.ResNameMemoryStrategy, beforeID, errors.New("not recreated"))
		}

		return nil
	}
}

func testAccMemoryStrategyImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s,%s", rs.Primary.Attributes["memory_id"], rs.Primary.Attributes["memory_strategy_id"]), nil
	}
}

func testAccMemoryStrategyConfig(rName, strategyType, description, namespace string) string {
	return acctest.ConfigCompose(memoryConfig(rName), fmt.Sprintf(`	

resource "aws_bedrockagentcore_memory_strategy" "test" {
  name        = %[1]q
  memory_id   = aws_bedrockagentcore_memory.test.id
  type        = %[2]q
  description = %[3]q
  namespaces  = [%[4]q]
}
`, rName, strategyType, description, namespace))
}

func testAccMemoryStrategyConfig_duplicateType(rName string) string {
	return fmt.Sprintf(`
%s

resource "aws_bedrockagentcore_memory_strategy" "test2" {
  name        = "%s_duplicate"
  memory_id   = aws_bedrockagentcore_memory.test.id
  type        = "USER_PREFERENCE"
  description = "Duplicate user preference strategy"
  namespaces  = ["preferences2"]
}
`, testAccMemoryStrategyConfig(rName, "USER_PREFERENCE", "User preference strategy", "preferences"), rName)
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
