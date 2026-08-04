// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	tfquerycheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/querycheck"
	tfqueryfilter "github.com/hashicorp/terraform-provider-aws/internal/acctest/queryfilter"
	tfstatecheck "github.com/hashicorp/terraform-provider-aws/internal/acctest/statecheck"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgentCoreHarness_List_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName1 := "aws_bedrockagentcore_harness.test[0]"
	resourceName2 := "aws_bedrockagentcore_harness.test[1]"
	rName := testAccRandomHarnessName(t)
	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		CheckDestroy:             testAccCheckHarnessDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/Harness/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrARN), checkHarnessARN(rName+"_0")),

					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrARN), checkHarnessARN(rName+"_1")),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Harness/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_bedrockagentcore_harness.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_bedrockagentcore_harness.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.NotNull()),
					tfquerycheck.ExpectNoResourceObject("aws_bedrockagentcore_harness.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),

					tfquerycheck.ExpectIdentityFunc("aws_bedrockagentcore_harness.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_bedrockagentcore_harness.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), knownvalue.NotNull()),
					tfquerycheck.ExpectNoResourceObject("aws_bedrockagentcore_harness.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreHarness_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName1 := "aws_bedrockagentcore_harness.test[0]"
	rName := testAccRandomHarnessName(t)
	identity1 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		CheckDestroy:             testAccCheckHarnessDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/Harness/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrARN), checkHarnessARN(rName+"_0")),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Harness/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
					acctest.CtResourceTags: config.MapVariable(map[string]config.Variable{
						acctest.CtKey1: config.StringVariable(acctest.CtValue1),
					}),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_bedrockagentcore_harness.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_bedrockagentcore_harness.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.NotNull()),
					querycheck.ExpectResourceKnownValues("aws_bedrockagentcore_harness.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New("allowed_tools"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrARN), checkHarnessARN(rName+"_0")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("authorizer_configuration"), knownvalue.ListSizeExact(0)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrEnvironment), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("environment_artifact"), knownvalue.ListSizeExact(0)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("environment_variables"), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrExecutionRoleARN), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("harness_id"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("harness_name"), knownvalue.StringExact(rName+"_0")),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("max_iterations"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("max_tokens"), knownvalue.Null()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("memory"), knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"agentcore_memory_configuration": knownvalue.ListSizeExact(0),
								"managed_memory_configuration": knownvalue.ListExact([]knownvalue.Check{
									knownvalue.ObjectExact(map[string]knownvalue.Check{
										"arn":                   tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`memory/harness_`+rName+`_0_[a-zA-Z0-9]+-[a-zA-Z0-9]+`)),
										"encryption_key_arn":    knownvalue.Null(),
										"event_expiry_duration": knownvalue.Int32Exact(30),
										"strategies": knownvalue.SetExact([]knownvalue.Check{
											knownvalue.StringExact("SEMANTIC"),
											knownvalue.StringExact("SUMMARIZATION"),
										}),
									}),
								}),
							}),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("model"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("skill"), knownvalue.ListSizeExact(0)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("system_prompt"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
							acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
						})),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("timeout_seconds"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("tool"), knownvalue.ListSizeExact(0)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("truncation"), knownvalue.NotNull()),
					}),
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreHarness_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName1 := "aws_bedrockagentcore_harness.test[0]"
	resourceName2 := "aws_bedrockagentcore_harness.test[1]"
	rName := testAccRandomHarnessName(t)
	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		CheckDestroy:             testAccCheckHarnessDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/Harness/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrARN), checkHarnessARNAlternateRegion(rName+"_0")),

					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrARN), checkHarnessARNAlternateRegion(rName+"_1")),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/Harness/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"region":         config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_bedrockagentcore_harness.test", identity1.Checks()),

					tfquerycheck.ExpectIdentityFunc("aws_bedrockagentcore_harness.test", identity2.Checks()),
				},
			},
		},
	})
}
