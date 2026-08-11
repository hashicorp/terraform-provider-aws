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
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEnvironment), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
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
				ImportStateVerifyIgnore:              []string{"memory"},
				ImportStateCheck: acctest.ComposeAggregateImportStateCheckFunc(
					acctest.ImportCheckResourceAttr("memory.#", "1"),
					acctest.ImportCheckResourceAttr("memory.0.agentcore_memory_configuration.#", "0"),
					acctest.ImportCheckResourceAttr("memory.0.managed_memory_configuration.#", "1"),
					acctest.ImportMatchResourceAttr("memory.0.managed_memory_configuration.0.arn", regexache.MustCompile(`^arn:[^:]+:bedrock-agentcore:[^:]+:\d{12}:memory/harness_`+rName+`_[a-zA-Z0-9]+-[a-zA-Z0-9]+$`)),
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
				ImportStateVerifyIgnore:              []string{"memory"},
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
				ImportStateVerifyIgnore:              []string{"memory"},
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
				ImportStateVerifyIgnore:              []string{"memory"},
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
					"memory",
					"model.0.bedrock_model_config.0.temperature",
					"model.0.bedrock_model_config.0.top_p",
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
				ImportStateVerifyIgnore:              []string{"memory"},
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
				ImportStateVerifyIgnore:              []string{"memory"},
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
				ImportStateVerifyIgnore:              []string{"memory"},
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
				},
			},
			{
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "harness_id"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "harness_id",
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
								names.AttrARN:           tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`memory/harness_`+rName+`_[a-zA-Z0-9]+-[a-zA-Z0-9]+`)),
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
						names.AttrARN:           tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`memory/harness_`+rName+`_[a-zA-Z0-9]+-[a-zA-Z0-9]+`)),
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
						names.AttrARN:           tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`memory/harness_`+rName+`_[a-zA-Z0-9]+-[a-zA-Z0-9]+`)),
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
						names.AttrARN:           tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`memory/harness_`+rName+`_[a-zA-Z0-9]+-[a-zA-Z0-9]+`)),
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
				},
			},
			{
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "harness_id"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "harness_id",
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
				ImportStateVerifyIgnore:              []string{"memory"},
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
				ImportStateVerifyIgnore:              []string{"memory"},
			},
		},
	})
}

func TestAccBedrockAgentCoreHarness_tags(t *testing.T) {
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
				Config: testAccHarnessConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
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
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "harness_id"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "harness_id",
				ImportStateVerifyIgnore:              []string{"memory"},
			},
			{
				Config: testAccHarnessConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
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
				Config: testAccHarnessConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
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

func memoryConfigStateChecks(t *testing.T, resourceName string, memType memoryConfigType) []statecheck.StateCheck {
	t.Helper()
	memoryPath := tfjsonpath.New("memory")
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
  name                  = %[1]q
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
  name                  = %[1]q
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
  name                  = %[1]q
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

func testAccHarnessConfig_tags1(rName, tagKey1, tagValue1 string) string {
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

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccHarnessConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
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

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
