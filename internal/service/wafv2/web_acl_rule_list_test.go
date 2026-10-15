// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package wafv2_test

import (
	"testing"

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

func TestAccWAFV2WebACLRule_List_basic(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_wafv2_web_acl_rule.test0"
	resourceName2 := "aws_wafv2_web_acl_rule.test1"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	webACLARN := tfstatecheck.StateValue()
	name1 := tfstatecheck.StateValue()
	name2 := tfstatecheck.StateValue()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/WebACLRule/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					webACLARN.GetStateValue("aws_wafv2_web_acl.test", tfjsonpath.New(names.AttrARN)),
					name1.GetStateValue(resourceName1, tfjsonpath.New(names.AttrName)),
					name2.GetStateValue(resourceName2, tfjsonpath.New(names.AttrName)),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/WebACLRule/list_basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectIdentity("aws_wafv2_web_acl_rule.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						"web_acl_arn":       webACLARN.ValueCheck(),
						names.AttrName:      name1.ValueCheck(),
					}),
					querycheck.ExpectIdentity("aws_wafv2_web_acl_rule.test", map[string]knownvalue.Check{
						names.AttrAccountID: tfknownvalue.AccountID(),
						names.AttrRegion:    knownvalue.StringExact(acctest.Region()),
						"web_acl_arn":       webACLARN.ValueCheck(),
						names.AttrName:      name2.ValueCheck(),
					}),
				},
			},
		},
	})
}

func TestAccWAFV2WebACLRule_List_includeResource(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName := "aws_wafv2_web_acl_rule.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	identity1 := tfstatecheck.Identity()

	acctest.ParallelTest(ctx, t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/WebACLRule/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/WebACLRule/list_include_resource/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_wafv2_web_acl_rule.test", identity1.Checks()),
					querycheck.ExpectResourceKnownValues("aws_wafv2_web_acl_rule.test", tfqueryfilter.ByResourceIdentityFunc(identity1.Checks()), []querycheck.KnownValueCheck{
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrPriority), knownvalue.Int32Exact(0)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("web_acl_arn"), knownvalue.NotNull()),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrAction), knownvalue.ListSizeExact(1)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("allow"), knownvalue.ListSizeExact(0)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("block"), knownvalue.ListSizeExact(1)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("captcha"), knownvalue.ListSizeExact(0)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("challenge"), knownvalue.ListSizeExact(0)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New(names.AttrAction).AtSliceIndex(0).AtMapKey("count"), knownvalue.ListSizeExact(0)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("statement"), knownvalue.ListSizeExact(1)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("statement").AtSliceIndex(0).AtMapKey("geo_match_statement"), knownvalue.ListSizeExact(1)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("statement").AtSliceIndex(0).AtMapKey("geo_match_statement").AtSliceIndex(0).AtMapKey("country_codes"), knownvalue.SetSizeExact(2)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("visibility_config"), knownvalue.ListSizeExact(1)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("visibility_config").AtSliceIndex(0).AtMapKey("cloudwatch_metrics_enabled"), knownvalue.Bool(false)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("visibility_config").AtSliceIndex(0).AtMapKey(names.AttrMetricName), knownvalue.StringExact(rName)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("visibility_config").AtSliceIndex(0).AtMapKey("sampled_requests_enabled"), knownvalue.Bool(false)),
						tfquerycheck.KnownValueCheck(tfjsonpath.New("rule_label"), knownvalue.ListSizeExact(0)),
					}),
				},
			},
		},
	})
}

func TestAccWAFV2WebACLRule_List_regionOverride(t *testing.T) {
	ctx := acctest.Context(t)

	resourceName1 := "aws_wafv2_web_acl_rule.test0"
	resourceName2 := "aws_wafv2_web_acl_rule.test1"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
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
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/WebACLRule/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"region":        config.StringVariable(acctest.AlternateRegion()),
				},
				ConfigStateChecks: []statecheck.StateCheck{
					identity1.GetIdentity(resourceName1),
					identity2.GetIdentity(resourceName2),
				},
			},
			{
				Query:           true,
				ConfigDirectory: config.StaticDirectory("testdata/WebACLRule/list_region_override/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"region":        config.StringVariable(acctest.AlternateRegion()),
				},
				QueryResultChecks: []querycheck.QueryResultCheck{
					tfquerycheck.ExpectIdentityFunc("aws_wafv2_web_acl_rule.test", identity1.Checks()),
					tfquerycheck.ExpectIdentityFunc("aws_wafv2_web_acl_rule.test", identity2.Checks()),
				},
			},
		},
	})
}
