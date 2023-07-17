// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafv2_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/wafv2"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfwafv2 "github.com/hashicorp/terraform-provider-aws/internal/service/wafv2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccWAFV2RuleGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v wafv2.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_basic(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
	var v wafv2.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_basic(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccRuleGroupConfig_basicUpdate(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "50"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                              "rule-1",
						"priority":                          "1",
						"action.#":                          "1",
						"action.0.allow.#":                  "0",
						"action.0.block.#":                  "0",
						"action.0.count.#":                  "1",
						"action.0.captcha.#":                "0",
						"action.0.challenge.#":              "0",
						"statement.#":                       "1",
						"statement.0.geo_match_statement.#": "1",
						"statement.0.geo_match_statement.0.country_codes.#": "2",
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
	var v wafv2.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"
	ruleName2 := fmt.Sprintf("%s-2", ruleGroupName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_basicUpdate(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "50"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                 "rule-1",
						"priority":             "1",
						"action.#":             "1",
						"action.0.allow.#":     "0",
						"action.0.block.#":     "0",
						"action.0.count.#":     "1",
						"action.0.captcha.#":   "0",
						"action.0.challenge.#": "0",
						"visibility_config.0.cloudwatch_metrics_enabled": "false",
						"visibility_config.0.metric_name":                "friendly-rule-metric-name",
						"visibility_config.0.sampled_requests_enabled":   "false",
						"statement.#":                                       "1",
						"statement.0.geo_match_statement.#":                 "1",
						"statement.0.geo_match_statement.0.country_codes.#": "2",
					}),
				),
			},
			{
				// Test step verifies addition of a rule block with the first block unchanged
				Config: testAccRuleGroupConfig_updateMultiples(ruleGroupName, "rule-1", ruleName2, 1, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "50"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                 "rule-1",
						"priority":             "1",
						"action.#":             "1",
						"action.0.allow.#":     "0",
						"action.0.block.#":     "0",
						"action.0.count.#":     "1",
						"action.0.captcha.#":   "0",
						"action.0.challenge.#": "0",
						"visibility_config.#":  "1",
						"visibility_config.0.cloudwatch_metrics_enabled": "false",
						"visibility_config.0.metric_name":                "rule-1",
						"visibility_config.0.sampled_requests_enabled":   "false",
						"statement.#":                                       "1",
						"statement.0.geo_match_statement.#":                 "1",
						"statement.0.geo_match_statement.0.country_codes.#": "2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                 ruleName2,
						"priority":             "2",
						"action.#":             "1",
						"action.0.allow.#":     "0",
						"action.0.block.#":     "1",
						"action.0.count.#":     "0",
						"action.0.captcha.#":   "0",
						"action.0.challenge.#": "0",
						"visibility_config.#":  "1",
						"visibility_config.0.cloudwatch_metrics_enabled": "false",
						"visibility_config.0.metric_name":                ruleName2,
						"visibility_config.0.sampled_requests_enabled":   "false",
						"statement.#": "1",
						"statement.0.size_constraint_statement.#":                                 "1",
						"statement.0.size_constraint_statement.0.comparison_operator":             "LT",
						"statement.0.size_constraint_statement.0.field_to_match.#":                "1",
						"statement.0.size_constraint_statement.0.field_to_match.0.query_string.#": "1",
						"statement.0.size_constraint_statement.0.size":                            "50",
						"statement.0.size_constraint_statement.0.text_transformation.#":           "2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.statement.0.size_constraint_statement.0.text_transformation.*", map[string]string{
						"priority": "2",
						"type":     "CMD_LINE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.statement.0.size_constraint_statement.0.text_transformation.*", map[string]string{
						"priority": "5",
						"type":     "NONE",
					}),
				),
			},
			{
				// Test step to verify a change in priority for rule #1 and a change in name and priority for rule #2
				Config: testAccRuleGroupConfig_updateMultiples(ruleGroupName, "rule-1", "updated", 5, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "50"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                 "rule-1",
						"priority":             "5",
						"action.#":             "1",
						"action.0.allow.#":     "0",
						"action.0.block.#":     "0",
						"action.0.count.#":     "1",
						"action.0.captcha.#":   "0",
						"action.0.challenge.#": "0",
						"visibility_config.#":  "1",
						"visibility_config.0.cloudwatch_metrics_enabled": "false",
						"visibility_config.0.metric_name":                "rule-1",
						"visibility_config.0.sampled_requests_enabled":   "false",
						"statement.#":                                       "1",
						"statement.0.geo_match_statement.#":                 "1",
						"statement.0.geo_match_statement.0.country_codes.#": "2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                 "updated",
						"priority":             "10",
						"action.#":             "1",
						"action.0.allow.#":     "0",
						"action.0.block.#":     "1",
						"action.0.count.#":     "0",
						"action.0.captcha.#":   "0",
						"action.0.challenge.#": "0",
						"visibility_config.#":  "1",
						"visibility_config.0.cloudwatch_metrics_enabled": "false",
						"visibility_config.0.metric_name":                "updated",
						"visibility_config.0.sampled_requests_enabled":   "false",
						"statement.#": "1",
						"statement.0.size_constraint_statement.#":                                 "1",
						"statement.0.size_constraint_statement.0.comparison_operator":             "LT",
						"statement.0.size_constraint_statement.0.field_to_match.#":                "1",
						"statement.0.size_constraint_statement.0.field_to_match.0.query_string.#": "1",
						"statement.0.size_constraint_statement.0.size":                            "50",
						"statement.0.size_constraint_statement.0.text_transformation.#":           "2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.statement.0.size_constraint_statement.0.text_transformation.*", map[string]string{
						"priority": "2",
						"type":     "CMD_LINE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.statement.0.size_constraint_statement.0.text_transformation.*", map[string]string{
						"priority": "5",
						"type":     "NONE",
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
	var v wafv2.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_byteMatchStatement(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                        "1",
						"statement.0.byte_match_statement.#": "1",
						"statement.0.byte_match_statement.0.positional_constraint": "CONTAINS",
						"statement.0.byte_match_statement.0.search_string":         "word",
						"statement.0.byte_match_statement.0.text_transformation.#": "2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.statement.0.byte_match_statement.0.text_transformation.*", map[string]string{
						"priority": "5",
						"type":     "NONE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.statement.0.byte_match_statement.0.text_transformation.*", map[string]string{
						"priority": "2",
						"type":     "LOWERCASE",
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_byteMatchStatementUpdate(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                        "1",
						"statement.0.byte_match_statement.#": "1",
						"statement.0.byte_match_statement.0.positional_constraint": "EXACTLY",
						"statement.0.byte_match_statement.0.search_string":         "sentence",
						"statement.0.byte_match_statement.0.text_transformation.#": "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.statement.0.byte_match_statement.0.text_transformation.*", map[string]string{
						"priority": "3",
						"type":     "CMD_LINE",
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
	var v wafv2.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_byteMatchStatementFieldToMatchAllQueryArguments(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                         "1",
						"statement.0.byte_match_statement.#":                  "1",
						"statement.0.byte_match_statement.0.field_to_match.#": "1",
						"statement.0.byte_match_statement.0.field_to_match.0.all_query_arguments.#":   "1",
						"statement.0.byte_match_statement.0.field_to_match.0.body.#":                  "0",
						"statement.0.byte_match_statement.0.field_to_match.0.cookies.#":               "0",
						"statement.0.byte_match_statement.0.field_to_match.0.headers.#":               "0",
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.#":             "0",
						"statement.0.byte_match_statement.0.field_to_match.0.method.#":                "0",
						"statement.0.byte_match_statement.0.field_to_match.0.query_string.#":          "0",
						"statement.0.byte_match_statement.0.field_to_match.0.single_header.#":         "0",
						"statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.#": "0",
						"statement.0.byte_match_statement.0.field_to_match.0.uri_path.#":              "0",
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_byteMatchStatementFieldToMatchBody(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                         "1",
						"statement.0.byte_match_statement.#":                  "1",
						"statement.0.byte_match_statement.0.field_to_match.#": "1",
						"statement.0.byte_match_statement.0.field_to_match.0.all_query_arguments.#":   "0",
						"statement.0.byte_match_statement.0.field_to_match.0.body.#":                  "1",
						"statement.0.byte_match_statement.0.field_to_match.0.cookies.#":               "0",
						"statement.0.byte_match_statement.0.field_to_match.0.headers.#":               "0",
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.#":             "0",
						"statement.0.byte_match_statement.0.field_to_match.0.method.#":                "0",
						"statement.0.byte_match_statement.0.field_to_match.0.query_string.#":          "0",
						"statement.0.byte_match_statement.0.field_to_match.0.single_header.#":         "0",
						"statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.#": "0",
						"statement.0.byte_match_statement.0.field_to_match.0.uri_path.#":              "0",
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_byteMatchStatementFieldToMatchCookies(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                         "1",
						"statement.0.byte_match_statement.#":                  "1",
						"statement.0.byte_match_statement.0.field_to_match.#": "1",
						"statement.0.byte_match_statement.0.field_to_match.0.all_query_arguments.#":                        "0",
						"statement.0.byte_match_statement.0.field_to_match.0.body.#":                                       "0",
						"statement.0.byte_match_statement.0.field_to_match.0.cookies.#":                                    "1",
						"statement.0.byte_match_statement.0.field_to_match.0.cookies.0.match_scope":                        "ALL",
						"statement.0.byte_match_statement.0.field_to_match.0.cookies.0.oversize_handling":                  "NO_MATCH",
						"statement.0.byte_match_statement.0.field_to_match.0.cookies.0.match_pattern.#":                    "1",
						"statement.0.byte_match_statement.0.field_to_match.0.cookies.0.match_pattern.0.included_cookies.0": "test",
						"statement.0.byte_match_statement.0.field_to_match.0.cookies.0.match_pattern.0.included_cookies.1": "cookie_test",
						"statement.0.byte_match_statement.0.field_to_match.0.headers.#":                                    "0",
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.#":                                  "0",
						"statement.0.byte_match_statement.0.field_to_match.0.method.#":                                     "0",
						"statement.0.byte_match_statement.0.field_to_match.0.query_string.#":                               "0",
						"statement.0.byte_match_statement.0.field_to_match.0.single_header.#":                              "0",
						"statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.#":                      "0",
						"statement.0.byte_match_statement.0.field_to_match.0.uri_path.#":                                   "0",
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_byteMatchStatementFieldToMatchJSONBody(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                                                                      "1",
						"statement.0.byte_match_statement.#":                                                               "1",
						"statement.0.byte_match_statement.0.field_to_match.#":                                              "1",
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.#":                                  "1",
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.0.match_scope":                      "VALUE",
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.0.invalid_fallback_behavior":        "MATCH",
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.0.oversize_handling":                "CONTINUE",
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.0.match_pattern.#":                  "1",
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.0.match_pattern.0.included_paths.#": "2",
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.0.match_pattern.0.included_paths.0": "/dogs/0/name",
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.0.match_pattern.0.included_paths.1": "/dogs/1/name",
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_byteMatchStatementFieldToMatchHeadersMatchPatternAll(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                         "1",
						"statement.0.byte_match_statement.#":                  "1",
						"statement.0.byte_match_statement.0.field_to_match.#": "1",
						"statement.0.byte_match_statement.0.field_to_match.0.all_query_arguments.#":                        "0",
						"statement.0.byte_match_statement.0.field_to_match.0.body.#":                                       "0",
						"statement.0.byte_match_statement.0.field_to_match.0.cookies.#":                                    "0",
						"statement.0.byte_match_statement.0.field_to_match.0.headers.#":                                    "1",
						"statement.0.byte_match_statement.0.field_to_match.0.headers.0.oversize_handling":                  "MATCH",
						"statement.0.byte_match_statement.0.field_to_match.0.headers.0.match_scope":                        "ALL",
						"statement.0.byte_match_statement.0.field_to_match.0.headers.0.match_pattern.#":                    "1",
						"statement.0.byte_match_statement.0.field_to_match.0.headers.0.match_pattern.0.all.#":              "1",
						"statement.0.byte_match_statement.0.field_to_match.0.headers.0.match_pattern.0.included_headers.#": "0",
						"statement.0.byte_match_statement.0.field_to_match.0.headers.0.match_pattern.0.excluded_headers.#": "0",
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.#":                                  "0",
						"statement.0.byte_match_statement.0.field_to_match.0.method.#":                                     "0",
						"statement.0.byte_match_statement.0.field_to_match.0.query_string.#":                               "0",
						"statement.0.byte_match_statement.0.field_to_match.0.single_header.#":                              "0",
						"statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.#":                      "0",
						"statement.0.byte_match_statement.0.field_to_match.0.uri_path.#":                                   "0",
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_byteMatchStatementFieldToMatchHeadersMatchPatternIncludedHeaders(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                         "1",
						"statement.0.byte_match_statement.#":                  "1",
						"statement.0.byte_match_statement.0.field_to_match.#": "1",
						"statement.0.byte_match_statement.0.field_to_match.0.all_query_arguments.#":                        "0",
						"statement.0.byte_match_statement.0.field_to_match.0.body.#":                                       "0",
						"statement.0.byte_match_statement.0.field_to_match.0.cookies.#":                                    "0",
						"statement.0.byte_match_statement.0.field_to_match.0.headers.#":                                    "1",
						"statement.0.byte_match_statement.0.field_to_match.0.headers.0.oversize_handling":                  "MATCH",
						"statement.0.byte_match_statement.0.field_to_match.0.headers.0.match_scope":                        "ALL",
						"statement.0.byte_match_statement.0.field_to_match.0.headers.0.match_pattern.#":                    "1",
						"statement.0.byte_match_statement.0.field_to_match.0.headers.0.match_pattern.0.all.#":              "0",
						"statement.0.byte_match_statement.0.field_to_match.0.headers.0.match_pattern.0.included_headers.#": "2",
						"statement.0.byte_match_statement.0.field_to_match.0.headers.0.match_pattern.0.included_headers.0": "session",
						"statement.0.byte_match_statement.0.field_to_match.0.headers.0.match_pattern.0.included_headers.1": "session-id",
						"statement.0.byte_match_statement.0.field_to_match.0.headers.0.match_pattern.0.excluded_headers.#": "0",
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.#":                                  "0",
						"statement.0.byte_match_statement.0.field_to_match.0.method.#":                                     "0",
						"statement.0.byte_match_statement.0.field_to_match.0.query_string.#":                               "0",
						"statement.0.byte_match_statement.0.field_to_match.0.single_header.#":                              "0",
						"statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.#":                      "0",
						"statement.0.byte_match_statement.0.field_to_match.0.uri_path.#":                                   "0",
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_byteMatchStatementFieldToMatchHeadersMatchPatternExcludedHeaders(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                         "1",
						"statement.0.byte_match_statement.#":                  "1",
						"statement.0.byte_match_statement.0.field_to_match.#": "1",
						"statement.0.byte_match_statement.0.field_to_match.0.all_query_arguments.#":                        "0",
						"statement.0.byte_match_statement.0.field_to_match.0.body.#":                                       "0",
						"statement.0.byte_match_statement.0.field_to_match.0.cookies.#":                                    "0",
						"statement.0.byte_match_statement.0.field_to_match.0.headers.#":                                    "1",
						"statement.0.byte_match_statement.0.field_to_match.0.headers.0.oversize_handling":                  "MATCH",
						"statement.0.byte_match_statement.0.field_to_match.0.headers.0.match_scope":                        "ALL",
						"statement.0.byte_match_statement.0.field_to_match.0.headers.0.match_pattern.#":                    "1",
						"statement.0.byte_match_statement.0.field_to_match.0.headers.0.match_pattern.0.all.#":              "0",
						"statement.0.byte_match_statement.0.field_to_match.0.headers.0.match_pattern.0.excluded_headers.#": "2",
						"statement.0.byte_match_statement.0.field_to_match.0.headers.0.match_pattern.0.excluded_headers.0": "session",
						"statement.0.byte_match_statement.0.field_to_match.0.headers.0.match_pattern.0.excluded_headers.1": "session-id",
						"statement.0.byte_match_statement.0.field_to_match.0.headers.0.match_pattern.0.included_headers.#": "0",
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.#":                                  "0",
						"statement.0.byte_match_statement.0.field_to_match.0.method.#":                                     "0",
						"statement.0.byte_match_statement.0.field_to_match.0.query_string.#":                               "0",
						"statement.0.byte_match_statement.0.field_to_match.0.single_header.#":                              "0",
						"statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.#":                      "0",
						"statement.0.byte_match_statement.0.field_to_match.0.uri_path.#":                                   "0",
					}),
				),
			},
			{
				Config:      testAccRuleGroupConfig_byteMatchStatementFieldToMatchHeadersInvalidConfiguration(ruleGroupName),
				ExpectError: regexp.MustCompile(`argument "oversize_handling" is required`),
			},
			{
				Config: testAccRuleGroupConfig_byteMatchStatementFieldToMatchMethod(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                         "1",
						"statement.0.byte_match_statement.#":                  "1",
						"statement.0.byte_match_statement.0.field_to_match.#": "1",
						"statement.0.byte_match_statement.0.field_to_match.0.all_query_arguments.#":   "0",
						"statement.0.byte_match_statement.0.field_to_match.0.body.#":                  "0",
						"statement.0.byte_match_statement.0.field_to_match.0.cookies.#":               "0",
						"statement.0.byte_match_statement.0.field_to_match.0.headers.#":               "0",
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.#":             "0",
						"statement.0.byte_match_statement.0.field_to_match.0.method.#":                "1",
						"statement.0.byte_match_statement.0.field_to_match.0.query_string.#":          "0",
						"statement.0.byte_match_statement.0.field_to_match.0.single_header.#":         "0",
						"statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.#": "0",
						"statement.0.byte_match_statement.0.field_to_match.0.uri_path.#":              "0",
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_byteMatchStatementFieldToMatchQueryString(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                         "1",
						"statement.0.byte_match_statement.#":                  "1",
						"statement.0.byte_match_statement.0.field_to_match.#": "1",
						"statement.0.byte_match_statement.0.field_to_match.0.all_query_arguments.#":   "0",
						"statement.0.byte_match_statement.0.field_to_match.0.body.#":                  "0",
						"statement.0.byte_match_statement.0.field_to_match.0.cookies.#":               "0",
						"statement.0.byte_match_statement.0.field_to_match.0.headers.#":               "0",
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.#":             "0",
						"statement.0.byte_match_statement.0.field_to_match.0.method.#":                "0",
						"statement.0.byte_match_statement.0.field_to_match.0.query_string.#":          "1",
						"statement.0.byte_match_statement.0.field_to_match.0.single_header.#":         "0",
						"statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.#": "0",
						"statement.0.byte_match_statement.0.field_to_match.0.uri_path.#":              "0",
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_byteMatchStatementFieldToMatchSingleHeader(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                         "1",
						"statement.0.byte_match_statement.#":                  "1",
						"statement.0.byte_match_statement.0.field_to_match.#": "1",
						"statement.0.byte_match_statement.0.field_to_match.0.all_query_arguments.#":   "0",
						"statement.0.byte_match_statement.0.field_to_match.0.body.#":                  "0",
						"statement.0.byte_match_statement.0.field_to_match.0.cookies.#":               "0",
						"statement.0.byte_match_statement.0.field_to_match.0.headers.#":               "0",
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.#":             "0",
						"statement.0.byte_match_statement.0.field_to_match.0.method.#":                "0",
						"statement.0.byte_match_statement.0.field_to_match.0.query_string.#":          "0",
						"statement.0.byte_match_statement.0.field_to_match.0.single_header.#":         "1",
						"statement.0.byte_match_statement.0.field_to_match.0.single_header.0.name":    "a-forty-character-long-header-name-40-40",
						"statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.#": "0",
						"statement.0.byte_match_statement.0.field_to_match.0.uri_path.#":              "0",
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_byteMatchStatementFieldToMatchSingleQueryArgument(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                         "1",
						"statement.0.byte_match_statement.#":                  "1",
						"statement.0.byte_match_statement.0.field_to_match.#": "1",
						"statement.0.byte_match_statement.0.field_to_match.0.all_query_arguments.#":        "0",
						"statement.0.byte_match_statement.0.field_to_match.0.body.#":                       "0",
						"statement.0.byte_match_statement.0.field_to_match.0.cookies.#":                    "0",
						"statement.0.byte_match_statement.0.field_to_match.0.headers.#":                    "0",
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.#":                  "0",
						"statement.0.byte_match_statement.0.field_to_match.0.method.#":                     "0",
						"statement.0.byte_match_statement.0.field_to_match.0.query_string.#":               "0",
						"statement.0.byte_match_statement.0.field_to_match.0.single_header.#":              "0",
						"statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.#":      "1",
						"statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.0.name": "a-max-30-characters-long-name-",
						"statement.0.byte_match_statement.0.field_to_match.0.uri_path.#":                   "0",
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_byteMatchStatementFieldToMatchURIPath(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                         "1",
						"statement.0.byte_match_statement.#":                  "1",
						"statement.0.byte_match_statement.0.field_to_match.#": "1",
						"statement.0.byte_match_statement.0.field_to_match.0.all_query_arguments.#":   "0",
						"statement.0.byte_match_statement.0.field_to_match.0.body.#":                  "0",
						"statement.0.byte_match_statement.0.field_to_match.0.cookies.#":               "0",
						"statement.0.byte_match_statement.0.field_to_match.0.headers.#":               "0",
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.#":             "0",
						"statement.0.byte_match_statement.0.field_to_match.0.method.#":                "0",
						"statement.0.byte_match_statement.0.field_to_match.0.query_string.#":          "0",
						"statement.0.byte_match_statement.0.field_to_match.0.single_header.#":         "0",
						"statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.#": "0",
						"statement.0.byte_match_statement.0.field_to_match.0.uri_path.#":              "1",
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
	var before, after wafv2.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	ruleGroupNewName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_basic(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &before),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
				),
			},
			{
				Config: testAccRuleGroupConfig_basic(ruleGroupNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &after),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupNewName),
					resource.TestCheckResourceAttr(resourceName, "description", ruleGroupNewName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
				),
			},
		},
	})
}

func TestAccWAFV2RuleGroup_changeCapacityForceNew(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after wafv2.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_basic(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &before),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
				),
			},
			{
				Config: testAccRuleGroupConfig_updateCapacity(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &after),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "3"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
				),
			},
		},
	})
}

func TestAccWAFV2RuleGroup_changeMetricNameForceNew(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after wafv2.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_basic(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &before),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
				),
			},
			{
				Config: testAccRuleGroupConfig_updateMetricName(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &after),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "updated-friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
				),
			},
		},
	})
}

func TestAccWAFV2RuleGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v wafv2.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, wafv2.EndpointsID),
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
	var v wafv2.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_labels(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"rule_label.#":      "2",
						"rule_label.0.name": "Hashicorp:Test:Label1",
						"rule_label.1.name": "Hashicorp:Test:Label2",
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_noLabels(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"rule_label.#": "0",
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
	var v wafv2.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_geoMatchStatement(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                             "1",
						"statement.0.geo_match_statement.#":                       "1",
						"statement.0.geo_match_statement.0.country_codes.#":       "2",
						"statement.0.geo_match_statement.0.country_codes.0":       "US",
						"statement.0.geo_match_statement.0.country_codes.1":       "NL",
						"statement.0.geo_match_statement.0.forwarded_ip_config.#": "0",
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_geoMatchStatementUpdate(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                             "1",
						"statement.0.geo_match_statement.#":                       "1",
						"statement.0.geo_match_statement.0.country_codes.#":       "3",
						"statement.0.geo_match_statement.0.country_codes.0":       "ZM",
						"statement.0.geo_match_statement.0.country_codes.1":       "EE",
						"statement.0.geo_match_statement.0.country_codes.2":       "MM",
						"statement.0.geo_match_statement.0.forwarded_ip_config.#": "0",
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
	var v wafv2.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_geoMatchStatementForwardedIP(ruleGroupName, "MATCH", "X-Forwarded-For"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                                               "1",
						"statement.0.geo_match_statement.#":                                         "1",
						"statement.0.geo_match_statement.0.country_codes.#":                         "2",
						"statement.0.geo_match_statement.0.country_codes.0":                         "US",
						"statement.0.geo_match_statement.0.country_codes.1":                         "NL",
						"statement.0.geo_match_statement.0.forwarded_ip_config.#":                   "1",
						"statement.0.geo_match_statement.0.forwarded_ip_config.0.fallback_behavior": "MATCH",
						"statement.0.geo_match_statement.0.forwarded_ip_config.0.header_name":       "X-Forwarded-For",
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_geoMatchStatementForwardedIP(ruleGroupName, "NO_MATCH", "Updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                                               "1",
						"statement.0.geo_match_statement.#":                                         "1",
						"statement.0.geo_match_statement.0.country_codes.#":                         "2",
						"statement.0.geo_match_statement.0.country_codes.0":                         "US",
						"statement.0.geo_match_statement.0.country_codes.1":                         "NL",
						"statement.0.geo_match_statement.0.forwarded_ip_config.#":                   "1",
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
	var v wafv2.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_labelMatchStatement(ruleGroupName, "LABEL", "Hashicorp:Test:Label1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                               "1",
						"statement.0.label_match_statement.#":       "1",
						"statement.0.label_match_statement.0.scope": "LABEL",
						"statement.0.label_match_statement.0.key":   "Hashicorp:Test:Label1",
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_labelMatchStatement(ruleGroupName, "NAMESPACE", "awswaf:managed:aws:bot-control:"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                               "1",
						"statement.0.label_match_statement.#":       "1",
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
	var v wafv2.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_ipsetReferenceStatement(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#": "1",
						"statement.0.ip_set_reference_statement.#":                              "1",
						"statement.0.ip_set_reference_statement.0.ip_set_forwarded_ip_config.#": "0",
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]*regexp.Regexp{
						"statement.0.ip_set_reference_statement.0.arn": regexp.MustCompile(`regional/ipset/.+$`),
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
	var v wafv2.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_ipsetReferenceStatementIPSetForwardedIP(ruleGroupName, "MATCH", "X-Forwarded-For", "FIRST"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#": "1",
						"statement.0.ip_set_reference_statement.#": "1",
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]*regexp.Regexp{
						"statement.0.ip_set_reference_statement.0.arn": regexp.MustCompile(`regional/ipset/.+$`),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.0.ip_set_reference_statement.0.ip_set_forwarded_ip_config.#":                   "1",
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
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#": "1",
						"statement.0.ip_set_reference_statement.#": "1",
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]*regexp.Regexp{
						"statement.0.ip_set_reference_statement.0.arn": regexp.MustCompile(`regional/ipset/.+$`),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.0.ip_set_reference_statement.0.ip_set_forwarded_ip_config.#":                   "1",
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
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#": "1",
						"statement.0.ip_set_reference_statement.#": "1",
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]*regexp.Regexp{
						"statement.0.ip_set_reference_statement.0.arn": regexp.MustCompile(`regional/ipset/.+$`),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.0.ip_set_reference_statement.0.ip_set_forwarded_ip_config.#":                   "1",
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
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#": "1",
						"statement.0.ip_set_reference_statement.#":                              "1",
						"statement.0.ip_set_reference_statement.0.ip_set_forwarded_ip_config.#": "0",
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]*regexp.Regexp{
						"statement.0.ip_set_reference_statement.0.arn": regexp.MustCompile(`regional/ipset/.+$`),
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
	var v wafv2.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_logicalStatementAnd(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                             "1",
						"statement.0.and_statement.#":             "1",
						"statement.0.and_statement.0.statement.#": "2",
						"statement.0.and_statement.0.statement.0.geo_match_statement.#": "1",
						"statement.0.and_statement.0.statement.1.geo_match_statement.#": "1",
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_logicalStatementNotAnd(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                                         "1",
						"statement.0.not_statement.#":                                         "1",
						"statement.0.not_statement.0.statement.#":                             "1",
						"statement.0.not_statement.0.statement.0.and_statement.#":             "1",
						"statement.0.not_statement.0.statement.0.and_statement.0.statement.#": "2",
						"statement.0.not_statement.0.statement.0.and_statement.0.statement.0.geo_match_statement.#": "1",
						"statement.0.not_statement.0.statement.0.and_statement.0.statement.1.geo_match_statement.#": "1",
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_logicalStatementOrNotAnd(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                                        "1",
						"statement.0.or_statement.#":                                         "1",
						"statement.0.or_statement.0.statement.#":                             "2",
						"statement.0.or_statement.0.statement.0.not_statement.#":             "1",
						"statement.0.or_statement.0.statement.0.not_statement.0.statement.#": "1",
						"statement.0.or_statement.0.statement.0.not_statement.0.statement.0.geo_match_statement.#": "1",
						"statement.0.or_statement.0.statement.1.and_statement.#":                                   "1",
						"statement.0.or_statement.0.statement.1.and_statement.0.statement.#":                       "2",
						"statement.0.or_statement.0.statement.1.and_statement.0.statement.0.geo_match_statement.#": "1",
						"statement.0.or_statement.0.statement.1.and_statement.0.statement.1.geo_match_statement.#": "1",
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
	var v wafv2.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_minimal(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
				),
			},
		},
	})
}

func TestAccWAFV2RuleGroup_regexMatchStatement(t *testing.T) {
	ctx := acctest.Context(t)
	var v wafv2.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_regexMatchStatement(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                               "1",
						"statement.0.regex_match_statement.#":                       "1",
						"statement.0.regex_match_statement.0.regex_string":          "[a-z]([a-z0-9_-]*[a-z0-9])?",
						"statement.0.regex_match_statement.0.field_to_match.#":      "1",
						"statement.0.regex_match_statement.0.text_transformation.#": "1",
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
	var v wafv2.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_regexPatternSetReferenceStatement(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#": "1",
						"statement.0.regex_pattern_set_reference_statement.#":                       "1",
						"statement.0.regex_pattern_set_reference_statement.0.field_to_match.#":      "1",
						"statement.0.regex_pattern_set_reference_statement.0.text_transformation.#": "1",
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]*regexp.Regexp{
						"statement.0.regex_pattern_set_reference_statement.0.arn": regexp.MustCompile(`regional/regexpatternset/.+$`),
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
	var v wafv2.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_actionAllow(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"action.#":         "1",
						"action.0.allow.#": "1",
						"action.0.allow.0.custom_request_handling.#": "0",
						"action.0.block.#":                           "0",
						"action.0.count.#":                           "0",
						"action.0.captcha.#":                         "0",
						"action.0.challenge.#":                       "0",
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_actionBlock(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"action.#":                           "1",
						"action.0.allow.#":                   "0",
						"action.0.block.#":                   "1",
						"action.0.block.0.custom_response.#": "0",
						"action.0.count.#":                   "0",
						"action.0.captcha.#":                 "0",
						"action.0.challenge.#":               "0",
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_actionCount(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"action.#":         "1",
						"action.0.allow.#": "0",
						"action.0.block.#": "0",
						"action.0.count.#": "1",
						"action.0.count.0.custom_request_handling.#": "0",
						"action.0.captcha.#":                         "0",
						"action.0.challenge.#":                       "0",
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
	var v wafv2.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_actionAllowCustomRequestHandling(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"action.#":         "1",
						"action.0.allow.#": "1",
						"action.0.allow.0.custom_request_handling.#":                       "1",
						"action.0.allow.0.custom_request_handling.0.insert_header.#":       "2",
						"action.0.allow.0.custom_request_handling.0.insert_header.0.name":  "x-hdr1",
						"action.0.allow.0.custom_request_handling.0.insert_header.0.value": "test-val1",
						"action.0.allow.0.custom_request_handling.0.insert_header.1.name":  "x-hdr2",
						"action.0.allow.0.custom_request_handling.0.insert_header.1.value": "test-val2",
						"action.0.block.#":     "0",
						"action.0.count.#":     "0",
						"action.0.captcha.#":   "0",
						"action.0.challenge.#": "0",
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_actionCountCustomRequestHandling(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"action.#":         "1",
						"action.0.allow.#": "0",
						"action.0.block.#": "0",
						"action.0.count.#": "1",
						"action.0.count.0.custom_request_handling.#":                       "1",
						"action.0.count.0.custom_request_handling.0.insert_header.#":       "2",
						"action.0.count.0.custom_request_handling.0.insert_header.0.name":  "x-hdr1",
						"action.0.count.0.custom_request_handling.0.insert_header.0.value": "test-val1",
						"action.0.count.0.custom_request_handling.0.insert_header.1.name":  "x-hdr2",
						"action.0.count.0.custom_request_handling.0.insert_header.1.value": "test-val2",
						"action.0.captcha.#":   "0",
						"action.0.challenge.#": "0",
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
	var v wafv2.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_actionBlockCustomResponse(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"action.#":                           "1",
						"action.0.allow.#":                   "0",
						"action.0.block.#":                   "1",
						"action.0.block.0.custom_response.#": "1",
						"action.0.block.0.custom_response.0.response_code":           "429",
						"action.0.block.0.custom_response.0.response_header.#":       "2",
						"action.0.block.0.custom_response.0.response_header.0.name":  "x-hdr1",
						"action.0.block.0.custom_response.0.response_header.0.value": "test-val1",
						"action.0.block.0.custom_response.0.response_header.1.name":  "x-hdr2",
						"action.0.block.0.custom_response.0.response_header.1.value": "test-val2",
						"action.0.count.#":     "0",
						"action.0.captcha.#":   "0",
						"action.0.challenge.#": "0",
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_actionBlockCustomResponseBody(ruleGroupName, "test_body_1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "custom_response_body.*", map[string]string{
						"key":          "test_body_1",
						"content":      "test response 1",
						"content_type": "TEXT_PLAIN",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "custom_response_body.*", map[string]string{
						"key":          "test_body_2",
						"content":      "<html><body>test response 2</body></html>",
						"content_type": "TEXT_HTML",
					}),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"action.#":                           "1",
						"action.0.allow.#":                   "0",
						"action.0.block.#":                   "1",
						"action.0.block.0.custom_response.#": "1",
						"action.0.block.0.custom_response.0.response_code":            "429",
						"action.0.block.0.custom_response.0.custom_response_body_key": "test_body_1",
						"action.0.count.#":     "0",
						"action.0.captcha.#":   "0",
						"action.0.challenge.#": "0",
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_actionBlockCustomResponseBody(ruleGroupName, "test_body_2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "custom_response_body.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "custom_response_body.*", map[string]string{
						"key":          "test_body_1",
						"content":      "test response 1",
						"content_type": "TEXT_PLAIN",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "custom_response_body.*", map[string]string{
						"key":          "test_body_2",
						"content":      "<html><body>test response 2</body></html>",
						"content_type": "TEXT_HTML",
					}),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"action.#":                           "1",
						"action.0.allow.#":                   "0",
						"action.0.block.#":                   "1",
						"action.0.block.0.custom_response.#": "1",
						"action.0.block.0.custom_response.0.response_code":            "429",
						"action.0.block.0.custom_response.0.custom_response_body_key": "test_body_2",
						"action.0.count.#":     "0",
						"action.0.captcha.#":   "0",
						"action.0.challenge.#": "0",
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
	var v wafv2.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_sizeConstraintStatement(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#": "1",
						"statement.0.size_constraint_statement.#":                           "1",
						"statement.0.size_constraint_statement.0.comparison_operator":       "GT",
						"statement.0.size_constraint_statement.0.size":                      "100",
						"statement.0.size_constraint_statement.0.field_to_match.#":          "1",
						"statement.0.size_constraint_statement.0.field_to_match.0.method.#": "1",
						"statement.0.size_constraint_statement.0.text_transformation.#":     "1",
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_sizeConstraintStatementUpdate(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#": "1",
						"statement.0.size_constraint_statement.#":                                 "1",
						"statement.0.size_constraint_statement.0.comparison_operator":             "LT",
						"statement.0.size_constraint_statement.0.size":                            "50",
						"statement.0.size_constraint_statement.0.field_to_match.#":                "1",
						"statement.0.size_constraint_statement.0.field_to_match.0.query_string.#": "1",
						"statement.0.size_constraint_statement.0.text_transformation.#":           "2",
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
	var v wafv2.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_sqliMatchStatement(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                         "1",
						"statement.0.sqli_match_statement.#":                  "1",
						"statement.0.sqli_match_statement.0.field_to_match.#": "1",
						"statement.0.sqli_match_statement.0.field_to_match.0.all_query_arguments.#": "1",
						"statement.0.sqli_match_statement.0.text_transformation.#":                  "2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.statement.0.sqli_match_statement.0.text_transformation.*", map[string]string{
						"priority": "5",
						"type":     "URL_DECODE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.statement.0.sqli_match_statement.0.text_transformation.*", map[string]string{
						"priority": "2",
						"type":     "LOWERCASE",
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_sqliMatchStatementUpdate(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                                "1",
						"statement.0.sqli_match_statement.#":                         "1",
						"statement.0.sqli_match_statement.0.field_to_match.#":        "1",
						"statement.0.sqli_match_statement.0.field_to_match.0.body.#": "1",
						"statement.0.sqli_match_statement.0.text_transformation.#":   "3",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.statement.0.sqli_match_statement.0.text_transformation.*", map[string]string{
						"priority": "5",
						"type":     "URL_DECODE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.statement.0.sqli_match_statement.0.text_transformation.*", map[string]string{
						"priority": "4",
						"type":     "HTML_ENTITY_DECODE",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.statement.0.sqli_match_statement.0.text_transformation.*", map[string]string{
						"priority": "3",
						"type":     "COMPRESS_WHITE_SPACE",
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
	var v wafv2.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_oneTag(ruleGroupName, "Tag1", "Value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag1", "Value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccRuleGroupImportStateIdFunc(resourceName),
			},
			{
				Config: testAccRuleGroupConfig_twoTags(ruleGroupName, "Tag1", "Value1Updated", "Tag2", "Value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag1", "Value1Updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag2", "Value2"),
				),
			},
			{
				Config: testAccRuleGroupConfig_oneTag(ruleGroupName, "Tag2", "Value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag2", "Value2"),
				),
			},
		},
	})
}

func TestAccWAFV2RuleGroup_xssMatchStatement(t *testing.T) {
	ctx := acctest.Context(t)
	var v wafv2.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_xssMatchStatement(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                               "1",
						"statement.0.xss_match_statement.#":                         "1",
						"statement.0.xss_match_statement.0.field_to_match.#":        "1",
						"statement.0.xss_match_statement.0.field_to_match.0.body.#": "1",
						"statement.0.xss_match_statement.0.text_transformation.#":   "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.statement.0.xss_match_statement.0.text_transformation.*", map[string]string{
						"priority": "2",
						"type":     "NONE",
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_xssMatchStatementUpdate(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                               "1",
						"statement.0.xss_match_statement.#":                         "1",
						"statement.0.xss_match_statement.0.field_to_match.#":        "1",
						"statement.0.xss_match_statement.0.field_to_match.0.body.#": "1",
						"statement.0.xss_match_statement.0.text_transformation.#":   "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*.statement.0.xss_match_statement.0.text_transformation.*", map[string]string{
						"priority": "2",
						"type":     "URL_DECODE",
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
	var v wafv2.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_rateBasedStatement(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#": "1",
						"statement.0.rate_based_statement.0.aggregate_key_type":     "IP",
						"statement.0.rate_based_statement.0.forwarded_ip_config.#":  "0",
						"statement.0.rate_based_statement.0.limit":                  "50000",
						"statement.0.rate_based_statement.0.scope_down_statement.#": "0",
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_rateBasedStatement_forwardedIPConfig(ruleGroupName, "MATCH", "X-Forwarded-For"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                        "1",
						"statement.0.rate_based_statement.#": "1",
						"statement.0.rate_based_statement.0.aggregate_key_type":                      "FORWARDED_IP",
						"statement.0.rate_based_statement.0.forwarded_ip_config.#":                   "1",
						"statement.0.rate_based_statement.0.forwarded_ip_config.0.fallback_behavior": "MATCH",
						"statement.0.rate_based_statement.0.forwarded_ip_config.0.header_name":       "X-Forwarded-For",
						"statement.0.rate_based_statement.0.limit":                                   "50000",
						"statement.0.rate_based_statement.0.scope_down_statement.#":                  "0",
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_rateBasedStatement_forwardedIPConfig(ruleGroupName, "NO_MATCH", "Updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                        "1",
						"statement.0.rate_based_statement.#": "1",
						"statement.0.rate_based_statement.0.aggregate_key_type":                      "FORWARDED_IP",
						"statement.0.rate_based_statement.0.forwarded_ip_config.#":                   "1",
						"statement.0.rate_based_statement.0.forwarded_ip_config.0.fallback_behavior": "NO_MATCH",
						"statement.0.rate_based_statement.0.forwarded_ip_config.0.header_name":       "Updated",
						"statement.0.rate_based_statement.0.limit":                                   "50000",
						"statement.0.rate_based_statement.0.scope_down_statement.#":                  "0",
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_rateBasedStatement_update(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                        "1",
						"statement.0.rate_based_statement.#": "1",
						"statement.0.rate_based_statement.0.aggregate_key_type":                                           "IP",
						"statement.0.rate_based_statement.0.forwarded_ip_config.#":                                        "0",
						"statement.0.rate_based_statement.0.limit":                                                        "10000",
						"statement.0.rate_based_statement.0.scope_down_statement.#":                                       "1",
						"statement.0.rate_based_statement.0.scope_down_statement.0.geo_match_statement.#":                 "1",
						"statement.0.rate_based_statement.0.scope_down_statement.0.geo_match_statement.0.country_codes.#": "2",
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
	var v wafv2.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_multipleNestedRateBasedStatements(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                                                                                      "1",
						"statement.0.rate_based_statement.#":                                                                               "1",
						"statement.0.rate_based_statement.0.limit":                                                                         "300",
						"statement.0.rate_based_statement.0.aggregate_key_type":                                                            "IP",
						"statement.0.rate_based_statement.0.scope_down_statement.#":                                                        "1",
						"statement.0.rate_based_statement.0.scope_down_statement.0.not_statement.#":                                        "1",
						"statement.0.rate_based_statement.0.scope_down_statement.0.not_statement.0.statement.#":                            "1",
						"statement.0.rate_based_statement.0.scope_down_statement.0.not_statement.0.statement.0.or_statement.#":             "1",
						"statement.0.rate_based_statement.0.scope_down_statement.0.not_statement.0.statement.0.or_statement.0.statement.#": "3",
						"statement.0.rate_based_statement.0.scope_down_statement.0.not_statement.0.statement.0.or_statement.0.statement.0.regex_pattern_set_reference_statement.#": "1",
						"statement.0.rate_based_statement.0.scope_down_statement.0.not_statement.0.statement.0.or_statement.0.statement.1.regex_match_statement.#":                 "1",
						"statement.0.rate_based_statement.0.scope_down_statement.0.not_statement.0.statement.0.or_statement.0.statement.2.ip_set_reference_statement.#":            "1",
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
	var v wafv2.RuleGroup
	ruleGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_multipleNestedOperatorStatements(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                                                                    "1",
						"statement.0.and_statement.#":                                                                    "1",
						"statement.0.and_statement.0.statement.#":                                                        "2",
						"statement.0.and_statement.0.statement.0.not_statement.#":                                        "1",
						"statement.0.and_statement.0.statement.0.not_statement.0.statement.#":                            "1",
						"statement.0.and_statement.0.statement.0.not_statement.0.statement.0.or_statement.#":             "1",
						"statement.0.and_statement.0.statement.0.not_statement.0.statement.0.or_statement.0.statement.#": "3",
						"statement.0.and_statement.0.statement.0.not_statement.0.statement.0.or_statement.0.statement.0.regex_pattern_set_reference_statement.#": "1",
						"statement.0.and_statement.0.statement.0.not_statement.0.statement.0.or_statement.0.statement.1.regex_match_statement.#":                 "1",
						"statement.0.and_statement.0.statement.0.not_statement.0.statement.0.or_statement.0.statement.2.ip_set_reference_statement.#":            "1",
						"statement.0.and_statement.0.statement.1.geo_match_statement.#":                                                                          "1",
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
	conn := acctest.Provider.Meta().(*conns.AWSClient).WAFV2Conn(ctx)

	input := &wafv2.ListRuleGroupsInput{
		Scope: aws.String(wafv2.ScopeRegional),
	}

	_, err := conn.ListRuleGroupsWithContext(ctx, input)

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

			conn := acctest.Provider.Meta().(*conns.AWSClient).WAFV2Conn(ctx)

			_, err := tfwafv2.FindRuleGroupByThreePartKey(ctx, conn, rs.Primary.ID, rs.Primary.Attributes["name"], rs.Primary.Attributes["scope"])

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

func testAccCheckRuleGroupExists(ctx context.Context, n string, v *wafv2.RuleGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAFv2 RuleGroup ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFV2Conn(ctx)

		output, err := tfwafv2.FindRuleGroupByThreePartKey(ctx, conn, rs.Primary.ID, rs.Primary.Attributes["name"], rs.Primary.Attributes["scope"])

		if err != nil {
			return err
		}

		*v = *output.RuleGroup

		return nil
	}
}

func testAccRuleGroupConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity    = 2
  name        = "%s"
  description = "%s"
  scope       = "REGIONAL"

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name, name)
}

func testAccRuleGroupConfig_basicUpdate(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity    = 50
  name        = "%s"
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
`, name)
}

func testAccRuleGroupConfig_updateMultiples(name string, ruleName1, ruleName2 string, priority1, priority2 int) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity    = 50
  name        = "%[1]s"
  description = "Updated"
  scope       = "REGIONAL"

  rule {
    name     = "%[2]s"
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
      metric_name                = "%[2]s"
      sampled_requests_enabled   = false
    }
  }

  rule {
    name     = "%[4]s"
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
      metric_name                = "%[4]s"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name, ruleName1, priority1, ruleName2, priority2)
}

func testAccRuleGroupConfig_updateCapacity(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity    = 3
  name        = "%s"
  description = "%s"
  scope       = "REGIONAL"

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name, name)
}

func testAccRuleGroupConfig_updateMetricName(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity    = 2
  name        = "%s"
  description = "%s"
  scope       = "REGIONAL"

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "updated-friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name, name)
}

func testAccRuleGroupConfig_minimal(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name     = "%s"
  scope    = "REGIONAL"

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, name)
}

func testAccRuleGroupConfig_actionAllow(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name     = "%s"
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
`, name)
}

func testAccRuleGroupConfig_actionAllowCustomRequestHandling(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name     = "%s"
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
`, name)
}

func testAccRuleGroupConfig_actionBlock(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name     = "%s"
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
`, name)
}

func testAccRuleGroupConfig_actionBlockCustomResponse(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name     = "%s"
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
`, name)
}

func testAccRuleGroupConfig_actionBlockCustomResponseBody(name string, customBodyKey string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name     = "%[1]s"
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
          custom_response_body_key = "%[2]s"
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
`, name, customBodyKey)
}

func testAccRuleGroupConfig_actionCount(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name     = "%s"
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
`, name)
}

func testAccRuleGroupConfig_actionCountCustomRequestHandling(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name     = "%s"
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
`, name)
}

func testAccRuleGroupConfig_byteMatchStatement(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 300
  name     = "%s"
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
`, name)
}

func testAccRuleGroupConfig_byteMatchStatementUpdate(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 30
  name     = "%s"
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
`, name)
}

func testAccRuleGroupConfig_byteMatchStatementFieldToMatchAllQueryArguments(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 30
  name     = "%s"
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
`, name)
}

func testAccRuleGroupConfig_byteMatchStatementFieldToMatchBody(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 15
  name     = "%s"
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
`, name)
}

func testAccRuleGroupConfig_byteMatchStatementFieldToMatchJSONBody(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 20
  name     = "%s"
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
`, name)
}

func testAccRuleGroupConfig_byteMatchStatementFieldToMatchHeadersInvalidConfiguration(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 50
  name     = "%s"
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
`, name)
}

func testAccRuleGroupConfig_byteMatchStatementFieldToMatchHeadersMatchPatternAll(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 50
  name     = "%s"
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
`, name)
}

func testAccRuleGroupConfig_byteMatchStatementFieldToMatchHeadersMatchPatternIncludedHeaders(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 50
  name     = "%s"
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
`, name)
}

func testAccRuleGroupConfig_byteMatchStatementFieldToMatchHeadersMatchPatternExcludedHeaders(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 50
  name     = "%s"
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
`, name)
}

func testAccRuleGroupConfig_byteMatchStatementFieldToMatchMethod(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 15
  name     = "%s"
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
`, name)
}

func testAccRuleGroupConfig_byteMatchStatementFieldToMatchQueryString(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 15
  name     = "%s"
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
`, name)
}

func testAccRuleGroupConfig_byteMatchStatementFieldToMatchCookies(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 15
  name     = "%s"
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
`, name)
}

func testAccRuleGroupConfig_byteMatchStatementFieldToMatchSingleHeader(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 15
  name     = "%s"
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
`, name)
}

func testAccRuleGroupConfig_byteMatchStatementFieldToMatchSingleQueryArgument(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 30
  name     = "%s"
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
`, name)
}

func testAccRuleGroupConfig_byteMatchStatementFieldToMatchURIPath(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 15
  name     = "%s"
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
`, name)
}

func testAccRuleGroupConfig_ipsetReferenceStatement(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_ip_set" "test" {
  name               = "ip-set-%s"
  scope              = "REGIONAL"
  ip_address_version = "IPV4"
  addresses          = ["1.1.1.1/32", "2.2.2.2/32"]
}

resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name     = "%s"
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
`, name, name)
}

func testAccRuleGroupConfig_ipsetReferenceStatementIPSetForwardedIP(name, fallbackBehavior, headerName, position string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_ip_set" "test" {
  name               = "ip-set-%[1]s"
  scope              = "REGIONAL"
  ip_address_version = "IPV4"
  addresses          = ["1.1.1.1/32", "2.2.2.2/32"]
}

resource "aws_wafv2_rule_group" "test" {
  capacity = 5
  name     = "%[1]s"
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
          fallback_behavior = "%[2]s"
          header_name       = "%[3]s"
          position          = "%[4]s"
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
`, name, fallbackBehavior, headerName, position)
}

func testAccRuleGroupConfig_geoMatchStatement(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name     = "%s"
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
`, name)
}

func testAccRuleGroupConfig_geoMatchStatementForwardedIP(name, fallbackBehavior, headerName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name     = "%s"
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
          fallback_behavior = "%s"
          header_name       = "%s"
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
`, name, fallbackBehavior, headerName)
}

func testAccRuleGroupConfig_geoMatchStatementUpdate(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name     = "%s"
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
`, name)
}

func testAccRuleGroupConfig_labelMatchStatement(name string, scope string, key string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name     = "%[1]s"
  scope    = "REGIONAL"
  rule {
    name     = "rule-1"
    priority = 1
    action {
      allow {}
    }
    statement {
      label_match_statement {
        scope = "%[2]s"
        key   = "%[3]s"
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
`, name, scope, key)
}

func testAccRuleGroupConfig_labels(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name     = "%s"
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
`, name)
}

func testAccRuleGroupConfig_noLabels(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name     = "%s"
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
`, name)
}

func testAccRuleGroupConfig_logicalStatementAnd(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name     = "%s"
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
`, name)
}

func testAccRuleGroupConfig_logicalStatementNotAnd(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name     = "%s"
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
`, name)
}

func testAccRuleGroupConfig_logicalStatementOrNotAnd(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 3
  name     = "%s"
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
`, name)
}

func testAccRuleGroupConfig_regexMatchStatement(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 50
  name     = "%s"
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
`, name)
}

func testAccRuleGroupConfig_regexPatternSetReferenceStatement(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_regex_pattern_set" "test" {
  name  = "regex-pattern-set-%s"
  scope = "REGIONAL"

  regular_expression {
    regex_string = "[a-z]([a-z0-9_-]*[a-z0-9])?"
  }
}

resource "aws_wafv2_rule_group" "test" {
  capacity = 50
  name     = "%s"
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
`, name, name)
}

func testAccRuleGroupConfig_sizeConstraintStatement(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 30
  name     = "%s"
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
`, name)
}

func testAccRuleGroupConfig_sizeConstraintStatementUpdate(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 30
  name     = "%s"
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
`, name)
}

func testAccRuleGroupConfig_sqliMatchStatement(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 300
  name     = "%s"
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
`, name)
}

func testAccRuleGroupConfig_sqliMatchStatementUpdate(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 300
  name     = "%s"
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
`, name)
}

func testAccRuleGroupConfig_xssMatchStatement(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 300
  name     = "%s"
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
`, name)
}

func testAccRuleGroupConfig_xssMatchStatementUpdate(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 300
  name     = "%s"
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
`, name)
}

func testAccRuleGroupConfig_rateBasedStatement(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 3
  name     = "%s"
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      count {}
    }

    statement {
      rate_based_statement {
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
`, name)
}

func testAccRuleGroupConfig_rateBasedStatement_forwardedIPConfig(name, fallbackBehavior, headerName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 3
  name     = "%s"
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
`, name, fallbackBehavior, headerName)
}

func testAccRuleGroupConfig_rateBasedStatement_update(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 3
  name     = "%s"
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
`, name)
}

func testAccRuleGroupConfig_oneTag(name, tagKey, tagValue string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity    = 2
  name        = "%s"
  description = "%s"
  scope       = "REGIONAL"

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }

  tags = {
    "%s" = "%s"
  }
}
`, name, name, tagKey, tagValue)
}

func testAccRuleGroupConfig_twoTags(name, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity    = 2
  name        = "%s"
  description = "%s"
  scope       = "REGIONAL"

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }

  tags = {
    "%s" = "%s"
    "%s" = "%s"
  }
}
`, name, name, tag1Key, tag1Value, tag2Key, tag2Value)
}

func testAccRuleGroupConfig_multipleNestedRateBasedStatements(name string) string {
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
`, name)
}

func testAccRuleGroupConfig_multipleNestedOperatorStatements(name string) string {
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
`, name)
}

func testAccRuleGroupImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s/%s", rs.Primary.ID, rs.Primary.Attributes["name"], rs.Primary.Attributes["scope"]), nil
	}
}
