// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrockagentcorecontrol/types"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfbedrockagentcore "github.com/hashicorp/terraform-provider-aws/internal/service/bedrockagentcore"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type memoryConfigType string

const (
	memoryNone         memoryConfigType = "none"
	memoryAgentCore    memoryConfigType = "agentcore"
	memoryDisabled     memoryConfigType = "disabled"
	memoryManagedEmpty memoryConfigType = "managed_empty"
	memoryManaged      memoryConfigType = "managed"
)

func testAccRandomHarnessName(t *testing.T) string {
	return strings.ReplaceAll(acctest.RandomWithPrefix(t, acctest.ResourcePrefix), "-", "_")
}

func checkHarnessARN(name string) knownvalue.Check {
	return tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`harness/`+name+`-[a-zA-Z0-9]{10}`))
}

func checkHarnessARNAlternateRegion(name string) knownvalue.Check {
	return tfknownvalue.RegionalARNAlternateRegionRegexp("bedrock-agentcore", regexache.MustCompile(`harness/`+name+`-[a-zA-Z0-9]{10}`))
}

func TestAccBedrockAgentCoreHarness_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var harness awstypes.Harness
	rName := testAccRandomHarnessName(t)
	resourceName := "aws_bedrockagentcore_harness.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckHarness(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHarnessDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Harness/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("allowed_tools"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.StringExact("*"),
					})),
					tfstatecheck.ExpectRegionalARNFormat(resourceName, tfjsonpath.New(names.AttrARN), "bedrock-agentcore", "harness/{harness_id}"),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("authorizer_configuration"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEnvironment), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("environment_actual"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"agentcore_runtime_environment": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"agent_runtime_arn":            tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`runtime/harness_`+rName+`-[a-zA-Z0-9]+`)),
									"agent_runtime_id":             knownvalue.StringRegexp(regexache.MustCompile(`^harness_` + rName + `-[a-zA-Z0-9]+$`)),
									"agent_runtime_name":           knownvalue.StringExact("harness_" + rName),
									"filesystem_configuration":     knownvalue.Null(),
									"lifecycle_configuration":      knownvalue.NotNull(),
									names.AttrNetworkConfiguration: knownvalue.NotNull(),
								}),
							}),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("environment_artifact"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("environment_variables"), knownvalue.Null()),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrExecutionRoleARN), "aws_iam_role.test", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("harness_id"), knownvalue.StringRegexp(regexache.MustCompile(`^`+rName+`-[a-zA-Z0-9]{10}$`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("harness_name"), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("max_iterations"), knownvalue.Int32Exact(75)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("max_tokens"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory_actual").AtSliceIndex(0), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"agentcore_memory_configuration": knownvalue.Null(),
						"disabled":                       knownvalue.Null(),
						"managed_memory_configuration": knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrARN:           tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`memory/`+rName+`-[a-zA-Z0-9]{10}`)),
								"encryption_key_arn":    knownvalue.Null(),
								"event_expiry_duration": knownvalue.Int32Exact(30),
								"strategies": knownvalue.SetExact([]knownvalue.Check{
									knownvalue.StringExact("SEMANTIC"),
									knownvalue.StringExact("SUMMARIZATION"),
								}),
							}),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("model"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"bedrock_model_config": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"max_tokens":  knownvalue.Null(),
									"model_id":    knownvalue.StringExact("anthropic.claude-sonnet-4-20250514"),
									"temperature": knownvalue.Null(),
									"top_p":       knownvalue.Null(),
								}),
							}),
							"gemini_model_config": knownvalue.ListSizeExact(0),
							"openai_model_config": knownvalue.ListSizeExact(0),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("skill"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("system_prompt"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"text": knownvalue.StringExact("You are a helpful assistant."),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("timeout_seconds"), knownvalue.Int32Exact(3600)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("tool"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("truncation"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"strategy": knownvalue.StringExact("sliding_window"),
							"config": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"sliding_window": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectExact(map[string]knownvalue.Check{
											"messages_count": knownvalue.Int32Exact(150),
										}),
									}),
									"summarization": knownvalue.Null(),
								}),
							}),
						}),
					})),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Harness/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "harness_id"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "harness_id",
				ImportStateVerifyIgnore:              []string{names.AttrEnvironment, "memory"},
				ImportStateCheck: acctest.ComposeAggregateImportStateCheckFunc(
					acctest.ImportCheckResourceAttr("memory.#", "1"),
					acctest.ImportCheckResourceAttr("memory.0.agentcore_memory_configuration.#", "0"),
					acctest.ImportCheckResourceAttr("memory.0.managed_memory_configuration.#", "1"),
					acctest.ImportMatchResourceAttr("memory.0.managed_memory_configuration.0.arn", regexache.MustCompile(`^arn:[^:]+:bedrock-agentcore:[^:]+:\d{12}:memory/`+rName+`-[a-zA-Z0-9]{10}$`)),
					acctest.ImportCheckResourceAttr("memory.0.managed_memory_configuration.0.encryption_key_arn", ""),
					acctest.ImportCheckResourceAttr("memory.0.managed_memory_configuration.0.event_expiry_duration", "30"),
					acctest.ImportCheckResourceAttr("memory.0.managed_memory_configuration.0.strategies.#", "2"),
					importCheckSetContains("memory.0.managed_memory_configuration.0.strategies", "SEMANTIC"),
					importCheckSetContains("memory.0.managed_memory_configuration.0.strategies", "SUMMARIZATION"),
				),
			},
		},
	})
}

func TestAccBedrockAgentCoreHarness_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var harness awstypes.Harness
	rName := testAccRandomHarnessName(t)
	resourceName := "aws_bedrockagentcore_harness.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckHarness(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHarnessDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Harness/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfbedrockagentcore.ResourceHarness, resourceName),
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

func TestAccBedrockAgentCoreHarness_update_systemPrompt(t *testing.T) {
	ctx := acctest.Context(t)
	var harness awstypes.Harness
	rName := testAccRandomHarnessName(t)
	resourceName := "aws_bedrockagentcore_harness.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckHarness(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHarnessDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHarnessConfig_systemPrompt(rName, "You are a helpful assistant."),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccHarnessConfig_systemPrompt(rName, "You are a coding assistant."),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "harness_id"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "harness_id",
				ImportStateVerifyIgnore:              []string{names.AttrEnvironment, "memory"},
			},
		},
	})
}

func TestAccBedrockAgentCoreHarness_update_allowedTools(t *testing.T) {
	ctx := acctest.Context(t)
	var harness awstypes.Harness
	rName := testAccRandomHarnessName(t)
	resourceName := "aws_bedrockagentcore_harness.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckHarness(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHarnessDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHarnessConfig_allowedTools(rName, `["*"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccHarnessConfig_allowedTools(rName, `["@builtin"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "harness_id"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "harness_id",
				ImportStateVerifyIgnore:              []string{names.AttrEnvironment, "memory"},
			},
		},
	})
}

func TestAccBedrockAgentCoreHarness_update_limits(t *testing.T) {
	ctx := acctest.Context(t)
	var harness awstypes.Harness
	rName := testAccRandomHarnessName(t)
	resourceName := "aws_bedrockagentcore_harness.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckHarness(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHarnessDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHarnessConfig_limits(rName, 10, 4096, 300),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccHarnessConfig_limits(rName, 20, 8192, 600),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "harness_id"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "harness_id",
				ImportStateVerifyIgnore:              []string{names.AttrEnvironment, "memory"},
			},
		},
	})
}

func TestAccBedrockAgentCoreHarness_model_bedrock(t *testing.T) {
	ctx := acctest.Context(t)
	var harness awstypes.Harness
	rName := testAccRandomHarnessName(t)
	resourceName := "aws_bedrockagentcore_harness.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckHarness(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHarnessDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHarnessConfig_bedrockModel(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "harness_id"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "harness_id",
				ImportStateVerifyIgnore: []string{
					names.AttrEnvironment,
					"memory",
					"model.0.bedrock_model_config.0.temperature",
					"model.0.bedrock_model_config.0.top_p",
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreHarness_model_liteLLM(t *testing.T) {
	ctx := acctest.Context(t)
	var harness awstypes.Harness
	rName := testAccRandomHarnessName(t)
	resourceName := "aws_bedrockagentcore_harness.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckHarness(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHarnessDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHarnessConfig_liteLLMModel(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
					resource.TestCheckResourceAttr(resourceName, "model.0.litellm_model_config.0.api_base", "https://api.example.com/v1"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreHarness_truncation_slidingWindow(t *testing.T) {
	ctx := acctest.Context(t)
	var harness awstypes.Harness
	rName := testAccRandomHarnessName(t)
	resourceName := "aws_bedrockagentcore_harness.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckHarness(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHarnessDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHarnessConfig_truncationSlidingWindow(rName, 50),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccHarnessConfig_truncationSlidingWindow(rName, 100),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "harness_id"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "harness_id",
				ImportStateVerifyIgnore:              []string{names.AttrEnvironment, "memory"},
			},
		},
	})
}

func TestAccBedrockAgentCoreHarness_truncation_summarization(t *testing.T) {
	ctx := acctest.Context(t)
	var harness awstypes.Harness
	rName := testAccRandomHarnessName(t)
	resourceName := "aws_bedrockagentcore_harness.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckHarness(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHarnessDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHarnessConfig_truncationSummarization(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "harness_id"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "harness_id",
				ImportStateVerifyIgnore:              []string{names.AttrEnvironment, "memory"},
			},
		},
	})
}

func TestAccBedrockAgentCoreHarness_tools_inlineFunction(t *testing.T) {
	ctx := acctest.Context(t)
	var harness awstypes.Harness
	rName := testAccRandomHarnessName(t)
	resourceName := "aws_bedrockagentcore_harness.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckHarness(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHarnessDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHarnessConfig_toolInlineFunction(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "harness_id"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "harness_id",
				ImportStateVerifyIgnore:              []string{names.AttrEnvironment, "memory"},
			},
		},
	})
}

func TestAccBedrockAgentCoreHarness_environmentVariables(t *testing.T) {
	ctx := acctest.Context(t)
	var harness awstypes.Harness
	rName := testAccRandomHarnessName(t)
	resourceName := "aws_bedrockagentcore_harness.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckHarness(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHarnessDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHarnessConfig_environmentVariables(rName, "KEY1", acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccHarnessConfig_environmentVariables(rName, "KEY2", acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "harness_id"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "harness_id",
				ImportStateVerifyIgnore: []string{
					names.AttrEnvironment,
					"environment_variables",
					"memory",
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreHarness_Memory_agentCoreMemoryConfiguration_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var harness awstypes.Harness
	rName := testAccRandomHarnessName(t)
	resourceName := "aws_bedrockagentcore_harness.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckHarness(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHarnessDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHarnessConfig_Memory_agentCoreMemoryConfiguration_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory").AtSliceIndex(0), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"agentcore_memory_configuration": knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrARN:      knownvalue.NotNull(),
								"actor_id":         knownvalue.Null(),
								"messages_count":   knownvalue.Null(),
								"retrieval_config": knownvalue.ListSizeExact(0),
							}),
						}),
						"disabled":                     knownvalue.ListSizeExact(0),
						"managed_memory_configuration": knownvalue.ListSizeExact(0),
					})),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("memory").AtSliceIndex(0).AtMapKey("agentcore_memory_configuration").AtSliceIndex(0).AtMapKey(names.AttrARN), "aws_bedrockagentcore_memory.test", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory_actual").AtSliceIndex(0), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"agentcore_memory_configuration": knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrARN:      knownvalue.NotNull(),
								"actor_id":         knownvalue.Null(),
								"messages_count":   knownvalue.Null(),
								"retrieval_config": knownvalue.Null(),
							}),
						}),
						"disabled":                     knownvalue.Null(),
						"managed_memory_configuration": knownvalue.Null(),
					})),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("memory_actual").AtSliceIndex(0).AtMapKey("agentcore_memory_configuration").AtSliceIndex(0).AtMapKey(names.AttrARN), "aws_bedrockagentcore_memory.test", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
			{
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "harness_id"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "harness_id",
				ImportStateVerifyIgnore:              []string{names.AttrEnvironment},
			},
		},
	})
}

func TestAccBedrockAgentCoreHarness_Memory_agentCoreMemoryConfiguration_options(t *testing.T) {
	ctx := acctest.Context(t)
	var harness awstypes.Harness
	rName := testAccRandomHarnessName(t)
	resourceName := "aws_bedrockagentcore_harness.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckHarness(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHarnessDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHarnessConfig_Memory_agentCoreMemoryConfiguration_options(rName, "actor1", 10, "/namespace1/{actorId}", 0.25, 5),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory").AtSliceIndex(0).AtMapKey("agentcore_memory_configuration").AtSliceIndex(0), knownvalue.ObjectExact(map[string]knownvalue.Check{
						names.AttrARN:    knownvalue.NotNull(),
						"actor_id":       knownvalue.StringExact("actor1"),
						"messages_count": knownvalue.Int32Exact(10),
						"retrieval_config": knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"map_block_key":   knownvalue.StringExact("/namespace1/{actorId}"),
								"relevance_score": knownvalue.Float32Exact(0.25),
								"strategy_id":     knownvalue.Null(),
								"top_k":           knownvalue.Int32Exact(5),
							}),
						}),
					})),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("memory").AtSliceIndex(0).AtMapKey("agentcore_memory_configuration").AtSliceIndex(0).AtMapKey(names.AttrARN), "aws_bedrockagentcore_memory.test", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
			{
				Config: testAccHarnessConfig_Memory_agentCoreMemoryConfiguration_options(rName, "actor2", 20, "/namespace2/{actorId}", 0.35, 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory").AtSliceIndex(0).AtMapKey("agentcore_memory_configuration").AtSliceIndex(0), knownvalue.ObjectExact(map[string]knownvalue.Check{
						names.AttrARN:    knownvalue.NotNull(),
						"actor_id":       knownvalue.StringExact("actor2"),
						"messages_count": knownvalue.Int32Exact(20),
						"retrieval_config": knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"map_block_key":   knownvalue.StringExact("/namespace2/{actorId}"),
								"relevance_score": knownvalue.Float32Exact(0.35),
								"strategy_id":     knownvalue.Null(),
								"top_k":           knownvalue.Int32Exact(10),
							}),
						}),
					})),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("memory").AtSliceIndex(0).AtMapKey("agentcore_memory_configuration").AtSliceIndex(0).AtMapKey(names.AttrARN), "aws_bedrockagentcore_memory.test", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
			{
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "harness_id"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "harness_id",
				ImportStateVerifyIgnore: []string{
					names.AttrEnvironment,
					"memory.0.agentcore_memory_configuration.0.retrieval_config.0.relevance_score",
				},
				ImportStateCheck: acctest.ComposeAggregateImportStateCheckFunc(
					// TODO: float32 precision issue
					acctest.ImportMatchResourceAttr("memory.0.agentcore_memory_configuration.0.retrieval_config.0.relevance_score", regexache.MustCompile(`^0\.34`)),
				),
			},
		},
	})
}

func TestAccBedrockAgentCoreHarness_Memory_agentCoreMemoryConfiguration_addRetrievalConfig(t *testing.T) {
	ctx := acctest.Context(t)
	var harness awstypes.Harness
	rName := testAccRandomHarnessName(t)
	resourceName := "aws_bedrockagentcore_harness.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckHarness(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHarnessDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHarnessConfig_Memory_agentCoreMemoryConfiguration_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory").AtSliceIndex(0).AtMapKey("agentcore_memory_configuration").AtSliceIndex(0).AtMapKey("retrieval_config"), knownvalue.ListSizeExact(0)),
				},
			},
			{
				Config: testAccHarnessConfig_Memory_agentCoreMemoryConfiguration_retrievalConfig(rName, "/namespace1", 0.25, 5),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory").AtSliceIndex(0).AtMapKey("agentcore_memory_configuration").AtSliceIndex(0).AtMapKey("retrieval_config"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"map_block_key":   knownvalue.StringExact("/namespace1"),
							"relevance_score": knownvalue.Float32Exact(0.25),
							"strategy_id":     knownvalue.Null(),
							"top_k":           knownvalue.Int32Exact(5),
						}),
					})),
				},
			},
			{
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "harness_id"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "harness_id",
				ImportStateVerifyIgnore: []string{
					names.AttrEnvironment,
					"memory.0.agentcore_memory_configuration.0.retrieval_config.0.relevance_score",
				},
				ImportStateCheck: acctest.ComposeAggregateImportStateCheckFunc(
					// TODO: float32 precision issue
					acctest.ImportMatchResourceAttr("memory.0.agentcore_memory_configuration.0.retrieval_config.0.relevance_score", regexache.MustCompile(`^0\.25`)),
				),
			},
		},
	})
}

func TestAccBedrockAgentCoreHarness_Memory_agentCoreMemoryConfiguration_removeRetrievalConfig(t *testing.T) {
	ctx := acctest.Context(t)
	var harness awstypes.Harness
	rName := testAccRandomHarnessName(t)
	resourceName := "aws_bedrockagentcore_harness.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckHarness(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHarnessDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHarnessConfig_Memory_agentCoreMemoryConfiguration_retrievalConfig(rName, "/namespace1", 0.25, 5),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory").AtSliceIndex(0).AtMapKey("agentcore_memory_configuration").AtSliceIndex(0).AtMapKey("retrieval_config"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"map_block_key":   knownvalue.StringExact("/namespace1"),
							"relevance_score": knownvalue.Float32Exact(0.25),
							"strategy_id":     knownvalue.Null(),
							"top_k":           knownvalue.Int32Exact(5),
						}),
					})),
				},
			},
			{
				Config: testAccHarnessConfig_Memory_agentCoreMemoryConfiguration_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory").AtSliceIndex(0).AtMapKey("agentcore_memory_configuration").AtSliceIndex(0).AtMapKey("retrieval_config"), knownvalue.ListSizeExact(0)),
				},
			},
			{
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "harness_id"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "harness_id",
				ImportStateVerifyIgnore:              []string{names.AttrEnvironment},
			},
		},
	})
}

func TestAccBedrockAgentCoreHarness_Memory_managedMemoryConfiguration_empty(t *testing.T) {
	ctx := acctest.Context(t)
	var harness awstypes.Harness
	rName := testAccRandomHarnessName(t)
	resourceName := "aws_bedrockagentcore_harness.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckHarness(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHarnessDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHarnessConfig_Memory_managedMemoryConfiguration_empty(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory").AtSliceIndex(0), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"agentcore_memory_configuration": knownvalue.ListSizeExact(0),
						"disabled":                       knownvalue.ListSizeExact(0),
						"managed_memory_configuration": knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrARN:           tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`memory/`+rName+`-[a-zA-Z0-9]{10}`)),
								"encryption_key_arn":    knownvalue.Null(),
								"event_expiry_duration": knownvalue.Int32Exact(30),
								"strategies": knownvalue.SetExact([]knownvalue.Check{
									knownvalue.StringExact("SEMANTIC"),
									knownvalue.StringExact("SUMMARIZATION"),
								}),
							}),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory_actual").AtSliceIndex(0), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"agentcore_memory_configuration": knownvalue.Null(),
						"disabled":                       knownvalue.Null(),
						"managed_memory_configuration": knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								names.AttrARN:           tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`memory/`+rName+`-[a-zA-Z0-9]{10}`)),
								"encryption_key_arn":    knownvalue.Null(),
								"event_expiry_duration": knownvalue.Int32Exact(30),
								"strategies": knownvalue.SetExact([]knownvalue.Check{
									knownvalue.StringExact("SEMANTIC"),
									knownvalue.StringExact("SUMMARIZATION"),
								}),
							}),
						}),
					})),
				},
			},
			{
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "harness_id"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "harness_id",
				ImportStateVerifyIgnore:              []string{names.AttrEnvironment},
			},
		},
	})
}

func TestAccBedrockAgentCoreHarness_Memory_managedMemoryConfiguration_update(t *testing.T) {
	ctx := acctest.Context(t)
	var harness awstypes.Harness
	rName := testAccRandomHarnessName(t)
	resourceName := "aws_bedrockagentcore_harness.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckHarness(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHarnessDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHarnessConfig_Memory_managedMemoryConfiguration_options(rName, 7, `["SEMANTIC"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory").AtSliceIndex(0).AtMapKey("managed_memory_configuration").AtSliceIndex(0), knownvalue.ObjectExact(map[string]knownvalue.Check{
						names.AttrARN:           tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`memory/`+rName+`-[a-zA-Z0-9]{10}`)),
						"encryption_key_arn":    knownvalue.Null(),
						"event_expiry_duration": knownvalue.Int32Exact(7),
						"strategies": knownvalue.SetExact([]knownvalue.Check{
							knownvalue.StringExact("SEMANTIC"),
						}),
					})),
				},
			},
			{
				Config: testAccHarnessConfig_Memory_managedMemoryConfiguration_options(rName, 14, `["SEMANTIC", "SUMMARIZATION"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory").AtSliceIndex(0).AtMapKey("managed_memory_configuration").AtSliceIndex(0), knownvalue.ObjectExact(map[string]knownvalue.Check{
						names.AttrARN:           tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`memory/`+rName+`-[a-zA-Z0-9]{10}`)),
						"encryption_key_arn":    knownvalue.Null(),
						"event_expiry_duration": knownvalue.Int32Exact(14),
						"strategies": knownvalue.SetExact([]knownvalue.Check{
							knownvalue.StringExact("SEMANTIC"),
							knownvalue.StringExact("SUMMARIZATION"),
						}),
					})),
				},
			},
			{
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "harness_id"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "harness_id",
				ImportStateVerifyIgnore:              []string{names.AttrEnvironment},
			},
		},
	})
}

func TestAccBedrockAgentCoreHarness_Memory_managedMemoryConfiguration_encryptionKey(t *testing.T) {
	ctx := acctest.Context(t)
	var harness awstypes.Harness
	rName := testAccRandomHarnessName(t)
	resourceName := "aws_bedrockagentcore_harness.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckHarness(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHarnessDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHarnessConfig_Memory_managedMemoryConfiguration_encryptionKey(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory").AtSliceIndex(0).AtMapKey("managed_memory_configuration").AtSliceIndex(0), knownvalue.ObjectExact(map[string]knownvalue.Check{
						names.AttrARN:           tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`memory/`+rName+`-[a-zA-Z0-9]{10}`)),
						"encryption_key_arn":    knownvalue.NotNull(),
						"event_expiry_duration": knownvalue.Int32Exact(30),
						"strategies": knownvalue.SetExact([]knownvalue.Check{
							knownvalue.StringExact("SEMANTIC"),
							knownvalue.StringExact("SUMMARIZATION"),
						}),
					})),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("memory").AtSliceIndex(0).AtMapKey("managed_memory_configuration").AtSliceIndex(0).AtMapKey("encryption_key_arn"), "aws_kms_key.test", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
			{
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "harness_id"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "harness_id",
				ImportStateVerifyIgnore:              []string{names.AttrEnvironment},
			},
		},
	})
}

func TestAccBedrockAgentCoreHarness_Memory_disabled(t *testing.T) {
	ctx := acctest.Context(t)
	var harness awstypes.Harness
	rName := testAccRandomHarnessName(t)
	resourceName := "aws_bedrockagentcore_harness.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckHarness(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHarnessDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHarnessConfig_Memory_disabled(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory").AtSliceIndex(0), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"agentcore_memory_configuration": knownvalue.ListSizeExact(0),
						"disabled": knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{}),
						}),
						"managed_memory_configuration": knownvalue.ListSizeExact(0),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory_actual").AtSliceIndex(0), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"agentcore_memory_configuration": knownvalue.Null(),
						"disabled": knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{}),
						}),
						"managed_memory_configuration": knownvalue.Null(),
					})),
				},
			},
			{
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "harness_id"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "harness_id",
				ImportStateVerifyIgnore:              []string{names.AttrEnvironment},
			},
		},
	})
}

func TestAccBedrockAgentCoreHarness_Memory_changeType(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		from memoryConfigType
		to   memoryConfigType
	}{
		{memoryNone, memoryAgentCore},
		{memoryNone, memoryDisabled},
		{memoryNone, memoryManagedEmpty},
		{memoryNone, memoryManaged},

		{memoryAgentCore, memoryDisabled},
		{memoryAgentCore, memoryManagedEmpty},
		{memoryAgentCore, memoryManaged},
		// {memoryAgentCore, memoryNone},

		{memoryDisabled, memoryAgentCore},
		{memoryDisabled, memoryManagedEmpty},
		{memoryDisabled, memoryManaged},
		{memoryDisabled, memoryNone},

		{memoryManagedEmpty, memoryAgentCore},
		{memoryManagedEmpty, memoryDisabled},
		{memoryManagedEmpty, memoryManaged},
		{memoryManagedEmpty, memoryNone},
	}

	for _, tc := range testcases { //nolint:paralleltest // false positive
		t.Run(fmt.Sprintf("%s_to_%s", tc.from, tc.to), func(t *testing.T) {
			ctx := acctest.Context(t)
			var harness awstypes.Harness
			rName := testAccRandomHarnessName(t)
			resourceName := "aws_bedrockagentcore_harness.test"

			fromConfig := testAccHarnessConfig_Memory_byType(t, rName, tc.from)
			toConfig := testAccHarnessConfig_Memory_byType(t, rName, tc.to)

			acctest.ParallelTest(ctx, t, resource.TestCase{
				PreCheck: func() {
					acctest.PreCheck(ctx, t)
					acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
					testAccPreCheckHarness(ctx, t)
				},
				ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				CheckDestroy:             testAccCheckHarnessDestroy(ctx, t),
				Steps: []resource.TestStep{
					{
						Config: fromConfig,
						Check: resource.ComposeAggregateTestCheckFunc(
							testAccCheckHarnessExists(ctx, t, resourceName, &harness),
						),
						ConfigStateChecks: memoryConfigStateChecks(t, resourceName, tc.from),
					},
					{
						Config: toConfig,
						Check: resource.ComposeAggregateTestCheckFunc(
							testAccCheckHarnessExists(ctx, t, resourceName, &harness),
						),
						ConfigPlanChecks: resource.ConfigPlanChecks{
							PreApply: []plancheck.PlanCheck{
								plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
							},
						},
						ConfigStateChecks: memoryConfigStateChecks(t, resourceName, tc.to),
					},
				},
			})
		})
	}
}

func TestAccBedrockAgentCoreHarness_environmentArtifact(t *testing.T) {
	ctx := acctest.Context(t)
	var harness awstypes.Harness
	rName := testAccRandomHarnessName(t)
	resourceName := "aws_bedrockagentcore_harness.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckHarness(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHarnessDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHarnessConfig_environmentArtifact(rName, "2.0.20230404.0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccHarnessConfig_environmentArtifact(rName, "2.0.20230515.0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "harness_id"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "harness_id",
				ImportStateVerifyIgnore:              []string{names.AttrEnvironment, "memory"},
			},
		},
	})
}

func TestAccBedrockAgentCoreHarness_authorizerConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var harness awstypes.Harness
	rName := testAccRandomHarnessName(t)
	resourceName := "aws_bedrockagentcore_harness.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckHarness(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHarnessDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHarnessConfig_authorizerConfiguration(rName, "https://accounts.google.com/.well-known/openid-configuration", "weather", "sports", "client-999", "client-888", "openid", names.AttrEmail),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccHarnessConfig_authorizerConfiguration(rName, "https://login.microsoftonline.com/common/v2.0/.well-known/openid-configuration", "finance", "technology", "client-111", "client-222", "openid", names.AttrProfile),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "harness_id"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "harness_id",
				ImportStateVerifyIgnore:              []string{names.AttrEnvironment, "memory"},
			},
		},
	})
}

func TestAccBedrockAgentCoreHarness_Environment_Network_VPC(t *testing.T) {
	t.Skip("Tests with VPC network mode are failing due to a lingering ENI issue.")

	ctx := acctest.Context(t)
	var harness awstypes.Harness
	rName := testAccRandomHarnessName(t)
	resourceName := "aws_bedrockagentcore_harness.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckHarness(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHarnessDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHarnessConfig_environment_Network_VPC(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEnvironment), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"agentcore_runtime_environment": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"agent_runtime_arn":        tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`runtime/harness_`+rName+`-[a-zA-Z0-9]+`)),
									"agent_runtime_id":         knownvalue.StringRegexp(regexache.MustCompile(`^harness_` + rName + `-[a-zA-Z0-9]+$`)),
									"agent_runtime_name":       knownvalue.StringExact("harness_" + rName),
									"filesystem_configuration": knownvalue.ListSizeExact(0),
									"lifecycle_configuration": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectExact(map[string]knownvalue.Check{
											"idle_runtime_session_timeout": knownvalue.Int32Exact(900),
											"max_lifetime":                 knownvalue.Int32Exact(28800),
										}),
									}),
									names.AttrNetworkConfiguration: knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectExact(map[string]knownvalue.Check{
											"network_mode": knownvalue.StringExact("VPC"),
											"network_mode_config": knownvalue.ListExact([]knownvalue.Check{
												knownvalue.ObjectExact(map[string]knownvalue.Check{
													"require_service_s3_endpoint": knownvalue.Null(),
													names.AttrSecurityGroups:      knownvalue.SetSizeExact(1),
													names.AttrSubnets:             knownvalue.SetSizeExact(2),
												}),
											}),
										}),
									}),
								}),
							}),
						}),
					})),
					environmentActualVPCStateCheck(resourceName, rName),
				},
			},
			{
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "harness_id"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "harness_id",
				ImportStateVerifyIgnore:              []string{"memory"},
			},
		},
	})
}

func TestAccBedrockAgentCoreHarness_Environment_Network_Public(t *testing.T) {
	ctx := acctest.Context(t)
	var harness awstypes.Harness
	rName := testAccRandomHarnessName(t)
	resourceName := "aws_bedrockagentcore_harness.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckHarness(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHarnessDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHarnessConfig_environment_Network_Public(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEnvironment), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"agentcore_runtime_environment": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"agent_runtime_arn":        tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`runtime/harness_`+rName+`-[a-zA-Z0-9]+`)),
									"agent_runtime_id":         knownvalue.StringRegexp(regexache.MustCompile(`^harness_` + rName + `-[a-zA-Z0-9]+$`)),
									"agent_runtime_name":       knownvalue.StringExact("harness_" + rName),
									"filesystem_configuration": knownvalue.ListSizeExact(0),
									"lifecycle_configuration": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectExact(map[string]knownvalue.Check{
											"idle_runtime_session_timeout": knownvalue.Int32Exact(900),
											"max_lifetime":                 knownvalue.Int32Exact(28800),
										}),
									}),
									names.AttrNetworkConfiguration: knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectExact(map[string]knownvalue.Check{
											"network_mode":        knownvalue.StringExact("PUBLIC"),
											"network_mode_config": knownvalue.ListSizeExact(0),
										}),
									}),
								}),
							}),
						}),
					})),
					environmentActualPublicStateCheck(resourceName, rName),
				},
			},
			{
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "harness_id"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "harness_id",
				ImportStateVerifyIgnore:              []string{"memory"},
			},
		},
	})
}

func TestAccBedrockAgentCoreHarness_Environment_lifecycleConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var harness awstypes.Harness
	rName := testAccRandomHarnessName(t)
	resourceName := "aws_bedrockagentcore_harness.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckHarness(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHarnessDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHarnessConfig_environment_lifecycleConfiguration(rName, 600, 14400),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEnvironment).AtSliceIndex(0).AtMapKey("agentcore_runtime_environment").AtSliceIndex(0).AtMapKey("lifecycle_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"idle_runtime_session_timeout": knownvalue.Int32Exact(600),
							"max_lifetime":                 knownvalue.Int32Exact(14400),
						}),
					})),
				},
			},
			{
				Config: testAccHarnessConfig_environment_lifecycleConfiguration(rName, 1200, 21600),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEnvironment).AtSliceIndex(0).AtMapKey("agentcore_runtime_environment").AtSliceIndex(0).AtMapKey("lifecycle_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"idle_runtime_session_timeout": knownvalue.Int32Exact(1200),
							"max_lifetime":                 knownvalue.Int32Exact(21600),
						}),
					})),
				},
			},
			{
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "harness_id"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "harness_id",
				ImportStateVerifyIgnore:              []string{"memory"},
			},
		},
	})
}

func TestAccBedrockAgentCoreHarness_Environment_FilesystemConfiguration_sessionStorage(t *testing.T) {
	t.Skip("Tests with VPC network mode are failing due to a lingering ENI issue.")

	ctx := acctest.Context(t)
	var harness awstypes.Harness
	rName := testAccRandomHarnessName(t)
	resourceName := "aws_bedrockagentcore_harness.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckHarness(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHarnessDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHarnessConfig_environment_FilesystemConfiguration_sessionStorage(rName, "/mnt/storage"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEnvironment).AtSliceIndex(0).AtMapKey("agentcore_runtime_environment").AtSliceIndex(0).AtMapKey("filesystem_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"efs_access_point":      knownvalue.ListSizeExact(0),
							"s3_files_access_point": knownvalue.ListSizeExact(0),
							"session_storage": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"mount_path": knownvalue.StringExact("/mnt/storage"),
								}),
							}),
						}),
					})),
				},
			},
			{
				Config: testAccHarnessConfig_environment_FilesystemConfiguration_sessionStorage(rName, "/mnt/data"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEnvironment).AtSliceIndex(0).AtMapKey("agentcore_runtime_environment").AtSliceIndex(0).AtMapKey("filesystem_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"efs_access_point":      knownvalue.ListSizeExact(0),
							"s3_files_access_point": knownvalue.ListSizeExact(0),
							"session_storage": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"mount_path": knownvalue.StringExact("/mnt/data"),
								}),
							}),
						}),
					})),
				},
			},
			{
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "harness_id"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "harness_id",
				ImportStateVerifyIgnore:              []string{"memory"},
			},
		},
	})
}

func TestAccBedrockAgentCoreHarness_Environment_addEnvironment(t *testing.T) {
	ctx := acctest.Context(t)
	var harness awstypes.Harness
	rName := testAccRandomHarnessName(t)
	resourceName := "aws_bedrockagentcore_harness.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckHarness(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHarnessDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHarnessConfig_environment_none(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEnvironment), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("environment_actual").AtSliceIndex(0), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"agentcore_runtime_environment": knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"agent_runtime_arn":        tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`runtime/harness_`+rName+`-[a-zA-Z0-9]+`)),
								"agent_runtime_id":         knownvalue.StringRegexp(regexache.MustCompile(`^harness_` + rName + `-[a-zA-Z0-9]+$`)),
								"agent_runtime_name":       knownvalue.StringExact("harness_" + rName),
								"filesystem_configuration": knownvalue.Null(),
								"lifecycle_configuration": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.ObjectExact(map[string]knownvalue.Check{
										"idle_runtime_session_timeout": knownvalue.Int32Exact(900),
										"max_lifetime":                 knownvalue.Int32Exact(28800),
									}),
								}),
								names.AttrNetworkConfiguration: knownvalue.ListExact([]knownvalue.Check{
									knownvalue.ObjectExact(map[string]knownvalue.Check{
										"network_mode":        knownvalue.StringExact("PUBLIC"),
										"network_mode_config": knownvalue.Null(),
									}),
								}),
							}),
						}),
					})),
				},
			},
			{
				Config: testAccHarnessConfig_environment_lifecycleConfiguration(rName, 600, 14400),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEnvironment).AtSliceIndex(0), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"agentcore_runtime_environment": knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"agent_runtime_arn":        tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`runtime/harness_`+rName+`-[a-zA-Z0-9]+`)),
								"agent_runtime_id":         knownvalue.StringRegexp(regexache.MustCompile(`^harness_` + rName + `-[a-zA-Z0-9]+$`)),
								"agent_runtime_name":       knownvalue.StringExact("harness_" + rName),
								"filesystem_configuration": knownvalue.ListSizeExact(0),
								"lifecycle_configuration": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.ObjectExact(map[string]knownvalue.Check{
										"idle_runtime_session_timeout": knownvalue.Int32Exact(600),
										"max_lifetime":                 knownvalue.Int32Exact(14400),
									}),
								}),
								names.AttrNetworkConfiguration: knownvalue.ListExact([]knownvalue.Check{
									knownvalue.ObjectExact(map[string]knownvalue.Check{
										"network_mode":        knownvalue.StringExact("PUBLIC"),
										"network_mode_config": knownvalue.ListSizeExact(0),
									}),
								}),
							}),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("environment_actual").AtSliceIndex(0), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"agentcore_runtime_environment": knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"agent_runtime_arn":        tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`runtime/harness_`+rName+`-[a-zA-Z0-9]+`)),
								"agent_runtime_id":         knownvalue.StringRegexp(regexache.MustCompile(`^harness_` + rName + `-[a-zA-Z0-9]+$`)),
								"agent_runtime_name":       knownvalue.StringExact("harness_" + rName),
								"filesystem_configuration": knownvalue.Null(),
								"lifecycle_configuration": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.ObjectExact(map[string]knownvalue.Check{
										"idle_runtime_session_timeout": knownvalue.Int32Exact(600),
										"max_lifetime":                 knownvalue.Int32Exact(14400),
									}),
								}),
								names.AttrNetworkConfiguration: knownvalue.ListExact([]knownvalue.Check{
									knownvalue.ObjectExact(map[string]knownvalue.Check{
										"network_mode":        knownvalue.StringExact("PUBLIC"),
										"network_mode_config": knownvalue.Null(),
									}),
								}),
							}),
						}),
					})),
				},
			},
			{
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "harness_id"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "harness_id",
				ImportStateVerifyIgnore:              []string{"memory"},
			},
		},
	})
}

func TestAccBedrockAgentCoreHarness_Environment_removeEnvironment(t *testing.T) {
	ctx := acctest.Context(t)
	var harness awstypes.Harness
	rName := testAccRandomHarnessName(t)
	resourceName := "aws_bedrockagentcore_harness.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckHarness(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHarnessDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHarnessConfig_environment_lifecycleConfiguration(rName, 600, 14400),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEnvironment).AtSliceIndex(0), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"agentcore_runtime_environment": knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"agent_runtime_arn":        tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`runtime/harness_`+rName+`-[a-zA-Z0-9]+`)),
								"agent_runtime_id":         knownvalue.StringRegexp(regexache.MustCompile(`^harness_` + rName + `-[a-zA-Z0-9]+$`)),
								"agent_runtime_name":       knownvalue.StringExact("harness_" + rName),
								"filesystem_configuration": knownvalue.ListSizeExact(0),
								"lifecycle_configuration": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.ObjectExact(map[string]knownvalue.Check{
										"idle_runtime_session_timeout": knownvalue.Int32Exact(600),
										"max_lifetime":                 knownvalue.Int32Exact(14400),
									}),
								}),
								names.AttrNetworkConfiguration: knownvalue.ListExact([]knownvalue.Check{
									knownvalue.ObjectExact(map[string]knownvalue.Check{
										"network_mode":        knownvalue.StringExact("PUBLIC"),
										"network_mode_config": knownvalue.ListSizeExact(0),
									}),
								}),
							}),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("environment_actual").AtSliceIndex(0), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"agentcore_runtime_environment": knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"agent_runtime_arn":        tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`runtime/harness_`+rName+`-[a-zA-Z0-9]+`)),
								"agent_runtime_id":         knownvalue.StringRegexp(regexache.MustCompile(`^harness_` + rName + `-[a-zA-Z0-9]+$`)),
								"agent_runtime_name":       knownvalue.StringExact("harness_" + rName),
								"filesystem_configuration": knownvalue.Null(),
								"lifecycle_configuration": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.ObjectExact(map[string]knownvalue.Check{
										"idle_runtime_session_timeout": knownvalue.Int32Exact(600),
										"max_lifetime":                 knownvalue.Int32Exact(14400),
									}),
								}),
								names.AttrNetworkConfiguration: knownvalue.ListExact([]knownvalue.Check{
									knownvalue.ObjectExact(map[string]knownvalue.Check{
										"network_mode":        knownvalue.StringExact("PUBLIC"),
										"network_mode_config": knownvalue.Null(),
									}),
								}),
							}),
						}),
					})),
				},
			},
			{
				Config: testAccHarnessConfig_environment_none(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEnvironment), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("environment_actual").AtSliceIndex(0), knownvalue.ObjectExact(map[string]knownvalue.Check{
						"agentcore_runtime_environment": knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"agent_runtime_arn":        tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`runtime/harness_`+rName+`-[a-zA-Z0-9]+`)),
								"agent_runtime_id":         knownvalue.StringRegexp(regexache.MustCompile(`^harness_` + rName + `-[a-zA-Z0-9]+$`)),
								"agent_runtime_name":       knownvalue.StringExact("harness_" + rName),
								"filesystem_configuration": knownvalue.Null(),
								"lifecycle_configuration": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.ObjectExact(map[string]knownvalue.Check{
										"idle_runtime_session_timeout": knownvalue.Int32Exact(600),
										"max_lifetime":                 knownvalue.Int32Exact(14400),
									}),
								}),
								names.AttrNetworkConfiguration: knownvalue.ListExact([]knownvalue.Check{
									knownvalue.ObjectExact(map[string]knownvalue.Check{
										"network_mode":        knownvalue.StringExact("PUBLIC"),
										"network_mode_config": knownvalue.Null(),
									}),
								}),
							}),
						}),
					})),
				},
			},
			{
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "harness_id"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "harness_id",
				ImportStateVerifyIgnore:              []string{names.AttrEnvironment, "memory"},
			},
		},
	})
}

func TestAccBedrockAgentCoreHarness_Environment_FilesystemConfiguration_s3FilesAccessPoint(t *testing.T) {
	t.Skip("Tests with VPC network mode are failing due to a lingering ENI issue.")

	ctx := acctest.Context(t)
	var harness awstypes.Harness
	rName := testAccRandomHarnessName(t)
	resourceName := "aws_bedrockagentcore_harness.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckHarness(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHarnessDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHarnessConfig_environment_FilesystemConfiguration_s3FilesAccessPoint(rName, "/mnt/s3data"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEnvironment).AtSliceIndex(0).AtMapKey("agentcore_runtime_environment").AtSliceIndex(0).AtMapKey("filesystem_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"efs_access_point": knownvalue.ListSizeExact(0),
							"s3_files_access_point": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"access_point_arn": knownvalue.NotNull(),
									"mount_path":       knownvalue.StringExact("/mnt/s3data"),
								}),
							}),
							"session_storage": knownvalue.ListSizeExact(0),
						}),
					})),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrEnvironment).AtSliceIndex(0).AtMapKey("agentcore_runtime_environment").AtSliceIndex(0).AtMapKey("filesystem_configuration").AtSliceIndex(0).AtMapKey("s3_files_access_point").AtSliceIndex(0).AtMapKey("access_point_arn"), "aws_s3files_access_point.test", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
			{
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "harness_id"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "harness_id",
				ImportStateVerifyIgnore:              []string{"memory"},
			},
		},
	})
}

func TestAccBedrockAgentCoreHarness_Environment_FilesystemConfiguration_efsAccessPoint(t *testing.T) {
	t.Skip("Tests with VPC network mode are failing due to a lingering ENI issue.")

	ctx := acctest.Context(t)
	var harness awstypes.Harness
	rName := testAccRandomHarnessName(t)
	resourceName := "aws_bedrockagentcore_harness.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckHarness(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHarnessDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHarnessConfig_environment_FilesystemConfiguration_efsAccessPoint(rName, "/mnt/efsdata"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEnvironment).AtSliceIndex(0).AtMapKey("agentcore_runtime_environment").AtSliceIndex(0).AtMapKey("filesystem_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"efs_access_point": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"access_point_arn": knownvalue.NotNull(),
									"mount_path":       knownvalue.StringExact("/mnt/efsdata"),
								}),
							}),
							"s3_files_access_point": knownvalue.ListSizeExact(0),
							"session_storage":       knownvalue.ListSizeExact(0),
						}),
					})),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrEnvironment).AtSliceIndex(0).AtMapKey("agentcore_runtime_environment").AtSliceIndex(0).AtMapKey("filesystem_configuration").AtSliceIndex(0).AtMapKey("efs_access_point").AtSliceIndex(0).AtMapKey("access_point_arn"), "aws_efs_access_point.test", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
			{
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "harness_id"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "harness_id",
				ImportStateVerifyIgnore:              []string{"memory"},
			},
		},
	})
}

func TestAccBedrockAgentCoreHarness_Environment_FilesystemConfiguration_multiple(t *testing.T) {
	t.Skip("Tests with VPC network mode are failing due to a lingering ENI issue.")

	ctx := acctest.Context(t)
	var harness awstypes.Harness
	rName := testAccRandomHarnessName(t)
	resourceName := "aws_bedrockagentcore_harness.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckHarness(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHarnessDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHarnessConfig_environment_FilesystemConfiguration_multiple(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEnvironment).AtSliceIndex(0).AtMapKey("agentcore_runtime_environment").AtSliceIndex(0).AtMapKey("filesystem_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"efs_access_point":      knownvalue.ListSizeExact(0),
							"s3_files_access_point": knownvalue.ListSizeExact(0),
							"session_storage": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"mount_path": knownvalue.StringExact("/mnt/session"),
								}),
							}),
						}),
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"efs_access_point": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"access_point_arn": knownvalue.NotNull(),
									"mount_path":       knownvalue.StringExact("/mnt/efsdata"),
								}),
							}),
							"s3_files_access_point": knownvalue.ListSizeExact(0),
							"session_storage":       knownvalue.ListSizeExact(0),
						}),
					})),
					statecheck.CompareValuePairs(
						resourceName, tfjsonpath.New(names.AttrEnvironment).AtSliceIndex(0).AtMapKey("agentcore_runtime_environment").AtSliceIndex(0).AtMapKey("filesystem_configuration").AtSliceIndex(1).AtMapKey("efs_access_point").AtSliceIndex(0).AtMapKey("access_point_arn"),
						"aws_efs_access_point.test", tfjsonpath.New(names.AttrARN),
						compare.ValuesSame()),
				},
			},
			{
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "harness_id"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "harness_id",
				ImportStateVerifyIgnore:              []string{"memory"},
			},
		},
	})
}

func TestAccBedrockAgentCoreHarness_Environment_FilesystemConfiguration_addFilesystem(t *testing.T) {
	t.Skip("Tests with VPC network mode are failing due to a lingering ENI issue.")

	ctx := acctest.Context(t)
	var harness awstypes.Harness
	rName := testAccRandomHarnessName(t)
	resourceName := "aws_bedrockagentcore_harness.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckHarness(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHarnessDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHarnessConfig_environment_FilesystemConfiguration_vpcNoFilesystem(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEnvironment).AtSliceIndex(0).AtMapKey("agentcore_runtime_environment").AtSliceIndex(0).AtMapKey("filesystem_configuration"), knownvalue.ListSizeExact(0)),
				},
			},
			{
				Config: testAccHarnessConfig_environment_FilesystemConfiguration_sessionStorage(rName, "/mnt/storage"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEnvironment).AtSliceIndex(0).AtMapKey("agentcore_runtime_environment").AtSliceIndex(0).AtMapKey("filesystem_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"efs_access_point":      knownvalue.ListSizeExact(0),
							"s3_files_access_point": knownvalue.ListSizeExact(0),
							"session_storage": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"mount_path": knownvalue.StringExact("/mnt/storage"),
								}),
							}),
						}),
					})),
				},
			},
			{
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "harness_id"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "harness_id",
				ImportStateVerifyIgnore:              []string{"memory"},
			},
		},
	})
}

func TestAccBedrockAgentCoreHarness_Environment_FilesystemConfiguration_removeFilesystem(t *testing.T) {
	t.Skip("Tests with VPC network mode are failing due to a lingering ENI issue.")

	ctx := acctest.Context(t)
	var harness awstypes.Harness
	rName := testAccRandomHarnessName(t)
	resourceName := "aws_bedrockagentcore_harness.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckHarness(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHarnessDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHarnessConfig_environment_FilesystemConfiguration_sessionStorage(rName, "/mnt/storage"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEnvironment).AtSliceIndex(0).AtMapKey("agentcore_runtime_environment").AtSliceIndex(0).AtMapKey("filesystem_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"efs_access_point":      knownvalue.ListSizeExact(0),
							"s3_files_access_point": knownvalue.ListSizeExact(0),
							"session_storage": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"mount_path": knownvalue.StringExact("/mnt/storage"),
								}),
							}),
						}),
					})),
				},
			},
			{
				Config: testAccHarnessConfig_environment_FilesystemConfiguration_vpcNoFilesystem(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEnvironment).AtSliceIndex(0).AtMapKey("agentcore_runtime_environment").AtSliceIndex(0).AtMapKey("filesystem_configuration"), knownvalue.ListSizeExact(0)),
				},
			},
			{
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "harness_id"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "harness_id",
				ImportStateVerifyIgnore:              []string{"memory"},
			},
		},
	})
}

func TestAccBedrockAgentCoreHarness_Environment_Network_updatePublicToVPC(t *testing.T) {
	t.Skip("Tests with VPC network mode are failing due to a lingering ENI issue.")

	ctx := acctest.Context(t)
	var harness awstypes.Harness
	rName := testAccRandomHarnessName(t)
	resourceName := "aws_bedrockagentcore_harness.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckHarness(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHarnessDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHarnessConfig_environment_Network_Public(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEnvironment), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"agentcore_runtime_environment": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"agent_runtime_arn":        tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`runtime/harness_`+rName+`-[a-zA-Z0-9]+`)),
									"agent_runtime_id":         knownvalue.StringRegexp(regexache.MustCompile(`^harness_` + rName + `-[a-zA-Z0-9]+$`)),
									"agent_runtime_name":       knownvalue.StringExact("harness_" + rName),
									"filesystem_configuration": knownvalue.ListSizeExact(0),
									"lifecycle_configuration": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectExact(map[string]knownvalue.Check{
											"idle_runtime_session_timeout": knownvalue.Int32Exact(900),
											"max_lifetime":                 knownvalue.Int32Exact(28800),
										}),
									}),
									names.AttrNetworkConfiguration: knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectExact(map[string]knownvalue.Check{
											"network_mode":        knownvalue.StringExact("PUBLIC"),
											"network_mode_config": knownvalue.ListSizeExact(0),
										}),
									}),
								}),
							}),
						}),
					})),
					environmentActualPublicStateCheck(resourceName, rName),
				},
			},
			{
				Config: testAccHarnessConfig_environment_Network_VPC(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEnvironment), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"agentcore_runtime_environment": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"agent_runtime_arn":        tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`runtime/harness_`+rName+`-[a-zA-Z0-9]+`)),
									"agent_runtime_id":         knownvalue.StringRegexp(regexache.MustCompile(`^harness_` + rName + `-[a-zA-Z0-9]+$`)),
									"agent_runtime_name":       knownvalue.StringExact("harness_" + rName),
									"filesystem_configuration": knownvalue.ListSizeExact(0),
									"lifecycle_configuration": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectExact(map[string]knownvalue.Check{
											"idle_runtime_session_timeout": knownvalue.Int32Exact(900),
											"max_lifetime":                 knownvalue.Int32Exact(28800),
										}),
									}),
									names.AttrNetworkConfiguration: knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectExact(map[string]knownvalue.Check{
											"network_mode": knownvalue.StringExact("VPC"),
											"network_mode_config": knownvalue.ListExact([]knownvalue.Check{
												knownvalue.ObjectExact(map[string]knownvalue.Check{
													"require_service_s3_endpoint": knownvalue.Null(),
													names.AttrSecurityGroups:      knownvalue.SetSizeExact(1),
													names.AttrSubnets:             knownvalue.SetSizeExact(2),
												}),
											}),
										}),
									}),
								}),
							}),
						}),
					})),
					environmentActualVPCStateCheck(resourceName, rName),
				},
			},
			{
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "harness_id"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "harness_id",
				ImportStateVerifyIgnore:              []string{"memory"},
			},
		},
	})
}

func TestAccBedrockAgentCoreHarness_Environment_Network_updateVPCToPublic(t *testing.T) {
	t.Skip("Tests with VPC network mode are failing due to a lingering ENI issue.")

	ctx := acctest.Context(t)
	var harness awstypes.Harness
	rName := testAccRandomHarnessName(t)
	resourceName := "aws_bedrockagentcore_harness.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckHarness(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHarnessDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHarnessConfig_environment_Network_VPC(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEnvironment), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"agentcore_runtime_environment": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"agent_runtime_arn":        tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`runtime/harness_`+rName+`-[a-zA-Z0-9]+`)),
									"agent_runtime_id":         knownvalue.StringRegexp(regexache.MustCompile(`^harness_` + rName + `-[a-zA-Z0-9]+$`)),
									"agent_runtime_name":       knownvalue.StringExact("harness_" + rName),
									"filesystem_configuration": knownvalue.ListSizeExact(0),
									"lifecycle_configuration": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectExact(map[string]knownvalue.Check{
											"idle_runtime_session_timeout": knownvalue.Int32Exact(900),
											"max_lifetime":                 knownvalue.Int32Exact(28800),
										}),
									}),
									names.AttrNetworkConfiguration: knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectExact(map[string]knownvalue.Check{
											"network_mode": knownvalue.StringExact("VPC"),
											"network_mode_config": knownvalue.ListExact([]knownvalue.Check{
												knownvalue.ObjectExact(map[string]knownvalue.Check{
													"require_service_s3_endpoint": knownvalue.Null(),
													names.AttrSecurityGroups:      knownvalue.SetSizeExact(1),
													names.AttrSubnets:             knownvalue.SetSizeExact(2),
												}),
											}),
										}),
									}),
								}),
							}),
						}),
					})),
					environmentActualVPCStateCheck(resourceName, rName),
				},
			},
			{
				Config: testAccHarnessConfig_environment_Network_VPCToPublic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEnvironment), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"agentcore_runtime_environment": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"agent_runtime_arn":        tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`runtime/harness_`+rName+`-[a-zA-Z0-9]+`)),
									"agent_runtime_id":         knownvalue.StringRegexp(regexache.MustCompile(`^harness_` + rName + `-[a-zA-Z0-9]+$`)),
									"agent_runtime_name":       knownvalue.StringExact("harness_" + rName),
									"filesystem_configuration": knownvalue.ListSizeExact(0),
									"lifecycle_configuration": knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectExact(map[string]knownvalue.Check{
											"idle_runtime_session_timeout": knownvalue.Int32Exact(900),
											"max_lifetime":                 knownvalue.Int32Exact(28800),
										}),
									}),
									names.AttrNetworkConfiguration: knownvalue.ListExact([]knownvalue.Check{
										knownvalue.ObjectExact(map[string]knownvalue.Check{
											"network_mode":        knownvalue.StringExact("PUBLIC"),
											"network_mode_config": knownvalue.ListSizeExact(0),
										}),
									}),
								}),
							}),
						}),
					})),
					environmentActualPublicStateCheck(resourceName, rName),
				},
			},
			{
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "harness_id"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "harness_id",
				ImportStateVerifyIgnore:              []string{"memory"},
			},
		},
	})
}

func TestAccBedrockAgentCoreHarness_skill_git(t *testing.T) {
	ctx := acctest.Context(t)
	var harness awstypes.Harness
	rName := testAccRandomHarnessName(t)
	resourceName := "aws_bedrockagentcore_harness.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.BedrockEndpointID)
			testAccPreCheckHarness(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckHarnessDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccHarnessConfig_skillGit(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
					resource.TestCheckResourceAttr(resourceName, "skill.0.git.0.url", "https://github.com/example/skill.git"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

// Helper functions.

func testAccCheckHarnessDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_bedrockagentcore_harness" {
				continue
			}

			_, err := tfbedrockagentcore.FindHarnessByID(ctx, conn, rs.Primary.Attributes["harness_id"])
			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Bedrock Agent Core Harness %s still exists", rs.Primary.Attributes["harness_id"])
		}

		return nil
	}
}

func testAccCheckHarnessExists(ctx context.Context, t *testing.T, n string, v *awstypes.Harness) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

		resp, err := tfbedrockagentcore.FindHarnessByID(ctx, conn, rs.Primary.Attributes["harness_id"])
		if err != nil {
			return err
		}

		*v = *resp

		return nil
	}
}

func testAccPreCheckHarness(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).BedrockAgentCoreClient(ctx)

	input := bedrockagentcorecontrol.ListHarnessesInput{}

	_, err := conn.ListHarnesses(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func environmentActualPublicStateCheck(resourceName, rName string) statecheck.StateCheck {
	return statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("environment_actual").AtSliceIndex(0), knownvalue.ObjectExact(map[string]knownvalue.Check{
		"agentcore_runtime_environment": knownvalue.ListExact([]knownvalue.Check{
			knownvalue.ObjectExact(map[string]knownvalue.Check{
				"agent_runtime_arn":        tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`runtime/harness_`+rName+`-[a-zA-Z0-9]+`)),
				"agent_runtime_id":         knownvalue.StringRegexp(regexache.MustCompile(`^harness_` + rName + `-[a-zA-Z0-9]+$`)),
				"agent_runtime_name":       knownvalue.StringExact("harness_" + rName),
				"filesystem_configuration": knownvalue.Null(),
				"lifecycle_configuration": knownvalue.ListExact([]knownvalue.Check{
					knownvalue.ObjectExact(map[string]knownvalue.Check{
						"idle_runtime_session_timeout": knownvalue.Int32Exact(900),
						"max_lifetime":                 knownvalue.Int32Exact(28800),
					}),
				}),
				names.AttrNetworkConfiguration: knownvalue.ListExact([]knownvalue.Check{
					knownvalue.ObjectExact(map[string]knownvalue.Check{
						"network_mode":        knownvalue.StringExact("PUBLIC"),
						"network_mode_config": knownvalue.Null(),
					}),
				}),
			}),
		}),
	}))
}

func environmentActualVPCStateCheck(resourceName, rName string) statecheck.StateCheck {
	return statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("environment_actual").AtSliceIndex(0), knownvalue.ObjectExact(map[string]knownvalue.Check{
		"agentcore_runtime_environment": knownvalue.ListExact([]knownvalue.Check{
			knownvalue.ObjectExact(map[string]knownvalue.Check{
				"agent_runtime_arn":        tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`runtime/harness_`+rName+`-[a-zA-Z0-9]+`)),
				"agent_runtime_id":         knownvalue.StringRegexp(regexache.MustCompile(`^harness_` + rName + `-[a-zA-Z0-9]+$`)),
				"agent_runtime_name":       knownvalue.StringExact("harness_" + rName),
				"filesystem_configuration": knownvalue.Null(),
				"lifecycle_configuration": knownvalue.ListExact([]knownvalue.Check{
					knownvalue.ObjectExact(map[string]knownvalue.Check{
						"idle_runtime_session_timeout": knownvalue.Int32Exact(900),
						"max_lifetime":                 knownvalue.Int32Exact(28800),
					}),
				}),
				names.AttrNetworkConfiguration: knownvalue.ListExact([]knownvalue.Check{
					knownvalue.ObjectExact(map[string]knownvalue.Check{
						"network_mode": knownvalue.StringExact("VPC"),
						"network_mode_config": knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"require_service_s3_endpoint": knownvalue.Null(),
								names.AttrSecurityGroups:      knownvalue.SetSizeExact(1),
								names.AttrSubnets:             knownvalue.SetSizeExact(2),
							}),
						}),
					}),
				}),
			}),
		}),
	}))
}

func memoryConfigStateChecks(t *testing.T, resourceName string, memType memoryConfigType) []statecheck.StateCheck {
	t.Helper()
	memoryPath := tfjsonpath.New("memory")
	memoryActualPath := tfjsonpath.New("memory_actual")
	switch memType {
	case memoryNone:
		return []statecheck.StateCheck{
			statecheck.ExpectKnownValue(resourceName, memoryPath, knownvalue.ListSizeExact(0)),
		}
	case memoryAgentCore:
		return []statecheck.StateCheck{
			statecheck.ExpectKnownValue(resourceName, memoryPath.AtSliceIndex(0), knownvalue.ObjectExact(map[string]knownvalue.Check{
				"agentcore_memory_configuration": knownvalue.ListExact([]knownvalue.Check{
					knownvalue.ObjectExact(map[string]knownvalue.Check{
						names.AttrARN:      knownvalue.NotNull(),
						"actor_id":         knownvalue.Null(),
						"messages_count":   knownvalue.Null(),
						"retrieval_config": knownvalue.ListSizeExact(0),
					}),
				}),
				"disabled":                     knownvalue.ListSizeExact(0),
				"managed_memory_configuration": knownvalue.ListSizeExact(0),
			})),
			statecheck.ExpectKnownValue(resourceName, memoryActualPath.AtSliceIndex(0), knownvalue.ObjectExact(map[string]knownvalue.Check{
				"agentcore_memory_configuration": knownvalue.ListExact([]knownvalue.Check{
					knownvalue.ObjectExact(map[string]knownvalue.Check{
						names.AttrARN:      knownvalue.NotNull(),
						"actor_id":         knownvalue.Null(),
						"messages_count":   knownvalue.Null(),
						"retrieval_config": knownvalue.Null(),
					}),
				}),
				"disabled":                     knownvalue.Null(),
				"managed_memory_configuration": knownvalue.Null(),
			})),
		}
	case memoryDisabled:
		return []statecheck.StateCheck{
			statecheck.ExpectKnownValue(resourceName, memoryPath.AtSliceIndex(0), knownvalue.ObjectExact(map[string]knownvalue.Check{
				"agentcore_memory_configuration": knownvalue.ListSizeExact(0),
				"disabled": knownvalue.ListExact([]knownvalue.Check{
					knownvalue.ObjectExact(map[string]knownvalue.Check{}),
				}),
				"managed_memory_configuration": knownvalue.ListSizeExact(0),
			})),
			statecheck.ExpectKnownValue(resourceName, memoryActualPath.AtSliceIndex(0), knownvalue.ObjectExact(map[string]knownvalue.Check{
				"agentcore_memory_configuration": knownvalue.Null(),
				"disabled": knownvalue.ListExact([]knownvalue.Check{
					knownvalue.ObjectExact(map[string]knownvalue.Check{}),
				}),
				"managed_memory_configuration": knownvalue.Null(),
			})),
		}
	case memoryManagedEmpty:
		return []statecheck.StateCheck{
			statecheck.ExpectKnownValue(resourceName, memoryPath.AtSliceIndex(0), knownvalue.ObjectExact(map[string]knownvalue.Check{
				"agentcore_memory_configuration": knownvalue.ListSizeExact(0),
				"disabled":                       knownvalue.ListSizeExact(0),
				"managed_memory_configuration": knownvalue.ListExact([]knownvalue.Check{
					knownvalue.ObjectExact(map[string]knownvalue.Check{
						names.AttrARN:           knownvalue.NotNull(),
						"encryption_key_arn":    knownvalue.Null(),
						"event_expiry_duration": knownvalue.Int32Exact(30),
						"strategies": knownvalue.SetExact([]knownvalue.Check{
							knownvalue.StringExact("SEMANTIC"),
							knownvalue.StringExact("SUMMARIZATION"),
						}),
					}),
				}),
			})),
			statecheck.ExpectKnownValue(resourceName, memoryActualPath.AtSliceIndex(0), knownvalue.ObjectExact(map[string]knownvalue.Check{
				"agentcore_memory_configuration": knownvalue.Null(),
				"disabled":                       knownvalue.Null(),
				"managed_memory_configuration": knownvalue.ListExact([]knownvalue.Check{
					knownvalue.ObjectExact(map[string]knownvalue.Check{
						names.AttrARN:           knownvalue.NotNull(),
						"encryption_key_arn":    knownvalue.Null(),
						"event_expiry_duration": knownvalue.Int32Exact(30),
						"strategies": knownvalue.SetExact([]knownvalue.Check{
							knownvalue.StringExact("SEMANTIC"),
							knownvalue.StringExact("SUMMARIZATION"),
						}),
					}),
				}),
			})),
		}
	case memoryManaged:
		return []statecheck.StateCheck{
			statecheck.ExpectKnownValue(resourceName, memoryPath.AtSliceIndex(0), knownvalue.ObjectExact(map[string]knownvalue.Check{
				"agentcore_memory_configuration": knownvalue.ListSizeExact(0),
				"disabled":                       knownvalue.ListSizeExact(0),
				"managed_memory_configuration": knownvalue.ListExact([]knownvalue.Check{
					knownvalue.ObjectExact(map[string]knownvalue.Check{
						names.AttrARN:           knownvalue.NotNull(),
						"encryption_key_arn":    knownvalue.Null(),
						"event_expiry_duration": knownvalue.Int32Exact(7),
						"strategies": knownvalue.SetExact([]knownvalue.Check{
							knownvalue.StringExact("SEMANTIC"),
						}),
					}),
				}),
			})),
			statecheck.ExpectKnownValue(resourceName, memoryActualPath.AtSliceIndex(0), knownvalue.ObjectExact(map[string]knownvalue.Check{
				"agentcore_memory_configuration": knownvalue.Null(),
				"disabled":                       knownvalue.Null(),
				"managed_memory_configuration": knownvalue.ListExact([]knownvalue.Check{
					knownvalue.ObjectExact(map[string]knownvalue.Check{
						names.AttrARN:           knownvalue.NotNull(),
						"encryption_key_arn":    knownvalue.Null(),
						"event_expiry_duration": knownvalue.Int32Exact(7),
						"strategies": knownvalue.SetExact([]knownvalue.Check{
							knownvalue.StringExact("SEMANTIC"),
						}),
					}),
				}),
			})),
		}
	default:
		t.Fatalf("unknown memory config type: %s", memType)
		return nil
	}
}

func importCheckSetContains(prefix, value string) resource.ImportStateCheckFunc {
	return func(is []*terraform.InstanceState) error {
		if len(is) != 1 {
			return fmt.Errorf("expected 1 instance state, got %d", len(is))
		}

		rs := is[0]
		for k, v := range rs.Attributes {
			if strings.HasPrefix(k, prefix+".") && k != prefix+".#" && v == value {
				return nil
			}
		}
		return fmt.Errorf("set %q does not contain value %q", prefix, value)
	}
}

// Config generators.

func testAccHarnessConfig_iamRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test.json
}

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

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": [
      "bedrock:InvokeModel",
      "bedrock:InvokeModelWithResponseStream"
    ],
    "Resource": "*"
  }
}
EOF
}
`, rName)
}

func testAccHarnessConfig_systemPrompt(rName, prompt string) string {
	return acctest.ConfigCompose(testAccHarnessConfig_iamRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_harness" "test" {
  harness_name       = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  model {
    bedrock_model_config {
      model_id = "anthropic.claude-sonnet-4-20250514"
    }
  }

  system_prompt {
    text = %[2]q
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName, prompt))
}

func testAccHarnessConfig_allowedTools(rName, tools string) string {
	return acctest.ConfigCompose(testAccHarnessConfig_iamRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_harness" "test" {
  harness_name       = %[1]q
  execution_role_arn = aws_iam_role.test.arn
  allowed_tools      = %[2]s

  model {
    bedrock_model_config {
      model_id = "anthropic.claude-sonnet-4-20250514"
    }
  }

  system_prompt {
    text = "You are a helpful assistant."
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName, tools))
}

func testAccHarnessConfig_limits(rName string, maxIter, maxTokens, timeout int) string {
	return acctest.ConfigCompose(testAccHarnessConfig_iamRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_harness" "test" {
  harness_name       = %[1]q
  execution_role_arn = aws_iam_role.test.arn
  max_iterations     = %[2]d
  max_tokens         = %[3]d
  timeout_seconds    = %[4]d

  model {
    bedrock_model_config {
      model_id = "anthropic.claude-sonnet-4-20250514"
    }
  }

  system_prompt {
    text = "You are a helpful assistant."
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName, maxIter, maxTokens, timeout))
}

func testAccHarnessConfig_bedrockModel(rName string) string {
	return acctest.ConfigCompose(testAccHarnessConfig_iamRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_harness" "test" {
  harness_name       = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  model {
    bedrock_model_config {
      model_id    = "anthropic.claude-sonnet-4-20250514"
      temperature = 0.7
      top_p       = 0.9
    }
  }

  system_prompt {
    text = "You are a helpful assistant."
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName))
}

func testAccHarnessConfig_liteLLMModel(rName string) string {
	return acctest.ConfigCompose(testAccHarnessConfig_iamRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_harness" "test" {
  harness_name       = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  model {
    litellm_model_config {
      model_id    = "anthropic/claude-sonnet-4-20250514"
      api_base    = "https://api.example.com/v1"
      temperature = 0.7
      top_p       = 0.9
    }
  }

  system_prompt {
    text = "You are a helpful assistant."
  }
}
`, rName))
}

func testAccHarnessConfig_truncationSlidingWindow(rName string, messagesCount int) string {
	return acctest.ConfigCompose(testAccHarnessConfig_iamRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_harness" "test" {
  harness_name       = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  model {
    bedrock_model_config {
      model_id = "anthropic.claude-sonnet-4-20250514"
    }
  }

  system_prompt {
    text = "You are a helpful assistant."
  }

  truncation {
    strategy = "sliding_window"

    config {
      sliding_window {
        messages_count = %[2]d
      }
    }
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName, messagesCount))
}

func testAccHarnessConfig_truncationSummarization(rName string) string {
	return acctest.ConfigCompose(testAccHarnessConfig_iamRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_harness" "test" {
  harness_name       = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  model {
    bedrock_model_config {
      model_id = "anthropic.claude-sonnet-4-20250514"
    }
  }

  system_prompt {
    text = "You are a helpful assistant."
  }

  truncation {
    strategy = "summarization"

    config {
      summarization {
        summary_ratio            = 0.5
        preserve_recent_messages = 5
      }
    }
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName))
}

func testAccHarnessConfig_toolInlineFunction(rName string) string {
	return acctest.ConfigCompose(testAccHarnessConfig_iamRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_harness" "test" {
  harness_name       = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  model {
    bedrock_model_config {
      model_id = "anthropic.claude-sonnet-4-20250514"
    }
  }

  system_prompt {
    text = "You are a helpful assistant."
  }

  tool {
    type = "inline_function"
    name = "get_weather"

    config {
      inline_function {
        description = "Gets the weather for a given location"
        input_schema = jsonencode({
          type = "object"
          properties = {
            location = {
              type        = "string"
              description = "The city and state"
            }
          }
          required = ["location"]
        })
      }
    }
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName))
}

func testAccHarnessConfig_environmentVariables(rName, key, value string) string {
	return acctest.ConfigCompose(testAccHarnessConfig_iamRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_harness" "test" {
  harness_name       = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  environment_variables = {
    %[2]q = %[3]q
  }

  model {
    bedrock_model_config {
      model_id = "anthropic.claude-sonnet-4-20250514"
    }
  }

  system_prompt {
    text = "You are a helpful assistant."
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName, key, value))
}

func testAccHarnessConfig_Memory_agentCoreMemoryConfiguration_basic(rName string) string {
	return acctest.ConfigCompose(testAccHarnessConfig_iamRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_harness" "test" {
  harness_name       = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  memory {
    agentcore_memory_configuration {
      arn = aws_bedrockagentcore_memory.test.arn
    }
  }

  model {
    bedrock_model_config {
      model_id = "anthropic.claude-sonnet-4-20250514"
    }
  }

  system_prompt {
    text = "You are a helpful assistant."
  }

  depends_on = [aws_iam_role_policy.test]
}

resource "aws_bedrockagentcore_memory" "test" {
  name                  = "%[1]s_mem"
  event_expiry_duration = 7
}
`, rName))
}

func testAccHarnessConfig_Memory_agentCoreMemoryConfiguration_options(rName, actorID string, messagesCount int, namespacePathTemplate string, relevanceScore float32, topK int) string {
	return acctest.ConfigCompose(testAccHarnessConfig_iamRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_harness" "test" {
  harness_name       = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  memory {
    agentcore_memory_configuration {
      arn            = aws_bedrockagentcore_memory.test.arn
      actor_id       = %[2]q
      messages_count = %[3]d

      retrieval_config {
        map_block_key   = %[4]q
        relevance_score = %[5]f
        top_k           = %[6]d
      }
    }
  }

  model {
    bedrock_model_config {
      model_id = "anthropic.claude-sonnet-4-20250514"
    }
  }

  system_prompt {
    text = "You are a helpful assistant."
  }

  depends_on = [aws_iam_role_policy.test]
}

resource "aws_bedrockagentcore_memory" "test" {
  name                  = "%[1]s_mem"
  event_expiry_duration = 7
}
`, rName, actorID, messagesCount, namespacePathTemplate, relevanceScore, topK))
}

func testAccHarnessConfig_Memory_agentCoreMemoryConfiguration_retrievalConfig(rName, namespacePathTemplate string, relevanceScore float32, topK int) string {
	return acctest.ConfigCompose(testAccHarnessConfig_iamRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_harness" "test" {
  harness_name       = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  memory {
    agentcore_memory_configuration {
      arn = aws_bedrockagentcore_memory.test.arn

      retrieval_config {
        map_block_key   = %[2]q
        relevance_score = %[3]f
        top_k           = %[4]d
      }
    }
  }

  model {
    bedrock_model_config {
      model_id = "anthropic.claude-sonnet-4-20250514"
    }
  }

  system_prompt {
    text = "You are a helpful assistant."
  }

  depends_on = [aws_iam_role_policy.test]
}

resource "aws_bedrockagentcore_memory" "test" {
  name                  = "%[1]s_mem"
  event_expiry_duration = 7
}
`, rName, namespacePathTemplate, relevanceScore, topK))
}

func testAccHarnessConfig_Memory_managedMemoryConfiguration_empty(rName string) string {
	return acctest.ConfigCompose(testAccHarnessConfig_iamRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_harness" "test" {
  harness_name       = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  memory {
    managed_memory_configuration {}
  }

  model {
    bedrock_model_config {
      model_id = "anthropic.claude-sonnet-4-20250514"
    }
  }

  system_prompt {
    text = "You are a helpful assistant."
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName))
}

func testAccHarnessConfig_Memory_managedMemoryConfiguration_options(rName string, eventExpiryDuration int, strategies string) string {
	return acctest.ConfigCompose(testAccHarnessConfig_iamRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_harness" "test" {
  harness_name       = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  memory {
    managed_memory_configuration {
      event_expiry_duration = %[2]d
      strategies            = %[3]s
    }
  }

  model {
    bedrock_model_config {
      model_id = "anthropic.claude-sonnet-4-20250514"
    }
  }

  system_prompt {
    text = "You are a helpful assistant."
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName, eventExpiryDuration, strategies))
}

func testAccHarnessConfig_Memory_managedMemoryConfiguration_encryptionKey(rName string) string {
	return acctest.ConfigCompose(testAccHarnessConfig_iamRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_harness" "test" {
  harness_name       = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  memory {
    managed_memory_configuration {
      encryption_key_arn = aws_kms_key.test.arn
    }
  }

  model {
    bedrock_model_config {
      model_id = "anthropic.claude-sonnet-4-20250514"
    }
  }

  system_prompt {
    text = "You are a helpful assistant."
  }

  depends_on = [aws_iam_role_policy.test]
}

resource "aws_kms_key" "test" {
  description = %[1]q
}
`, rName))
}

func testAccHarnessConfig_Memory_byType(t *testing.T, rName string, memType memoryConfigType) string {
	t.Helper()
	switch memType {
	case memoryNone:
		return testAccHarnessConfig_Memory_byType_none(rName)
	case memoryAgentCore:
		return testAccHarnessConfig_Memory_agentCoreMemoryConfiguration_basic(rName)
	case memoryDisabled:
		return testAccHarnessConfig_Memory_disabled(rName)
	case memoryManagedEmpty:
		return testAccHarnessConfig_Memory_managedMemoryConfiguration_empty(rName)
	case memoryManaged:
		return testAccHarnessConfig_Memory_managedMemoryConfiguration_options(rName, 7, `["SEMANTIC"]`)
	default:
		t.Fatalf("unknown memory config type: %s", memType)
		return ""
	}
}

func testAccHarnessConfig_Memory_byType_none(rName string) string {
	return acctest.ConfigCompose(testAccHarnessConfig_iamRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_harness" "test" {
  harness_name       = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  model {
    bedrock_model_config {
      model_id = "anthropic.claude-sonnet-4-20250514"
    }
  }

  system_prompt {
    text = "You are a helpful assistant."
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName))
}

func testAccHarnessConfig_Memory_disabled(rName string) string {
	return acctest.ConfigCompose(testAccHarnessConfig_iamRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_harness" "test" {
  harness_name       = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  memory {
    disabled {}
  }

  model {
    bedrock_model_config {
      model_id = "anthropic.claude-sonnet-4-20250514"
    }
  }

  system_prompt {
    text = "You are a helpful assistant."
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName))
}

func testAccHarnessConfig_environmentArtifact(rName, imageTag string) string {
	return acctest.ConfigCompose(testAccHarnessConfig_iamRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_harness" "test" {
  harness_name       = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  environment_artifact {
    container_configuration {
      container_uri = data.aws_ecr_image.test.image_uri
    }
  }

  model {
    bedrock_model_config {
      model_id = "anthropic.claude-sonnet-4-20250514"
    }
  }

  system_prompt {
    text = "You are a helpful assistant."
  }

  depends_on = [aws_iam_role_policy.test]
}

data "aws_ecr_image" "test" {
  registry_id     = "137112412989"
  repository_name = "amazonlinux"
  image_tag       = %[2]q
}
`, rName, imageTag))
}

func testAccHarnessConfig_authorizerConfiguration(rName, discoveryUrl, audience1, audience2, client1, client2, scope1, scope2 string) string {
	return acctest.ConfigCompose(testAccHarnessConfig_iamRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_harness" "test" {
  harness_name       = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  authorizer_configuration {
    custom_jwt_authorizer {
      discovery_url    = %[2]q
      allowed_audience = [%[3]q, %[4]q]
      allowed_clients  = [%[5]q, %[6]q]
      allowed_scopes   = [%[7]q, %[8]q]
    }
  }

  model {
    bedrock_model_config {
      model_id = "anthropic.claude-sonnet-4-20250514"
    }
  }

  system_prompt {
    text = "You are a helpful assistant."
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName, discoveryUrl, audience1, audience2, client1, client2, scope1, scope2))
}

func testAccHarnessConfig_environment_none(rName string) string {
	return acctest.ConfigCompose(testAccHarnessConfig_iamRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_harness" "test" {
  harness_name       = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  model {
    bedrock_model_config {
      model_id = "anthropic.claude-sonnet-4-20250514"
    }
  }

  system_prompt {
    text = "You are a helpful assistant."
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName))
}

func testAccHarnessConfig_environment_lifecycleConfiguration(rName string, idleTimeout, maxLifetime int) string {
	return acctest.ConfigCompose(testAccHarnessConfig_iamRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_harness" "test" {
  harness_name       = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  model {
    bedrock_model_config {
      model_id = "anthropic.claude-sonnet-4-20250514"
    }
  }

  system_prompt {
    text = "You are a helpful assistant."
  }

  environment {
    agentcore_runtime_environment {
      lifecycle_configuration {
        idle_runtime_session_timeout = %[2]d
        max_lifetime                 = %[3]d
      }

      network_configuration {
        network_mode = "PUBLIC"
      }
    }
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName, idleTimeout, maxLifetime))
}

func testAccHarnessConfig_environment_FilesystemConfiguration_sessionStorage(rName, mountPath string) string {
	return acctest.ConfigCompose(testAccHarnessConfig_iamRole(rName), acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_bedrockagentcore_harness" "test" {
  harness_name       = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  model {
    bedrock_model_config {
      model_id = "anthropic.claude-sonnet-4-20250514"
    }
  }

  system_prompt {
    text = "You are a helpful assistant."
  }

  environment {
    agentcore_runtime_environment {
      filesystem_configuration {
        session_storage {
          mount_path = %[2]q
        }
      }

      network_configuration {
        network_mode = "VPC"
        network_mode_config {
          security_groups = [aws_security_group.test.id]
          subnets         = aws_subnet.test[*].id
        }
      }
    }
  }

  depends_on = [aws_iam_role_policy.test]
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
  name   = %[1]q
}
`, rName, mountPath))
}

func testAccHarnessConfig_environment_FilesystemConfiguration_s3FilesAccessPoint(rName, mountPath string) string {
	bucketName := strings.ReplaceAll(rName, "_", "-")
	return acctest.ConfigCompose(
		testAccHarnessConfig_iamRole(rName),
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_bedrockagentcore_harness" "test" {
  harness_name       = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  model {
    bedrock_model_config {
      model_id = "anthropic.claude-sonnet-4-20250514"
    }
  }

  system_prompt {
    text = "You are a helpful assistant."
  }

  environment {
    agentcore_runtime_environment {
      filesystem_configuration {
        s3_files_access_point {
          access_point_arn = aws_s3files_access_point.test.arn
          mount_path       = %[2]q
        }
      }

      network_configuration {
        network_mode = "VPC"
        network_mode_config {
          security_groups = [aws_security_group.test.id]
          subnets         = aws_subnet.test[*].id
        }
      }
    }
  }

  depends_on = [aws_iam_role_policy.test, aws_iam_role_policy.test_s3files]
}

resource "aws_iam_role_policy" "test_s3files" {
  role = aws_iam_role.test.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = ["s3files:*"]
      Resource = "*"
    }]
  })
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
  name   = %[1]q
}

data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {}

resource "aws_s3_bucket" "test" {
  bucket        = %[3]q
  force_destroy = true
}

resource "aws_s3_bucket_versioning" "test" {
  bucket = aws_s3_bucket.test.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_iam_role" "s3files" {
  name = "%[1]s_s3f"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Sid       = "AllowS3FilesAssumeRole"
      Effect    = "Allow"
      Principal = { Service = "elasticfilesystem.amazonaws.com" }
      Action    = "sts:AssumeRole"
      Condition = {
        StringEquals = {
          "aws:SourceAccount" = data.aws_caller_identity.current.account_id
        }
        ArnLike = {
          "aws:SourceArn" = "arn:${data.aws_partition.current.partition}:s3files:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:file-system/*"
        }
      }
    }]
  })
}

resource "aws_iam_role_policy" "s3files" {
  role = aws_iam_role.s3files.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = ["s3:ListBucket", "s3:ListBucketVersions"]
        Resource = aws_s3_bucket.test.arn
      },
      {
        Effect   = "Allow"
        Action   = ["s3:AbortMultipartUpload", "s3:DeleteObject*", "s3:GetObject*", "s3:List*", "s3:PutObject*"]
        Resource = "${aws_s3_bucket.test.arn}/*"
      },
    ]
  })
}

resource "aws_s3files_file_system" "test" {
  bucket   = aws_s3_bucket.test.arn
  role_arn = aws_iam_role.s3files.arn

  depends_on = [aws_iam_role_policy.s3files, aws_s3_bucket_versioning.test]
}

resource "aws_s3files_access_point" "test" {
  file_system_id = aws_s3files_file_system.test.id

  posix_user {
    gid = 1000
    uid = 1000
  }

  root_directory {
    path = "/"

    creation_permissions {
      owner_gid   = 1000
      owner_uid   = 1000
      permissions = "755"
    }
  }

  depends_on = [aws_s3files_mount_target.test]
}

resource "aws_s3files_mount_target" "test" {
  count           = 2
  file_system_id  = aws_s3files_file_system.test.id
  subnet_id       = aws_subnet.test[count.index].id
  security_groups = [aws_security_group.test.id]
}
`, rName, mountPath, bucketName))
}

func testAccHarnessConfig_environment_FilesystemConfiguration_vpcNoFilesystem(rName string) string {
	return acctest.ConfigCompose(testAccHarnessConfig_iamRole(rName), acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_bedrockagentcore_harness" "test" {
  harness_name       = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  model {
    bedrock_model_config {
      model_id = "anthropic.claude-sonnet-4-20250514"
    }
  }

  system_prompt {
    text = "You are a helpful assistant."
  }

  environment {
    agentcore_runtime_environment {
      network_configuration {
        network_mode = "VPC"
        network_mode_config {
          security_groups = [aws_security_group.test.id]
          subnets         = aws_subnet.test[*].id
        }
      }
    }
  }

  depends_on = [aws_iam_role_policy.test]
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
  name   = %[1]q
}
`, rName))
}

func testAccHarnessConfig_environment_FilesystemConfiguration_efsAccessPoint(rName, mountPath string) string {
	return acctest.ConfigCompose(
		testAccHarnessConfig_iamRole(rName),
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_bedrockagentcore_harness" "test" {
  harness_name       = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  model {
    bedrock_model_config {
      model_id = "anthropic.claude-sonnet-4-20250514"
    }
  }

  system_prompt {
    text = "You are a helpful assistant."
  }

  environment {
    agentcore_runtime_environment {
      filesystem_configuration {
        efs_access_point {
          access_point_arn = aws_efs_access_point.test.arn
          mount_path       = %[2]q
        }
      }

      network_configuration {
        network_mode = "VPC"
        network_mode_config {
          security_groups = [aws_security_group.test.id]
          subnets         = aws_subnet.test[*].id
        }
      }
    }
  }

  depends_on = [aws_iam_role_policy.test, aws_iam_role_policy.test_efs, aws_efs_mount_target.test]
}

resource "aws_iam_role_policy" "test_efs" {
  role = aws_iam_role.test.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = ["elasticfilesystem:DescribeAccessPoints", "elasticfilesystem:DescribeMountTargets", "elasticfilesystem:DescribeFileSystems"]
      Resource = "*"
    }]
  })
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
  name   = %[1]q
}

resource "aws_efs_file_system" "test" {
  creation_token = %[1]q
}

resource "aws_efs_mount_target" "test" {
  count           = 2
  file_system_id  = aws_efs_file_system.test.id
  subnet_id       = aws_subnet.test[count.index].id
  security_groups = [aws_security_group.test.id]
}

resource "aws_efs_access_point" "test" {
  file_system_id = aws_efs_file_system.test.id

  posix_user {
    gid = 1000
    uid = 1000
  }

  root_directory {
    path = "/data"

    creation_info {
      owner_gid   = 1000
      owner_uid   = 1000
      permissions = "755"
    }
  }
}
`, rName, mountPath))
}

func testAccHarnessConfig_environment_FilesystemConfiguration_multiple(rName string) string {
	return acctest.ConfigCompose(testAccHarnessConfig_iamRole(rName), acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_bedrockagentcore_harness" "test" {
  harness_name       = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  model {
    bedrock_model_config {
      model_id = "anthropic.claude-sonnet-4-20250514"
    }
  }

  system_prompt {
    text = "You are a helpful assistant."
  }

  environment {
    agentcore_runtime_environment {
      filesystem_configuration {
        session_storage {
          mount_path = "/mnt/session"
        }
      }

      filesystem_configuration {
        efs_access_point {
          access_point_arn = aws_efs_access_point.test.arn
          mount_path       = "/mnt/efsdata"
        }
      }

      network_configuration {
        network_mode = "VPC"
        network_mode_config {
          security_groups = [aws_security_group.test.id]
          subnets         = aws_subnet.test[*].id
        }
      }
    }
  }

  depends_on = [aws_iam_role_policy.test, aws_iam_role_policy.test_efs, aws_efs_mount_target.test]
}

resource "aws_iam_role_policy" "test_efs" {
  role = aws_iam_role.test.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = ["elasticfilesystem:DescribeAccessPoints", "elasticfilesystem:DescribeMountTargets", "elasticfilesystem:DescribeFileSystems"]
      Resource = "*"
    }]
  })
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
  name   = %[1]q
}

resource "aws_efs_file_system" "test" {
  creation_token = %[1]q
}

resource "aws_efs_mount_target" "test" {
  count           = 2
  file_system_id  = aws_efs_file_system.test.id
  subnet_id       = aws_subnet.test[count.index].id
  security_groups = [aws_security_group.test.id]
}

resource "aws_efs_access_point" "test" {
  file_system_id = aws_efs_file_system.test.id

  posix_user {
    gid = 1000
    uid = 1000
  }

  root_directory {
    path = "/data"

    creation_info {
      owner_gid   = 1000
      owner_uid   = 1000
      permissions = "755"
    }
  }
}
`, rName))
}

func testAccHarnessConfig_environment_Network_VPC(rName string) string {
	return acctest.ConfigCompose(testAccHarnessConfig_iamRole(rName), acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_bedrockagentcore_harness" "test" {
  harness_name       = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  model {
    bedrock_model_config {
      model_id = "anthropic.claude-sonnet-4-20250514"
    }
  }

  system_prompt {
    text = "You are a coding assistant."
  }

  environment {
    agentcore_runtime_environment {
      network_configuration {
        network_mode = "VPC"
        network_mode_config {
          security_groups = [aws_security_group.test.id]
          subnets         = aws_subnet.test[*].id
        }
      }
    }
  }

  depends_on = [aws_iam_role_policy.test]
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
  name   = %[1]q
}
`, rName))
}

func testAccHarnessConfig_environment_Network_Public(rName string) string {
	return acctest.ConfigCompose(testAccHarnessConfig_iamRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_harness" "test" {
  harness_name       = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  model {
    bedrock_model_config {
      model_id = "anthropic.claude-sonnet-4-20250514"
    }
  }

  system_prompt {
    text = "You are a coding assistant."
  }

  environment {
    agentcore_runtime_environment {
      network_configuration {
        network_mode = "PUBLIC"
      }
    }
  }
}
`, rName))
}

// testAccHarnessConfig_environment_Network_VPCToPublic is used in the VPC-to-Public update test.
// It includes the VPC/subnet infrastructure (so it doesn't get destroyed mid-test) but configures
// the harness with PUBLIC network mode.
func testAccHarnessConfig_environment_Network_VPCToPublic(rName string) string {
	return acctest.ConfigCompose(testAccHarnessConfig_iamRole(rName), acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_bedrockagentcore_harness" "test" {
  harness_name       = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  model {
    bedrock_model_config {
      model_id = "anthropic.claude-sonnet-4-20250514"
    }
  }

  system_prompt {
    text = "You are a coding assistant."
  }

  environment {
    agentcore_runtime_environment {
      network_configuration {
        network_mode = "PUBLIC"
      }
    }
  }

  depends_on = [aws_iam_role_policy.test]
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
  name   = %[1]q
}
`, rName))
}

func testAccHarnessConfig_skillGit(rName string) string {
	return acctest.ConfigCompose(testAccHarnessConfig_iamRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_harness" "test" {
  harness_name       = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  model {
    bedrock_model_config {
      model_id = "anthropic.claude-sonnet-4-20250514"
    }
  }

  skill {
    git {
      url  = "https://github.com/example/skill.git"
      path = "skills"
    }
  }

  system_prompt {
    text = "You are a helpful assistant."
  }
}
`, rName))
}
