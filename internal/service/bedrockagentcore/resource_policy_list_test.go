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

func TestAccBedrockAgentCoreResourcePolicy_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	acctest.SkipIfEnvVarNotSet(t, "AWS_BEDROCK_AGENTCORE_RUNTIME_IMAGE_V1_URI")
	resourceName1 := "aws_bedrockagentcore_resource_policy.test[0]"
	resourceName2 := "aws_bedrockagentcore_resource_policy.test[1]"
	rName := randomWithPrefixAndUnderscore(t)

	identity1 := tfstatecheck.Identity()
	identity2 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/ResourcePolicy/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"AWS_BEDROCK_AGENTCORE_RUNTIME_IMAGE_V1_URI": config.StringVariable(acctest.SkipIfEnvVarNotSet(t, "AWS_BEDROCK_AGENTCORE_RUNTIME_IMAGE_V1_URI")),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrResourceARN), tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`runtime/.+`))),

					identity2.GetIdentity(resourceName2),
					statecheck.ExpectKnownValue(resourceName2, tfjsonpath.New(names.AttrResourceARN), tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`runtime/.+`))),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/ResourcePolicy/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(2),
					"AWS_BEDROCK_AGENTCORE_RUNTIME_IMAGE_V1_URI": config.StringVariable(acctest.SkipIfEnvVarNotSet(t, "AWS_BEDROCK_AGENTCORE_RUNTIME_IMAGE_V1_URI")),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_bedrockagentcore_resource_policy.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_bedrockagentcore_resource_policy.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.NotNull()),
					tfquerycheck.ExpectNoResourceObject("aws_bedrockagentcore_resource_policy.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks())),

					tfquerycheck.ExpectIdentityFunc("aws_bedrockagentcore_resource_policy.test", identity2.Checks()),
					querycheck.ExpectResourceDisplayName("aws_bedrockagentcore_resource_policy.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks()), knownvalue.NotNull()),
					tfquerycheck.ExpectNoResourceObject("aws_bedrockagentcore_resource_policy.test", tfqueryfilter.ByResourceIdentityFunc(identity2.Checks())),
				},
			},
		},
	})
}

func TestAccBedrockAgentCoreResourcePolicy_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	acctest.SkipIfEnvVarNotSet(t, "AWS_BEDROCK_AGENTCORE_RUNTIME_IMAGE_V1_URI")
	resourceName1 := "aws_bedrockagentcore_resource_policy.test[0]"
	rName := randomWithPrefixAndUnderscore(t)

	identity1 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BedrockAgentCoreServiceID),
		CheckDestroy:             testAccCheckResourcePolicyDestroy(ctx, t),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Setup
			{
				ConfigDirectory: config.StaticDirectory("testdata/ResourcePolicy/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
					"AWS_BEDROCK_AGENTCORE_RUNTIME_IMAGE_V1_URI": config.StringVariable(acctest.SkipIfEnvVarNotSet(t, "AWS_BEDROCK_AGENTCORE_RUNTIME_IMAGE_V1_URI")),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					statecheck.ExpectKnownValue(resourceName1, tfjsonpath.New(names.AttrResourceARN), tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`runtime/.+`))),
				},
			},

			// Step 2: Query
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/ResourcePolicy/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:  config.StringVariable(rName),
					"resource_count": config.IntegerVariable(1),
					"AWS_BEDROCK_AGENTCORE_RUNTIME_IMAGE_V1_URI": config.StringVariable(acctest.SkipIfEnvVarNotSet(t, "AWS_BEDROCK_AGENTCORE_RUNTIME_IMAGE_V1_URI")),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_bedrockagentcore_resource_policy.test", identity1.Checks()),
					querycheck.ExpectResourceDisplayName("aws_bedrockagentcore_resource_policy.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), knownvalue.NotNull()),
					querycheck.ExpectResourceKnownValues("aws_bedrockagentcore_resource_policy.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrResourceARN), tfknownvalue.RegionalARNRegexp("bedrock-agentcore", regexache.MustCompile(`runtime/.+`))),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrPolicy), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					}),
				},
			},
		},
	})
}
