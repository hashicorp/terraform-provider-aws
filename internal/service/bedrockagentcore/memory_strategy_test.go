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

func TestAccBedrockAgentCoreMemoryStrategy_basic(t *testing.T) {
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
				Config: testAccMemoryStrategyConfig_basicNamespaceTemplates(rName, awstypes.MemoryStrategyTypeSemantic, "default"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory_execution_role_arn"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory_strategy_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("namespaces"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("default"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("namespace_templates"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("default"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(awstypes.MemoryStrategyTypeSemantic)),
				},
			},
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
				Config: testAccMemoryStrategyConfig_basicNamespaceTemplates(rName, awstypes.MemoryStrategyTypeSemantic, "default"),
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

func TestAccBedrockAgentCoreMemoryStrategy_namespacesToNamespaceTemplates(t *testing.T) {
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
				Config: testAccMemoryStrategyConfig_basicNamespaces(rName, awstypes.MemoryStrategyTypeSemantic, "{sessionId}"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("namespaces"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("{sessionId}"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("namespace_templates"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("{sessionId}"),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccMemoryStrategyImportStateIDFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "memory_strategy_id",
				ImportStateVerifyIgnore:              []string{"memory_execution_role_arn"},
			},
			{
				Config: testAccMemoryStrategyConfig_basicNamespaceTemplates(rName, awstypes.MemoryStrategyTypeSemantic, "{sessionId}"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
			{
				Config: testAccMemoryStrategyConfig_basicNamespaces(rName, awstypes.MemoryStrategyTypeSemantic, "default"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("namespaces"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("default"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("namespace_templates"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("default"),
					})),
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreMemoryStrategy_namespaceTemplatesToNamespaces(t *testing.T) {
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
				Config: testAccMemoryStrategyConfig_basicNamespaceTemplates(rName, awstypes.MemoryStrategyTypeSemantic, "{sessionId}"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("namespaces"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("{sessionId}"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("namespace_templates"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("{sessionId}"),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    testAccMemoryStrategyImportStateIDFunc(resourceName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "memory_strategy_id",
				ImportStateVerifyIgnore:              []string{"memory_execution_role_arn"},
			},
			{
				Config: testAccMemoryStrategyConfig_basicNamespaces(rName, awstypes.MemoryStrategyTypeSemantic, "{sessionId}"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
			{
				Config: testAccMemoryStrategyConfig_basicNamespaceTemplates(rName, awstypes.MemoryStrategyTypeSemantic, "default"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("namespaces"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("default"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("namespace_templates"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("default"),
					})),
				},
			},
		},
	})
}

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
				Config: testAccMemoryStrategyConfig_withExecutionRole(rName, awstypes.MemoryStrategyTypeEpisodic, "Episodic strategy", "/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("Episodic strategy")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory_execution_role_arn"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory_strategy_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("namespace_templates"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(awstypes.MemoryStrategyTypeEpisodic)),
				},
			},
			// Step 2: Update episodic description (in-place)
			{
				Config: testAccMemoryStrategyConfig_withExecutionRole(rName, awstypes.MemoryStrategyTypeEpisodic, "Updated episodic strategy", "/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("Updated episodic strategy")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory_execution_role_arn"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory_strategy_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("namespace_templates"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("/strategies/{memoryStrategyId}/actors/{actorId}/sessions/{sessionId}"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(awstypes.MemoryStrategyTypeEpisodic)),
				},
			},
			// Step 3: Change type episodic→semantic (replacement)
			{
				Config: testAccMemoryStrategyConfig_withExecutionRole(rName, awstypes.MemoryStrategyTypeSemantic, "desc1", "default"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("desc1")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory_execution_role_arn"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory_strategy_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("namespace_templates"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("default"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(awstypes.MemoryStrategyTypeSemantic)),
				},
			},
			// Step 4: Update description + namespace (in-place)
			{
				Config: testAccMemoryStrategyConfig_withExecutionRole(rName, awstypes.MemoryStrategyTypeSemantic, "desc2", "custom"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("desc2")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory_execution_role_arn"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory_strategy_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("namespace_templates"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("custom"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(awstypes.MemoryStrategyTypeSemantic)),
				},
			},
			// Step 5: Change type semantic→user_preference (replacement)
			{
				Config: testAccMemoryStrategyConfig_withExecutionRole(rName, awstypes.MemoryStrategyTypeUserPreference, "User preference strategy", "preferences"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("User preference strategy")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory_execution_role_arn"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory_strategy_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("namespace_templates"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("preferences"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(awstypes.MemoryStrategyTypeUserPreference)),
				},
			},
			// Step 6: Try to create ANOTHER user_preference strategy → should ERROR
			{
				Config:      testAccMemoryStrategyConfig_duplicateType(rName, awstypes.MemoryStrategyTypeUserPreference),
				ExpectError: regexache.MustCompile("Found multiple strategies of type"),
			},
			// Step 7: Change type user_preference→summarization (replacement)
			{
				Config: testAccMemoryStrategyConfig_withExecutionRole(rName, awstypes.MemoryStrategyTypeSummarization, "Summarization strategy", "{sessionId}"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("Summarization strategy")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory_execution_role_arn"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory_strategy_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("namespace_templates"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.StringExact("{sessionId}"),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(awstypes.MemoryStrategyTypeSummarization)),
				},
			},
			// Step 8: Import test - verify composite ID import works
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
				Config: testAccMemoryStrategyConfig_customConsolidationOnly(rName, awstypes.OverrideTypeSemanticOverride, "Focus on semantic relationships", "us.amazon.nova-2-lite-v1:0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"consolidation": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
							"append_to_prompt": knownvalue.StringExact("Focus on semantic relationships"),
							"model_id":         knownvalue.StringExact("us.amazon.nova-2-lite-v1:0"),
						})}),
						"extraction":   knownvalue.ListSizeExact(0),
						names.AttrType: tfknownvalue.StringExact(awstypes.OverrideTypeSemanticOverride),
					})})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(awstypes.MemoryStrategyTypeCustom)),
				},
			},
			// Step 3: Add extraction block and update consolidation properties (same override type)
			{
				Config: testAccMemoryStrategyConfig_custom(rName, awstypes.OverrideTypeSemanticOverride, "Updated semantic consolidation", "amazon.nova-lite-v1:0", "Extract semantic meaning", "us.amazon.nova-2-lite-v1:0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"consolidation": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
							"append_to_prompt": knownvalue.StringExact("Updated semantic consolidation"),
							"model_id":         knownvalue.StringExact("amazon.nova-lite-v1:0"),
						})}),
						"extraction": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
							"append_to_prompt": knownvalue.StringExact("Extract semantic meaning"),
							"model_id":         knownvalue.StringExact("us.amazon.nova-2-lite-v1:0"),
						})}),
						names.AttrType: tfknownvalue.StringExact(awstypes.OverrideTypeSemanticOverride),
					})})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(awstypes.MemoryStrategyTypeCustom)),
				},
			},
			// Step 4: Try to remove consolidation block → should ERROR
			{
				Config:      testAccMemoryStrategyConfig_customExtractionOnly(rName, awstypes.OverrideTypeSemanticOverride, "Extract semantic meaning", "us.amazon.nova-2-lite-v1:0"),
				ExpectError: regexache.MustCompile("Removing the previously configured \"consolidation\" block"),
			},
			//// Step 5: Change override type → should replace resource
			{
				Config: testAccMemoryStrategyConfig_custom(rName, awstypes.OverrideTypeUserPreferenceOverride, "Store user preferences", "amazon.nova-lite-v1:0", "Extract user preferences", "us.amazon.nova-2-lite-v1:0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"consolidation": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
							"append_to_prompt": knownvalue.StringExact("Store user preferences"),
							"model_id":         knownvalue.StringExact("amazon.nova-lite-v1:0"),
						})}),
						"extraction": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
							"append_to_prompt": knownvalue.StringExact("Extract user preferences"),
							"model_id":         knownvalue.StringExact("us.amazon.nova-2-lite-v1:0"),
						})}),
						names.AttrType: tfknownvalue.StringExact(awstypes.OverrideTypeUserPreferenceOverride),
					})})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(awstypes.MemoryStrategyTypeCustom)),
				},
			},
			//// Step 6: SUMMARY_OVERRIDE with extraction block → ValidateConfig error
			{
				Config:      testAccMemoryStrategyConfig_custom(rName, awstypes.OverrideTypeSummaryOverride, "Summary consolidation", "amazon.nova-lite-v1:0", "Summary extraction", "us.amazon.nova-2-lite-v1:0"),
				ExpectError: regexache.MustCompile(`Attribute "configuration\[0\].extraction" must not be configured`),
			},
			//// Step 7: SUMMARY_OVERRIDE with no extraction block → should succeed
			{
				Config: testAccMemoryStrategyConfig_customConsolidationOnly(rName, awstypes.OverrideTypeSummaryOverride, "Summary consolidation only", "amazon.nova-lite-v1:0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckMemoryStrategyExists(ctx, t, resourceName, &m),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration), knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
						"consolidation": knownvalue.ListExact([]knownvalue.Check{knownvalue.ObjectExact(map[string]knownvalue.Check{
							"append_to_prompt": knownvalue.StringExact("Summary consolidation only"),
							"model_id":         knownvalue.StringExact("amazon.nova-lite-v1:0"),
						})}),
						"extraction":   knownvalue.ListSizeExact(0),
						names.AttrType: tfknownvalue.StringExact(awstypes.OverrideTypeSummaryOverride),
					})})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrType), tfknownvalue.StringExact(awstypes.MemoryStrategyTypeCustom)),
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

func testAccMemoryStrategyConfig_basicNamespaces(rName string, strategyType awstypes.MemoryStrategyType, nss ...string) string {
	return acctest.ConfigCompose(testAccMemoryConfig_basic(rName), fmt.Sprintf(`	
resource "aws_bedrockagentcore_memory_strategy" "test" {
  name       = %[1]q
  memory_id  = aws_bedrockagentcore_memory.test.id
  type       = %[2]q
  namespaces = [%[3]s]
}
`, rName, strategyType, acctest.ListOfStrings(nss...)))
}

func testAccMemoryStrategyConfig_basicNamespaceTemplates(rName string, strategyType awstypes.MemoryStrategyType, nss ...string) string {
	return acctest.ConfigCompose(testAccMemoryConfig_basic(rName), fmt.Sprintf(`	
resource "aws_bedrockagentcore_memory_strategy" "test" {
  name                = %[1]q
  memory_id           = aws_bedrockagentcore_memory.test.id
  type                = %[2]q
  namespace_templates = [%[3]s]
}
`, rName, strategyType, acctest.ListOfStrings(nss...)))
}

func testAccMemoryStrategyConfig_withExecutionRole(rName string, strategyType awstypes.MemoryStrategyType, description string, nss ...string) string {
	return acctest.ConfigCompose(testAccMemoryConfig_memoryExecutionRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_memory_strategy" "test" {
  name                      = %[1]q
  memory_id                 = aws_bedrockagentcore_memory.test.id
  memory_execution_role_arn = aws_bedrockagentcore_memory.test.memory_execution_role_arn
  type                      = %[2]q
  description               = %[3]q
  namespace_templates       = [%[4]s]
}
`, rName, strategyType, description, acctest.ListOfStrings(nss...)))
}

func testAccMemoryStrategyConfig_duplicateType(rName string, strategyType awstypes.MemoryStrategyType) string {
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
  namespace_templates       = [%[3]q]
}
`, rName, strategyType, duplicateNamespace))
}

func testAccMemoryStrategyConfig_custom(rName string, overrideType awstypes.OverrideType, consolidationPrompt, consolidationModel, extractionPrompt, extractionModel string) string {
	return acctest.ConfigCompose(testAccMemoryConfig_memoryExecutionRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_memory_strategy" "test" {
  name                      = %[1]q
  memory_id                 = aws_bedrockagentcore_memory.test.id
  memory_execution_role_arn = aws_bedrockagentcore_memory.test.memory_execution_role_arn
  type                      = "CUSTOM"
  description               = "Test custom strategy"
  namespace_templates       = ["{sessionId}"]

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

func testAccMemoryStrategyConfig_customConsolidationOnly(rName string, overrideType awstypes.OverrideType, consolidationPrompt, consolidationModel string) string {
	return acctest.ConfigCompose(testAccMemoryConfig_memoryExecutionRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_memory_strategy" "test" {
  name                      = %[1]q
  memory_id                 = aws_bedrockagentcore_memory.test.id
  memory_execution_role_arn = aws_bedrockagentcore_memory.test.memory_execution_role_arn
  type                      = "CUSTOM"
  description               = "Test custom strategy"
  namespace_templates       = ["{sessionId}"]

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

func testAccMemoryStrategyConfig_customExtractionOnly(rName string, overrideType awstypes.OverrideType, extractionPrompt, extractionModel string) string {
	return acctest.ConfigCompose(testAccMemoryConfig_memoryExecutionRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_memory_strategy" "test" {
  name                      = %[1]q
  memory_id                 = aws_bedrockagentcore_memory.test.id
  memory_execution_role_arn = aws_bedrockagentcore_memory.test.memory_execution_role_arn
  type                      = "CUSTOM"
  description               = "Test custom strategy"
  namespace_templates       = ["{sessionId}"]

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
  namespace_templates       = ["default"]
}
`, rName))
}
