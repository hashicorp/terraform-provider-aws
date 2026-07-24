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
	"github.com/hashicorp/terraform-plugin-testing/config"
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
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), checkHarnessARN(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("harness_id"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("harness_name"), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
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
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("model").AtSliceIndex(0).AtMapKey("bedrock_model_config").AtSliceIndex(0).AtMapKey("api_format"), knownvalue.StringExact("converse_stream")),
				},
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
				ImportStateVerifyIgnore:              []string{"model.0.bedrock_model_config.0.temperature", "model.0.bedrock_model_config.0.top_p"},
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
		},
	})
}

func TestAccBedrockAgentCoreHarness_memory(t *testing.T) {
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
				Config: testAccHarnessConfig_memory(rName, 0.25),
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
				Config: testAccHarnessConfig_memory(rName, 0.35),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckHarnessExists(ctx, t, resourceName, &harness),
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
      api_format  = "converse_stream"
      model_id    = "anthropic.claude-sonnet-4-20250514"
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
}
`, rName, key, value))
}

func testAccHarnessConfig_memory(rName string, relevanceScore float32) string {
	return acctest.ConfigCompose(testAccHarnessConfig_iamRole(rName), fmt.Sprintf(`
resource "aws_bedrockagentcore_harness" "test" {
  harness_name       = %[1]q
  execution_role_arn = aws_iam_role.test.arn

  memory {
    agentcore_memory_configuration {
      arn = aws_bedrockagentcore_memory.test.arn

      retrieval_config {
        map_block_key   = "key1"
        relevance_score = %[2]f
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
}

resource "aws_bedrockagentcore_memory" "test" {
  name                  = %[1]q
  event_expiry_duration = 7
}
`, rName, relevanceScore))
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
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}
