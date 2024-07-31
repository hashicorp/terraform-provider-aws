// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafv2_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfwafv2 "github.com/hashicorp/terraform-provider-aws/internal/service/wafv2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(names.WAFV2ServiceID, testAccErrorCheckSkip)
}

func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"Your request contains fields that belong to a feature you are not allowed to use",
		"The scope is not valid., field: SCOPE_VALUE, parameter: CLOUDFRONT",
	)
}

func TestAccWAFV2WebACL_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_basic(webACLName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "application_integration_url", ""),
					resource.TestCheckResourceAttr(resourceName, "association_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "captcha_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "challenge_config.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, webACLName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, string(awstypes.ScopeRegional)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "token_domains.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", acctest.CtFalse),
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

func TestAccWAFV2WebACL_Update_rule(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"
	ruleName1 := fmt.Sprintf("%s-1", webACLName)
	ruleName2 := fmt.Sprintf("%s-2", webACLName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_basicRule(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Updated"),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, string(awstypes.ScopeRegional)),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrName:        ruleName1,
						names.AttrPriority:    acctest.Ct10,
						"action.#":            acctest.Ct1,
						"action.0.allow.#":    acctest.Ct0,
						"action.0.block.#":    acctest.Ct0,
						"action.0.count.#":    acctest.Ct1,
						"visibility_config.#": acctest.Ct1,
						"visibility_config.0.cloudwatch_metrics_enabled": acctest.CtFalse,
						"visibility_config.0.metric_name":                fmt.Sprintf("%s-metric-name-1", webACLName),
						"visibility_config.0.sampled_requests_enabled":   acctest.CtFalse,
						"statement.#": acctest.Ct1,
						"statement.0.size_constraint_statement.#":                                          acctest.Ct1,
						"statement.0.size_constraint_statement.0.comparison_operator":                      "LT",
						"statement.0.size_constraint_statement.0.field_to_match.#":                         acctest.Ct1,
						"statement.0.size_constraint_statement.0.field_to_match.0.all_query_arguments.#":   acctest.Ct0,
						"statement.0.size_constraint_statement.0.field_to_match.0.body.#":                  acctest.Ct0,
						"statement.0.size_constraint_statement.0.field_to_match.0.cookies.#":               acctest.Ct0,
						"statement.0.size_constraint_statement.0.field_to_match.0.header_order.#":          acctest.Ct0,
						"statement.0.size_constraint_statement.0.field_to_match.0.headers.#":               acctest.Ct0,
						"statement.0.size_constraint_statement.0.field_to_match.0.ja3_fingerprint.#":       acctest.Ct0,
						"statement.0.size_constraint_statement.0.field_to_match.0.json_body.#":             acctest.Ct0,
						"statement.0.size_constraint_statement.0.field_to_match.0.method.#":                acctest.Ct0,
						"statement.0.size_constraint_statement.0.field_to_match.0.query_string.#":          acctest.Ct1,
						"statement.0.size_constraint_statement.0.field_to_match.0.single_header.#":         acctest.Ct0,
						"statement.0.size_constraint_statement.0.field_to_match.0.single_query_argument.#": acctest.Ct0,
						"statement.0.size_constraint_statement.0.field_to_match.0.uri_path.#":              acctest.Ct0,
						"statement.0.size_constraint_statement.0.size":                                     "50",
						"statement.0.size_constraint_statement.0.text_transformation.#":                    acctest.Ct2,
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
				// Test step to verify additional rule block with first rule block unchanged
				Config: testAccWebACLConfig_updateRuleNamePriorityMetric(webACLName, ruleName1, ruleName2, 10, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Updated"),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, string(awstypes.ScopeRegional)),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrName:        ruleName1,
						names.AttrPriority:    acctest.Ct10,
						"action.#":            acctest.Ct1,
						"action.0.allow.#":    acctest.Ct0,
						"action.0.block.#":    acctest.Ct0,
						"action.0.count.#":    acctest.Ct1,
						"visibility_config.#": acctest.Ct1,
						"visibility_config.0.cloudwatch_metrics_enabled": acctest.CtFalse,
						"visibility_config.0.metric_name":                ruleName1,
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
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrName:        ruleName2,
						names.AttrPriority:    "5",
						"action.#":            acctest.Ct1,
						"action.0.allow.#":    acctest.Ct1,
						"action.0.block.#":    acctest.Ct0,
						"action.0.count.#":    acctest.Ct0,
						"visibility_config.#": acctest.Ct1,
						"visibility_config.0.cloudwatch_metrics_enabled": acctest.CtFalse,
						"visibility_config.0.metric_name":                ruleName2,
						"visibility_config.0.sampled_requests_enabled":   acctest.CtFalse,
						"statement.#":                                       acctest.Ct1,
						"statement.0.geo_match_statement.#":                 acctest.Ct1,
						"statement.0.geo_match_statement.0.country_codes.#": acctest.Ct2,
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

func TestAccWAFV2WebACL_Update_ruleProperties(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"
	ruleName1 := fmt.Sprintf("%s-1", webACLName)
	ruleName2 := fmt.Sprintf("%s-2", webACLName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_updateRuleNamePriorityMetric(webACLName, ruleName1, ruleName2, 5, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Updated"),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, string(awstypes.ScopeRegional)),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrName:        ruleName1,
						names.AttrPriority:    "5",
						"action.#":            acctest.Ct1,
						"action.0.allow.#":    acctest.Ct0,
						"action.0.block.#":    acctest.Ct0,
						"action.0.count.#":    acctest.Ct1,
						"visibility_config.#": acctest.Ct1,
						"visibility_config.0.cloudwatch_metrics_enabled": acctest.CtFalse,
						"visibility_config.0.metric_name":                ruleName1,
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
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrName:        ruleName2,
						names.AttrPriority:    acctest.Ct10,
						"action.#":            acctest.Ct1,
						"action.0.allow.#":    acctest.Ct1,
						"action.0.block.#":    acctest.Ct0,
						"action.0.count.#":    acctest.Ct0,
						"visibility_config.#": acctest.Ct1,
						"visibility_config.0.cloudwatch_metrics_enabled": acctest.CtFalse,
						"visibility_config.0.metric_name":                ruleName2,
						"visibility_config.0.sampled_requests_enabled":   acctest.CtFalse,
						"statement.#":                                       acctest.Ct1,
						"statement.0.geo_match_statement.#":                 acctest.Ct1,
						"statement.0.geo_match_statement.0.country_codes.#": acctest.Ct2,
					}),
				),
			},
			{
				Config: testAccWebACLConfig_updateRuleNamePriorityMetric(webACLName, ruleName1, ruleName2, 10, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Updated"),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, string(awstypes.ScopeRegional)),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrName:        ruleName1,
						names.AttrPriority:    acctest.Ct10,
						"action.#":            acctest.Ct1,
						"action.0.allow.#":    acctest.Ct0,
						"action.0.block.#":    acctest.Ct0,
						"action.0.count.#":    acctest.Ct1,
						"visibility_config.#": acctest.Ct1,
						"visibility_config.0.cloudwatch_metrics_enabled": acctest.CtFalse,
						"visibility_config.0.metric_name":                ruleName1,
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
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrName:        ruleName2,
						names.AttrPriority:    "5",
						"action.#":            acctest.Ct1,
						"action.0.allow.#":    acctest.Ct1,
						"action.0.block.#":    acctest.Ct0,
						"action.0.count.#":    acctest.Ct0,
						"visibility_config.#": acctest.Ct1,
						"visibility_config.0.cloudwatch_metrics_enabled": acctest.CtFalse,
						"visibility_config.0.metric_name":                ruleName2,
						"visibility_config.0.sampled_requests_enabled":   acctest.CtFalse,
						"statement.#":                                       acctest.Ct1,
						"statement.0.geo_match_statement.#":                 acctest.Ct1,
						"statement.0.geo_match_statement.0.country_codes.#": acctest.Ct2,
					}),
				),
			},
			{
				Config: testAccWebACLConfig_updateRuleNamePriorityMetric(webACLName, ruleName1, "updated", 10, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Updated"),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, string(awstypes.ScopeRegional)),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrName:        ruleName1,
						names.AttrPriority:    acctest.Ct10,
						"action.#":            acctest.Ct1,
						"action.0.allow.#":    acctest.Ct0,
						"action.0.block.#":    acctest.Ct0,
						"action.0.count.#":    acctest.Ct1,
						"visibility_config.#": acctest.Ct1,
						"visibility_config.0.cloudwatch_metrics_enabled": acctest.CtFalse,
						"visibility_config.0.metric_name":                ruleName1,
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
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrName:        "updated",
						names.AttrPriority:    "5",
						"action.#":            acctest.Ct1,
						"action.0.allow.#":    acctest.Ct1,
						"action.0.block.#":    acctest.Ct0,
						"action.0.count.#":    acctest.Ct0,
						"visibility_config.#": acctest.Ct1,
						"visibility_config.0.cloudwatch_metrics_enabled": acctest.CtFalse,
						"visibility_config.0.metric_name":                "updated",
						"visibility_config.0.sampled_requests_enabled":   acctest.CtFalse,
						"statement.#":                                       acctest.Ct1,
						"statement.0.geo_match_statement.#":                 acctest.Ct1,
						"statement.0.geo_match_statement.0.country_codes.#": acctest.Ct2,
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

func TestAccWAFV2WebACL_Update_nameForceNew(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after awstypes.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	ruleGroupNewName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_basic(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &before),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, webACLName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, string(awstypes.ScopeRegional)),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", acctest.CtFalse),
				),
			},
			{
				Config: testAccWebACLConfig_basic(ruleGroupNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &after),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ruleGroupNewName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ruleGroupNewName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, string(awstypes.ScopeRegional)),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccWAFV2WebACL_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_minimal(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfwafv2.ResourceWebACL(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWAFV2WebACL_ManagedRuleGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_managedRuleGroupStatement(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "application_integration_url", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrName:              "rule-1",
						"action.#":                  acctest.Ct0,
						"override_action.#":         acctest.Ct1,
						"override_action.0.count.#": acctest.Ct0,
						"override_action.0.none.#":  acctest.Ct1,
						"statement.#":               acctest.Ct1,
						"statement.0.managed_rule_group_statement.#":                        acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.name":                   "AWSManagedRulesCommonRuleSet",
						"statement.0.managed_rule_group_statement.0.rule_action_override.#": acctest.Ct0,
						"statement.0.managed_rule_group_statement.0.scope_down_statement.#": acctest.Ct0,
						"statement.0.managed_rule_group_statement.0.vendor_name":            "AWS",
					}),
				),
			},
			{
				Config: testAccWebACLConfig_managedRuleGroupStatementUpdate(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrName:              "rule-1",
						"action.#":                  acctest.Ct0,
						"override_action.#":         acctest.Ct1,
						"override_action.0.count.#": acctest.Ct1,
						"override_action.0.none.#":  acctest.Ct0,
						"statement.#":               acctest.Ct1,
						"statement.0.managed_rule_group_statement.#":                                                              acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.name":                                                         "AWSManagedRulesCommonRuleSet",
						"statement.0.managed_rule_group_statement.0.rule_action_override.#":                                       acctest.Ct2,
						"statement.0.managed_rule_group_statement.0.rule_action_override.0.name":                                  "SizeRestrictions_QUERYSTRING",
						"statement.0.managed_rule_group_statement.0.rule_action_override.1.name":                                  "NoUserAgent_HEADER",
						"statement.0.managed_rule_group_statement.0.scope_down_statement.#":                                       acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.scope_down_statement.0.geo_match_statement.#":                 acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.scope_down_statement.0.geo_match_statement.0.country_codes.#": acctest.Ct2,
						"statement.0.managed_rule_group_statement.0.scope_down_statement.0.geo_match_statement.0.country_codes.0": "US",
						"statement.0.managed_rule_group_statement.0.scope_down_statement.0.geo_match_statement.0.country_codes.1": "NL",
						"statement.0.managed_rule_group_statement.0.vendor_name":                                                  "AWS",
					}),
				),
			},
			{
				Config: testAccWebACLConfig_managedRuleGroupStatementRuleActionOverrides(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrName:              "rule-1",
						"action.#":                  acctest.Ct0,
						"override_action.#":         acctest.Ct1,
						"override_action.0.count.#": acctest.Ct1,
						"override_action.0.none.#":  acctest.Ct0,
						"statement.#":               acctest.Ct1,
						"statement.0.managed_rule_group_statement.#":                                                              acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.name":                                                         "AWSManagedRulesCommonRuleSet",
						"statement.0.managed_rule_group_statement.0.vendor_name":                                                  "AWS",
						"statement.0.managed_rule_group_statement.0.rule_action_override.#":                                       acctest.Ct2,
						"statement.0.managed_rule_group_statement.0.rule_action_override.0.action_to_use.#":                       acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.rule_action_override.0.action_to_use.0.allow.#":               acctest.Ct0,
						"statement.0.managed_rule_group_statement.0.rule_action_override.0.action_to_use.0.block.#":               acctest.Ct0,
						"statement.0.managed_rule_group_statement.0.rule_action_override.0.action_to_use.0.captcha.#":             acctest.Ct0,
						"statement.0.managed_rule_group_statement.0.rule_action_override.0.action_to_use.0.challenge.#":           acctest.Ct0,
						"statement.0.managed_rule_group_statement.0.rule_action_override.0.action_to_use.0.count.#":               acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.rule_action_override.0.name":                                  "SizeRestrictions_QUERYSTRING",
						"statement.0.managed_rule_group_statement.0.rule_action_override.1.action_to_use.#":                       acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.rule_action_override.1.action_to_use.0.allow.#":               acctest.Ct0,
						"statement.0.managed_rule_group_statement.0.rule_action_override.1.action_to_use.0.block.#":               acctest.Ct0,
						"statement.0.managed_rule_group_statement.0.rule_action_override.1.action_to_use.0.captcha.#":             acctest.Ct0,
						"statement.0.managed_rule_group_statement.0.rule_action_override.1.action_to_use.0.challenge.#":           acctest.Ct0,
						"statement.0.managed_rule_group_statement.0.rule_action_override.1.action_to_use.0.count.#":               acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.rule_action_override.1.name":                                  "NoUserAgent_HEADER",
						"statement.0.managed_rule_group_statement.0.scope_down_statement.#":                                       acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.scope_down_statement.0.geo_match_statement.#":                 acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.scope_down_statement.0.geo_match_statement.0.country_codes.#": acctest.Ct2,
						"statement.0.managed_rule_group_statement.0.scope_down_statement.0.geo_match_statement.0.country_codes.0": "US",
						"statement.0.managed_rule_group_statement.0.scope_down_statement.0.geo_match_statement.0.country_codes.1": "NL",
					}),
				),
			},
			{
				Config: testAccWebACLConfig_managedRuleGroupStatementVersionVersion10(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrName:              "rule-1",
						"action.#":                  acctest.Ct0,
						"override_action.#":         acctest.Ct1,
						"override_action.0.count.#": acctest.Ct0,
						"override_action.0.none.#":  acctest.Ct1,
						"statement.#":               acctest.Ct1,
						"statement.0.managed_rule_group_statement.#":                        acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.name":                   "AWSManagedRulesCommonRuleSet",
						"statement.0.managed_rule_group_statement.0.rule_action_override.#": acctest.Ct0,
						"statement.0.managed_rule_group_statement.0.scope_down_statement.#": acctest.Ct0,
						"statement.0.managed_rule_group_statement.0.vendor_name":            "AWS",
						"statement.0.managed_rule_group_statement.0.version":                "Version_1.0",
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

func TestAccWAFV2WebACL_ManagedRuleGroup_ManagedRuleGroupConfig(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_managedRuleGroupStatementManagedRuleGroupConfig(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrName:              "rule-1",
						"action.#":                  acctest.Ct0,
						"override_action.#":         acctest.Ct1,
						"override_action.0.count.#": acctest.Ct0,
						"override_action.0.none.#":  acctest.Ct1,
						"statement.#":               acctest.Ct1,
						"statement.0.managed_rule_group_statement.#":                                                          acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.login_path":                  "/login",
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.1.payload_type":                "JSON",
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.2.password_field.0.identifier": "/password",
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.3.username_field.0.identifier": "/username",
						"statement.0.managed_rule_group_statement.0.name":                                                     "AWSManagedRulesATPRuleSet",
						"statement.0.managed_rule_group_statement.0.rule_action_override.#":                                   acctest.Ct0,
						"statement.0.managed_rule_group_statement.0.scope_down_statement.#":                                   acctest.Ct0,
						"statement.0.managed_rule_group_statement.0.vendor_name":                                              "AWS"}),
				),
			},
			{
				Config: testAccWebACLConfig_managedRuleGroupStatementManagedRuleGroupConfigUpdate(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrName:              "rule-1",
						"action.#":                  acctest.Ct0,
						"override_action.#":         acctest.Ct1,
						"override_action.0.count.#": acctest.Ct0,
						"override_action.0.none.#":  acctest.Ct1,
						"statement.#":               acctest.Ct1,
						"statement.0.managed_rule_group_statement.#":                                                          acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.login_path":                  "/app-login",
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.1.payload_type":                "JSON",
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.2.password_field.0.identifier": "/app-password",
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.3.username_field.0.identifier": "/app-username",
						"statement.0.managed_rule_group_statement.0.name":                                                     "AWSManagedRulesATPRuleSet",
						"statement.0.managed_rule_group_statement.0.rule_action_override.#":                                   acctest.Ct0,
						"statement.0.managed_rule_group_statement.0.scope_down_statement.#":                                   acctest.Ct0,
						"statement.0.managed_rule_group_statement.0.vendor_name":                                              "AWS",
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

func TestAccWAFV2WebACL_ManagedRuleGroup_ManagedRuleGroupConfig_ACFPRuleSet(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_managedRuleGroupStatementManagedRuleGroupConfig_acfpRuleSet(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestMatchResourceAttr(resourceName, "application_integration_url", regexache.MustCompile(`https:\/\/.*\.sdk\.awswaf\.com.*`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrName:              "rule-1",
						"action.#":                  acctest.Ct0,
						"override_action.#":         acctest.Ct1,
						"override_action.0.count.#": acctest.Ct0,
						"override_action.0.none.#":  acctest.Ct1,
						"statement.#":               acctest.Ct1,
						"statement.0.managed_rule_group_statement.#":                                                                acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_acfp_rule_set.#": acctest.Ct1,
						// "statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_acfp_rule_set.0.enable_regex_in_path":                             "false",
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_acfp_rule_set.0.creation_path":                                            "/creation",
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_acfp_rule_set.0.registration_page_path":                                   "/registration",
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_acfp_rule_set.0.request_inspection.#":                                     acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_acfp_rule_set.0.request_inspection.0.email_field.#":                       acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_acfp_rule_set.0.request_inspection.0.email_field.0.identifier":            "/email",
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_acfp_rule_set.0.request_inspection.0.password_field.#":                    acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_acfp_rule_set.0.request_inspection.0.password_field.0.identifier":         "/password",
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_acfp_rule_set.0.request_inspection.0.payload_type":                        "JSON",
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_acfp_rule_set.0.request_inspection.0.username_field.#":                    acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_acfp_rule_set.0.request_inspection.0.username_field.0.identifier":         "/username",
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_acfp_rule_set.0.request_inspection.0.phone_number_fields.#":               acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_acfp_rule_set.0.request_inspection.0.phone_number_fields.0.identifiers.#": acctest.Ct2,
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_acfp_rule_set.0.request_inspection.0.phone_number_fields.0.identifiers.0": "/phone1",
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_acfp_rule_set.0.request_inspection.0.phone_number_fields.0.identifiers.1": "/phone2",
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_acfp_rule_set.0.request_inspection.0.address_fields.#":                    acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_acfp_rule_set.0.request_inspection.0.address_fields.0.identifiers.#":      acctest.Ct2,
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_acfp_rule_set.0.request_inspection.0.address_fields.0.identifiers.0":      "home",
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_acfp_rule_set.0.request_inspection.0.address_fields.0.identifiers.1":      "work",
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_acfp_rule_set.0.response_inspection.#":                                    acctest.Ct0,
						"statement.0.managed_rule_group_statement.0.name":                   "AWSManagedRulesACFPRuleSet",
						"statement.0.managed_rule_group_statement.0.rule_action_override.#": acctest.Ct0,
						"statement.0.managed_rule_group_statement.0.scope_down_statement.#": acctest.Ct0,
						"statement.0.managed_rule_group_statement.0.vendor_name":            "AWS",
					}),
				),
			},
			{
				Config: testAccWebACLConfig_managedRuleGroupStatementManagedRuleGroupConfig_acfpRuleSetUpdate(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrName:              "rule-1",
						"action.#":                  acctest.Ct0,
						"override_action.#":         acctest.Ct1,
						"override_action.0.count.#": acctest.Ct0,
						"override_action.0.none.#":  acctest.Ct1,
						"statement.#":               acctest.Ct1,
						"statement.0.managed_rule_group_statement.#":                                                                                                                         acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_acfp_rule_set.#":                                                          acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_acfp_rule_set.0.enable_regex_in_path":                                     acctest.CtTrue,
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_acfp_rule_set.0.creation_path":                                            "/creation",
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_acfp_rule_set.0.registration_page_path":                                   "/registration",
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_acfp_rule_set.0.request_inspection.#":                                     acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_acfp_rule_set.0.request_inspection.0.email_field.#":                       acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_acfp_rule_set.0.request_inspection.0.email_field.0.identifier":            "/email",
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_acfp_rule_set.0.request_inspection.0.password_field.#":                    acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_acfp_rule_set.0.request_inspection.0.password_field.0.identifier":         "/pass",
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_acfp_rule_set.0.request_inspection.0.payload_type":                        "JSON",
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_acfp_rule_set.0.request_inspection.0.username_field.#":                    acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_acfp_rule_set.0.request_inspection.0.username_field.0.identifier":         "/user",
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_acfp_rule_set.0.request_inspection.0.phone_number_fields.#":               acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_acfp_rule_set.0.request_inspection.0.phone_number_fields.0.identifiers.#": acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_acfp_rule_set.0.request_inspection.0.phone_number_fields.0.identifiers.0": "/phone3",
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_acfp_rule_set.0.request_inspection.0.address_fields.#":                    acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_acfp_rule_set.0.request_inspection.0.address_fields.0.identifiers.#":      acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_acfp_rule_set.0.request_inspection.0.address_fields.0.identifiers.0":      "mobile",
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_acfp_rule_set.0.response_inspection.#":                                    acctest.Ct0,
						"statement.0.managed_rule_group_statement.0.name":                   "AWSManagedRulesACFPRuleSet",
						"statement.0.managed_rule_group_statement.0.rule_action_override.#": acctest.Ct0,
						"statement.0.managed_rule_group_statement.0.scope_down_statement.#": acctest.Ct0,
						"statement.0.managed_rule_group_statement.0.vendor_name":            "AWS",
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

func TestAccWAFV2WebACL_ManagedRuleGroup_ManagedRuleGroupConfig_ATPRuleSet(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_managedRuleGroupStatementManagedRuleGroupConfig_atpRuleSet(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestMatchResourceAttr(resourceName, "application_integration_url", regexache.MustCompile(`https:\/\/.*\.sdk\.awswaf\.com.*`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrName:              "rule-1",
						"action.#":                  acctest.Ct0,
						"override_action.#":         acctest.Ct1,
						"override_action.0.count.#": acctest.Ct0,
						"override_action.0.none.#":  acctest.Ct1,
						"statement.#":               acctest.Ct1,
						"statement.0.managed_rule_group_statement.#":                                                               acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_atp_rule_set.#": acctest.Ct1,
						// "statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_atp_rule_set.0.enable_regex_in_path":                             "false",
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_atp_rule_set.0.login_path":                                       "/api/1/signin",
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_atp_rule_set.0.request_inspection.#":                             acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_atp_rule_set.0.request_inspection.0.password_field.#":            acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_atp_rule_set.0.request_inspection.0.password_field.0.identifier": "/password",
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_atp_rule_set.0.request_inspection.0.payload_type":                "JSON",
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_atp_rule_set.0.request_inspection.0.username_field.#":            acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_atp_rule_set.0.request_inspection.0.username_field.0.identifier": "/username",
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_atp_rule_set.0.response_inspection.#":                            acctest.Ct0,
						"statement.0.managed_rule_group_statement.0.name":                   "AWSManagedRulesATPRuleSet",
						"statement.0.managed_rule_group_statement.0.rule_action_override.#": acctest.Ct0,
						"statement.0.managed_rule_group_statement.0.scope_down_statement.#": acctest.Ct0,
						"statement.0.managed_rule_group_statement.0.vendor_name":            "AWS",
					}),
				),
			},
			{
				Config: testAccWebACLConfig_managedRuleGroupStatementManagedRuleGroupConfig_atpRuleSetUpdate(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrName:              "rule-1",
						"action.#":                  acctest.Ct0,
						"override_action.#":         acctest.Ct1,
						"override_action.0.count.#": acctest.Ct0,
						"override_action.0.none.#":  acctest.Ct1,
						"statement.#":               acctest.Ct1,
						"statement.0.managed_rule_group_statement.#":                                                                                                                acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_atp_rule_set.#":                                                  acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_atp_rule_set.0.enable_regex_in_path":                             acctest.CtTrue,
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_atp_rule_set.0.login_path":                                       "/api/2/signin",
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_atp_rule_set.0.request_inspection.#":                             acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_atp_rule_set.0.request_inspection.0.password_field.#":            acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_atp_rule_set.0.request_inspection.0.password_field.0.identifier": "/pass",
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_atp_rule_set.0.request_inspection.0.payload_type":                "JSON",
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_atp_rule_set.0.request_inspection.0.username_field.#":            acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_atp_rule_set.0.request_inspection.0.username_field.0.identifier": "/user",
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_atp_rule_set.0.response_inspection.#":                            acctest.Ct0,
						"statement.0.managed_rule_group_statement.0.name":                                                                                                           "AWSManagedRulesATPRuleSet",
						"statement.0.managed_rule_group_statement.0.rule_action_override.#":                                                                                         acctest.Ct0,
						"statement.0.managed_rule_group_statement.0.scope_down_statement.#":                                                                                         acctest.Ct0,
						"statement.0.managed_rule_group_statement.0.vendor_name":                                                                                                    "AWS",
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

func TestAccWAFV2WebACL_ManagedRuleGroup_ManagedRuleGroupConfig_BotControl(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_managedRuleGroupStatementManagedRuleGroupConfig_botControl(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestMatchResourceAttr(resourceName, "application_integration_url", regexache.MustCompile(`https:\/\/.*\.sdk\.awswaf\.com.*`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrName:              "rule-1",
						"action.#":                  acctest.Ct0,
						"override_action.#":         acctest.Ct1,
						"override_action.0.count.#": acctest.Ct0,
						"override_action.0.none.#":  acctest.Ct1,
						"statement.#":               acctest.Ct1,
						"statement.0.managed_rule_group_statement.#":             acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.name":        "AWSManagedRulesBotControlRuleSet",
						"statement.0.managed_rule_group_statement.0.vendor_name": "AWS",
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_bot_control_rule_set.0.inspection_level":        "TARGETED",
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_bot_control_rule_set.0.enable_machine_learning": acctest.CtFalse,
					}),
				),
			},
		},
	})
}

func TestAccWAFV2WebACL_ManagedRuleGroup_specifyVersion(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_managedRuleGroupStatementVersionVersion10(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrName:              "rule-1",
						"action.#":                  acctest.Ct0,
						"override_action.#":         acctest.Ct1,
						"override_action.0.count.#": acctest.Ct0,
						"override_action.0.none.#":  acctest.Ct1,
						"statement.#":               acctest.Ct1,
						"statement.0.managed_rule_group_statement.#":                        acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.name":                   "AWSManagedRulesCommonRuleSet",
						"statement.0.managed_rule_group_statement.0.rule_action_override.#": acctest.Ct0,
						"statement.0.managed_rule_group_statement.0.scope_down_statement.#": acctest.Ct0,
						"statement.0.managed_rule_group_statement.0.vendor_name":            "AWS",
						"statement.0.managed_rule_group_statement.0.version":                "Version_1.0",
					}),
				),
			},
			{
				Config: testAccWebACLConfig_managedRuleGroupStatement(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrName:              "rule-1",
						"action.#":                  acctest.Ct0,
						"override_action.#":         acctest.Ct1,
						"override_action.0.count.#": acctest.Ct0,
						"override_action.0.none.#":  acctest.Ct1,
						"statement.#":               acctest.Ct1,
						"statement.0.managed_rule_group_statement.#":                        acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.name":                   "AWSManagedRulesCommonRuleSet",
						"statement.0.managed_rule_group_statement.0.rule_action_override.#": acctest.Ct0,
						"statement.0.managed_rule_group_statement.0.scope_down_statement.#": acctest.Ct0,
						"statement.0.managed_rule_group_statement.0.vendor_name":            "AWS",
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

func TestAccWAFV2WebACL_minimal(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_minimal(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, string(awstypes.ScopeRegional)),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccWAFV2WebACL_RateBased_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_rateBasedStatement(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrName:                       "rule-1",
						"action.#":                           acctest.Ct1,
						"action.0.allow.#":                   acctest.Ct0,
						"action.0.block.#":                   acctest.Ct0,
						"action.0.count.#":                   acctest.Ct1,
						"statement.#":                        acctest.Ct1,
						"statement.0.rate_based_statement.#": acctest.Ct1,
						"statement.0.rate_based_statement.0.aggregate_key_type":     "IP",
						"statement.0.rate_based_statement.0.evaluation_window_sec":  "300",
						"statement.0.rate_based_statement.0.forwarded_ip_config.#":  acctest.Ct0,
						"statement.0.rate_based_statement.0.limit":                  "50000",
						"statement.0.rate_based_statement.0.scope_down_statement.#": acctest.Ct0,
					}),
				),
			},
			{
				Config: testAccWebACLConfig_rateBasedStatementUpdate(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrName:                       "rule-1",
						"action.#":                           acctest.Ct1,
						"action.0.allow.#":                   acctest.Ct0,
						"action.0.block.#":                   acctest.Ct0,
						"action.0.count.#":                   acctest.Ct1,
						"statement.#":                        acctest.Ct1,
						"statement.0.rate_based_statement.#": acctest.Ct1,
						"statement.0.rate_based_statement.0.aggregate_key_type":                                           "IP",
						"statement.0.rate_based_statement.0.evaluation_window_sec":                                        "600",
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
				ImportStateIdFunc: testAccWebACLImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccWAFV2WebACL_ByteMatchStatement_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_byteMatchStatement(webACLName, string(awstypes.PositionalConstraintContainsWord), acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                         acctest.Ct1,
						"statement.0.byte_match_statement.#":                  acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.#": acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.0.all_query_arguments.#": acctest.Ct1,
						"statement.0.byte_match_statement.0.positional_constraint":                  "CONTAINS_WORD",
						"statement.0.byte_match_statement.0.search_string":                          acctest.CtValue1,
						"statement.0.byte_match_statement.0.text_transformation.#":                  acctest.Ct1,
						"statement.0.byte_match_statement.0.text_transformation.0.priority":         acctest.Ct0,
						"statement.0.byte_match_statement.0.text_transformation.0.type":             "NONE",
					}),
				),
			},
			{
				Config: testAccWebACLConfig_byteMatchStatement(webACLName, string(awstypes.PositionalConstraintExactly), acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                         acctest.Ct1,
						"statement.0.byte_match_statement.#":                  acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.#": acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.0.all_query_arguments.#": acctest.Ct1,
						"statement.0.byte_match_statement.0.positional_constraint":                  "EXACTLY",
						"statement.0.byte_match_statement.0.search_string":                          acctest.CtValue2,
						"statement.0.byte_match_statement.0.text_transformation.#":                  acctest.Ct1,
						"statement.0.byte_match_statement.0.text_transformation.0.priority":         acctest.Ct0,
						"statement.0.byte_match_statement.0.text_transformation.0.type":             "NONE",
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

func TestAccWAFV2WebACL_ByteMatchStatement_ja3fingerprint(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_byteMatchStatementJA3Fingerprint(webACLName, string(awstypes.FallbackBehaviorMatch)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.0.byte_match_statement.0.field_to_match.0.ja3_fingerprint.#":                   acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.0.ja3_fingerprint.0.fallback_behavior": "MATCH",
					}),
				),
			},
			{
				Config: testAccWebACLConfig_byteMatchStatementJA3Fingerprint(webACLName, string(awstypes.FallbackBehaviorNoMatch)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.0.byte_match_statement.0.field_to_match.0.ja3_fingerprint.#":                   acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.0.ja3_fingerprint.0.fallback_behavior": "NO_MATCH",
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

func TestAccWAFV2WebACL_ByteMatchStatement_jsonBody(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_byteMatchStatementJSONBody(webACLName, string(awstypes.JsonMatchScopeValue), string(awstypes.FallbackBehaviorMatch), string(awstypes.OversizeHandlingNoMatch), `included_paths = ["/dogs/0/name", "/dogs/1/name"]`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.#":                                  acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.0.match_scope":                      "VALUE",
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.0.invalid_fallback_behavior":        "MATCH",
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.0.oversize_handling":                "NO_MATCH",
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.0.match_pattern.#":                  acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.0.match_pattern.0.all.#":            acctest.Ct0,
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.0.match_pattern.0.included_paths.#": acctest.Ct2,
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.0.match_pattern.0.included_paths.0": "/dogs/0/name",
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.0.match_pattern.0.included_paths.1": "/dogs/1/name",
					}),
				),
			},
			{
				Config: testAccWebACLConfig_byteMatchStatementJSONBody(webACLName, string(awstypes.JsonMatchScopeAll), string(awstypes.FallbackBehaviorNoMatch), string(awstypes.OversizeHandlingContinue), "all {}"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.#":                                  acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.0.match_scope":                      "ALL",
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.0.invalid_fallback_behavior":        "NO_MATCH",
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.0.oversize_handling":                "CONTINUE",
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.0.match_pattern.#":                  acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.0.match_pattern.0.all.#":            acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.0.json_body.0.match_pattern.0.included_paths.#": acctest.Ct0,
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

func TestAccWAFV2WebACL_ByteMatchStatement_body(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_byteMatchStatementBody(webACLName, string(awstypes.OversizeHandlingNoMatch)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.0.byte_match_statement.0.field_to_match.0.body.#":                   acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.0.body.0.oversize_handling": "NO_MATCH",
					}),
				),
			},
			{
				Config: testAccWebACLConfig_byteMatchStatementBody(webACLName, string(awstypes.OversizeHandlingContinue)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.0.byte_match_statement.0.field_to_match.0.body.#":                   acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.0.body.0.oversize_handling": "CONTINUE",
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

func TestAccWAFV2WebACL_ByteMatchStatement_headerOrder(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_byteMatchStatementHeaderOrder(webACLName, string(awstypes.OversizeHandlingMatch)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.0.byte_match_statement.0.field_to_match.0.header_order.#":                   acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.0.header_order.0.oversize_handling": "MATCH",
					}),
				),
			},
			{
				Config: testAccWebACLConfig_byteMatchStatementHeaderOrder(webACLName, string(awstypes.OversizeHandlingNoMatch)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.0.byte_match_statement.0.field_to_match.0.header_order.#":                   acctest.Ct1,
						"statement.0.byte_match_statement.0.field_to_match.0.header_order.0.oversize_handling": "NO_MATCH",
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

func TestAccWAFV2WebACL_GeoMatch_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"
	countryCode := fmt.Sprintf("%q", "US")
	countryCodes := fmt.Sprintf("%s, %q", countryCode, "CA")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_geoMatchStatement(webACLName, countryCode),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, "REGIONAL"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrName:                      "rule-1",
						"action.#":                          acctest.Ct1,
						"action.0.allow.#":                  acctest.Ct0,
						"action.0.block.#":                  acctest.Ct1,
						"action.0.count.#":                  acctest.Ct0,
						names.AttrPriority:                  acctest.Ct1,
						"statement.#":                       acctest.Ct1,
						"statement.0.geo_match_statement.#": acctest.Ct1,
						"statement.0.geo_match_statement.0.country_codes.#":       acctest.Ct1,
						"statement.0.geo_match_statement.0.country_codes.0":       "US",
						"statement.0.geo_match_statement.0.forwarded_ip_config.#": acctest.Ct0,
						"visibility_config.#":                            acctest.Ct1,
						"visibility_config.0.cloudwatch_metrics_enabled": acctest.CtFalse,
						"visibility_config.0.metric_name":                "friendly-rule-metric-name",
						"visibility_config.0.sampled_requests_enabled":   acctest.CtFalse,
					}),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", acctest.CtFalse),
				),
			},
			{
				Config: testAccWebACLConfig_geoMatchStatement(webACLName, countryCodes),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, "REGIONAL"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrName:                      "rule-1",
						"action.#":                          acctest.Ct1,
						"action.0.allow.#":                  acctest.Ct0,
						"action.0.block.#":                  acctest.Ct1,
						"action.0.count.#":                  acctest.Ct0,
						names.AttrPriority:                  acctest.Ct1,
						"statement.#":                       acctest.Ct1,
						"statement.0.geo_match_statement.#": acctest.Ct1,
						"statement.0.geo_match_statement.0.country_codes.#":       acctest.Ct2,
						"statement.0.geo_match_statement.0.country_codes.0":       "US",
						"statement.0.geo_match_statement.0.country_codes.1":       "CA",
						"statement.0.geo_match_statement.0.forwarded_ip_config.#": acctest.Ct0,
						"visibility_config.#":                                     acctest.Ct1,
						"visibility_config.0.cloudwatch_metrics_enabled":          acctest.CtFalse,
						"visibility_config.0.metric_name":                         "friendly-rule-metric-name",
						"visibility_config.0.sampled_requests_enabled":            acctest.CtFalse,
					}),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", acctest.CtFalse),
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

func TestAccWAFV2WebACL_GeoMatch_forwardedIP(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_geoMatchStatementForwardedIP(webACLName, "MATCH", "X-Forwarded-For"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                            acctest.Ct1,
						"statement.0.or_statement.#":             acctest.Ct1,
						"statement.0.or_statement.0.statement.#": acctest.Ct2,
						"statement.0.or_statement.0.statement.0.geo_match_statement.#":                                         acctest.Ct1,
						"statement.0.or_statement.0.statement.0.geo_match_statement.0.country_codes.#":                         acctest.Ct1,
						"statement.0.or_statement.0.statement.0.geo_match_statement.0.forwarded_ip_config.#":                   acctest.Ct0,
						"statement.0.or_statement.0.statement.1.geo_match_statement.#":                                         acctest.Ct1,
						"statement.0.or_statement.0.statement.1.geo_match_statement.0.country_codes.#":                         acctest.Ct1,
						"statement.0.or_statement.0.statement.1.geo_match_statement.0.forwarded_ip_config.#":                   acctest.Ct1,
						"statement.0.or_statement.0.statement.1.geo_match_statement.0.forwarded_ip_config.0.fallback_behavior": "MATCH",
						"statement.0.or_statement.0.statement.1.geo_match_statement.0.forwarded_ip_config.0.header_name":       "X-Forwarded-For",
					}),
				),
			},
			{
				Config: testAccWebACLConfig_geoMatchStatementForwardedIP(webACLName, "NO_MATCH", "Updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                            acctest.Ct1,
						"statement.0.or_statement.#":             acctest.Ct1,
						"statement.0.or_statement.0.statement.#": acctest.Ct2,
						"statement.0.or_statement.0.statement.0.geo_match_statement.#":                                         acctest.Ct1,
						"statement.0.or_statement.0.statement.0.geo_match_statement.0.country_codes.#":                         acctest.Ct1,
						"statement.0.or_statement.0.statement.0.geo_match_statement.0.forwarded_ip_config.#":                   acctest.Ct0,
						"statement.0.or_statement.0.statement.1.geo_match_statement.#":                                         acctest.Ct1,
						"statement.0.or_statement.0.statement.1.geo_match_statement.0.country_codes.#":                         acctest.Ct1,
						"statement.0.or_statement.0.statement.1.geo_match_statement.0.forwarded_ip_config.#":                   acctest.Ct1,
						"statement.0.or_statement.0.statement.1.geo_match_statement.0.forwarded_ip_config.0.fallback_behavior": "NO_MATCH",
						"statement.0.or_statement.0.statement.1.geo_match_statement.0.forwarded_ip_config.0.header_name":       "Updated",
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

func TestAccWAFV2WebACL_LabelMatchStatement(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_labelMatchStatement(webACLName, "LABEL", "Hashicorp:Test:Label1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
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
				Config: testAccWebACLConfig_labelMatchStatement(webACLName, "NAMESPACE", "awswaf:managed:aws:bot-control:"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
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
				ImportStateIdFunc: testAccWebACLImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccWAFV2WebACL_RuleLabels(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_ruleLabels(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"rule_label.#":      acctest.Ct2,
						"rule_label.0.name": "Hashicorp:Test:Label1",
						"rule_label.1.name": "Hashicorp:Test:Label2",
					}),
				),
			},
			{
				Config: testAccWebACLConfig_noRuleLabels(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
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
				ImportStateIdFunc: testAccWebACLImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccWAFV2WebACL_IPSetReference_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_ipsetReference(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#": acctest.Ct1,
						"statement.0.ip_set_reference_statement.#":                              acctest.Ct1,
						"statement.0.ip_set_reference_statement.0.ip_set_forwarded_ip_config.#": acctest.Ct0,
						"visibility_config.#":                            acctest.Ct1,
						"visibility_config.0.cloudwatch_metrics_enabled": acctest.CtFalse,
						"visibility_config.0.metric_name":                "friendly-rule-metric-name",
						"visibility_config.0.sampled_requests_enabled":   acctest.CtFalse,
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]*regexp.Regexp{
						"statement.0.ip_set_reference_statement.0.arn": regexache.MustCompile(`regional/ipset/.+$`),
					}),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", acctest.CtFalse),
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

func TestAccWAFV2WebACL_IPSetReference_forwardedIP(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_ipsetReferenceForwardedIP(webACLName, "MATCH", "X-Forwarded-For", "FIRST"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
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
				Config: testAccWebACLConfig_ipsetReferenceForwardedIP(webACLName, "NO_MATCH", "X-Forwarded-For", "LAST"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
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
				Config: testAccWebACLConfig_ipsetReferenceForwardedIP(webACLName, "MATCH", "Updated", "ANY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
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
				Config: testAccWebACLConfig_ipsetReference(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
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
				ImportStateIdFunc: testAccWebACLImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccWAFV2WebACL_RateBased_customKeys(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_rateBasedStatement_customKeysBasic(webACLName, "cookie", "testcookie"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
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
				Config: testAccWebACLConfig_rateBasedStatement_customKeysForwardedIP(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
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
				Config: testAccWebACLConfig_rateBasedStatement_customKeysHTTPMethod(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
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
				Config: testAccWebACLConfig_rateBasedStatement_customKeysBasic(webACLName, names.AttrHeader, "x-forwrded-for"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
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
				Config: testAccWebACLConfig_rateBasedStatement_customKeysIP(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
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
				Config: testAccWebACLConfig_rateBasedStatement_customKeysBasic(webACLName, "query_argument", names.AttrKey),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
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
				Config: testAccWebACLConfig_rateBasedStatement_customKeysMinimal(webACLName, "query_string"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
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
				Config: testAccWebACLConfig_rateBasedStatement_customKeysMinimal(webACLName, "uri_path"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
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
				Config: testAccWebACLConfig_rateBasedStatement_customKeysMaxKeys(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
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
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccWebACLImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccWAFV2WebACL_RateBased_forwardedIP(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_rateBasedStatementForwardedIP(webACLName, "MATCH", "X-Forwarded-For"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrName:                       "rule-1",
						"action.#":                           acctest.Ct1,
						"action.0.allow.#":                   acctest.Ct0,
						"action.0.block.#":                   acctest.Ct0,
						"action.0.count.#":                   acctest.Ct1,
						"statement.#":                        acctest.Ct1,
						"statement.0.rate_based_statement.#": acctest.Ct1,
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
				Config: testAccWebACLConfig_rateBasedStatementForwardedIP(webACLName, "NO_MATCH", "Updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrName:                       "rule-1",
						"action.#":                           acctest.Ct1,
						"action.0.allow.#":                   acctest.Ct0,
						"action.0.block.#":                   acctest.Ct0,
						"action.0.count.#":                   acctest.Ct1,
						"statement.#":                        acctest.Ct1,
						"statement.0.rate_based_statement.#": acctest.Ct1,
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
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccWebACLImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccWAFV2WebACL_RuleGroupReference_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_ruleGroupReferenceStatement(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrName:                                 "rule-1",
						"override_action.#":                            acctest.Ct1,
						"override_action.0.count.#":                    acctest.Ct1,
						"override_action.0.none.#":                     acctest.Ct0,
						"statement.#":                                  acctest.Ct1,
						"statement.0.rule_group_reference_statement.#": acctest.Ct1,
						"statement.0.rule_group_reference_statement.0.rule_action_override.#": acctest.Ct0,
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]*regexp.Regexp{
						"statement.0.rule_group_reference_statement.0.arn": regexache.MustCompile(`regional/rulegroup/.+$`),
					}),
				),
			},
			{
				Config: testAccWebACLConfig_ruleGroupReferenceStatementUpdate(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrName:                                 "rule-1",
						"override_action.#":                            acctest.Ct1,
						"override_action.0.count.#":                    acctest.Ct1,
						"override_action.0.none.#":                     acctest.Ct0,
						"statement.#":                                  acctest.Ct1,
						"statement.0.rule_group_reference_statement.#": acctest.Ct1,
						"statement.0.rule_group_reference_statement.0.rule_action_override.#":      acctest.Ct2,
						"statement.0.rule_group_reference_statement.0.rule_action_override.0.name": "rule-to-exclude-b",
						"statement.0.rule_group_reference_statement.0.rule_action_override.1.name": "rule-to-exclude-a",
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]*regexp.Regexp{
						"statement.0.rule_group_reference_statement.0.arn": regexache.MustCompile(`regional/rulegroup/.+$`),
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

// Ensure magically-added (i.e., AWS-added) rule for Shield with CF distribution DDoS auto
// mitigation does not cause diff and provider doesn't attempt to remove.
// See https://github.com/hashicorp/terraform-provider-aws/issues/22869
func TestAccWAFV2WebACL_RuleGroupReference_shieldMitigation(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_ruleGroupForShieldMitigation(webACLName, "desc1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "desc1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct0),
				),
			},
			{
				// Currently, there is no way to use provider to enable automatic application layer
				// DDoS mitigation with Shield for CF distribution. Doing so adds an out-of-band rule
				// similar to the one added below.
				PreConfig: func() {
					conn := acctest.Provider.Meta().(*conns.AWSClient).WAFV2Client(ctx)

					input := &wafv2.ListWebACLsInput{
						Scope: awstypes.ScopeRegional,
					}

					aclID := ""
					lockToken := ""

					err := tfwafv2.ListWebACLsPages(ctx, conn, input, func(page *wafv2.ListWebACLsOutput, lastPage bool) bool {
						if page == nil {
							return !lastPage
						}

						for _, acl := range page.WebACLs {
							if aws.ToString(acl.Name) == webACLName {
								aclID = aws.ToString(acl.Id)
								lockToken = aws.ToString(acl.LockToken)

								return false
							}
						}

						return !lastPage
					})

					if err != nil {
						t.Fatalf("finding WebACL (%s): %s", webACLName, err)
					}

					if aclID == "" {
						t.Fatalf("couldn't find WebACL (%s)", webACLName)
					}

					in := &wafv2.ListRuleGroupsInput{
						Scope: awstypes.ScopeRegional,
					}

					rgARN := ""

					err = tfwafv2.ListRuleGroupsPages(ctx, conn, in, func(page *wafv2.ListRuleGroupsOutput, lastPage bool) bool {
						if page == nil {
							return !lastPage
						}

						for _, rg := range page.RuleGroups {
							if aws.ToString(rg.Name) == fmt.Sprintf("rule-group-%s", webACLName) {
								rgARN = aws.ToString(rg.ARN)

								return false
							}
						}

						return !lastPage
					})

					if err != nil {
						t.Fatalf("finding rule group (%s): %s", webACLName, err)
					}

					if rgARN == "" {
						t.Fatalf("couldn't find Rule Group (%s)", webACLName)
					}

					_, err = conn.UpdateWebACL(ctx, &wafv2.UpdateWebACLInput{
						DefaultAction: &awstypes.DefaultAction{
							Allow: &awstypes.AllowAction{},
						},
						Id:        aws.String(aclID),
						LockToken: aws.String(lockToken),
						Name:      aws.String(webACLName),
						Rules: []awstypes.Rule{{
							Name:     aws.String("ShieldMitigationRuleGroup_012345678901_5e665b1c-1641-4b7a-8db1-567871a18b2a_uniqueid"),
							Priority: int32(11),
							OverrideAction: &awstypes.OverrideAction{
								None: &awstypes.NoneAction{},
							},
							Statement: &awstypes.Statement{
								RuleGroupReferenceStatement: &awstypes.RuleGroupReferenceStatement{
									ARN: aws.String(rgARN),
								},
							},
							VisibilityConfig: &awstypes.VisibilityConfig{
								CloudWatchMetricsEnabled: true,
								MetricName:               aws.String("ShieldMitigationRuleGroup_012345678901_5e665b1c-1641-4b7a-8db1-567871a18b2a_uniqueid"),
								SampledRequestsEnabled:   true,
							},
						}},
						Scope: awstypes.ScopeRegional,
						VisibilityConfig: &awstypes.VisibilityConfig{
							CloudWatchMetricsEnabled: true,
							MetricName:               aws.String("friendly-metric-name"),
							SampledRequestsEnabled:   false,
						},
					})
					if err != nil {
						t.Fatalf("adding rule in PreConfig: %s", err)
					}

					time.Sleep(15 * time.Second) // mitigate possible eventual consistency lag

					output, err := tfwafv2.FindWebACLByThreePartKey(ctx, conn, aclID, webACLName, "REGIONAL")
					if err != nil {
						t.Fatalf("finding WebACL (%s) in PreConfig: %s", webACLName, err)
					}

					if len(output.WebACL.Rules) < 1 {
						t.Fatalf("out-of-band added rule (%s) not found, cannot test handling of rule", webACLName)
					}
				},
				Config: testAccWebACLConfig_ruleGroupForShieldMitigation(webACLName, "desc1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "desc1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct0),
				),
			},
			{
				Config: testAccWebACLConfig_ruleGroupForShieldMitigation(webACLName, "desc2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "desc2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct0),
				),
			},
		},
	})
}

// Ensure magically-added (i.e., AWS-added) rule for Shield with CF distribution DDoS auto
// mitigation does not cause diff and provider doesn't attempt to remove.
// See https://github.com/hashicorp/terraform-provider-aws/issues/22869
func TestAccWAFV2WebACL_RuleGroupReference_manageShieldMitigationRule(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_ruleGroupShieldMitigation(webACLName, "desc1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "desc1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
				),
			},
			{
				Config:   testAccWebACLConfig_ruleGroupShieldMitigation(webACLName, "desc1"),
				PlanOnly: true,
			},
			{
				Config: testAccWebACLConfig_ruleGroupShieldMitigation(webACLName, "desc2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "desc2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
				),
			},
			{
				Config:   testAccWebACLConfig_ruleGroupShieldMitigation(webACLName, "desc2"),
				PlanOnly: true,
			},
		},
	})
}

func TestAccWAFV2WebACL_Custom_requestHandling(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_customRequestHandlingAllow(webACLName, "x-hdr1", "x-hdr2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, "REGIONAL"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrName:     "rule-1",
						"action.#":         acctest.Ct1,
						"action.0.allow.#": acctest.Ct1,
						"action.0.allow.0.custom_request_handling.#":                       acctest.Ct1,
						"action.0.allow.0.custom_request_handling.0.insert_header.#":       acctest.Ct2,
						"action.0.allow.0.custom_request_handling.0.insert_header.0.name":  "x-hdr1",
						"action.0.allow.0.custom_request_handling.0.insert_header.0.value": "test-value-1",
						"action.0.allow.0.custom_request_handling.0.insert_header.1.name":  "x-hdr2",
						"action.0.allow.0.custom_request_handling.0.insert_header.1.value": "test-value-2",
						"action.0.block.#":     acctest.Ct0,
						"action.0.captcha.#":   acctest.Ct0,
						"action.0.challenge.#": acctest.Ct0,
						"action.0.count.#":     acctest.Ct0,
						names.AttrPriority:     acctest.Ct1,
					}),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccWebACLImportStateIdFunc(resourceName),
			},
			{
				Config: testAccWebACLConfig_customRequestHandlingCount(webACLName, "x-hdr1", "x-hdr2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, "REGIONAL"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrName:         "rule-1",
						"action.#":             acctest.Ct1,
						"action.0.allow.#":     acctest.Ct0,
						"action.0.block.#":     acctest.Ct0,
						"action.0.captcha.#":   acctest.Ct0,
						"action.0.challenge.#": acctest.Ct0,
						"action.0.count.#":     acctest.Ct1,
						"action.0.count.0.custom_request_handling.#":                       acctest.Ct1,
						"action.0.count.0.custom_request_handling.0.insert_header.#":       acctest.Ct2,
						"action.0.count.0.custom_request_handling.0.insert_header.0.name":  "x-hdr1",
						"action.0.count.0.custom_request_handling.0.insert_header.0.value": "test-value-1",
						"action.0.count.0.custom_request_handling.0.insert_header.1.name":  "x-hdr2",
						"action.0.count.0.custom_request_handling.0.insert_header.1.value": "test-value-2",
						names.AttrPriority: acctest.Ct1,
					}),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", acctest.CtFalse),
				),
			},
			{
				Config: testAccWebACLConfig_customRequestHandlingCaptcha(webACLName, "x-hdr1", "x-hdr2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, "REGIONAL"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrName:         "rule-1",
						"action.#":             acctest.Ct1,
						"action.0.allow.#":     acctest.Ct0,
						"action.0.block.#":     acctest.Ct0,
						"action.0.captcha.#":   acctest.Ct1,
						"action.0.challenge.#": acctest.Ct0,
						"action.0.captcha.0.custom_request_handling.#":                       acctest.Ct1,
						"action.0.captcha.0.custom_request_handling.0.insert_header.#":       acctest.Ct2,
						"action.0.captcha.0.custom_request_handling.0.insert_header.0.name":  "x-hdr1",
						"action.0.captcha.0.custom_request_handling.0.insert_header.0.value": "test-value-1",
						"action.0.captcha.0.custom_request_handling.0.insert_header.1.name":  "x-hdr2",
						"action.0.captcha.0.custom_request_handling.0.insert_header.1.value": "test-value-2",
						"action.0.count.#": acctest.Ct0,
						names.AttrPriority: acctest.Ct1,
						"captcha_config.#": acctest.Ct1,
						"captcha_config.0.immunity_time_property.0.immunity_time": "240",
					}),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "captcha_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "captcha_config.0.immunity_time_property.0.immunity_time", "120"),
					resource.TestCheckResourceAttr(resourceName, "challenge_config.0.immunity_time_property.0.immunity_time", "300"),
				),
			},
			{
				Config: testAccWebACLConfig_customRequestHandlingChallenge(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, "REGIONAL"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrName:         "rule-1",
						"action.#":             acctest.Ct1,
						"action.0.allow.#":     acctest.Ct0,
						"action.0.block.#":     acctest.Ct0,
						"action.0.captcha.#":   acctest.Ct0,
						"action.0.challenge.#": acctest.Ct1,
						"action.0.count.#":     acctest.Ct0,
						names.AttrPriority:     acctest.Ct1,
					}),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccWAFV2WebACL_Custom_response(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_customResponse(webACLName, 401, 403, "x-hdr1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.0.custom_response.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.0.custom_response.0.response_code", "401"),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, "REGIONAL"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrName:                       "rule-1",
						"action.#":                           acctest.Ct1,
						"action.0.allow.#":                   acctest.Ct0,
						"action.0.block.#":                   acctest.Ct1,
						"action.0.block.0.custom_response.#": acctest.Ct1,
						"action.0.block.0.custom_response.0.response_code":           "403",
						"action.0.block.0.custom_response.0.response_header.#":       acctest.Ct1,
						"action.0.block.0.custom_response.0.response_header.0.name":  "x-hdr1",
						"action.0.block.0.custom_response.0.response_header.0.value": "custom-response-header-value",
						"action.0.count.#": acctest.Ct0,
						names.AttrPriority: acctest.Ct1,
					}),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", acctest.CtFalse),
				),
			},
			{
				Config: testAccWebACLConfig_customResponse(webACLName, 404, 429, "x-hdr2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.0.custom_response.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.0.custom_response.0.response_code", "404"),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, "REGIONAL"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrName:                       "rule-1",
						"action.#":                           acctest.Ct1,
						"action.0.allow.#":                   acctest.Ct0,
						"action.0.block.#":                   acctest.Ct1,
						"action.0.block.0.custom_response.#": acctest.Ct1,
						"action.0.block.0.custom_response.0.response_code":           "429",
						"action.0.block.0.custom_response.0.response_header.#":       acctest.Ct1,
						"action.0.block.0.custom_response.0.response_header.0.name":  "x-hdr2",
						"action.0.block.0.custom_response.0.response_header.0.value": "custom-response-header-value",
						"action.0.count.#": acctest.Ct0,
						names.AttrPriority: acctest.Ct1,
					}),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", acctest.CtFalse),
				),
			},
			{
				Config: testAccWebACLConfig_customResponseBody(webACLName, 404, 429),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, "custom_response_body.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "custom_response_body.0.key", "test_body"),
					resource.TestCheckResourceAttr(resourceName, "custom_response_body.0.content", "<html><body>Oops<body></html>"),
					resource.TestCheckResourceAttr(resourceName, "custom_response_body.0.content_type", "TEXT_HTML"),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.0.custom_response.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.0.custom_response.0.response_code", "404"),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, "REGIONAL"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrName:                       "rule-1",
						"action.#":                           acctest.Ct1,
						"action.0.allow.#":                   acctest.Ct0,
						"action.0.block.#":                   acctest.Ct1,
						"action.0.block.0.custom_response.#": acctest.Ct1,
						"action.0.block.0.custom_response.0.response_code":            "429",
						"action.0.block.0.custom_response.0.custom_response_body_key": "test_body",
						"action.0.count.#": acctest.Ct0,
						names.AttrPriority: acctest.Ct1,
					}),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", acctest.CtFalse),
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

func TestAccWAFV2WebACL_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_tags1(webACLName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccWebACLImportStateIdFunc(resourceName),
			},
			{
				Config: testAccWebACLConfig_tags2(webACLName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccWebACLConfig_tags1(webACLName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/13862
func TestAccWAFV2WebACL_RateBased_maxNested(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_multipleNestedRateBasedStatements(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                        acctest.Ct1,
						"statement.0.rate_based_statement.#": acctest.Ct1,
						"statement.0.rate_based_statement.0.aggregate_key_type":                                                                                                    "IP",
						"statement.0.rate_based_statement.0.evaluation_window_sec":                                                                                                 "300",
						"statement.0.rate_based_statement.0.limit":                                                                                                                 "300",
						"statement.0.rate_based_statement.0.scope_down_statement.#":                                                                                                acctest.Ct1,
						"statement.0.rate_based_statement.0.scope_down_statement.0.not_statement.#":                                                                                acctest.Ct1,
						"statement.0.rate_based_statement.0.scope_down_statement.0.not_statement.0.statement.#":                                                                    acctest.Ct1,
						"statement.0.rate_based_statement.0.scope_down_statement.0.not_statement.0.statement.0.or_statement.#":                                                     acctest.Ct1,
						"statement.0.rate_based_statement.0.scope_down_statement.0.not_statement.0.statement.0.or_statement.0.statement.#":                                         acctest.Ct3,
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

func TestAccWAFV2WebACL_Operators_maxNested(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_multipleNestedOperatorStatements(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
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

func TestAccWAFV2WebACL_tokenDomains(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain1 := "mywebsite.com"
	domain2 := "myotherwebsite.com"
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckScopeRegional(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_tokenDomains(webACLName, domain1, domain2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, webACLName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, string(awstypes.ScopeRegional)),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "token_domains.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "token_domains.*", domain1),
					resource.TestCheckTypeSetElemAttr(resourceName, "token_domains.*", domain2),
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
				ImportStateIdFunc: testAccWebACLImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccWAFV2WebACL_associationConfigCloudFront(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckWAFV2CloudFrontScope(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_associationConfigCloudFront(webACLName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`global/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "association_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "association_config.0.request_body.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "association_config.0.request_body.0.cloudfront.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "association_config.0.request_body.0.cloudfront.0.default_size_inspection_limit", "KB_64"),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, webACLName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, string(awstypes.ScopeCloudfront)),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", acctest.CtFalse),
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

func TestAccWAFV2WebACL_associationConfigRegional(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_associationConfigRegional(webACLName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "association_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "association_config.0.request_body.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "association_config.0.request_body.0.api_gateway.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "association_config.0.request_body.0.api_gateway.0.default_size_inspection_limit", "KB_16"),
					resource.TestCheckResourceAttr(resourceName, "association_config.0.request_body.0.cognito_user_pool.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "association_config.0.request_body.0.cognito_user_pool.0.default_size_inspection_limit", "KB_32"),
					resource.TestCheckResourceAttr(resourceName, "association_config.0.request_body.0.app_runner_service.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "association_config.0.request_body.0.app_runner_service.0.default_size_inspection_limit", "KB_48"),
					resource.TestCheckResourceAttr(resourceName, "association_config.0.request_body.0.verified_access_instance.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "association_config.0.request_body.0.verified_access_instance.0.default_size_inspection_limit", "KB_64"),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, webACLName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, string(awstypes.ScopeRegional)),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", acctest.CtFalse),
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

func TestAccWAFV2WebACL_CloudFrontScope(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckWAFV2CloudFrontScope(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_CloudFrontScope(webACLName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "wafv2", regexache.MustCompile(`global/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, webACLName),
					resource.TestCheckResourceAttr(resourceName, names.AttrScope, string(awstypes.ScopeCloudfront)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrName: "rule-1",
						"statement.#":  acctest.Ct1,
						"statement.0.managed_rule_group_statement.#":                                                                                                                       acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_atp_rule_set.#":                                                         acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_atp_rule_set.0.login_path":                                              "/api/1/signin",
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_atp_rule_set.0.request_inspection.#":                                    acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_atp_rule_set.0.request_inspection.0.password_field.#":                   acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_atp_rule_set.0.request_inspection.0.password_field.0.identifier":        "/password",
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_atp_rule_set.0.request_inspection.0.payload_type":                       "JSON",
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_atp_rule_set.0.request_inspection.0.username_field.#":                   acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_atp_rule_set.0.request_inspection.0.username_field.0.identifier":        "/username",
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_atp_rule_set.0.response_inspection.#":                                   acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_atp_rule_set.0.response_inspection.0.body_contains.#":                   acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_atp_rule_set.0.response_inspection.0.body_contains.0.success_strings.#": acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_atp_rule_set.0.response_inspection.0.body_contains.0.success_strings.0": "Login successful",
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_atp_rule_set.0.response_inspection.0.body_contains.0.failure_strings.#": acctest.Ct1,
						"statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_atp_rule_set.0.response_inspection.0.body_contains.0.failure_strings.0": "Login failed",
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

func testAccCheckWebACLDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_wafv2_web_acl" {
				continue
			}

			_, err := tfwafv2.FindWebACLByThreePartKey(ctx, conn, rs.Primary.ID, rs.Primary.Attributes[names.AttrName], rs.Primary.Attributes[names.AttrScope])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("WAFv2 WebACL %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckWebACLExists(ctx context.Context, n string, v *awstypes.WebACL) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFV2Client(ctx)

		output, err := tfwafv2.FindWebACLByThreePartKey(ctx, conn, rs.Primary.ID, rs.Primary.Attributes[names.AttrName], rs.Primary.Attributes[names.AttrScope])

		if err != nil {
			return err
		}

		*v = *output.WebACL

		return nil
	}
}

func testAccWebACLImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s/%s", rs.Primary.ID, rs.Primary.Attributes[names.AttrName], rs.Primary.Attributes[names.AttrScope]), nil
	}
}

func testAccWebACLConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccWebACLConfig_basicRule(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = "Updated"
  scope       = "REGIONAL"

  default_action {
    block {}
  }

  rule {
    name     = "%[1]s-1"
    priority = 10

    action {
      count {}
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
      metric_name                = "%[1]s-metric-name-1"
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

func testAccWebACLConfig_updateRuleNamePriorityMetric(rName, ruleName1, ruleName2 string, priority1, priority2 int) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = "Updated"
  scope       = "REGIONAL"

  default_action {
    block {}
  }

  rule {
    name     = %[2]q
    priority = %[3]d

    action {
      count {}
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
      metric_name                = %[2]q
      sampled_requests_enabled   = false
    }
  }

  rule {
    name     = %[4]q
    priority = %[5]d

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

func testAccWebACLConfig_byteMatchStatement(rName, positionalConstraint, searchString string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

  rule {
    name     = "rule-1"
    priority = 1

    action {
      count {}
    }

    statement {
      byte_match_statement {
        field_to_match {
          all_query_arguments {}
        }
        positional_constraint = %[2]q
        search_string         = %[3]q
        text_transformation {
          priority = 0
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
`, rName, positionalConstraint, searchString)
}

func testAccWebACLConfig_byteMatchStatementJA3Fingerprint(rName, fallbackBehavior string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

  rule {
    name     = "rule-1"
    priority = 1

    action {
      count {}
    }

    statement {
      byte_match_statement {
        field_to_match {
          ja3_fingerprint {
            fallback_behavior = %[2]q
          }
        }
        positional_constraint = "EXACTLY"
        search_string         = "abcdef1234567890abcdef1234567890"
        text_transformation {
          priority = 0
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
`, rName, fallbackBehavior)
}

func testAccWebACLConfig_byteMatchStatementJSONBody(rName, matchScope, invalidFallbackBehavior, oversizeHandling, matchPattern string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

  rule {
    name     = "rule-1"
    priority = 1

    action {
      count {}
    }

    statement {
      byte_match_statement {
        field_to_match {
          json_body {
            match_scope               = %[2]q
            invalid_fallback_behavior = %[3]q
            oversize_handling         = %[4]q
            match_pattern {
              %[5]s
            }
          }
        }
        positional_constraint = "CONTAINS_WORD"
        search_string         = "Buddy"
        text_transformation {
          priority = 0
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
`, rName, matchScope, invalidFallbackBehavior, oversizeHandling, matchPattern)
}

func testAccWebACLConfig_byteMatchStatementBody(rName, oversizeHandling string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

  rule {
    name     = "rule-1"
    priority = 1

    action {
      count {}
    }

    statement {
      byte_match_statement {
        field_to_match {
          body {
            oversize_handling = %[2]q
          }
        }
        positional_constraint = "CONTAINS_WORD"
        search_string         = "Buddy"
        text_transformation {
          priority = 0
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
`, rName, oversizeHandling)
}

func testAccWebACLConfig_byteMatchStatementHeaderOrder(rName, oversizeHandling string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

  rule {
    name     = "rule-1"
    priority = 1

    action {
      count {}
    }

    statement {
      byte_match_statement {
        search_string = "host:user-agent:accept:authorization:referer"
        field_to_match {
          header_order {
            oversize_handling = %[2]q
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
`, rName, oversizeHandling)
}

func testAccWebACLConfig_geoMatchStatement(rName, countryCodes string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

  rule {
    name     = "rule-1"
    priority = 1

    action {
      block {}
    }

    statement {
      geo_match_statement {
        country_codes = [%[2]s]
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName, countryCodes)
}

func testAccWebACLConfig_labelMatchStatement(rName, labelScope string, labelKey string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"
  default_action {
    allow {}
  }
  rule {
    name     = "rule-1"
    priority = 1
    action {
      count {}
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
  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }
  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName, labelScope, labelKey)
}

func testAccWebACLConfig_ruleLabels(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"
  default_action {
    allow {}
  }
  rule {
    name     = "rule-1"
    priority = 1
    action {
      block {}
    }
    rule_label {
      name = "Hashicorp:Test:Label1"
    }
    rule_label {
      name = "Hashicorp:Test:Label2"
    }
    statement {
      geo_match_statement {
        country_codes = ["US", "CA"]
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

func testAccWebACLConfig_noRuleLabels(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"
  default_action {
    allow {}
  }
  rule {
    name     = "rule-1"
    priority = 1
    action {
      block {}
    }
    statement {
      geo_match_statement {
        country_codes = ["US", "CA"]
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

func testAccWebACLConfig_customRequestHandlingAllow(rName, firstHeader string, secondHeader string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

  rule {
    name     = "rule-1"
    priority = 1

    action {
      allow {
        custom_request_handling {
          insert_header {
            name  = %[2]q
            value = "test-value-1"
          }

          insert_header {
            name  = %[3]q
            value = "test-value-2"
          }
        }
      }
    }

    statement {
      geo_match_statement {
        country_codes = ["US", "CA"]
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
`, rName, firstHeader, secondHeader)
}

func testAccWebACLConfig_customRequestHandlingCaptcha(rName, firstHeader string, secondHeader string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

  rule {
    name     = "rule-1"
    priority = 1

    action {
      captcha {
        custom_request_handling {
          insert_header {
            name  = %[2]q
            value = "test-value-1"
          }

          insert_header {
            name  = %[3]q
            value = "test-value-2"
          }
        }
      }
    }

    statement {
      geo_match_statement {
        country_codes = ["US", "CA"]
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }

    captcha_config {
      immunity_time_property {
        immunity_time = 240
      }
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }

  captcha_config {
    immunity_time_property {
      immunity_time = 120
    }
  }

  challenge_config {
    immunity_time_property {
      immunity_time = 300
    }
  }
}
`, rName, firstHeader, secondHeader)
}

func testAccWebACLConfig_customRequestHandlingChallenge(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

  rule {
    name     = "rule-1"
    priority = 1

    action {
      challenge {}
    }

    statement {
      geo_match_statement {
        country_codes = ["US", "CA"]
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

func testAccWebACLConfig_customRequestHandlingCount(rName, firstHeader string, secondHeader string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

  rule {
    name     = "rule-1"
    priority = 1

    action {
      count {
        custom_request_handling {
          insert_header {
            name  = %[2]q
            value = "test-value-1"
          }

          insert_header {
            name  = %[3]q
            value = "test-value-2"
          }
        }
      }
    }

    statement {
      geo_match_statement {
        country_codes = ["US", "CA"]
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
`, rName, firstHeader, secondHeader)
}

func testAccWebACLConfig_customResponse(rName string, defaultStatusCode int, countryBlockStatusCode int, countryHeaderName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  default_action {
    block {
      custom_response {
        response_code = %[2]d
      }
    }
  }

  rule {
    name     = "rule-1"
    priority = 1

    action {
      block {
        custom_response {
          response_code = %[3]d

          response_header {
            name  = %[4]q
            value = "custom-response-header-value"
          }
        }
      }
    }

    statement {
      geo_match_statement {
        country_codes = ["US", "CA"]
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
`, rName, defaultStatusCode, countryBlockStatusCode, countryHeaderName)
}

func testAccWebACLConfig_customResponseBody(rName string, defaultStatusCode int, countryBlockStatusCode int) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"
  default_action {
    block {
      custom_response {
        response_code = %[2]d
      }
    }
  }
  custom_response_body {
    key          = "test_body"
    content      = "<html><body>Oops<body></html>"
    content_type = "TEXT_HTML"
  }
  rule {
    name     = "rule-1"
    priority = 1
    action {
      block {
        custom_response {
          response_code            = %[3]d
          custom_response_body_key = "test_body"
        }
      }
    }
    statement {
      geo_match_statement {
        country_codes = ["US", "CA"]
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
`, rName, defaultStatusCode, countryBlockStatusCode)
}

func testAccWebACLConfig_geoMatchStatementForwardedIP(rName, fallbackBehavior, headerName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  default_action {
    block {}
  }

  rule {
    name     = "rule-1"
    priority = 1

    action {
      block {}
    }

    statement {
      or_statement {
        statement {
          geo_match_statement {
            country_codes = ["US"]
          }
        }

        statement {
          geo_match_statement {
            country_codes = ["CA"]
            forwarded_ip_config {
              fallback_behavior = %[2]q
              header_name       = %[3]q
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
`, rName, fallbackBehavior, headerName)
}

func testAccWebACLConfig_ipsetReference(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_ip_set" "test" {
  name               = "ip-set-%[1]s"
  scope              = "REGIONAL"
  ip_address_version = "IPV4"
  addresses          = ["1.1.1.1/32", "2.2.2.2/32"]
}

resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  default_action {
    block {}
  }

  rule {
    name     = "rule-1"
    priority = 1

    action {
      block {}
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

func testAccWebACLConfig_ipsetReferenceForwardedIP(rName, fallbackBehavior, headerName, position string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_ip_set" "test" {
  name               = "ip-set-%[1]s"
  scope              = "REGIONAL"
  ip_address_version = "IPV4"
  addresses          = ["1.1.1.1/32", "2.2.2.2/32"]
}

resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  default_action {
    block {}
  }

  rule {
    name     = "rule-1"
    priority = 1

    action {
      block {}
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

func testAccWebACLConfig_managedRuleGroupStatement(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

  rule {
    name     = "rule-1"
    priority = 1

    override_action {
      none {}
    }

    statement {
      managed_rule_group_statement {
        name        = "AWSManagedRulesCommonRuleSet"
        vendor_name = "AWS"
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccWebACLConfig_managedRuleGroupStatementManagedRuleGroupConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

  rule {
    name     = "rule-1"
    priority = 1

    override_action {
      none {}
    }

    statement {
      managed_rule_group_statement {
        name        = "AWSManagedRulesATPRuleSet"
        vendor_name = "AWS"

        managed_rule_group_configs {
          login_path = "/login"
        }
        managed_rule_group_configs {
          payload_type = "JSON"
        }
        managed_rule_group_configs {
          password_field {
            identifier = "/password"
          }
        }
        managed_rule_group_configs {
          username_field {
            identifier = "/username"
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

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccWebACLConfig_managedRuleGroupStatementManagedRuleGroupConfigUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

  rule {
    name     = "rule-1"
    priority = 1

    override_action {
      none {}
    }

    statement {
      managed_rule_group_statement {
        name        = "AWSManagedRulesATPRuleSet"
        vendor_name = "AWS"

        managed_rule_group_configs {
          login_path = "/app-login"
        }
        managed_rule_group_configs {
          payload_type = "JSON"
        }
        managed_rule_group_configs {
          password_field {
            identifier = "/app-password"
          }
        }
        managed_rule_group_configs {
          username_field {
            identifier = "/app-username"
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

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccWebACLConfig_managedRuleGroupStatementManagedRuleGroupConfig_acfpRuleSet(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

  rule {
    name     = "rule-1"
    priority = 1

    override_action {
      none {}
    }

    statement {
      managed_rule_group_statement {
        name        = "AWSManagedRulesACFPRuleSet"
        vendor_name = "AWS"

        managed_rule_group_configs {
          aws_managed_rules_acfp_rule_set {
            creation_path          = "/creation"
            registration_page_path = "/registration"
            request_inspection {
              email_field {
                identifier = "/email"
              }
              password_field {
                identifier = "/password"
              }
              phone_number_fields {
                identifiers = ["/phone1", "/phone2"]
              }
              address_fields {
                identifiers = ["home", "work"]
              }
              payload_type = "JSON"
              username_field {
                identifier = "/username"
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

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccWebACLConfig_managedRuleGroupStatementManagedRuleGroupConfig_atpRuleSet(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

  rule {
    name     = "rule-1"
    priority = 1

    override_action {
      none {}
    }

    statement {
      managed_rule_group_statement {
        name        = "AWSManagedRulesATPRuleSet"
        vendor_name = "AWS"

        managed_rule_group_configs {
          aws_managed_rules_atp_rule_set {
            login_path = "/api/1/signin"
            request_inspection {
              password_field {
                identifier = "/password"
              }
              payload_type = "JSON"
              username_field {
                identifier = "/username"
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

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccWebACLConfig_managedRuleGroupStatementManagedRuleGroupConfig_acfpRuleSetUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

  rule {
    name     = "rule-1"
    priority = 1

    override_action {
      none {}
    }

    statement {
      managed_rule_group_statement {
        name        = "AWSManagedRulesACFPRuleSet"
        vendor_name = "AWS"

        managed_rule_group_configs {
          aws_managed_rules_acfp_rule_set {
            enable_regex_in_path   = true
            creation_path          = "/creation"
            registration_page_path = "/registration"

            request_inspection {
              email_field {
                identifier = "/email"
              }
              password_field {
                identifier = "/pass"
              }
              phone_number_fields {
                identifiers = ["/phone3"]
              }
              address_fields {
                identifiers = ["mobile"]
              }
              payload_type = "JSON"
              username_field {
                identifier = "/user"
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

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccWebACLConfig_managedRuleGroupStatementManagedRuleGroupConfig_atpRuleSetUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

  rule {
    name     = "rule-1"
    priority = 1

    override_action {
      none {}
    }

    statement {
      managed_rule_group_statement {
        name        = "AWSManagedRulesATPRuleSet"
        vendor_name = "AWS"

        managed_rule_group_configs {
          aws_managed_rules_atp_rule_set {
            enable_regex_in_path = true
            login_path           = "/api/2/signin"

            request_inspection {
              password_field {
                identifier = "/pass"
              }
              payload_type = "JSON"
              username_field {
                identifier = "/user"
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

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccWebACLConfig_managedRuleGroupStatementManagedRuleGroupConfig_botControl(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

  rule {
    name     = "rule-1"
    priority = 1

    override_action {
      none {}
    }

    statement {
      managed_rule_group_statement {
        name        = "AWSManagedRulesBotControlRuleSet"
        vendor_name = "AWS"

        managed_rule_group_configs {
          aws_managed_rules_bot_control_rule_set {
            inspection_level        = "TARGETED"
            enable_machine_learning = false
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

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccWebACLConfig_managedRuleGroupStatementUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

  rule {
    name     = "rule-1"
    priority = 1

    override_action {
      count {}
    }

    statement {
      managed_rule_group_statement {
        name        = "AWSManagedRulesCommonRuleSet"
        vendor_name = "AWS"

        rule_action_override {
          action_to_use {
            count {}
          }

          name = "SizeRestrictions_QUERYSTRING"
        }

        rule_action_override {
          action_to_use {
            count {}
          }

          name = "NoUserAgent_HEADER"
        }

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

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccWebACLConfig_managedRuleGroupStatementRuleActionOverrides(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

  rule {
    name     = "rule-1"
    priority = 1

    override_action {
      count {}
    }

    statement {
      managed_rule_group_statement {
        name        = "AWSManagedRulesCommonRuleSet"
        vendor_name = "AWS"

        rule_action_override {
          action_to_use {
            count {}
          }

          name = "SizeRestrictions_QUERYSTRING"
        }

        rule_action_override {
          action_to_use {
            count {}
          }

          name = "NoUserAgent_HEADER"
        }

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

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccWebACLConfig_managedRuleGroupStatementVersionVersion10(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

  rule {
    name     = "rule-1"
    priority = 1

    override_action {
      none {}
    }

    statement {
      managed_rule_group_statement {
        name        = "AWSManagedRulesCommonRuleSet"
        vendor_name = "AWS"
        version     = "Version_1.0"
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccWebACLConfig_rateBasedStatement(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  default_action {
    block {}
  }

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

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccWebACLConfig_rateBasedStatementForwardedIP(rName, fallbackBehavior, headerName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  default_action {
    block {}
  }

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

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName, fallbackBehavior, headerName)
}

func testAccWebACLConfig_rateBasedStatement_customKeysBasic(rName, customKey, customKeyName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

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

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName, customKey, customKeyName)
}

func testAccWebACLConfig_rateBasedStatement_customKeysMinimal(rName, customKey string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

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

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName, customKey)
}

func testAccWebACLConfig_rateBasedStatement_customKeysIP(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

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

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }


  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccWebACLConfig_rateBasedStatement_customKeysForwardedIP(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

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

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccWebACLConfig_rateBasedStatement_customKeysHTTPMethod(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

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

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccWebACLConfig_rateBasedStatement_customKeysMaxKeys(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

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

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccWebACLConfig_rateBasedStatementUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  default_action {
    block {}
  }

  rule {
    name     = "rule-1"
    priority = 1

    action {
      count {}
    }

    statement {
      rate_based_statement {
        limit                 = 10000
        aggregate_key_type    = "IP"
        evaluation_window_sec = 600

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

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccWebACLConfig_ruleGroupReferenceStatement(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 10
  name     = "rule-group-%[1]s"
  scope    = "REGIONAL"

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}

resource "aws_wafv2_web_acl" "test" {
  name  = %[1]q
  scope = "REGIONAL"

  default_action {
    block {}
  }

  rule {
    name     = "rule-1"
    priority = 1

    override_action {
      count {}
    }

    statement {
      rule_group_reference_statement {
        arn = aws_wafv2_rule_group.test.arn
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccWebACLConfig_ruleGroupReferenceStatementUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 10
  name     = "rule-group-%[1]s"
  scope    = "REGIONAL"

  rule {
    name     = "rule-1"
    priority = 1

    action {
      count {}
    }

    statement {
      geo_match_statement {
        country_codes = ["NL"]
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  rule {
    name     = "rule-to-exclude-a"
    priority = 10

    action {
      allow {}
    }

    statement {
      geo_match_statement {
        country_codes = ["US"]
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  rule {
    name     = "rule-to-exclude-b"
    priority = 15

    action {
      allow {}
    }

    statement {
      geo_match_statement {
        country_codes = ["GB"]
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

resource "aws_wafv2_web_acl" "test" {
  name  = %[1]q
  scope = "REGIONAL"

  default_action {
    block {}
  }

  rule {
    name     = "rule-1"
    priority = 1

    override_action {
      count {}
    }

    statement {
      rule_group_reference_statement {
        arn = aws_wafv2_rule_group.test.arn

        rule_action_override {
          action_to_use {
            count {}
          }

          name = "rule-to-exclude-b"
        }

        rule_action_override {
          action_to_use {
            count {}
          }

          name = "rule-to-exclude-a"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "friendly-rule-metric-name"
      sampled_requests_enabled   = false
    }
  }

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccWebACLConfig_minimal(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name  = %[1]q
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccWebACLConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

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

func testAccWebACLConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

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

func testAccWebACLConfig_multipleNestedRateBasedStatements(rName string) string {
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

resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

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

func testAccWebACLConfig_multipleNestedOperatorStatements(rName string) string {
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

resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

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

func testAccWebACLConfig_ruleGroupShieldMitigation(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 10
  name     = "rule-group-%[1]s"
  scope    = "REGIONAL"

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}

data "aws_caller_identity" "current" {}

resource "aws_wafv2_web_acl" "test" {
  name  = %[1]q
  scope = "REGIONAL"

  default_action {
    block {}
  }

  description = %[2]q

  rule {
    name     = "ShieldMitigationRuleGroup_${data.aws_caller_identity.current.account_id}_5e665b1c-1641-4b7a-8db1-567871a18b2a_uniqueid"
    priority = 11

    override_action {
      none {}
    }

    statement {
      rule_group_reference_statement {
        arn = aws_wafv2_rule_group.test.arn
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = true
      metric_name                = "ShieldMitigationRuleGroup_${data.aws_caller_identity.current.account_id}_5e665b1c-1641-4b7a-8db1-567871a18b2a_uniqueid"
      sampled_requests_enabled   = true
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = true
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName, description)
}

func testAccWebACLConfig_ruleGroupForShieldMitigation(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 10
  name     = "rule-group-%[1]s"
  scope    = "REGIONAL"

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}

resource "aws_wafv2_web_acl" "test" {
  name  = %[1]q
  scope = "REGIONAL"

  default_action {
    block {}
  }

  description = %[2]q

  visibility_config {
    cloudwatch_metrics_enabled = true
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName, description)
}

func testAccWebACLConfig_tokenDomains(rName, domain1, domain2 string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

  token_domains = [%[2]q, %[3]q]
  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName, domain1, domain2)
}

func testAccWebACLConfig_CloudFrontScope(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "CLOUDFRONT"

  default_action {
    allow {}
  }

  rule {
    name     = "rule-1"
    priority = 1

    override_action {
      count {}
    }

    statement {
      managed_rule_group_statement {
        name        = "AWSManagedRulesATPRuleSet"
        vendor_name = "AWS"

        managed_rule_group_configs {
          aws_managed_rules_atp_rule_set {
            login_path = "/api/1/signin"
            request_inspection {
              password_field {
                identifier = "/password"
              }
              payload_type = "JSON"
              username_field {
                identifier = "/username"
              }
            }
            response_inspection {
              body_contains {
                success_strings = ["Login successful"]
                failure_strings = ["Login failed"]
              }
            }
          }
        }
      }
    }
    visibility_config {
      cloudwatch_metrics_enabled = true
      metric_name                = "AWSManagedRulesATPRuleSet_json"
      sampled_requests_enabled   = true
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

func testAccWebACLConfig_associationConfigCloudFront(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "CLOUDFRONT"

  default_action {
    allow {}
  }

  association_config {
    request_body {
      cloudfront {
        default_size_inspection_limit = "KB_64"
      }
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

func testAccWebACLConfig_associationConfigRegional(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = %[1]q
  description = %[1]q
  scope       = "REGIONAL"

  default_action {
    allow {}
  }

  association_config {
    request_body {
      api_gateway {
        default_size_inspection_limit = "KB_16"
      }
      cognito_user_pool {
        default_size_inspection_limit = "KB_32"
      }
      app_runner_service {
        default_size_inspection_limit = "KB_48"
      }
      verified_access_instance {
        default_size_inspection_limit = "KB_64"
      }
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
