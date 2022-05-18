package wafv2_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfwafv2 "github.com/hashicorp/terraform-provider-aws/internal/service/wafv2"
)

func TestAccWAFV2WebACL_basic(t *testing.T) {
	var v wafv2.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckScopeRegional(t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_Basic(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "description", webACLName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", "0"),
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
				ImportStateIdFunc: testAccWebACLImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccWAFV2WebACL_Update_rule(t *testing.T) {
	var v wafv2.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"
	ruleName1 := fmt.Sprintf("%s-1", webACLName)
	ruleName2 := fmt.Sprintf("%s-2", webACLName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckScopeRegional(t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_BasicRule(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                ruleName1,
						"priority":            "10",
						"action.#":            "1",
						"action.0.allow.#":    "0",
						"action.0.block.#":    "0",
						"action.0.count.#":    "1",
						"visibility_config.#": "1",
						"visibility_config.0.cloudwatch_metrics_enabled": "false",
						"visibility_config.0.metric_name":                fmt.Sprintf("%s-metric-name-1", webACLName),
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
				// Test step to verify additional rule block with first rule block unchanged
				Config: testAccWebACLConfig_UpdateRuleNamePriorityMetric(webACLName, ruleName1, ruleName2, 10, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                ruleName1,
						"priority":            "10",
						"action.#":            "1",
						"action.0.allow.#":    "0",
						"action.0.block.#":    "0",
						"action.0.count.#":    "1",
						"visibility_config.#": "1",
						"visibility_config.0.cloudwatch_metrics_enabled": "false",
						"visibility_config.0.metric_name":                ruleName1,
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
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                ruleName2,
						"priority":            "5",
						"action.#":            "1",
						"action.0.allow.#":    "1",
						"action.0.block.#":    "0",
						"action.0.count.#":    "0",
						"visibility_config.#": "1",
						"visibility_config.0.cloudwatch_metrics_enabled": "false",
						"visibility_config.0.metric_name":                ruleName2,
						"visibility_config.0.sampled_requests_enabled":   "false",
						"statement.#":                                       "1",
						"statement.0.geo_match_statement.#":                 "1",
						"statement.0.geo_match_statement.0.country_codes.#": "2",
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
	var v wafv2.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"
	ruleName1 := fmt.Sprintf("%s-1", webACLName)
	ruleName2 := fmt.Sprintf("%s-2", webACLName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckScopeRegional(t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_UpdateRuleNamePriorityMetric(webACLName, ruleName1, ruleName2, 5, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                ruleName1,
						"priority":            "5",
						"action.#":            "1",
						"action.0.allow.#":    "0",
						"action.0.block.#":    "0",
						"action.0.count.#":    "1",
						"visibility_config.#": "1",
						"visibility_config.0.cloudwatch_metrics_enabled": "false",
						"visibility_config.0.metric_name":                ruleName1,
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
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                ruleName2,
						"priority":            "10",
						"action.#":            "1",
						"action.0.allow.#":    "1",
						"action.0.block.#":    "0",
						"action.0.count.#":    "0",
						"visibility_config.#": "1",
						"visibility_config.0.cloudwatch_metrics_enabled": "false",
						"visibility_config.0.metric_name":                ruleName2,
						"visibility_config.0.sampled_requests_enabled":   "false",
						"statement.#":                                       "1",
						"statement.0.geo_match_statement.#":                 "1",
						"statement.0.geo_match_statement.0.country_codes.#": "2",
					}),
				),
			},
			{
				Config: testAccWebACLConfig_UpdateRuleNamePriorityMetric(webACLName, ruleName1, ruleName2, 10, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                ruleName1,
						"priority":            "10",
						"action.#":            "1",
						"action.0.allow.#":    "0",
						"action.0.block.#":    "0",
						"action.0.count.#":    "1",
						"visibility_config.#": "1",
						"visibility_config.0.cloudwatch_metrics_enabled": "false",
						"visibility_config.0.metric_name":                ruleName1,
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
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                ruleName2,
						"priority":            "5",
						"action.#":            "1",
						"action.0.allow.#":    "1",
						"action.0.block.#":    "0",
						"action.0.count.#":    "0",
						"visibility_config.#": "1",
						"visibility_config.0.cloudwatch_metrics_enabled": "false",
						"visibility_config.0.metric_name":                ruleName2,
						"visibility_config.0.sampled_requests_enabled":   "false",
						"statement.#":                                       "1",
						"statement.0.geo_match_statement.#":                 "1",
						"statement.0.geo_match_statement.0.country_codes.#": "2",
					}),
				),
			},
			{
				Config: testAccWebACLConfig_UpdateRuleNamePriorityMetric(webACLName, ruleName1, "updated", 10, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                ruleName1,
						"priority":            "10",
						"action.#":            "1",
						"action.0.allow.#":    "0",
						"action.0.block.#":    "0",
						"action.0.count.#":    "1",
						"visibility_config.#": "1",
						"visibility_config.0.cloudwatch_metrics_enabled": "false",
						"visibility_config.0.metric_name":                ruleName1,
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
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                "updated",
						"priority":            "5",
						"action.#":            "1",
						"action.0.allow.#":    "1",
						"action.0.block.#":    "0",
						"action.0.count.#":    "0",
						"visibility_config.#": "1",
						"visibility_config.0.cloudwatch_metrics_enabled": "false",
						"visibility_config.0.metric_name":                "updated",
						"visibility_config.0.sampled_requests_enabled":   "false",
						"statement.#":                                       "1",
						"statement.0.geo_match_statement.#":                 "1",
						"statement.0.geo_match_statement.0.country_codes.#": "2",
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
	var before, after wafv2.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	ruleGroupNewName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckScopeRegional(t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_Basic(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &before),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "description", webACLName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
				),
			},
			{
				Config: testAccWebACLConfig_Basic(ruleGroupNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &after),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupNewName),
					resource.TestCheckResourceAttr(resourceName, "description", ruleGroupNewName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
				),
			},
		},
	})
}

func TestAccWAFV2WebACL_disappears(t *testing.T) {
	var v wafv2.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckScopeRegional(t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_Minimal(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfwafv2.ResourceWebACL(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWAFV2WebACL_ManagedRuleGroup_basic(t *testing.T) {
	var v wafv2.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckScopeRegional(t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_ManagedRuleGroupStatement(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                      "rule-1",
						"action.#":                  "0",
						"override_action.#":         "1",
						"override_action.0.count.#": "0",
						"override_action.0.none.#":  "1",
						"statement.#":               "1",
						"statement.0.managed_rule_group_statement.#":                        "1",
						"statement.0.managed_rule_group_statement.0.name":                   "AWSManagedRulesCommonRuleSet",
						"statement.0.managed_rule_group_statement.0.vendor_name":            "AWS",
						"statement.0.managed_rule_group_statement.0.excluded_rule.#":        "0",
						"statement.0.managed_rule_group_statement.0.scope_down_statement.#": "0",
					}),
				),
			},
			{
				Config: testAccWebACLConfig_ManagedRuleGroupStatement_update(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                      "rule-1",
						"action.#":                  "0",
						"override_action.#":         "1",
						"override_action.0.count.#": "1",
						"override_action.0.none.#":  "0",
						"statement.#":               "1",
						"statement.0.managed_rule_group_statement.#":                                                              "1",
						"statement.0.managed_rule_group_statement.0.name":                                                         "AWSManagedRulesCommonRuleSet",
						"statement.0.managed_rule_group_statement.0.vendor_name":                                                  "AWS",
						"statement.0.managed_rule_group_statement.0.excluded_rule.#":                                              "2",
						"statement.0.managed_rule_group_statement.0.excluded_rule.0.name":                                         "SizeRestrictions_QUERYSTRING",
						"statement.0.managed_rule_group_statement.0.excluded_rule.1.name":                                         "NoUserAgent_HEADER",
						"statement.0.managed_rule_group_statement.0.scope_down_statement.#":                                       "1",
						"statement.0.managed_rule_group_statement.0.scope_down_statement.0.geo_match_statement.#":                 "1",
						"statement.0.managed_rule_group_statement.0.scope_down_statement.0.geo_match_statement.0.country_codes.#": "2",
						"statement.0.managed_rule_group_statement.0.scope_down_statement.0.geo_match_statement.0.country_codes.0": "US",
						"statement.0.managed_rule_group_statement.0.scope_down_statement.0.geo_match_statement.0.country_codes.1": "NL",
					}),
				),
			},
			{
				Config: testAccWebACLConfig_ManagedRuleGroupStatementWithVersionVersion_10(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                      "rule-1",
						"action.#":                  "0",
						"override_action.#":         "1",
						"override_action.0.count.#": "0",
						"override_action.0.none.#":  "1",
						"statement.#":               "1",
						"statement.0.managed_rule_group_statement.#":                        "1",
						"statement.0.managed_rule_group_statement.0.name":                   "AWSManagedRulesCommonRuleSet",
						"statement.0.managed_rule_group_statement.0.vendor_name":            "AWS",
						"statement.0.managed_rule_group_statement.0.version":                "Version_1.0",
						"statement.0.managed_rule_group_statement.0.excluded_rule.#":        "0",
						"statement.0.managed_rule_group_statement.0.scope_down_statement.#": "0",
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

func TestAccWAFV2WebACL_ManagedRuleGroup_specifyVersion(t *testing.T) {
	var v wafv2.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckScopeRegional(t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_ManagedRuleGroupStatementWithVersionVersion_10(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                      "rule-1",
						"action.#":                  "0",
						"override_action.#":         "1",
						"override_action.0.count.#": "0",
						"override_action.0.none.#":  "1",
						"statement.#":               "1",
						"statement.0.managed_rule_group_statement.#":                        "1",
						"statement.0.managed_rule_group_statement.0.name":                   "AWSManagedRulesCommonRuleSet",
						"statement.0.managed_rule_group_statement.0.vendor_name":            "AWS",
						"statement.0.managed_rule_group_statement.0.version":                "Version_1.0",
						"statement.0.managed_rule_group_statement.0.excluded_rule.#":        "0",
						"statement.0.managed_rule_group_statement.0.scope_down_statement.#": "0",
					}),
				),
			},
			{
				Config: testAccWebACLConfig_ManagedRuleGroupStatement(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                      "rule-1",
						"action.#":                  "0",
						"override_action.#":         "1",
						"override_action.0.count.#": "0",
						"override_action.0.none.#":  "1",
						"statement.#":               "1",
						"statement.0.managed_rule_group_statement.#":                        "1",
						"statement.0.managed_rule_group_statement.0.name":                   "AWSManagedRulesCommonRuleSet",
						"statement.0.managed_rule_group_statement.0.vendor_name":            "AWS",
						"statement.0.managed_rule_group_statement.0.excluded_rule.#":        "0",
						"statement.0.managed_rule_group_statement.0.scope_down_statement.#": "0",
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
	var v wafv2.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckScopeRegional(t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_Minimal(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
				),
			},
		},
	})
}

func TestAccWAFV2WebACL_RateBased_basic(t *testing.T) {
	var v wafv2.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckScopeRegional(t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_RateBasedStatement(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                               "rule-1",
						"action.#":                           "1",
						"action.0.allow.#":                   "0",
						"action.0.block.#":                   "0",
						"action.0.count.#":                   "1",
						"statement.#":                        "1",
						"statement.0.rate_based_statement.#": "1",
						"statement.0.rate_based_statement.0.aggregate_key_type":     "IP",
						"statement.0.rate_based_statement.0.forwarded_ip_config.#":  "0",
						"statement.0.rate_based_statement.0.limit":                  "50000",
						"statement.0.rate_based_statement.0.scope_down_statement.#": "0",
					}),
				),
			},
			{
				Config: testAccWebACLConfig_RateBasedStatement_update(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                               "rule-1",
						"action.#":                           "1",
						"action.0.allow.#":                   "0",
						"action.0.block.#":                   "0",
						"action.0.count.#":                   "1",
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
				ImportStateIdFunc: testAccWebACLImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccWAFV2WebACL_GeoMatch_basic(t *testing.T) {
	var v wafv2.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"
	countryCode := fmt.Sprintf("%q", "US")
	countryCodes := fmt.Sprintf("%s, %q", countryCode, "CA")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckScopeRegional(t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_GeoMatchStatement(webACLName, countryCode),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope", "REGIONAL"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                              "rule-1",
						"action.#":                          "1",
						"action.0.allow.#":                  "0",
						"action.0.block.#":                  "1",
						"action.0.count.#":                  "0",
						"priority":                          "1",
						"statement.#":                       "1",
						"statement.0.geo_match_statement.#": "1",
						"statement.0.geo_match_statement.0.country_codes.#":       "1",
						"statement.0.geo_match_statement.0.country_codes.0":       "US",
						"statement.0.geo_match_statement.0.forwarded_ip_config.#": "0",
						"visibility_config.#":                            "1",
						"visibility_config.0.cloudwatch_metrics_enabled": "false",
						"visibility_config.0.metric_name":                "friendly-rule-metric-name",
						"visibility_config.0.sampled_requests_enabled":   "false",
					}),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
				),
			},
			{
				Config: testAccWebACLConfig_GeoMatchStatement(webACLName, countryCodes),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope", "REGIONAL"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                              "rule-1",
						"action.#":                          "1",
						"action.0.allow.#":                  "0",
						"action.0.block.#":                  "1",
						"action.0.count.#":                  "0",
						"priority":                          "1",
						"statement.#":                       "1",
						"statement.0.geo_match_statement.#": "1",
						"statement.0.geo_match_statement.0.country_codes.#":       "2",
						"statement.0.geo_match_statement.0.country_codes.0":       "US",
						"statement.0.geo_match_statement.0.country_codes.1":       "CA",
						"statement.0.geo_match_statement.0.forwarded_ip_config.#": "0",
						"visibility_config.#":                                     "1",
						"visibility_config.0.cloudwatch_metrics_enabled":          "false",
						"visibility_config.0.metric_name":                         "friendly-rule-metric-name",
						"visibility_config.0.sampled_requests_enabled":            "false",
					}),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
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
	var v wafv2.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckScopeRegional(t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_GeoMatchStatement_forwardedIPConfig(webACLName, "MATCH", "X-Forwarded-For"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                            "1",
						"statement.0.or_statement.#":             "1",
						"statement.0.or_statement.0.statement.#": "2",
						"statement.0.or_statement.0.statement.0.geo_match_statement.#":                                         "1",
						"statement.0.or_statement.0.statement.0.geo_match_statement.0.country_codes.#":                         "1",
						"statement.0.or_statement.0.statement.0.geo_match_statement.0.forwarded_ip_config.#":                   "0",
						"statement.0.or_statement.0.statement.1.geo_match_statement.#":                                         "1",
						"statement.0.or_statement.0.statement.1.geo_match_statement.0.country_codes.#":                         "1",
						"statement.0.or_statement.0.statement.1.geo_match_statement.0.forwarded_ip_config.#":                   "1",
						"statement.0.or_statement.0.statement.1.geo_match_statement.0.forwarded_ip_config.0.fallback_behavior": "MATCH",
						"statement.0.or_statement.0.statement.1.geo_match_statement.0.forwarded_ip_config.0.header_name":       "X-Forwarded-For",
					}),
				),
			},
			{
				Config: testAccWebACLConfig_GeoMatchStatement_forwardedIPConfig(webACLName, "NO_MATCH", "Updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                            "1",
						"statement.0.or_statement.#":             "1",
						"statement.0.or_statement.0.statement.#": "2",
						"statement.0.or_statement.0.statement.0.geo_match_statement.#":                                         "1",
						"statement.0.or_statement.0.statement.0.geo_match_statement.0.country_codes.#":                         "1",
						"statement.0.or_statement.0.statement.0.geo_match_statement.0.forwarded_ip_config.#":                   "0",
						"statement.0.or_statement.0.statement.1.geo_match_statement.#":                                         "1",
						"statement.0.or_statement.0.statement.1.geo_match_statement.0.country_codes.#":                         "1",
						"statement.0.or_statement.0.statement.1.geo_match_statement.0.forwarded_ip_config.#":                   "1",
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
	var v wafv2.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckScopeRegional(t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_LabelMatchStatement(webACLName, "LABEL", "Hashicorp:Test:Label1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
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
				Config: testAccWebACLConfig_LabelMatchStatement(webACLName, "NAMESPACE", "awswaf:managed:aws:bot-control:"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
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
				ImportStateIdFunc: testAccWebACLImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccWAFV2WebACL_RuleLabels(t *testing.T) {
	var v wafv2.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckScopeRegional(t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_RuleLabels(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"rule_label.#":      "2",
						"rule_label.0.name": "Hashicorp:Test:Label1",
						"rule_label.1.name": "Hashicorp:Test:Label2",
					}),
				),
			},
			{
				Config: testAccWebACLConfig_NoRuleLabels(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
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
				ImportStateIdFunc: testAccWebACLImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccWAFV2WebACL_IPSetReference_basic(t *testing.T) {
	var v wafv2.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckScopeRegional(t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_IPSetReference(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#": "1",
						"statement.0.ip_set_reference_statement.#":                              "1",
						"statement.0.ip_set_reference_statement.0.ip_set_forwarded_ip_config.#": "0",
						"visibility_config.#":                            "1",
						"visibility_config.0.cloudwatch_metrics_enabled": "false",
						"visibility_config.0.metric_name":                "friendly-rule-metric-name",
						"visibility_config.0.sampled_requests_enabled":   "false",
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]*regexp.Regexp{
						"statement.0.ip_set_reference_statement.0.arn": regexp.MustCompile(`regional/ipset/.+$`),
					}),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
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
	var v wafv2.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckScopeRegional(t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_IPSetReference_forwardedIPConfig(webACLName, "MATCH", "X-Forwarded-For", "FIRST"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
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
				Config: testAccWebACLConfig_IPSetReference_forwardedIPConfig(webACLName, "NO_MATCH", "X-Forwarded-For", "LAST"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
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
				Config: testAccWebACLConfig_IPSetReference_forwardedIPConfig(webACLName, "MATCH", "Updated", "ANY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
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
				Config: testAccWebACLConfig_IPSetReference(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
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
				ImportStateIdFunc: testAccWebACLImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccWAFV2WebACL_RateBased_forwardedIP(t *testing.T) {
	var v wafv2.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckScopeRegional(t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_RateBasedStatement_forwardedIPConfig(webACLName, "MATCH", "X-Forwarded-For"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                               "rule-1",
						"action.#":                           "1",
						"action.0.allow.#":                   "0",
						"action.0.block.#":                   "0",
						"action.0.count.#":                   "1",
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
				Config: testAccWebACLConfig_RateBasedStatement_forwardedIPConfig(webACLName, "NO_MATCH", "Updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                               "rule-1",
						"action.#":                           "1",
						"action.0.allow.#":                   "0",
						"action.0.block.#":                   "0",
						"action.0.count.#":                   "1",
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
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccWebACLImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccWAFV2WebACL_RuleGroupReference_basic(t *testing.T) {
	var v wafv2.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckScopeRegional(t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_RuleGroupReferenceStatement(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                      "rule-1",
						"override_action.#":         "1",
						"override_action.0.count.#": "1",
						"override_action.0.none.#":  "0",
						"statement.#":               "1",
						"statement.0.rule_group_reference_statement.#":                 "1",
						"statement.0.rule_group_reference_statement.0.excluded_rule.#": "0",
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]*regexp.Regexp{
						"statement.0.rule_group_reference_statement.0.arn": regexp.MustCompile(`regional/rulegroup/.+$`),
					}),
				),
			},
			{
				Config: testAccWebACLConfig_RuleGroupReferenceStatement_update(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                      "rule-1",
						"override_action.#":         "1",
						"override_action.0.count.#": "1",
						"override_action.0.none.#":  "0",
						"statement.#":               "1",
						"statement.0.rule_group_reference_statement.#":                      "1",
						"statement.0.rule_group_reference_statement.0.excluded_rule.#":      "2",
						"statement.0.rule_group_reference_statement.0.excluded_rule.0.name": "rule-to-exclude-b",
						"statement.0.rule_group_reference_statement.0.excluded_rule.1.name": "rule-to-exclude-a",
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]*regexp.Regexp{
						"statement.0.rule_group_reference_statement.0.arn": regexp.MustCompile(`regional/rulegroup/.+$`),
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

func TestAccWAFV2WebACL_Custom_requestHandling(t *testing.T) {
	var v wafv2.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckScopeRegional(t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_CustomRequestHandling_allow(webACLName, "x-hdr1", "x-hdr2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope", "REGIONAL"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":             "rule-1",
						"action.#":         "1",
						"action.0.allow.#": "1",
						"action.0.allow.0.custom_request_handling.#":                       "1",
						"action.0.allow.0.custom_request_handling.0.insert_header.#":       "2",
						"action.0.allow.0.custom_request_handling.0.insert_header.0.name":  "x-hdr1",
						"action.0.allow.0.custom_request_handling.0.insert_header.0.value": "test-value-1",
						"action.0.allow.0.custom_request_handling.0.insert_header.1.name":  "x-hdr2",
						"action.0.allow.0.custom_request_handling.0.insert_header.1.value": "test-value-2",
						"action.0.block.#": "0",
						"action.0.count.#": "0",
						"priority":         "1",
					}),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
				),
			},
			{
				Config: testAccWebACLConfig_CustomRequestHandling_count(webACLName, "x-hdr1", "x-hdr2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "scope", "REGIONAL"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":             "rule-1",
						"action.#":         "1",
						"action.0.allow.#": "0",
						"action.0.block.#": "0",
						"action.0.count.#": "1",
						"action.0.count.0.custom_request_handling.#":                       "1",
						"action.0.count.0.custom_request_handling.0.insert_header.#":       "2",
						"action.0.count.0.custom_request_handling.0.insert_header.0.name":  "x-hdr1",
						"action.0.count.0.custom_request_handling.0.insert_header.0.value": "test-value-1",
						"action.0.count.0.custom_request_handling.0.insert_header.1.name":  "x-hdr2",
						"action.0.count.0.custom_request_handling.0.insert_header.1.value": "test-value-2",
						"priority": "1",
					}),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
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

func TestAccWAFV2WebACL_Custom_response(t *testing.T) {
	var v wafv2.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckScopeRegional(t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_CustomResponse(webACLName, 401, 403, "x-hdr1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.0.custom_response.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.0.custom_response.0.response_code", "401"),
					resource.TestCheckResourceAttr(resourceName, "scope", "REGIONAL"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                               "rule-1",
						"action.#":                           "1",
						"action.0.allow.#":                   "0",
						"action.0.block.#":                   "1",
						"action.0.block.0.custom_response.#": "1",
						"action.0.block.0.custom_response.0.response_code":           "403",
						"action.0.block.0.custom_response.0.response_header.#":       "1",
						"action.0.block.0.custom_response.0.response_header.0.name":  "x-hdr1",
						"action.0.block.0.custom_response.0.response_header.0.value": "custom-response-header-value",
						"action.0.count.#": "0",
						"priority":         "1",
					}),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
				),
			},
			{
				Config: testAccWebACLConfig_CustomResponse(webACLName, 404, 429, "x-hdr2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.0.custom_response.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.0.custom_response.0.response_code", "404"),
					resource.TestCheckResourceAttr(resourceName, "scope", "REGIONAL"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                               "rule-1",
						"action.#":                           "1",
						"action.0.allow.#":                   "0",
						"action.0.block.#":                   "1",
						"action.0.block.0.custom_response.#": "1",
						"action.0.block.0.custom_response.0.response_code":           "429",
						"action.0.block.0.custom_response.0.response_header.#":       "1",
						"action.0.block.0.custom_response.0.response_header.0.name":  "x-hdr2",
						"action.0.block.0.custom_response.0.response_header.0.value": "custom-response-header-value",
						"action.0.count.#": "0",
						"priority":         "1",
					}),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
				),
			},
			{
				Config: testAccWebACLConfig_CustomResponseBody(webACLName, 404, 429),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "custom_response_body.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "custom_response_body.0.key", "test_body"),
					resource.TestCheckResourceAttr(resourceName, "custom_response_body.0.content", "<html><body>Oops<body></html>"),
					resource.TestCheckResourceAttr(resourceName, "custom_response_body.0.content_type", "TEXT_HTML"),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.allow.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.0.custom_response.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.block.0.custom_response.0.response_code", "404"),
					resource.TestCheckResourceAttr(resourceName, "scope", "REGIONAL"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"name":                               "rule-1",
						"action.#":                           "1",
						"action.0.allow.#":                   "0",
						"action.0.block.#":                   "1",
						"action.0.block.0.custom_response.#": "1",
						"action.0.block.0.custom_response.0.response_code":            "429",
						"action.0.block.0.custom_response.0.custom_response_body_key": "test_body",
						"action.0.count.#": "0",
						"priority":         "1",
					}),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
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
	var v wafv2.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckScopeRegional(t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_OneTag(webACLName, "Tag1", "Value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag1", "Value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccWebACLImportStateIdFunc(resourceName),
			},
			{
				Config: testAccWebACLConfig_TwoTags(webACLName, "Tag1", "Value1Updated", "Tag2", "Value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag1", "Value1Updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag2", "Value2"),
				),
			},
			{
				Config: testAccWebACLConfig_OneTag(webACLName, "Tag2", "Value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag2", "Value2"),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/13862
func TestAccWAFV2WebACL_RateBased_maxNested(t *testing.T) {
	var v wafv2.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckScopeRegional(t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_multipleNestedRateBasedStatements(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
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
						"statement.0.rate_based_statement.0.scope_down_statement.0.not_statement.0.statement.0.or_statement.0.statement.#": "2",
						"statement.0.rate_based_statement.0.scope_down_statement.0.not_statement.0.statement.0.or_statement.0.statement.0.regex_pattern_set_reference_statement.#": "1",
						"statement.0.rate_based_statement.0.scope_down_statement.0.not_statement.0.statement.0.or_statement.0.statement.1.ip_set_reference_statement.#":            "1",
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
	var v wafv2.WebACL
	webACLName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckScopeRegional(t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafv2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_multipleNestedOperatorStatements(webACLName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/webacl/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "name", webACLName),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						"statement.#":                                                                                    "1",
						"statement.0.and_statement.#":                                                                    "1",
						"statement.0.and_statement.0.statement.#":                                                        "2",
						"statement.0.and_statement.0.statement.0.not_statement.#":                                        "1",
						"statement.0.and_statement.0.statement.0.not_statement.0.statement.#":                            "1",
						"statement.0.and_statement.0.statement.0.not_statement.0.statement.0.or_statement.#":             "1",
						"statement.0.and_statement.0.statement.0.not_statement.0.statement.0.or_statement.0.statement.#": "2",
						"statement.0.and_statement.0.statement.0.not_statement.0.statement.0.or_statement.0.statement.0.regex_pattern_set_reference_statement.#": "1",
						"statement.0.and_statement.0.statement.0.not_statement.0.statement.0.or_statement.0.statement.1.ip_set_reference_statement.#":            "1",
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

func testAccCheckWebACLDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_wafv2_web_acl" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFV2Conn
		resp, err := conn.GetWebACL(
			&wafv2.GetWebACLInput{
				Id:    aws.String(rs.Primary.ID),
				Name:  aws.String(rs.Primary.Attributes["name"]),
				Scope: aws.String(rs.Primary.Attributes["scope"]),
			})

		if err == nil {
			if resp == nil || resp.WebACL == nil {
				return fmt.Errorf("Error getting WAFv2 WebACL")
			}
			if aws.StringValue(resp.WebACL.Id) == rs.Primary.ID {
				return fmt.Errorf("WAFv2 WebACL %s still exists", rs.Primary.ID)
			}
		}

		// Return nil if the WebACL is already destroyed
		if tfawserr.ErrCodeEquals(err, wafv2.ErrCodeWAFNonexistentItemException) {
			return nil
		}

		return err
	}

	return nil
}

func testAccCheckWebACLExists(n string, v *wafv2.WebACL) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAFv2 WebACL ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFV2Conn
		resp, err := conn.GetWebACL(&wafv2.GetWebACLInput{
			Id:    aws.String(rs.Primary.ID),
			Name:  aws.String(rs.Primary.Attributes["name"]),
			Scope: aws.String(rs.Primary.Attributes["scope"]),
		})

		if err != nil {
			return err
		}

		if resp == nil || resp.WebACL == nil {
			return fmt.Errorf("Error getting WAFv2 WebACL")
		}

		if aws.StringValue(resp.WebACL.Id) == rs.Primary.ID {
			*v = *resp.WebACL
			return nil
		}

		return fmt.Errorf("WAFv2 WebACL (%s) not found", rs.Primary.ID)
	}
}

func testAccWebACLConfig_Basic(name string) string {
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
`, name)
}

func testAccWebACLConfig_BasicRule(name string) string {
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
`, name)
}

func testAccWebACLConfig_UpdateRuleNamePriorityMetric(name, ruleName1, ruleName2 string, priority1, priority2 int) string {
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
`, name, ruleName1, priority1, ruleName2, priority2)
}

func testAccWebACLConfig_GeoMatchStatement(name, countryCodes string) string {
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
`, name, countryCodes)
}

func testAccWebACLConfig_LabelMatchStatement(name, labelScope string, labelKey string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = "%[1]s"
  description = "%[1]s"
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
`, name, labelScope, labelKey)
}

func testAccWebACLConfig_RuleLabels(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = "%[1]s"
  description = "%[1]s"
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
`, name)
}

func testAccWebACLConfig_NoRuleLabels(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = "%[1]s"
  description = "%[1]s"
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
`, name)
}

func testAccWebACLConfig_CustomRequestHandling_count(name, firstHeader string, secondHeader string) string {
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
`, name, firstHeader, secondHeader)
}

func testAccWebACLConfig_CustomRequestHandling_allow(name, firstHeader string, secondHeader string) string {
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
`, name, firstHeader, secondHeader)
}

func testAccWebACLConfig_CustomResponse(name string, defaultStatusCode int, countryBlockStatusCode int, countryHeaderName string) string {
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
`, name, defaultStatusCode, countryBlockStatusCode, countryHeaderName)
}

func testAccWebACLConfig_CustomResponseBody(name string, defaultStatusCode int, countryBlockStatusCode int) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name        = "%[1]s"
  description = "%[1]s"
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
`, name, defaultStatusCode, countryBlockStatusCode)
}

func testAccWebACLConfig_GeoMatchStatement_forwardedIPConfig(name, fallbackBehavior, headerName string) string {
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
`, name, fallbackBehavior, headerName)
}

func testAccWebACLConfig_IPSetReference(name string) string {
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
`, name)
}

func testAccWebACLConfig_IPSetReference_forwardedIPConfig(name, fallbackBehavior, headerName, position string) string {
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
`, name, fallbackBehavior, headerName, position)
}

func testAccWebACLConfig_ManagedRuleGroupStatement(name string) string {
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
`, name)
}

func testAccWebACLConfig_ManagedRuleGroupStatement_update(name string) string {
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

        excluded_rule {
          name = "SizeRestrictions_QUERYSTRING"
        }

        excluded_rule {
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
`, name)
}

func testAccWebACLConfig_ManagedRuleGroupStatementWithVersionVersion_10(name string) string {
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
`, name)
}

func testAccWebACLConfig_RateBasedStatement(name string) string {
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
`, name)
}

func testAccWebACLConfig_RateBasedStatement_forwardedIPConfig(name, fallbackBehavior, headerName string) string {
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
`, name, fallbackBehavior, headerName)
}

func testAccWebACLConfig_RateBasedStatement_update(name string) string {
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
`, name)
}

func testAccWebACLConfig_RuleGroupReferenceStatement(name string) string {
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
`, name)
}

func testAccWebACLConfig_RuleGroupReferenceStatement_update(name string) string {
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

        excluded_rule {
          name = "rule-to-exclude-b"
        }

        excluded_rule {
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
`, name)
}

func testAccWebACLConfig_Minimal(name string) string {
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
`, name)
}

func testAccWebACLConfig_OneTag(name, tagKey, tagValue string) string {
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
`, name, tagKey, tagValue)
}

func testAccWebACLConfig_TwoTags(name, tag1Key, tag1Value, tag2Key, tag2Value string) string {
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
`, name, tag1Key, tag1Value, tag2Key, tag2Value)
}

func testAccWebACLConfig_multipleNestedRateBasedStatements(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_regex_pattern_set" "test" {
  name  = %[1]q
  scope = "REGIONAL"

  regular_expression {
    regex_string = "one"
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

func testAccWebACLConfig_multipleNestedOperatorStatements(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_regex_pattern_set" "test" {
  name  = %[1]q
  scope = "REGIONAL"

  regular_expression {
    regex_string = "one"
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

func testAccWebACLImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s/%s", rs.Primary.ID, rs.Primary.Attributes["name"], rs.Primary.Attributes["scope"]), nil
	}
}
