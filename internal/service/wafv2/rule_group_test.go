// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafv2_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfwafv2 "github.com/hashicorp/terraform-provider-aws/internal/service/wafv2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWAFV2RuleGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_basic(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, string(awstypes.ScopeRegional)),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccRuleGroupImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccWAFV2RuleGroup_nameGenerated(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RuleGroup
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_nameGenerated(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrNameGenerated(resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, id.UniqueIdPrefix),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccRuleGroupImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccWAFV2RuleGroup_namePrefix(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RuleGroup
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_namePrefix("tf-acc-test-prefix-"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.CheckResourceAttrNameFromPrefix(resourceName, names.AttrName, "tf-acc-test-prefix-"),
					resource.TestCheckResourceAttr(resourceName, names.AttrNamePrefix, "tf-acc-test-prefix-"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccRuleGroupImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccWAFV2RuleGroup_updateRule(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_basic(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, string(awstypes.ScopeRegional)),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				Config: testAccRuleGroupConfig_basicUpdate(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "50"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Updated"),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, string(awstypes.ScopeRegional)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrName:                      "rule-1",
						names.AttrPriority:                  acctest.Ct1,
						"action.#":                          acctest.Ct1,
						"action.0.allow.#":                  acctest.Ct0,
						"action.0.block.#":                  acctest.Ct0,
						"action.0.count.#":                  acctest.Ct1,
						"action.0.captcha.#":                acctest.Ct0,
						"action.0.challenge.#":              acctest.Ct0,
						"statement.#":                       acctest.Ct1,
						"statement.0.geo_match_statement.#": acctest.Ct1,
						"statement.0.geo_match_statement.0.country_codes.#": acctest.Ct2,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccRuleGroupImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccWAFV2RuleGroup_updateRuleProperties(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"
	ruleName2 := fmt.Sprintf("%s-2", ruleGroupName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_basicUpdate(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "50"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Updated"),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, string(awstypes.ScopeRegional)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrName:         "rule-1",
						names.AttrPriority:     acctest.Ct1,
						"action.#":             acctest.Ct1,
						"action.0.allow.#":     acctest.Ct0,
						"action.0.block.#":     acctest.Ct0,
						"action.0.count.#":     acctest.Ct1,
						"action.0.captcha.#":   acctest.Ct0,
						"action.0.challenge.#": acctest.Ct0,
						"visibility_config.0.cloudwatch_metrics_enabled": acctest.CtFalse,
						"visibility_config.0.metric_name":                "friendly-rule-metric-name",
						"visibility_config.0.sampled_requests_enabled":   acctest.CtFalse,
						"statement.#":                                       acctest.Ct1,
						"statement.0.geo_match_statement.#":                 acctest.Ct1,
						"statement.0.geo_match_statement.0.country_codes.#": acctest.Ct2,
					}),
				),
			},
			{
				// Test step verifies addition of a rule block with the first block unchanged
				Config: testAccRuleGroupConfig_updateMultiples(ruleGroupName, "rule-1", ruleName2, 1, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "50"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Updated"),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, string(awstypes.ScopeRegional)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrName:         "rule-1",
						names.AttrPriority:     acctest.Ct1,
						"action.#":             acctest.Ct1,
						"action.0.allow.#":     acctest.Ct0,
						"action.0.block.#":     acctest.Ct0,
						"action.0.count.#":     acctest.Ct1,
						"action.0.captcha.#":   acctest.Ct0,
						"action.0.challenge.#": acctest.Ct0,
						"visibility_config.#":  acctest.Ct1,
						"visibility_config.0.cloudwatch_metrics_enabled": acctest.CtFalse,
						"visibility_config.0.metric_name":                "rule-1",
						"visibility_config.0.sampled_requests_enabled":   acctest.CtFalse,
						"statement.#":                                       acctest.Ct1,
						"statement.0.geo_match_statement.#":                 acctest.Ct1,
						"statement.0.geo_match_statement.0.country_codes.#": acctest.Ct2,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrName:         ruleName2,
						names.AttrPriority:     acctest.Ct2,
						"action.#":             acctest.Ct1,
						"action.0.allow.#":     acctest.Ct0,
						"action.0.block.#":     acctest.Ct1,
						"action.0.count.#":     acctest.Ct0,
						"action.0.captcha.#":   acctest.Ct0,
						"action.0.challenge.#": acctest.Ct0,
						"visibility_config.#":  acctest.Ct1,
						"visibility_config.0.cloudwatch_metrics_enabled": acctest.CtFalse,
						"visibility_config.0.metric_name":                ruleName2,
						"visibility_config.0.sampled_requests_enabled":   acctest.CtFalse,
						"statement.#": acctest.Ct1,
						"statement.0.size_constraint_statement.#":                                 acctest.Ct1,
						"statement.0.size_constraint_statement.0.comparison_operator":             "LT",
						"statement.0.size_constraint_statement.0.field_to_match.#":                acctest.Ct1,
						"statement.0.size_constraint_statement.0.field_to_match.0.query_string.#": acctest.Ct1,
						"statement.0.size_constraint_statement.0.size":                            "50",
						"statement.0.size_constraint_statement.0.text_transformation.#":           acctest.Ct2,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.statement.0.size_constraint_statement.0.text_transformation.*", map[string]string{
						names.AttrPriority: acctest.Ct2,
						names.AttrType:     "CMD_LINE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.statement.0.size_constraint_statement.0.text_transformation.*", map[string]string{
						names.AttrPriority: "5",
						names.AttrType:     "NONE",
					}),
				),
			},
			{
				// Test step to verify a change in priority for rule #1 and a change in name and priority for rule #2
				Config: testAccRuleGroupConfig_updateMultiples(ruleGroupName, "rule-1", "updated", 5, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "50"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Updated"),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, string(awstypes.ScopeRegional)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrName:         "rule-1",
						names.AttrPriority:     "5",
						"action.#":             acctest.Ct1,
						"action.0.allow.#":     acctest.Ct0,
						"action.0.block.#":     acctest.Ct0,
						"action.0.count.#":     acctest.Ct1,
						"action.0.captcha.#":   acctest.Ct0,
						"action.0.challenge.#": acctest.Ct0,
						"visibility_config.#":  acctest.Ct1,
						"visibility_config.0.cloudwatch_metrics_enabled": acctest.CtFalse,
						"visibility_config.0.metric_name":                "rule-1",
						"visibility_config.0.sampled_requests_enabled":   acctest.CtFalse,
						"statement.#":                                       acctest.Ct1,
						"statement.0.geo_match_statement.#":                 acctest.Ct1,
						"statement.0.geo_match_statement.0.country_codes.#": acctest.Ct2,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrName:         "updated",
						names.AttrPriority:     acctest.Ct10,
						"action.#":             acctest.Ct1,
						"action.0.allow.#":     acctest.Ct0,
						"action.0.block.#":     acctest.Ct1,
						"action.0.count.#":     acctest.Ct0,
						"action.0.captcha.#":   acctest.Ct0,
						"action.0.challenge.#": acctest.Ct0,
						"visibility_config.#":  acctest.Ct1,
						"visibility_config.0.cloudwatch_metrics_enabled": acctest.CtFalse,
						"visibility_config.0.metric_name":                "updated",
						"visibility_config.0.sampled_requests_enabled":   acctest.CtFalse,
						"statement.#": acctest.Ct1,
						"statement.0.size_constraint_statement.#":                                 acctest.Ct1,
						"statement.0.size_constraint_statement.0.comparison_operator":             "LT",
						"statement.0.size_constraint_statement.0.field_to_match.#":                acctest.Ct1,
						"statement.0.size_constraint_statement.0.field_to_match.0.query_string.#": acctest.Ct1,
						"statement.0.size_constraint_statement.0.size":                            "50",
						"statement.0.size_constraint_statement.0.text_transformation.#":           acctest.Ct2,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.statement.0.size_constraint_statement.0.text_transformation.*", map[string]string{
						names.AttrPriority: acctest.Ct2,
						names.AttrType:     "CMD_LINE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.statement.0.size_constraint_statement.0.text_transformation.*", map[string]string{
						names.AttrPriority: "5",
						names.AttrType:     "NONE",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccRuleGroupImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccWAFV2RuleGroup_byteMatchStatement(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_byteMatchStatement(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                        acctest.Ct1,
						"statement.0.byte_match_statement.#": acctest.Ct1,
						"statement.0.byte_match_statement.0.positional_constraint": "CONTAINS",
						"statement.0.byte_match_statement.0.search_string":         "word",
						"statement.0.byte_match_statement.0.text_transformation.#": acctest.Ct2,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.statement.0.byte_match_statement.0.text_transformation.*", map[string]string{
						names.AttrPriority: "5",
						names.AttrType:     "NONE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.statement.0.byte_match_statement.0.text_transformation.*", map[string]string{
						names.AttrPriority: acctest.Ct2,
						names.AttrType:     "LOWERCASE",
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_byteMatchStatementUpdate(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                        acctest.Ct1,
						"statement.0.byte_match_statement.#": acctest.Ct1,
						"statement.0.byte_match_statement.0.positional_constraint": "EXACTLY",
						"statement.0.byte_match_statement.0.search_string":         "sentence",
						"statement.0.byte_match_statement.0.text_transformation.#": acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.statement.0.byte_match_statement.0.text_transformation.*", map[string]string{
						names.AttrPriority: acctest.Ct3,
						names.AttrType:     "CMD_LINE",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccRuleGroupImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccWAFV2RuleGroup_ByteMatchStatement_fieldToMatch(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_byteMatchStatementFieldToMatchAllQueryArguments(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                         acctest.Ct1,
						"statement.0.byte_match_statement.#":                  acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.#": acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.0.all_query_arguments.#":   acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.0.body.#":                  acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.cookies.#":               acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.header_order.#":          acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.headers.#":               acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.ja3_fingerprint.#":       acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.#":             acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.method.#":                acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.query_string.#":          acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.single_header.#":         acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.#": acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.uri_path.#":              acctest.Ct0,
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_byteMatchStatementFieldToMatchBody(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                         acctest.Ct1,
						"statement.0.byte_match_statement.#":                  acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.#": acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.0.all_query_arguments.#":   acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.body.#":                  acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.0.cookies.#":               acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.header_order.#":          acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.headers.#":               acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.ja3_fingerprint.#":       acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.#":             acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.method.#":                acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.query_string.#":          acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.single_header.#":         acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.#": acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.uri_path.#":              acctest.Ct0,
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_byteMatchStatementFieldToMatchCookies(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                         acctest.Ct1,
						"statement.0.byte_match_statement.#":                  acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.#": acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.0.all_query_arguments.#":                        acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.body.#":                                       acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.cookies.#":                                    acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.0.cookies.0.match_scope":                        "ALL",
						"statement.0.byte_match_statement.0.field_to_match.0.cookies.0.oversize_handling":                  "NO_MATCH",
						"statement.0.byte_match_statement.0.field_to_match.0.cookies.0.match_pattern.#":                    acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.0.cookies.0.match_pattern.0.included_cookies.0": "test",
						"statement.0.byte_match_statement.0.field_to_match.0.cookies.0.match_pattern.0.included_cookies.1": "cookie_test",
						"statement.0.byte_match_statement.0.field_to_match.0.header_order.#":                               acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.headers.#":                                    acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.ja3_fingerprint.#":                            acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.#":                                  acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.method.#":                                     acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.query_string.#":                               acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.single_header.#":                              acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.#":                      acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.uri_path.#":                                   acctest.Ct0,
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_byteMatchStatementFieldToMatchJSONBody(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                                                                      acctest.Ct1,
						"statement.0.byte_match_statement.#":                                                               acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.#":                                              acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.#":                                  acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.0.match_scope":                      "VALUE",
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.0.invalid_fallback_behavior":        "MATCH",
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.0.oversize_handling":                "CONTINUE",
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.0.match_pattern.#":                  acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.0.match_pattern.0.included_paths.#": acctest.Ct2,
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.0.match_pattern.0.included_paths.0": "/dogs/0/name",
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.0.match_pattern.0.included_paths.1": "/dogs/1/name",
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_byteMatchStatementFieldToMatchHeaderOrder(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                         acctest.Ct1,
						"statement.0.byte_match_statement.#":                  acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.#": acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.0.all_query_arguments.#":            acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.body.#":                           acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.cookies.#":                        acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.header_order.#":                   acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.0.header_order.0.oversize_handling": "MATCH",
						"statement.0.byte_match_statement.0.field_to_match.0.headers.#":                        acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.ja3_fingerprint.#":                acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.#":                      acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.method.#":                         acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.query_string.#":                   acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.single_header.#":                  acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.#":          acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.uri_path.#":                       acctest.Ct0,
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_byteMatchStatementFieldToMatchHeadersMatchPatternAll(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                         acctest.Ct1,
						"statement.0.byte_match_statement.#":                  acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.#": acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.0.all_query_arguments.#":                        acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.body.#":                                       acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.cookies.#":                                    acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.header_order.#":                               acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.headers.#":                                    acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.0.headers.0.oversize_handling":                  "MATCH",
						"statement.0.byte_match_statement.0.field_to_match.0.headers.0.match_scope":                        "ALL",
						"statement.0.byte_match_statement.0.field_to_match.0.headers.0.match_pattern.#":                    acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.0.headers.0.match_pattern.0.all.#":              acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.0.headers.0.match_pattern.0.included_headers.#": acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.headers.0.match_pattern.0.excluded_headers.#": acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.ja3_fingerprint.#":                            acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.#":                                  acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.method.#":                                     acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.query_string.#":                               acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.single_header.#":                              acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.#":                      acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.uri_path.#":                                   acctest.Ct0,
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_byteMatchStatementFieldToMatchHeadersMatchPatternIncludedHeaders(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                         acctest.Ct1,
						"statement.0.byte_match_statement.#":                  acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.#": acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.0.all_query_arguments.#":                        acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.body.#":                                       acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.cookies.#":                                    acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.header_order.#":                               acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.headers.#":                                    acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.0.headers.0.oversize_handling":                  "MATCH",
						"statement.0.byte_match_statement.0.field_to_match.0.headers.0.match_scope":                        "ALL",
						"statement.0.byte_match_statement.0.field_to_match.0.headers.0.match_pattern.#":                    acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.0.headers.0.match_pattern.0.all.#":              acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.headers.0.match_pattern.0.included_headers.#": acctest.Ct2,
						"statement.0.byte_match_statement.0.field_to_match.0.headers.0.match_pattern.0.included_headers.0": names.AttrSession,
						"statement.0.byte_match_statement.0.field_to_match.0.headers.0.match_pattern.0.included_headers.1": "session-id",
						"statement.0.byte_match_statement.0.field_to_match.0.headers.0.match_pattern.0.excluded_headers.#": acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.ja3_fingerprint.#":                            acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.#":                                  acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.method.#":                                     acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.query_string.#":                               acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.single_header.#":                              acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.#":                      acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.uri_path.#":                                   acctest.Ct0,
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_byteMatchStatementFieldToMatchHeadersMatchPatternExcludedHeaders(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                         acctest.Ct1,
						"statement.0.byte_match_statement.#":                  acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.#": acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.0.all_query_arguments.#":                        acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.body.#":                                       acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.cookies.#":                                    acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.header_order.#":                               acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.headers.#":                                    acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.0.headers.0.oversize_handling":                  "MATCH",
						"statement.0.byte_match_statement.0.field_to_match.0.headers.0.match_scope":                        "ALL",
						"statement.0.byte_match_statement.0.field_to_match.0.headers.0.match_pattern.#":                    acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.0.headers.0.match_pattern.0.all.#":              acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.headers.0.match_pattern.0.excluded_headers.#": acctest.Ct2,
						"statement.0.byte_match_statement.0.field_to_match.0.headers.0.match_pattern.0.excluded_headers.0": names.AttrSession,
						"statement.0.byte_match_statement.0.field_to_match.0.headers.0.match_pattern.0.excluded_headers.1": "session-id",
						"statement.0.byte_match_statement.0.field_to_match.0.headers.0.match_pattern.0.included_headers.#": acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.ja3_fingerprint.#":                            acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.#":                                  acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.method.#":                                     acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.query_string.#":                               acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.single_header.#":                              acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.#":                      acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.uri_path.#":                                   acctest.Ct0,
					}),
				),
			},
			{
				Config:      testAccRuleGroupConfig_byteMatchStatementFieldToMatchHeadersInvalidConfiguration(ruleGroupName),
				ExpectError: regexache.MustCompile(`argument "oversize_handling" is required`),
			},
			{
				Config: testAccRuleGroupConfig_byteMatchStatementFieldToMatchMethod(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                         acctest.Ct1,
						"statement.0.byte_match_statement.#":                  acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.#": acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.0.all_query_arguments.#":   acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.body.#":                  acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.cookies.#":               acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.header_order.#":          acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.headers.#":               acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.ja3_fingerprint.#":       acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.#":             acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.method.#":                acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.0.query_string.#":          acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.single_header.#":         acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.#": acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.uri_path.#":              acctest.Ct0,
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_byteMatchStatementFieldToMatchQueryString(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                         acctest.Ct1,
						"statement.0.byte_match_statement.#":                  acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.#": acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.0.all_query_arguments.#":   acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.body.#":                  acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.cookies.#":               acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.header_order.#":          acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.headers.#":               acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.ja3_fingerprint.#":       acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.#":             acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.method.#":                acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.query_string.#":          acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.0.single_header.#":         acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.#": acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.uri_path.#":              acctest.Ct0,
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_byteMatchStatementFieldToMatchSingleHeader(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                         acctest.Ct1,
						"statement.0.byte_match_statement.#":                  acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.#": acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.0.all_query_arguments.#":   acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.body.#":                  acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.cookies.#":               acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.header_order.#":          acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.headers.#":               acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.ja3_fingerprint.#":       acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.#":             acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.method.#":                acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.query_string.#":          acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.single_header.#":         acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.0.single_header.0.name":    "a-forty-character-long-header-name-40-40",
						"statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.#": acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.uri_path.#":              acctest.Ct0,
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_byteMatchStatementFieldToMatchSingleQueryArgument(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                         acctest.Ct1,
						"statement.0.byte_match_statement.#":                  acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.#": acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.0.all_query_arguments.#":        acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.body.#":                       acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.cookies.#":                    acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.header_order.#":               acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.headers.#":                    acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.ja3_fingerprint.#":            acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.#":                  acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.method.#":                     acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.query_string.#":               acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.single_header.#":              acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.#":      acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.0.name": "a-max-30-characters-long-name-",
						"statement.0.byte_match_statement.0.field_to_match.0.uri_path.#":                   acctest.Ct0,
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_byteMatchStatementFieldToMatchURIPath(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                         acctest.Ct1,
						"statement.0.byte_match_statement.#":                  acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.#": acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.0.all_query_arguments.#":   acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.body.#":                  acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.cookies.#":               acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.header_order.#":          acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.headers.#":               acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.ja3_fingerprint.#":       acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.#":             acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.method.#":                acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.query_string.#":          acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.single_header.#":         acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.#": acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.uri_path.#":              acctest.Ct1,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccRuleGroupImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccWAFV2RuleGroup_changeNameForceNew(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after awstypes.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	ruleGroupNewName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_basic(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &before),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, string(awstypes.ScopeRegional)),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", acctest.CtFalse),
				),
			},
			{
				Config: testAccRuleGroupConfig_basic(ruleGroupNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &after),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ruleGroupNewName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ruleGroupNewName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, string(awstypes.ScopeRegional)),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccWAFV2RuleGroup_changeCapacityForceNew(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after awstypes.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_basic(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &before),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, string(awstypes.ScopeRegional)),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", acctest.CtFalse),
				),
			},
			{
				Config: testAccRuleGroupConfig_updateCapacity(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &after),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, string(awstypes.ScopeRegional)),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccWAFV2RuleGroup_changeMetricNameForceNew(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after awstypes.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_basic(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &before),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, string(awstypes.ScopeRegional)),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", acctest.CtFalse),
				),
			},
			{
				Config: testAccRuleGroupConfig_updateMetricName(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &after),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, string(awstypes.ScopeRegional)),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "updated-friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccWAFV2RuleGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_minimal(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfwafv2.ResourceRuleGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWAFV2RuleGroup_RuleLabels(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_labels(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"rule_label.#":      acctest.Ct2,
						"rule_label.0.name": "Hashicorp:Test:Label1",
						"rule_label.1.name": "Hashicorp:Test:Label2",
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_noLabels(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"rule_label.#": acctest.Ct0,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccRuleGroupImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccWAFV2RuleGroup_geoMatchStatement(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_geoMatchStatement(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                             acctest.Ct1,
						"statement.0.geo_match_statement.#":                       acctest.Ct1,
						"statement.0.geo_match_statement.0.country_codes.#":       acctest.Ct2,
						"statement.0.geo_match_statement.0.country_codes.0":       "US",
						"statement.0.geo_match_statement.0.country_codes.1":       "NL",
						"statement.0.geo_match_statement.0.forwarded_ip_config.#": acctest.Ct0,
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_geoMatchStatementUpdate(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                             acctest.Ct1,
						"statement.0.geo_match_statement.#":                       acctest.Ct1,
						"statement.0.geo_match_statement.0.country_codes.#":       acctest.Ct3,
						"statement.0.geo_match_statement.0.country_codes.0":       "ZM",
						"statement.0.geo_match_statement.0.country_codes.1":       "EE",
						"statement.0.geo_match_statement.0.country_codes.2":       "MM",
						"statement.0.geo_match_statement.0.forwarded_ip_config.#": acctest.Ct0,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccRuleGroupImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccWAFV2RuleGroup_GeoMatchStatement_forwardedIP(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_geoMatchStatementForwardedIP(ruleGroupName, "MATCH", "X-Forwarded-For"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                                               acctest.Ct1,
						"statement.0.geo_match_statement.#":                                         acctest.Ct1,
						"statement.0.geo_match_statement.0.country_codes.#":                         acctest.Ct2,
						"statement.0.geo_match_statement.0.country_codes.0":                         "US",
						"statement.0.geo_match_statement.0.country_codes.1":                         "NL",
						"statement.0.geo_match_statement.0.forwarded_ip_config.#":                   acctest.Ct1,
						"statement.0.geo_match_statement.0.forwarded_ip_config.0.fallback_behavior": "MATCH",
						"statement.0.geo_match_statement.0.forwarded_ip_config.0.header_name":       "X-Forwarded-For",
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_geoMatchStatementForwardedIP(ruleGroupName, "NO_MATCH", "Updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                                               acctest.Ct1,
						"statement.0.geo_match_statement.#":                                         acctest.Ct1,
						"statement.0.geo_match_statement.0.country_codes.#":                         acctest.Ct2,
						"statement.0.geo_match_statement.0.country_codes.0":                         "US",
						"statement.0.geo_match_statement.0.country_codes.1":                         "NL",
						"statement.0.geo_match_statement.0.forwarded_ip_config.#":                   acctest.Ct1,
						"statement.0.geo_match_statement.0.forwarded_ip_config.0.fallback_behavior": "NO_MATCH",
						"statement.0.geo_match_statement.0.forwarded_ip_config.0.header_name":       "Updated",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccRuleGroupImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccWAFV2RuleGroup_LabelMatchStatement(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_labelMatchStatement(ruleGroupName, "LABEL", "Hashicorp:Test:Label1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                               acctest.Ct1,
						"statement.0.label_match_statement.#":       acctest.Ct1,
						"statement.0.label_match_statement.0.scope": "LABEL",
						"statement.0.label_match_statement.0.key":   "Hashicorp:Test:Label1",
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_labelMatchStatement(ruleGroupName, "NAMESPACE", "awswaf:managed:aws:bot-control:"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                               acctest.Ct1,
						"statement.0.label_match_statement.#":       acctest.Ct1,
						"statement.0.label_match_statement.0.scope": "NAMESPACE",
						"statement.0.label_match_statement.0.key":   "awswaf:managed:aws:bot-control:",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccRuleGroupImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccWAFV2RuleGroup_ipSetReferenceStatement(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_ipsetReferenceStatement(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#": acctest.Ct1,
						"statement.0.ip_set_reference_statement.#":                              acctest.Ct1,
						"statement.0.ip_set_reference_statement.0.ip_set_forwarded_ip_config.#": acctest.Ct0,
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]*regexp.Regexp{
						"statement.0.ip_set_reference_statement.0.arn": regexache.MustCompile(`regional/ipset/.+$`),
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccRuleGroupImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccWAFV2RuleGroup_IPSetReferenceStatement_ipsetForwardedIP(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_ipsetReferenceStatementIPSetForwardedIP(ruleGroupName, "MATCH", "X-Forwarded-For", "FIRST"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#": acctest.Ct1,
						"statement.0.ip_set_reference_statement.#": acctest.Ct1,
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]*regexp.Regexp{
						"statement.0.ip_set_reference_statement.0.arn": regexache.MustCompile(`regional/ipset/.+$`),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.0.ip_set_reference_statement.0.ip_set_forwarded_ip_config.#":                   acctest.Ct1,
						"statement.0.ip_set_reference_statement.0.ip_set_forwarded_ip_config.0.fallback_behavior": "MATCH",
						"statement.0.ip_set_reference_statement.0.ip_set_forwarded_ip_config.0.header_name":       "X-Forwarded-For",
						"statement.0.ip_set_reference_statement.0.ip_set_forwarded_ip_config.0.position":          "FIRST",
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_ipsetReferenceStatementIPSetForwardedIP(ruleGroupName, "NO_MATCH", "X-Forwarded-For", "LAST"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#": acctest.Ct1,
						"statement.0.ip_set_reference_statement.#": acctest.Ct1,
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]*regexp.Regexp{
						"statement.0.ip_set_reference_statement.0.arn": regexache.MustCompile(`regional/ipset/.+$`),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.0.ip_set_reference_statement.0.ip_set_forwarded_ip_config.#":                   acctest.Ct1,
						"statement.0.ip_set_reference_statement.0.ip_set_forwarded_ip_config.0.fallback_behavior": "NO_MATCH",
						"statement.0.ip_set_reference_statement.0.ip_set_forwarded_ip_config.0.header_name":       "X-Forwarded-For",
						"statement.0.ip_set_reference_statement.0.ip_set_forwarded_ip_config.0.position":          "LAST",
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_ipsetReferenceStatementIPSetForwardedIP(ruleGroupName, "MATCH", "Updated", "ANY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#": acctest.Ct1,
						"statement.0.ip_set_reference_statement.#": acctest.Ct1,
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]*regexp.Regexp{
						"statement.0.ip_set_reference_statement.0.arn": regexache.MustCompile(`regional/ipset/.+$`),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.0.ip_set_reference_statement.0.ip_set_forwarded_ip_config.#":                   acctest.Ct1,
						"statement.0.ip_set_reference_statement.0.ip_set_forwarded_ip_config.0.fallback_behavior": "MATCH",
						"statement.0.ip_set_reference_statement.0.ip_set_forwarded_ip_config.0.header_name":       "Updated",
						"statement.0.ip_set_reference_statement.0.ip_set_forwarded_ip_config.0.position":          "ANY",
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_ipsetReferenceStatement(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#": acctest.Ct1,
						"statement.0.ip_set_reference_statement.#":                              acctest.Ct1,
						"statement.0.ip_set_reference_statement.0.ip_set_forwarded_ip_config.#": acctest.Ct0,
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]*regexp.Regexp{
						"statement.0.ip_set_reference_statement.0.arn": regexache.MustCompile(`regional/ipset/.+$`),
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccRuleGroupImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccWAFV2RuleGroup_logicalRuleStatements(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_logicalStatementAnd(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                             acctest.Ct1,
						"statement.0.and_statement.#":             acctest.Ct1,
						"statement.0.and_statement.0.statement.#": acctest.Ct2,
						"statement.0.and_statement.0.statement.0.geo_match_statement.#": acctest.Ct1,
						"statement.0.and_statement.0.statement.1.geo_match_statement.#": acctest.Ct1,
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_logicalStatementNotAnd(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                                         acctest.Ct1,
						"statement.0.not_statement.#":                                         acctest.Ct1,
						"statement.0.not_statement.0.statement.#":                             acctest.Ct1,
						"statement.0.not_statement.0.statement.0.and_statement.#":             acctest.Ct1,
						"statement.0.not_statement.0.statement.0.and_statement.0.statement.#": acctest.Ct2,
						"statement.0.not_statement.0.statement.0.and_statement.0.statement.0.geo_match_statement.#": acctest.Ct1,
						"statement.0.not_statement.0.statement.0.and_statement.0.statement.1.geo_match_statement.#": acctest.Ct1,
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_logicalStatementOrNotAnd(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                                        acctest.Ct1,
						"statement.0.or_statement.#":                                         acctest.Ct1,
						"statement.0.or_statement.0.statement.#":                             acctest.Ct2,
						"statement.0.or_statement.0.statement.0.not_statement.#":             acctest.Ct1,
						"statement.0.or_statement.0.statement.0.not_statement.0.statement.#": acctest.Ct1,
						"statement.0.or_statement.0.statement.0.not_statement.0.statement.0.geo_match_statement.#": acctest.Ct1,
						"statement.0.or_statement.0.statement.1.and_statement.#":                                   acctest.Ct1,
						"statement.0.or_statement.0.statement.1.and_statement.0.statement.#":                       acctest.Ct2,
						"statement.0.or_statement.0.statement.1.and_statement.0.statement.0.geo_match_statement.#": acctest.Ct1,
						"statement.0.or_statement.0.statement.1.and_statement.0.statement.1.geo_match_statement.#": acctest.Ct1,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccRuleGroupImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccWAFV2RuleGroup_minimal(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_minimal(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, string(awstypes.ScopeRegional)),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccWAFV2RuleGroup_regexMatchStatement(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_regexMatchStatement(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                               acctest.Ct1,
						"statement.0.regex_match_statement.#":                       acctest.Ct1,
						"statement.0.regex_match_statement.0.regex_string":          "[a-z]([a-z0-9_-]*[a-z0-9])?",
						"statement.0.regex_match_statement.0.field_to_match.#":      acctest.Ct1,
						"statement.0.regex_match_statement.0.text_transformation.#": acctest.Ct1,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccRuleGroupImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccWAFV2RuleGroup_regexPatternSetReferenceStatement(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_regexPatternSetReferenceStatement(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#": acctest.Ct1,
						"statement.0.regex_pattern_set_reference_statement.#":                       acctest.Ct1,
						"statement.0.regex_pattern_set_reference_statement.0.field_to_match.#":      acctest.Ct1,
						"statement.0.regex_pattern_set_reference_statement.0.text_transformation.#": acctest.Ct1,
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]*regexp.Regexp{
						"statement.0.regex_pattern_set_reference_statement.0.arn": regexache.MustCompile(`regional/regexpatternset/.+$`),
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccRuleGroupImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccWAFV2RuleGroup_ruleAction(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_actionAllow(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, string(awstypes.ScopeRegional)),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"action.#":         acctest.Ct1,
						"action.0.allow.#": acctest.Ct1,
						"action.0.allow.0.custom_request_handling.#": acctest.Ct0,
						"action.0.block.#":                           acctest.Ct0,
						"action.0.count.#":                           acctest.Ct0,
						"action.0.captcha.#":                         acctest.Ct0,
						"action.0.challenge.#":                       acctest.Ct0,
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_actionBlock(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, string(awstypes.ScopeRegional)),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"action.#":                           acctest.Ct1,
						"action.0.allow.#":                   acctest.Ct0,
						"action.0.block.#":                   acctest.Ct1,
						"action.0.block.0.custom_response.#": acctest.Ct0,
						"action.0.count.#":                   acctest.Ct0,
						"action.0.captcha.#":                 acctest.Ct0,
						"action.0.challenge.#":               acctest.Ct0,
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_actionCount(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, string(awstypes.ScopeRegional)),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"action.#":         acctest.Ct1,
						"action.0.allow.#": acctest.Ct0,
						"action.0.block.#": acctest.Ct0,
						"action.0.count.#": acctest.Ct1,
						"action.0.count.0.custom_request_handling.#": acctest.Ct0,
						"action.0.captcha.#":                         acctest.Ct0,
						"action.0.challenge.#":                       acctest.Ct0,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccRuleGroupImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccWAFV2RuleGroup_RuleAction_customRequestHandling(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_actionAllowCustomRequestHandling(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, string(awstypes.ScopeRegional)),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"action.#":         acctest.Ct1,
						"action.0.allow.#": acctest.Ct1,
						"action.0.allow.0.custom_request_handling.#":                       acctest.Ct1,
						"action.0.allow.0.custom_request_handling.0.insert_header.#":       acctest.Ct2,
						"action.0.allow.0.custom_request_handling.0.insert_header.0.name":  "x-hdr1",
						"action.0.allow.0.custom_request_handling.0.insert_header.0.value": "test-val1",
						"action.0.allow.0.custom_request_handling.0.insert_header.1.name":  "x-hdr2",
						"action.0.allow.0.custom_request_handling.0.insert_header.1.value": "test-val2",
						"action.0.block.#":     acctest.Ct0,
						"action.0.count.#":     acctest.Ct0,
						"action.0.captcha.#":   acctest.Ct0,
						"action.0.challenge.#": acctest.Ct0,
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_actionCountCustomRequestHandling(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, string(awstypes.ScopeRegional)),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"action.#":         acctest.Ct1,
						"action.0.allow.#": acctest.Ct0,
						"action.0.block.#": acctest.Ct0,
						"action.0.count.#": acctest.Ct1,
						"action.0.count.0.custom_request_handling.#":                       acctest.Ct1,
						"action.0.count.0.custom_request_handling.0.insert_header.#":       acctest.Ct2,
						"action.0.count.0.custom_request_handling.0.insert_header.0.name":  "x-hdr1",
						"action.0.count.0.custom_request_handling.0.insert_header.0.value": "test-val1",
						"action.0.count.0.custom_request_handling.0.insert_header.1.name":  "x-hdr2",
						"action.0.count.0.custom_request_handling.0.insert_header.1.value": "test-val2",
						"action.0.captcha.#":   acctest.Ct0,
						"action.0.challenge.#": acctest.Ct0,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccRuleGroupImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccWAFV2RuleGroup_RuleAction_customResponse(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_actionBlockCustomResponse(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, string(awstypes.ScopeRegional)),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"action.#":                           acctest.Ct1,
						"action.0.allow.#":                   acctest.Ct0,
						"action.0.block.#":                   acctest.Ct1,
						"action.0.block.0.custom_response.#": acctest.Ct1,
						"action.0.block.0.custom_response.0.response_code":           "429",
						"action.0.block.0.custom_response.0.response_header.#":       acctest.Ct2,
						"action.0.block.0.custom_response.0.response_header.0.name":  "x-hdr1",
						"action.0.block.0.custom_response.0.response_header.0.value": "test-val1",
						"action.0.block.0.custom_response.0.response_header.1.name":  "x-hdr2",
						"action.0.block.0.custom_response.0.response_header.1.value": "test-val2",
						"action.0.count.#":     acctest.Ct0,
						"action.0.captcha.#":   acctest.Ct0,
						"action.0.challenge.#": acctest.Ct0,
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_actionBlockCustomResponseBody(ruleGroupName, "test_body_1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, string(awstypes.ScopeRegional)),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "custom_response_body.*", map[string]string{
						names.AttrKey:         "test_body_1",
						names.AttrContent:     "test response 1",
						names.AttrContentType: "TEXT_PLAIN",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "custom_response_body.*", map[string]string{
						names.AttrKey:         "test_body_2",
						names.AttrContent:     "<html><body>test response 2</body></html>",
						names.AttrContentType: "TEXT_HTML",
					}),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"action.#":                           acctest.Ct1,
						"action.0.allow.#":                   acctest.Ct0,
						"action.0.block.#":                   acctest.Ct1,
						"action.0.block.0.custom_response.#": acctest.Ct1,
						"action.0.block.0.custom_response.0.response_code":            "429",
						"action.0.block.0.custom_response.0.custom_response_body_key": "test_body_1",
						"action.0.count.#":     acctest.Ct0,
						"action.0.captcha.#":   acctest.Ct0,
						"action.0.challenge.#": acctest.Ct0,
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_actionBlockCustomResponseBody(ruleGroupName, "test_body_2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, string(awstypes.ScopeRegional)),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "custom_response_body.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "custom_response_body.*", map[string]string{
						names.AttrKey:         "test_body_1",
						names.AttrContent:     "test response 1",
						names.AttrContentType: "TEXT_PLAIN",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "custom_response_body.*", map[string]string{
						names.AttrKey:         "test_body_2",
						names.AttrContent:     "<html><body>test response 2</body></html>",
						names.AttrContentType: "TEXT_HTML",
					}),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"action.#":                           acctest.Ct1,
						"action.0.allow.#":                   acctest.Ct0,
						"action.0.block.#":                   acctest.Ct1,
						"action.0.block.0.custom_response.#": acctest.Ct1,
						"action.0.block.0.custom_response.0.response_code":            "429",
						"action.0.block.0.custom_response.0.custom_response_body_key": "test_body_2",
						"action.0.count.#":     acctest.Ct0,
						"action.0.captcha.#":   acctest.Ct0,
						"action.0.challenge.#": acctest.Ct0,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccRuleGroupImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccWAFV2RuleGroup_sizeConstraintStatement(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_sizeConstraintStatement(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#": acctest.Ct1,
						"statement.0.size_constraint_statement.#":                           acctest.Ct1,
						"statement.0.size_constraint_statement.0.comparison_operator":       "GT",
						"statement.0.size_constraint_statement.0.size":                      "100",
						"statement.0.size_constraint_statement.0.field_to_match.#":          acctest.Ct1,
						"statement.0.size_constraint_statement.0.field_to_match.0.method.#": acctest.Ct1,
						"statement.0.size_constraint_statement.0.text_transformation.#":     acctest.Ct1,
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_sizeConstraintStatementUpdate(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#": acctest.Ct1,
						"statement.0.size_constraint_statement.#":                                 acctest.Ct1,
						"statement.0.size_constraint_statement.0.comparison_operator":             "LT",
						"statement.0.size_constraint_statement.0.size":                            "50",
						"statement.0.size_constraint_statement.0.field_to_match.#":                acctest.Ct1,
						"statement.0.size_constraint_statement.0.field_to_match.0.query_string.#": acctest.Ct1,
						"statement.0.size_constraint_statement.0.text_transformation.#":           acctest.Ct2,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccRuleGroupImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccWAFV2RuleGroup_sqliMatchStatement(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_sqliMatchStatement(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                         acctest.Ct1,
						"statement.0.sqli_match_statement.#":                  acctest.Ct1,
						"statement.0.sqli_match_statement.0.field_to_match.#": acctest.Ct1,
						"statement.0.sqli_match_statement.0.field_to_match.0.all_query_arguments.#": acctest.Ct1,
						"statement.0.sqli_match_statement.0.text_transformation.#":                  acctest.Ct2,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.statement.0.sqli_match_statement.0.text_transformation.*", map[string]string{
						names.AttrPriority: "5",
						names.AttrType:     "URL_DECODE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.statement.0.sqli_match_statement.0.text_transformation.*", map[string]string{
						names.AttrPriority: acctest.Ct2,
						names.AttrType:     "LOWERCASE",
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_sqliMatchStatementUpdate(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                                acctest.Ct1,
						"statement.0.sqli_match_statement.#":                         acctest.Ct1,
						"statement.0.sqli_match_statement.0.field_to_match.#":        acctest.Ct1,
						"statement.0.sqli_match_statement.0.field_to_match.0.body.#": acctest.Ct1,
						"statement.0.sqli_match_statement.0.text_transformation.#":   acctest.Ct3,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.statement.0.sqli_match_statement.0.text_transformation.*", map[string]string{
						names.AttrPriority: "5",
						names.AttrType:     "URL_DECODE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.statement.0.sqli_match_statement.0.text_transformation.*", map[string]string{
						names.AttrPriority: acctest.Ct4,
						names.AttrType:     "HTML_ENTITY_DECODE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.statement.0.sqli_match_statement.0.text_transformation.*", map[string]string{
						names.AttrPriority: acctest.Ct3,
						names.AttrType:     "COMPRESS_WHITE_SPACE",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccRuleGroupImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccWAFV2RuleGroup_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_tags1(ruleGroupName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccRuleGroupImportStateIdFunc(resourceName),
			},
			{
				Config: testAccRuleGroupConfig_tags2(ruleGroupName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccRuleGroupConfig_tags1(ruleGroupName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccWAFV2RuleGroup_xssMatchStatement(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_xssMatchStatement(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                               acctest.Ct1,
						"statement.0.xss_match_statement.#":                         acctest.Ct1,
						"statement.0.xss_match_statement.0.field_to_match.#":        acctest.Ct1,
						"statement.0.xss_match_statement.0.field_to_match.0.body.#": acctest.Ct1,
						"statement.0.xss_match_statement.0.text_transformation.#":   acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.statement.0.xss_match_statement.0.text_transformation.*", map[string]string{
						names.AttrPriority: acctest.Ct2,
						names.AttrType:     "NONE",
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_xssMatchStatementUpdate(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                               acctest.Ct1,
						"statement.0.xss_match_statement.#":                         acctest.Ct1,
						"statement.0.xss_match_statement.0.field_to_match.#":        acctest.Ct1,
						"statement.0.xss_match_statement.0.field_to_match.0.body.#": acctest.Ct1,
						"statement.0.xss_match_statement.0.text_transformation.#":   acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.statement.0.xss_match_statement.0.text_transformation.*", map[string]string{
						names.AttrPriority: acctest.Ct2,
						names.AttrType:     "URL_DECODE",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccRuleGroupImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccWAFV2RuleGroup_rateBasedStatement(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_rateBasedStatement(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#": acctest.Ct1,
						"statement.0.rate_based_statement.0.custom_key.#":           acctest.Ct0,
						"statement.0.rate_based_statement.0.aggregate_key_type":     "IP",
						"statement.0.rate_based_statement.0.evaluation_window_sec":  "600",
						"statement.0.rate_based_statement.0.forwarded_ip_config.#":  acctest.Ct0,
						"statement.0.rate_based_statement.0.limit":                  "50000",
						"statement.0.rate_based_statement.0.scope_down_statement.#": acctest.Ct0,
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_rateBasedStatement_forwardedIPConfig(ruleGroupName, "MATCH", "X-Forwarded-For"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                                                acctest.Ct1,
						"statement.0.rate_based_statement.#":                                         acctest.Ct1,
						"statement.0.rate_based_statement.0.custom_key.#":                            acctest.Ct0,
						"statement.0.rate_based_statement.0.aggregate_key_type":                      "FORWARDED_IP",
						"statement.0.rate_based_statement.0.evaluation_window_sec":                   "300",
						"statement.0.rate_based_statement.0.forwarded_ip_config.#":                   acctest.Ct1,
						"statement.0.rate_based_statement.0.forwarded_ip_config.0.fallback_behavior": "MATCH",
						"statement.0.rate_based_statement.0.forwarded_ip_config.0.header_name":       "X-Forwarded-For",
						"statement.0.rate_based_statement.0.limit":                                   "50000",
						"statement.0.rate_based_statement.0.scope_down_statement.#":                  acctest.Ct0,
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_rateBasedStatement_forwardedIPConfig(ruleGroupName, "NO_MATCH", "Updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                                                acctest.Ct1,
						"statement.0.rate_based_statement.#":                                         acctest.Ct1,
						"statement.0.rate_based_statement.0.custom_key.#":                            acctest.Ct0,
						"statement.0.rate_based_statement.0.aggregate_key_type":                      "FORWARDED_IP",
						"statement.0.rate_based_statement.0.evaluation_window_sec":                   "300",
						"statement.0.rate_based_statement.0.forwarded_ip_config.#":                   acctest.Ct1,
						"statement.0.rate_based_statement.0.forwarded_ip_config.0.fallback_behavior": "NO_MATCH",
						"statement.0.rate_based_statement.0.forwarded_ip_config.0.header_name":       "Updated",
						"statement.0.rate_based_statement.0.limit":                                   "50000",
						"statement.0.rate_based_statement.0.scope_down_statement.#":                  acctest.Ct0,
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_rateBasedStatement_customKeysBasic(ruleGroupName, "cookie", "testcookie"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                                                    acctest.Ct1,
						"statement.0.rate_based_statement.#":                                             acctest.Ct1,
						"statement.0.rate_based_statement.0.custom_key.#":                                acctest.Ct1,
						"statement.0.rate_based_statement.0.aggregate_key_type":                          "CUSTOM_KEYS",
						"statement.0.rate_based_statement.0.evaluation_window_sec":                       "300",
						"statement.0.rate_based_statement.0.forwarded_ip_config.#":                       acctest.Ct0,
						"statement.0.rate_based_statement.0.limit":                                       "50000",
						"statement.0.rate_based_statement.0.scope_down_statement.#":                      acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.cookie.#":                       acctest.Ct1,
						"statement.0.rate_based_statement.0.custom_key.0.forwarded_ip.#":                 acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.http_method.#":                  acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.header.#":                       acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.ip.#":                           acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.label_namespace.#":              acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.query_argument.#":               acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.query_string.#":                 acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.uri_path.#":                     acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.cookie.0.text_transformation.#": acctest.Ct1,
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_rateBasedStatement_customKeysForwardedIP(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                                       acctest.Ct1,
						"statement.0.rate_based_statement.#":                                acctest.Ct1,
						"statement.0.rate_based_statement.0.custom_key.#":                   acctest.Ct2,
						"statement.0.rate_based_statement.0.aggregate_key_type":             "CUSTOM_KEYS",
						"statement.0.rate_based_statement.0.evaluation_window_sec":          "300",
						"statement.0.rate_based_statement.0.forwarded_ip_config.#":          acctest.Ct1,
						"statement.0.rate_based_statement.0.limit":                          "50000",
						"statement.0.rate_based_statement.0.scope_down_statement.#":         acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.cookie.#":          acctest.Ct1,
						"statement.0.rate_based_statement.0.custom_key.0.forwarded_ip.#":    acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.http_method.#":     acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.header.#":          acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.ip.#":              acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.label_namespace.#": acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.query_argument.#":  acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.query_string.#":    acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.uri_path.#":        acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.1.forwarded_ip.#":    acctest.Ct1,
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_rateBasedStatement_customKeysHTTPMethod(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                                       acctest.Ct1,
						"statement.0.rate_based_statement.#":                                acctest.Ct1,
						"statement.0.rate_based_statement.0.custom_key.#":                   acctest.Ct1,
						"statement.0.rate_based_statement.0.aggregate_key_type":             "CUSTOM_KEYS",
						"statement.0.rate_based_statement.0.evaluation_window_sec":          "300",
						"statement.0.rate_based_statement.0.forwarded_ip_config.#":          acctest.Ct0,
						"statement.0.rate_based_statement.0.limit":                          "50000",
						"statement.0.rate_based_statement.0.scope_down_statement.#":         acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.cookie.#":          acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.forwarded_ip.#":    acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.http_method.#":     acctest.Ct1,
						"statement.0.rate_based_statement.0.custom_key.0.header.#":          acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.ip.#":              acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.label_namespace.#": acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.query_argument.#":  acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.query_string.#":    acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.uri_path.#":        acctest.Ct0,
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_rateBasedStatement_customKeysBasic(ruleGroupName, names.AttrHeader, "x-forwrded-for"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                                                    acctest.Ct1,
						"statement.0.rate_based_statement.#":                                             acctest.Ct1,
						"statement.0.rate_based_statement.0.custom_key.#":                                acctest.Ct1,
						"statement.0.rate_based_statement.0.aggregate_key_type":                          "CUSTOM_KEYS",
						"statement.0.rate_based_statement.0.evaluation_window_sec":                       "300",
						"statement.0.rate_based_statement.0.forwarded_ip_config.#":                       acctest.Ct0,
						"statement.0.rate_based_statement.0.limit":                                       "50000",
						"statement.0.rate_based_statement.0.scope_down_statement.#":                      acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.cookie.#":                       acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.forwarded_ip.#":                 acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.http_method.#":                  acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.header.#":                       acctest.Ct1,
						"statement.0.rate_based_statement.0.custom_key.0.ip.#":                           acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.label_namespace.#":              acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.query_argument.#":               acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.query_string.#":                 acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.uri_path.#":                     acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.header.0.text_transformation.#": acctest.Ct1,
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_rateBasedStatement_customKeysIP(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                                       acctest.Ct1,
						"statement.0.rate_based_statement.#":                                acctest.Ct1,
						"statement.0.rate_based_statement.0.custom_key.#":                   acctest.Ct2,
						"statement.0.rate_based_statement.0.aggregate_key_type":             "CUSTOM_KEYS",
						"statement.0.rate_based_statement.0.evaluation_window_sec":          "300",
						"statement.0.rate_based_statement.0.forwarded_ip_config.#":          acctest.Ct0,
						"statement.0.rate_based_statement.0.limit":                          "50000",
						"statement.0.rate_based_statement.0.scope_down_statement.#":         acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.cookie.#":          acctest.Ct1,
						"statement.0.rate_based_statement.0.custom_key.0.forwarded_ip.#":    acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.http_method.#":     acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.header.#":          acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.ip.#":              acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.label_namespace.#": acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.query_argument.#":  acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.query_string.#":    acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.uri_path.#":        acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.1.ip.#":              acctest.Ct1,
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_rateBasedStatement_customKeysBasic(ruleGroupName, "query_argument", names.AttrKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                                                            acctest.Ct1,
						"statement.0.rate_based_statement.#":                                                     acctest.Ct1,
						"statement.0.rate_based_statement.0.custom_key.#":                                        acctest.Ct1,
						"statement.0.rate_based_statement.0.aggregate_key_type":                                  "CUSTOM_KEYS",
						"statement.0.rate_based_statement.0.evaluation_window_sec":                               "300",
						"statement.0.rate_based_statement.0.forwarded_ip_config.#":                               acctest.Ct0,
						"statement.0.rate_based_statement.0.limit":                                               "50000",
						"statement.0.rate_based_statement.0.scope_down_statement.#":                              acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.cookie.#":                               acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.forwarded_ip.#":                         acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.http_method.#":                          acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.header.#":                               acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.ip.#":                                   acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.label_namespace.#":                      acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.query_argument.#":                       acctest.Ct1,
						"statement.0.rate_based_statement.0.custom_key.0.query_string.#":                         acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.uri_path.#":                             acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.query_argument.0.text_transformation.#": acctest.Ct1,
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_rateBasedStatement_customKeysMinimal(ruleGroupName, "query_string"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                                                          acctest.Ct1,
						"statement.0.rate_based_statement.#":                                                   acctest.Ct1,
						"statement.0.rate_based_statement.0.custom_key.#":                                      acctest.Ct1,
						"statement.0.rate_based_statement.0.aggregate_key_type":                                "CUSTOM_KEYS",
						"statement.0.rate_based_statement.0.evaluation_window_sec":                             "300",
						"statement.0.rate_based_statement.0.forwarded_ip_config.#":                             acctest.Ct0,
						"statement.0.rate_based_statement.0.limit":                                             "50000",
						"statement.0.rate_based_statement.0.scope_down_statement.#":                            acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.cookie.#":                             acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.forwarded_ip.#":                       acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.http_method.#":                        acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.header.#":                             acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.ip.#":                                 acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.label_namespace.#":                    acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.query_argument.#":                     acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.query_string.#":                       acctest.Ct1,
						"statement.0.rate_based_statement.0.custom_key.0.uri_path.#":                           acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.query_string.0.text_transformation.#": acctest.Ct1,
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_rateBasedStatement_customKeysMinimal(ruleGroupName, "uri_path"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                                                      acctest.Ct1,
						"statement.0.rate_based_statement.#":                                               acctest.Ct1,
						"statement.0.rate_based_statement.0.custom_key.#":                                  acctest.Ct1,
						"statement.0.rate_based_statement.0.aggregate_key_type":                            "CUSTOM_KEYS",
						"statement.0.rate_based_statement.0.evaluation_window_sec":                         "300",
						"statement.0.rate_based_statement.0.forwarded_ip_config.#":                         acctest.Ct0,
						"statement.0.rate_based_statement.0.limit":                                         "50000",
						"statement.0.rate_based_statement.0.scope_down_statement.#":                        acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.cookie.#":                         acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.forwarded_ip.#":                   acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.http_method.#":                    acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.header.#":                         acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.ip.#":                             acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.label_namespace.#":                acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.query_argument.#":                 acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.query_string.#":                   acctest.Ct0,
						"statement.0.rate_based_statement.0.custom_key.0.uri_path.#":                       acctest.Ct1,
						"statement.0.rate_based_statement.0.custom_key.0.uri_path.0.text_transformation.#": acctest.Ct1,
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_rateBasedStatement_customKeysMaxKeys(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                               acctest.Ct1,
						"statement.0.rate_based_statement.#":                        acctest.Ct1,
						"statement.0.rate_based_statement.0.custom_key.#":           "5",
						"statement.0.rate_based_statement.0.aggregate_key_type":     "CUSTOM_KEYS",
						"statement.0.rate_based_statement.0.evaluation_window_sec":  "300",
						"statement.0.rate_based_statement.0.forwarded_ip_config.#":  acctest.Ct0,
						"statement.0.rate_based_statement.0.limit":                  "50000",
						"statement.0.rate_based_statement.0.scope_down_statement.#": acctest.Ct0,
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_rateBasedStatement_update(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                                                                     acctest.Ct1,
						"statement.0.rate_based_statement.#":                                                              acctest.Ct1,
						"statement.0.rate_based_statement.0.custom_key.#":                                                 acctest.Ct0,
						"statement.0.rate_based_statement.0.aggregate_key_type":                                           "IP",
						"statement.0.rate_based_statement.0.evaluation_window_sec":                                        "300",
						"statement.0.rate_based_statement.0.forwarded_ip_config.#":                                        acctest.Ct0,
						"statement.0.rate_based_statement.0.limit":                                                        "10000",
						"statement.0.rate_based_statement.0.scope_down_statement.#":                                       acctest.Ct1,
						"statement.0.rate_based_statement.0.scope_down_statement.0.geo_match_statement.#":                 acctest.Ct1,
						"statement.0.rate_based_statement.0.scope_down_statement.0.geo_match_statement.0.country_codes.#": acctest.Ct2,
						"statement.0.rate_based_statement.0.scope_down_statement.0.geo_match_statement.0.country_codes.0": "US",
						"statement.0.rate_based_statement.0.scope_down_statement.0.geo_match_statement.0.country_codes.1": "NL",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccRuleGroupImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccWAFV2RuleGroup_RateBased_maxNested(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_multipleNestedRateBasedStatements(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                                                                                      acctest.Ct1,
						"statement.0.rate_based_statement.#":                                                                               acctest.Ct1,
						"statement.0.rate_based_statement.0.limit":                                                                         "300",
						"statement.0.rate_based_statement.0.aggregate_key_type":                                                            "IP",
						"statement.0.rate_based_statement.0.scope_down_statement.#":                                                        acctest.Ct1,
						"statement.0.rate_based_statement.0.scope_down_statement.0.not_statement.#":                                        acctest.Ct1,
						"statement.0.rate_based_statement.0.scope_down_statement.0.not_statement.0.statement.#":                            acctest.Ct1,
						"statement.0.rate_based_statement.0.scope_down_statement.0.not_statement.0.statement.0.or_statement.#":             acctest.Ct1,
						"statement.0.rate_based_statement.0.scope_down_statement.0.not_statement.0.statement.0.or_statement.0.statement.#": acctest.Ct3,
						"statement.0.rate_based_statement.0.scope_down_statement.0.not_statement.0.statement.0.or_statement.0.statement.0.regex_pattern_set_reference_statement.#": acctest.Ct1,
						"statement.0.rate_based_statement.0.scope_down_statement.0.not_statement.0.statement.0.or_statement.0.statement.1.regex_match_statement.#":                 acctest.Ct1,
						"statement.0.rate_based_statement.0.scope_down_statement.0.not_statement.0.statement.0.or_statement.0.statement.2.ip_set_reference_statement.#":            acctest.Ct1,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccWebACLImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccWAFV2RuleGroup_Operators_maxNested(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_multipleNestedOperatorStatements(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                                                                    acctest.Ct1,
						"statement.0.and_statement.#":                                                                    acctest.Ct1,
						"statement.0.and_statement.0.statement.#":                                                        acctest.Ct2,
						"statement.0.and_statement.0.statement.0.not_statement.#":                                        acctest.Ct1,
						"statement.0.and_statement.0.statement.0.not_statement.0.statement.#":                            acctest.Ct1,
						"statement.0.and_statement.0.statement.0.not_statement.0.statement.0.or_statement.#":             acctest.Ct1,
						"statement.0.and_statement.0.statement.0.not_statement.0.statement.0.or_statement.0.statement.#": acctest.Ct3,
						"statement.0.and_statement.0.statement.0.not_statement.0.statement.0.or_statement.0.statement.0.regex_pattern_set_reference_statement.#": acctest.Ct1,
						"statement.0.and_statement.0.statement.0.not_statement.0.statement.0.or_statement.0.statement.1.regex_match_statement.#":                 acctest.Ct1,
						"statement.0.and_statement.0.statement.0.not_statement.0.statement.0.or_statement.0.statement.2.ip_set_reference_statement.#":            acctest.Ct1,
						"statement.0.and_statement.0.statement.1.geo_match_statement.#":                                                                          acctest.Ct1,
						"statement.0.and_statement.0.statement.1.geo_match_statement.0.country_codes.0":                                                          "NL",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccWebACLImportStateIdFunc(resourceName),
			},
		},
	})
}

func testAccPreCheckScopeRegional(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).WAFV2Client(ctx)

	input := &wafv2.ListRuleGroupsInput{
		Scope: awstypes.ScopeRegional,
	}

	_, err := conn.ListRuleGroups(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckRuleGroupDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_wafv2_rule_group" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).WAFV2Client(ctx)

			_, err := tfwafv2.FindRuleGroupByThreePartKey(ctx, conn, rs.Primary.ID, rs.Primary.Attributes[names.AttrName], rs.Primary.Attributes[names.AttrScope])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("WAFv2 RuleGroup %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckRuleGroupExists(ctx context.Context, n string, v *awstypes.RuleGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFV2Client(ctx)

		output, err := tfwafv2.FindRuleGroupByThreePartKey(ctx, conn, rs.Primary.ID, rs.Primary.Attributes[names.AttrName], rs.Primary.Attributes[names.AttrScope])

		if err != nil {
			return err
		}

		*v = *output.RuleGroup

		return nil
	}
}

func testAccRuleGroupImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s/%s", rs.Primary.ID, rs.Primary.Attributes[names.AttrName], rs.Primary.Attributes[names.AttrScope]), nil
	}
}

func testAccRuleGroupConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity    = 2
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_namePrefix(namePrefix string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity    = 2
  name_prefix = %[1]q
  description = "test"
  scope       = "REGIONAL"

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, namePrefix)
}

func testAccRuleGroupConfig_nameGenerated() string {
	return `
resource "aws_wafv2_rule_group" "test" {
  capacity    = 2
  description = "test"
  scope       = "REGIONAL"

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`
}

func testAccRuleGroupConfig_basicUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity    = 50
  name        = %[1]q
  description = "Updated"
  scope       = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      count {}
    }

    statement {
      geo_match_statement {
        country_codes = ["US", "NL"]
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_updateMultiples(rName string, ruleName1, ruleName2 string, priority1, priority2 int) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity    = 50
  name        = %[1]q
  description = "Updated"
  scope       = "REGIONAL"

  rule {
    name     = %[2]q
    priority = %[3]d

    action {
      count {}
    }

    statement {
      geo_match_statement {
        country_codes = ["US", "NL"]
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = %[2]q
      sampled_requests_enabled   = false
    }
  }

  rule {
    name     = %[4]q
    priority = %[5]d

    action {
      block {}
    }

    statement {
      size_constraint_statement {
        comparison_operator = "LT"
        size                = 50

        field_to_match {
          query_string {}
        }

        text_transformation {
          priority = 5
          type     = "NONE"
        }

        text_transformation {
          priority = 2
          type     = "CMD_LINE"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = %[4]q
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName, ruleName1, priority1, ruleName2, priority2)
}

func testAccRuleGroupConfig_updateCapacity(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity    = 3
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_updateMetricName(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity    = 2
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "updated-friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_minimal(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name     = %[1]q
  scope    = "REGIONAL"

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_actionAllow(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name     = %[1]q
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      geo_match_statement {
        country_codes = ["US", "NL"]
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_actionAllowCustomRequestHandling(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name     = %[1]q
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {
        custom_request_handling {
          insert_header {
            name  = "x-hdr1"
            value = "test-val1"
          }

          insert_header {
            name  = "x-hdr2"
            value = "test-val2"
          }
        }
      }
    }

    statement {
      geo_match_statement {
        country_codes = ["US", "NL"]
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_actionBlock(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name     = %[1]q
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      block {}
    }

    statement {
      geo_match_statement {
        country_codes = ["US", "NL"]
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_actionBlockCustomResponse(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name     = %[1]q
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      block {
        custom_response {
          response_code = 429
          response_header {
            name  = "x-hdr1"
            value = "test-val1"
          }

          response_header {
            name  = "x-hdr2"
            value = "test-val2"
          }
        }
      }
    }

    statement {
      geo_match_statement {
        country_codes = ["US", "NL"]
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_actionBlockCustomResponseBody(rName string, customBodyKey string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name     = %[1]q
  scope    = "REGIONAL"
  custom_response_body {
    key          = "test_body_1"
    content      = "test response 1"
    content_type = "TEXT_PLAIN"
  }
  custom_response_body {
    key          = "test_body_2"
    content      = "<html><body>test response 2</body></html>"
    content_type = "TEXT_HTML"
  }
  rule {
    name     = "rule-1"
    priority = 1
    action {
      block {
        custom_response {
          response_code            = 429
          custom_response_body_key = %[2]q
        }
      }
    }
    statement {
      geo_match_statement {
        country_codes = ["US", "NL"]
      }
    }
    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }
  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName, customBodyKey)
}

func testAccRuleGroupConfig_actionCount(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name     = %[1]q
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      count {}
    }

    statement {
      geo_match_statement {
        country_codes = ["US", "NL"]
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_actionCountCustomRequestHandling(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name     = %[1]q
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      count {
        custom_request_handling {
          insert_header {
            name  = "x-hdr1"
            value = "test-val1"
          }

          insert_header {
            name  = "x-hdr2"
            value = "test-val2"
          }
        }
      }
    }

    statement {
      geo_match_statement {
        country_codes = ["US", "NL"]
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_byteMatchStatement(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 300
  name     = %[1]q
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      byte_match_statement {
        positional_constraint = "CONTAINS"
        search_string         = "word"

        field_to_match {
          all_query_arguments {}
        }

        text_transformation {
          priority = 5
          type     = "NONE"
        }

        text_transformation {
          priority = 2
          type     = "LOWERCASE"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_byteMatchStatementUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 30
  name     = %[1]q
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      byte_match_statement {
        positional_constraint = "EXACTLY"
        search_string         = "sentence"

        field_to_match {
          all_query_arguments {}
        }

        text_transformation {
          priority = 3
          type     = "CMD_LINE"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_byteMatchStatementFieldToMatchAllQueryArguments(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 30
  name     = %[1]q
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      byte_match_statement {
        positional_constraint = "CONTAINS"
        search_string         = "word"

        field_to_match {
          all_query_arguments {}
        }

        text_transformation {
          priority = 5
          type     = "NONE"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_byteMatchStatementFieldToMatchBody(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 15
  name     = %[1]q
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      byte_match_statement {
        positional_constraint = "CONTAINS"
        search_string         = "word"

        field_to_match {
          body {}
        }

        text_transformation {
          priority = 1
          type     = "NONE"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_byteMatchStatementFieldToMatchJSONBody(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 20
  name     = %[1]q
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      byte_match_statement {
        positional_constraint = "CONTAINS"
        search_string         = "Clifford"

        field_to_match {
          json_body {
            match_scope               = "VALUE"
            invalid_fallback_behavior = "MATCH"
            oversize_handling         = "CONTINUE"
            match_pattern {
              included_paths = ["/dogs/0/name", "/dogs/1/name"]
            }
          }
        }

        text_transformation {
          priority = 1
          type     = "NONE"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_byteMatchStatementFieldToMatchHeadersInvalidConfiguration(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 50
  name     = %[1]q
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      byte_match_statement {
        positional_constraint = "CONTAINS"
        search_string         = "word"

        field_to_match {
          headers {
            match_scope = "ALL"
            match_pattern {
              all {}
            }
          }
        }

        text_transformation {
          priority = 1
          type     = "NONE"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_byteMatchStatementFieldToMatchHeaderOrder(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 50
  name     = %[1]q
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      byte_match_statement {
        search_string = "host:user-agent:accept:authorization:referer"
        field_to_match {
          header_order {
            oversize_handling = "MATCH"
          }
        }
        text_transformation {
          priority = 0
          type     = "NONE"
        }
        positional_constraint = "STARTS_WITH"
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_byteMatchStatementFieldToMatchHeadersMatchPatternAll(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 50
  name     = %[1]q
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      byte_match_statement {
        positional_constraint = "CONTAINS"
        search_string         = "word"

        field_to_match {
          headers {
            match_scope = "ALL"
            match_pattern {
              all {}
            }
            oversize_handling = "MATCH"
          }
        }

        text_transformation {
          priority = 1
          type     = "NONE"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_byteMatchStatementFieldToMatchHeadersMatchPatternIncludedHeaders(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 50
  name     = %[1]q
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      byte_match_statement {
        positional_constraint = "CONTAINS"
        search_string         = "word"

        field_to_match {
          headers {
            match_scope = "ALL"
            match_pattern {
              included_headers = ["session", "session-id"]
            }
            oversize_handling = "MATCH"
          }
        }

        text_transformation {
          priority = 1
          type     = "NONE"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_byteMatchStatementFieldToMatchHeadersMatchPatternExcludedHeaders(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 50
  name     = %[1]q
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      byte_match_statement {
        positional_constraint = "CONTAINS"
        search_string         = "word"

        field_to_match {
          headers {
            match_scope = "ALL"
            match_pattern {
              excluded_headers = ["session", "session-id"]
            }
            oversize_handling = "MATCH"
          }
        }

        text_transformation {
          priority = 1
          type     = "NONE"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_byteMatchStatementFieldToMatchMethod(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 15
  name     = %[1]q
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      byte_match_statement {
        positional_constraint = "CONTAINS"
        search_string         = "word"

        field_to_match {
          method {}
        }

        text_transformation {
          priority = 1
          type     = "NONE"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_byteMatchStatementFieldToMatchQueryString(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 15
  name     = %[1]q
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      byte_match_statement {
        positional_constraint = "CONTAINS"
        search_string         = "word"

        field_to_match {
          query_string {}
        }

        text_transformation {
          priority = 1
          type     = "NONE"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_byteMatchStatementFieldToMatchCookies(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 15
  name     = %[1]q
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      byte_match_statement {
        positional_constraint = "CONTAINS"
        search_string         = "word"

        field_to_match {
          cookies {
            match_pattern {
              included_cookies = ["test", "cookie_test"]
            }
            match_scope       = "ALL"
            oversize_handling = "NO_MATCH"
          }
        }

        text_transformation {
          priority = 1
          type     = "NONE"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_byteMatchStatementFieldToMatchSingleHeader(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 15
  name     = %[1]q
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      byte_match_statement {
        positional_constraint = "CONTAINS"
        search_string         = "word"

        field_to_match {
          single_header {
            name = "a-forty-character-long-header-name-40-40"
          }
        }

        text_transformation {
          priority = 1
          type     = "NONE"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_byteMatchStatementFieldToMatchSingleQueryArgument(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 30
  name     = %[1]q
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      byte_match_statement {
        positional_constraint = "CONTAINS"
        search_string         = "word"

        field_to_match {
          single_query_argument {
            name = "a-max-30-characters-long-name-"
          }
        }

        text_transformation {
          priority = 1
          type     = "NONE"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_byteMatchStatementFieldToMatchURIPath(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 15
  name     = %[1]q
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      byte_match_statement {
        positional_constraint = "CONTAINS"
        search_string         = "word"

        field_to_match {
          uri_path {}
        }

        text_transformation {
          priority = 1
          type     = "NONE"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_ipsetReferenceStatement(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_ip_set" "test" {
  name               = "ip-set-%[1]s"
  scope              = "REGIONAL"
  ip_address_version = "IPV4"
  addresses          = ["1.1.1.1/32", "2.2.2.2/32"]
}

resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name     = %[1]q
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      ip_set_reference_statement {
        arn = aws_wafv2_ip_set.test.arn
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_ipsetReferenceStatementIPSetForwardedIP(rName, fallbackBehavior, headerName, position string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_ip_set" "test" {
  name               = "ip-set-%[1]s"
  scope              = "REGIONAL"
  ip_address_version = "IPV4"
  addresses          = ["1.1.1.1/32", "2.2.2.2/32"]
}

resource "aws_wafv2_rule_group" "test" {
  capacity = 5
  name     = %[1]q
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      ip_set_reference_statement {
        arn = aws_wafv2_ip_set.test.arn
        ip_set_forwarded_ip_config {
          fallback_behavior = %[2]q
          header_name       = %[3]q
          position          = %[4]q
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName, fallbackBehavior, headerName, position)
}

func testAccRuleGroupConfig_geoMatchStatement(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name     = %[1]q
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      geo_match_statement {
        country_codes = ["US", "NL"]
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_geoMatchStatementForwardedIP(rName, fallbackBehavior, headerName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name     = %[1]q
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      geo_match_statement {
        country_codes = ["US", "NL"]
        forwarded_ip_config {
          fallback_behavior = %[2]q
          header_name       = %[3]q
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName, fallbackBehavior, headerName)
}

func testAccRuleGroupConfig_geoMatchStatementUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name     = %[1]q
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      geo_match_statement {
        country_codes = ["ZM", "EE", "MM"]
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_labelMatchStatement(rName string, scope string, key string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name     = %[1]q
  scope    = "REGIONAL"
  rule {
    name     = "rule-1"
    priority = 1
    action {
      allow {}
    }
    statement {
      label_match_statement {
        scope = %[2]q
        key   = %[3]q
      }
    }
    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }
  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName, scope, key)
}

func testAccRuleGroupConfig_labels(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name     = %[1]q
  scope    = "REGIONAL"
  rule {
    name     = "rule-1"
    priority = 1
    action {
      allow {}
    }
    rule_label {
      name = "Hashicorp:Test:Label1"
    }
    rule_label {
      name = "Hashicorp:Test:Label2"
    }
    statement {
      geo_match_statement {
        country_codes = ["US", "NL"]
      }
    }
    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }
  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_noLabels(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name     = %[1]q
  scope    = "REGIONAL"
  rule {
    name     = "rule-1"
    priority = 1
    action {
      allow {}
    }
    statement {
      geo_match_statement {
        country_codes = ["US", "NL"]
      }
    }
    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }
  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_logicalStatementAnd(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name     = %[1]q
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      and_statement {
        statement {
          geo_match_statement {
            country_codes = ["US"]
          }
        }

        statement {
          geo_match_statement {
            country_codes = ["NL"]
          }
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_logicalStatementNotAnd(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name     = %[1]q
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      not_statement {
        statement {
          and_statement {
            statement {
              geo_match_statement {
                country_codes = ["US"]
              }
            }

            statement {
              geo_match_statement {
                country_codes = ["NL"]
              }
            }
          }
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_logicalStatementOrNotAnd(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 3
  name     = %[1]q
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      or_statement {
        statement {
          not_statement {
            statement {
              geo_match_statement {
                country_codes = ["DE"]
              }
            }
          }
        }

        statement {
          and_statement {
            statement {
              geo_match_statement {
                country_codes = ["US"]
              }
            }

            statement {
              geo_match_statement {
                country_codes = ["NL"]
              }
            }
          }
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_regexMatchStatement(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 50
  name     = %[1]q
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      regex_match_statement {
        regex_string = "[a-z]([a-z0-9_-]*[a-z0-9])?"

        field_to_match {
          body {}
        }

        text_transformation {
          priority = 2
          type     = "NONE"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_regexPatternSetReferenceStatement(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_regex_pattern_set" "test" {
  name  = "regex-pattern-set-%[1]s"
  scope = "REGIONAL"

  regular_expression {
    regex_string = "[a-z]([a-z0-9_-]*[a-z0-9])?"
  }
}

resource "aws_wafv2_rule_group" "test" {
  capacity = 50
  name     = %[1]q
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      regex_pattern_set_reference_statement {
        arn = aws_wafv2_regex_pattern_set.test.arn

        field_to_match {
          body {}
        }

        text_transformation {
          priority = 2
          type     = "NONE"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_sizeConstraintStatement(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 30
  name     = %[1]q
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      size_constraint_statement {
        comparison_operator = "GT"
        size                = 100

        field_to_match {
          method {}
        }

        text_transformation {
          priority = 5
          type     = "NONE"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_sizeConstraintStatementUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 30
  name     = %[1]q
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      size_constraint_statement {
        comparison_operator = "LT"
        size                = 50

        field_to_match {
          query_string {}
        }

        text_transformation {
          priority = 5
          type     = "NONE"
        }

        text_transformation {
          priority = 2
          type     = "CMD_LINE"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_sqliMatchStatement(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 300
  name     = %[1]q
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      sqli_match_statement {
        field_to_match {
          all_query_arguments {}
        }

        text_transformation {
          priority = 5
          type     = "URL_DECODE"
        }

        text_transformation {
          priority = 2
          type     = "LOWERCASE"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_sqliMatchStatementUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 300
  name     = %[1]q
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      sqli_match_statement {
        field_to_match {
          body {}
        }

        text_transformation {
          priority = 5
          type     = "URL_DECODE"
        }

        text_transformation {
          priority = 4
          type     = "HTML_ENTITY_DECODE"
        }

        text_transformation {
          priority = 3
          type     = "COMPRESS_WHITE_SPACE"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_xssMatchStatement(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 300
  name     = %[1]q
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      block {}
    }

    statement {
      xss_match_statement {
        field_to_match {
          body {}
        }

        text_transformation {
          priority = 2
          type     = "NONE"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_xssMatchStatementUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 300
  name     = %[1]q
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {}
    }

    statement {
      xss_match_statement {
        field_to_match {
          body {}
        }

        text_transformation {
          priority = 2
          type     = "URL_DECODE"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_rateBasedStatement(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 100
  name     = %[1]q
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      count {}
    }

    statement {
      rate_based_statement {
        evaluation_window_sec = 600
        limit                 = 50000
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_rateBasedStatement_forwardedIPConfig(rName, fallbackBehavior, headerName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 100
  name     = %[1]q
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      count {}
    }

    statement {
      rate_based_statement {
        aggregate_key_type = "FORWARDED_IP"
        forwarded_ip_config {
          fallback_behavior = %[2]q
          header_name       = %[3]q
        }
        limit = 50000
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName, fallbackBehavior, headerName)
}

func testAccRuleGroupConfig_rateBasedStatement_customKeysBasic(rName, customKey, customKeyName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 100
  name     = %[1]q
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      count {}
    }

    statement {
      rate_based_statement {
        aggregate_key_type = "CUSTOM_KEYS"
        limit              = 50000

        custom_key {
          %[2]s {
            name = %[3]q

            text_transformation {
              type     = "NONE"
              priority = 0
            }
          }
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName, customKey, customKeyName)
}

func testAccRuleGroupConfig_rateBasedStatement_customKeysMinimal(rName, customKey string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 100
  name     = %[1]q
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      count {}
    }

    statement {
      rate_based_statement {
        aggregate_key_type = "CUSTOM_KEYS"
        limit              = 50000

        custom_key {
          %[2]s {
            text_transformation {
              type     = "NONE"
              priority = 0
            }
          }
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName, customKey)
}

func testAccRuleGroupConfig_rateBasedStatement_customKeysIP(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 100
  name     = %[1]q
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      count {}
    }

    statement {
      rate_based_statement {
        aggregate_key_type = "CUSTOM_KEYS"
        limit              = 50000

        custom_key {
          cookie {
            name = "cookie-name"

            text_transformation {
              type     = "NONE"
              priority = 0
            }
          }
        }

        custom_key {
          ip {}
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_rateBasedStatement_customKeysForwardedIP(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 100
  name     = %[1]q
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      count {}
    }

    statement {
      rate_based_statement {
        aggregate_key_type = "CUSTOM_KEYS"
        limit              = 50000

        forwarded_ip_config {
          fallback_behavior = "MATCH"
          header_name       = "x-forwarded-for"
        }

        custom_key {
          cookie {
            name = "cookie-name"

            text_transformation {
              type     = "NONE"
              priority = 0
            }
          }
        }

        custom_key {
          forwarded_ip {}
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_rateBasedStatement_customKeysHTTPMethod(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 100
  name     = %[1]q
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      count {}
    }

    statement {
      rate_based_statement {
        aggregate_key_type = "CUSTOM_KEYS"
        limit              = 50000

        custom_key {
          http_method {}
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_rateBasedStatement_customKeysMaxKeys(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 200
  name     = %[1]q
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      count {}
    }

    statement {
      rate_based_statement {
        aggregate_key_type = "CUSTOM_KEYS"
        limit              = 50000

        custom_key {
          cookie {
            name = "cookie-name"

            text_transformation {
              type     = "NONE"
              priority = 0
            }
          }
        }

        custom_key {
          header {
            name = "x-api-key"

            text_transformation {
              type     = "NONE"
              priority = 0
            }
          }
        }

        custom_key {
          query_string {
            text_transformation {
              type     = "NONE"
              priority = 0
            }
          }
        }

        custom_key {
          uri_path {
            text_transformation {
              type     = "NONE"
              priority = 0
            }
          }
        }

        custom_key {
          http_method {}
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_rateBasedStatement_update(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 3
  name     = %[1]q
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      count {}
    }

    statement {
      rate_based_statement {
        limit              = 10000
        aggregate_key_type = "IP"

        scope_down_statement {
          geo_match_statement {
            country_codes = ["US", "NL"]
          }
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity    = 2
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccRuleGroupConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity    = 2
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccRuleGroupConfig_multipleNestedRateBasedStatements(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_regex_pattern_set" "test" {
  name  = %[1]q
  scope = "REGIONAL"

  regular_expression {
    regex_string = "[a-z]([a-z0-9_-]*[a-z0-9])?"
  }
}

resource "aws_wafv2_ip_set" "test" {
  name               = %[1]q
  scope              = "REGIONAL"
  ip_address_version = "IPV4"
  addresses          = ["1.2.3.4/32", "5.6.7.8/32"]
}

resource "aws_wafv2_rule_group" "test" {
  capacity    = 300
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  rule {
    name     = "rule"
    priority = 0

    action {
      block {}
    }

    statement {
      rate_based_statement {
        limit              = 300
        aggregate_key_type = "IP"

        scope_down_statement {
          not_statement {
            statement {
              or_statement {
                statement {
                  regex_pattern_set_reference_statement {
                    arn = aws_wafv2_regex_pattern_set.test.arn

                    field_to_match {
                      uri_path {}
                    }

                    text_transformation {
                      type     = "LOWERCASE"
                      priority = 1
                    }
                  }
                }

                statement {
                  regex_match_statement {
                    regex_string = "[a-z]([a-z0-9_-]*[a-z0-9])?"

                    field_to_match {
                      uri_path {}
                    }

                    text_transformation {
                      type     = "LOWERCASE"
                      priority = 1
                    }
                  }
                }

                statement {
                  ip_set_reference_statement {
                    arn = aws_wafv2_ip_set.test.arn
                  }
                }
              }
            }
          }
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "rule"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "waf"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccRuleGroupConfig_multipleNestedOperatorStatements(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_regex_pattern_set" "test" {
  name  = %[1]q
  scope = "REGIONAL"

  regular_expression {
    regex_string = "[a-z]([a-z0-9_-]*[a-z0-9])?"
  }
}

resource "aws_wafv2_ip_set" "test" {
  name               = %[1]q
  scope              = "REGIONAL"
  ip_address_version = "IPV4"
  addresses          = ["1.2.3.4/32", "5.6.7.8/32"]
}

resource "aws_wafv2_rule_group" "test" {
  capacity    = 300
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  rule {
    name     = "rule"
    priority = 0

    action {
      block {}
    }

    statement {
      and_statement {
        statement {
          not_statement {
            statement {
              or_statement {
                statement {
                  regex_pattern_set_reference_statement {
                    arn = aws_wafv2_regex_pattern_set.test.arn

                    field_to_match {
                      uri_path {}
                    }

                    text_transformation {
                      type     = "LOWERCASE"
                      priority = 1
                    }
                  }
                }

                statement {
                  regex_match_statement {
                    regex_string = "[a-z]([a-z0-9_-]*[a-z0-9])?"

                    field_to_match {
                      uri_path {}
                    }

                    text_transformation {
                      type     = "LOWERCASE"
                      priority = 1
                    }
                  }
                }

                statement {
                  ip_set_reference_statement {
                    arn = aws_wafv2_ip_set.test.arn
                  }
                }
              }
            }
          }
        }

        statement {
          geo_match_statement {
            country_codes = ["NL"]
          }
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "rule"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "waf"
    sampled_requests_enabled   = false
  }
}
`, rName)
}
