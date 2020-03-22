package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAwsWafv2RuleGroup_basic(t *testing.T) {
	var v wafv2.RuleGroup
	ruleGroupName := fmt.Sprintf("rule-group-%s", acctest.RandString(5))
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2RuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2RuleGroupConfigBasic(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists("aws_wafv2_rule_group.test", &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag2", "Value2"),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfigBasicUpdate(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists("aws_wafv2_rule_group.test", &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated"),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "rule.1778736223.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.1778736223.name", "rule-1"),
					resource.TestCheckResourceAttr(resourceName, "rule.1778736223.priority", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.1778736223.action.0.allow.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.1778736223.action.0.block.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.1778736223.action.0.count.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAwsWafv2RuleGroupImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccAwsWafv2RuleGroup_minimal(t *testing.T) {
	var v wafv2.RuleGroup
	ruleGroupName := fmt.Sprintf("rule-group-%s", acctest.RandString(5))
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2RuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2RuleGroupConfigMinimal(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists("aws_wafv2_rule_group.test", &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
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

func TestAccAwsWafv2RuleGroup_RuleAction(t *testing.T) {
	var v wafv2.RuleGroup
	ruleGroupName := fmt.Sprintf("rule-group-%s", acctest.RandString(5))
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2RuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2RuleGroupConfigRuleActionAllow(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists("aws_wafv2_rule_group.test", &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.494879654.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.494879654.action.0.allow.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.494879654.action.0.block.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.494879654.action.0.count.#", "0"),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfigRuleActionBlock(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists("aws_wafv2_rule_group.test", &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.1223698756.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.1223698756.action.0.allow.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.1223698756.action.0.block.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.1223698756.action.0.count.#", "0"),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfigRuleActionCount(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists("aws_wafv2_rule_group.test", &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.1778736223.action.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.1778736223.action.0.allow.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.1778736223.action.0.block.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.1778736223.action.0.count.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAwsWafv2RuleGroupImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccAwsWafv2RuleGroup_ByteMatchStatement(t *testing.T) {
	var v wafv2.RuleGroup
	ruleGroupName := fmt.Sprintf("rule-group-%s", acctest.RandString(5))
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2RuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2RuleGroupConfigByteMatchStatement(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists("aws_wafv2_rule_group.test", &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.2327219868.statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.2327219868.statement.0.byte_match_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.2327219868.statement.0.byte_match_statement.0.positional_constraint", "CONTAINS"),
					resource.TestCheckResourceAttr(resourceName, "rule.2327219868.statement.0.byte_match_statement.0.search_string", "word"),
					resource.TestCheckResourceAttr(resourceName, "rule.2327219868.statement.0.byte_match_statement.0.text_transformation.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "rule.2327219868.statement.0.byte_match_statement.0.text_transformation.4292479961.priority", "5"),
					resource.TestCheckResourceAttr(resourceName, "rule.2327219868.statement.0.byte_match_statement.0.text_transformation.4292479961.type", "NONE"),
					resource.TestCheckResourceAttr(resourceName, "rule.2327219868.statement.0.byte_match_statement.0.text_transformation.2156930824.priority", "2"),
					resource.TestCheckResourceAttr(resourceName, "rule.2327219868.statement.0.byte_match_statement.0.text_transformation.2156930824.type", "LOWERCASE"),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfigByteMatchStatementUpdate(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists("aws_wafv2_rule_group.test", &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.1691194695.statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.1691194695.statement.0.byte_match_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.1691194695.statement.0.byte_match_statement.0.positional_constraint", "EXACTLY"),
					resource.TestCheckResourceAttr(resourceName, "rule.1691194695.statement.0.byte_match_statement.0.search_string", "sentence"),
					resource.TestCheckResourceAttr(resourceName, "rule.1691194695.statement.0.byte_match_statement.0.text_transformation.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.1691194695.statement.0.byte_match_statement.0.text_transformation.766585421.priority", "3"),
					resource.TestCheckResourceAttr(resourceName, "rule.1691194695.statement.0.byte_match_statement.0.text_transformation.766585421.type", "CMD_LINE"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAwsWafv2RuleGroupImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccAwsWafv2RuleGroup_ByteMatchStatement_FieldToMatch(t *testing.T) {
	var v wafv2.RuleGroup
	ruleGroupName := fmt.Sprintf("rule-group-%s", acctest.RandString(5))
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2RuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2RuleGroupConfigByteMatchStatementFieldToMatchAllQueryArguments(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists("aws_wafv2_rule_group.test", &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.720804986.statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.720804986.statement.0.byte_match_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.720804986.statement.0.byte_match_statement.0.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.720804986.statement.0.byte_match_statement.0.field_to_match.0.all_query_arguments.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.720804986.statement.0.byte_match_statement.0.field_to_match.0.body.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.720804986.statement.0.byte_match_statement.0.field_to_match.0.method.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.720804986.statement.0.byte_match_statement.0.field_to_match.0.query_string.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.720804986.statement.0.byte_match_statement.0.field_to_match.0.single_header.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.720804986.statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.720804986.statement.0.byte_match_statement.0.field_to_match.0.uri_path.#", "0"),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfigByteMatchStatementFieldToMatchBody(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists("aws_wafv2_rule_group.test", &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.1230325395.statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.1230325395.statement.0.byte_match_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.1230325395.statement.0.byte_match_statement.0.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.1230325395.statement.0.byte_match_statement.0.field_to_match.0.all_query_arguments.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.1230325395.statement.0.byte_match_statement.0.field_to_match.0.body.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.1230325395.statement.0.byte_match_statement.0.field_to_match.0.method.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.1230325395.statement.0.byte_match_statement.0.field_to_match.0.query_string.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.1230325395.statement.0.byte_match_statement.0.field_to_match.0.single_header.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.1230325395.statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.1230325395.statement.0.byte_match_statement.0.field_to_match.0.uri_path.#", "0"),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfigByteMatchStatementFieldToMatchMethod(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists("aws_wafv2_rule_group.test", &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.2217354537.statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.2217354537.statement.0.byte_match_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.2217354537.statement.0.byte_match_statement.0.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.2217354537.statement.0.byte_match_statement.0.field_to_match.0.all_query_arguments.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.2217354537.statement.0.byte_match_statement.0.field_to_match.0.body.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.2217354537.statement.0.byte_match_statement.0.field_to_match.0.method.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.2217354537.statement.0.byte_match_statement.0.field_to_match.0.query_string.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.2217354537.statement.0.byte_match_statement.0.field_to_match.0.single_header.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.2217354537.statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.2217354537.statement.0.byte_match_statement.0.field_to_match.0.uri_path.#", "0"),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfigByteMatchStatementFieldToMatchQueryString(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists("aws_wafv2_rule_group.test", &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.1873458199.statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.1873458199.statement.0.byte_match_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.1873458199.statement.0.byte_match_statement.0.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.1873458199.statement.0.byte_match_statement.0.field_to_match.0.all_query_arguments.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.1873458199.statement.0.byte_match_statement.0.field_to_match.0.body.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.1873458199.statement.0.byte_match_statement.0.field_to_match.0.method.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.1873458199.statement.0.byte_match_statement.0.field_to_match.0.query_string.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.1873458199.statement.0.byte_match_statement.0.field_to_match.0.single_header.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.1873458199.statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.1873458199.statement.0.byte_match_statement.0.field_to_match.0.uri_path.#", "0"),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfigByteMatchStatementFieldToMatchSingleHeader(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists("aws_wafv2_rule_group.test", &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.405778533.statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.405778533.statement.0.byte_match_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.405778533.statement.0.byte_match_statement.0.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.405778533.statement.0.byte_match_statement.0.field_to_match.0.all_query_arguments.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.405778533.statement.0.byte_match_statement.0.field_to_match.0.body.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.405778533.statement.0.byte_match_statement.0.field_to_match.0.method.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.405778533.statement.0.byte_match_statement.0.field_to_match.0.query_string.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.405778533.statement.0.byte_match_statement.0.field_to_match.0.single_header.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.405778533.statement.0.byte_match_statement.0.field_to_match.0.single_header.0.name", "a-forty-character-long-header-name-40-40"),
					resource.TestCheckResourceAttr(resourceName, "rule.405778533.statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.405778533.statement.0.byte_match_statement.0.field_to_match.0.uri_path.#", "0"),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfigByteMatchStatementFieldToMatchSingleQueryArgument(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists("aws_wafv2_rule_group.test", &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.2136468000.statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.2136468000.statement.0.byte_match_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.2136468000.statement.0.byte_match_statement.0.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.2136468000.statement.0.byte_match_statement.0.field_to_match.0.all_query_arguments.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.2136468000.statement.0.byte_match_statement.0.field_to_match.0.body.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.2136468000.statement.0.byte_match_statement.0.field_to_match.0.method.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.2136468000.statement.0.byte_match_statement.0.field_to_match.0.query_string.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.2136468000.statement.0.byte_match_statement.0.field_to_match.0.single_header.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.2136468000.statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.2136468000.statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.0.name", "a-max-30-characters-long-name-"),
					resource.TestCheckResourceAttr(resourceName, "rule.2136468000.statement.0.byte_match_statement.0.field_to_match.0.uri_path.#", "0"),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfigByteMatchStatementFieldToMatchUriPath(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists("aws_wafv2_rule_group.test", &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.566650652.statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.566650652.statement.0.byte_match_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.566650652.statement.0.byte_match_statement.0.field_to_match.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.566650652.statement.0.byte_match_statement.0.field_to_match.0.all_query_arguments.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.566650652.statement.0.byte_match_statement.0.field_to_match.0.body.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.566650652.statement.0.byte_match_statement.0.field_to_match.0.method.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.566650652.statement.0.byte_match_statement.0.field_to_match.0.query_string.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.566650652.statement.0.byte_match_statement.0.field_to_match.0.single_header.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.566650652.statement.0.byte_match_statement.0.field_to_match.0.single_query_argument.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule.566650652.statement.0.byte_match_statement.0.field_to_match.0.uri_path.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAwsWafv2RuleGroupImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccAwsWafv2RuleGroup_GeoMatchStatement(t *testing.T) {
	var v wafv2.RuleGroup
	ruleGroupName := fmt.Sprintf("rule-group-%s", acctest.RandString(5))
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2RuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2RuleGroupConfigGeoMatchStatement(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists("aws_wafv2_rule_group.test", &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.494879654.statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.494879654.statement.0.geo_match_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.494879654.statement.0.geo_match_statement.0.country_codes.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "rule.494879654.statement.0.geo_match_statement.0.country_codes.0", "US"),
					resource.TestCheckResourceAttr(resourceName, "rule.494879654.statement.0.geo_match_statement.0.country_codes.1", "NL"),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfigGeoMatchStatementUpdate(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists("aws_wafv2_rule_group.test", &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "rule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.3292761979.statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.3292761979.statement.0.geo_match_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule.3292761979.statement.0.geo_match_statement.0.country_codes.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "rule.3292761979.statement.0.geo_match_statement.0.country_codes.0", "ZM"),
					resource.TestCheckResourceAttr(resourceName, "rule.3292761979.statement.0.geo_match_statement.0.country_codes.1", "EE"),
					resource.TestCheckResourceAttr(resourceName, "rule.3292761979.statement.0.geo_match_statement.0.country_codes.2", "MM"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAwsWafv2RuleGroupImportStateIdFunc(resourceName),
			},
		},
	})
}

func TestAccAwsWafv2RuleGroup_changeNameForceNew(t *testing.T) {
	var before, after wafv2.RuleGroup
	ruleGroupName := fmt.Sprintf("rule-group-%s", acctest.RandString(5))
	ruleGroupNewName := fmt.Sprintf("rule-group-%s", acctest.RandString(5))
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2RuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2RuleGroupConfigBasic(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists("aws_wafv2_rule_group.test", &before),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfigBasic(ruleGroupNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists("aws_wafv2_rule_group.test", &after),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupNewName),
					resource.TestCheckResourceAttr(resourceName, "description", ruleGroupNewName),
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

func TestAccAwsWafv2RuleGroup_changeCapacityForceNew(t *testing.T) {
	var before, after wafv2.RuleGroup
	ruleGroupName := fmt.Sprintf("rule-group-%s", acctest.RandString(5))
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2RuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2RuleGroupConfigBasic(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists("aws_wafv2_rule_group.test", &before),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfigUpdate_capacity(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists("aws_wafv2_rule_group.test", &after),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "3"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", ruleGroupName),
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

func TestAccAwsWafv2RuleGroup_changeMetricNameForceNew(t *testing.T) {
	var before, after wafv2.RuleGroup
	ruleGroupName := fmt.Sprintf("rule-group-%s", acctest.RandString(5))
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2RuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2RuleGroupConfigBasic(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists("aws_wafv2_rule_group.test", &before),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "scope", wafv2.ScopeRegional),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", "false"),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfigUpdate_metricName(ruleGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists("aws_wafv2_rule_group.test", &after),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", ruleGroupName),
					resource.TestCheckResourceAttr(resourceName, "description", ruleGroupName),
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

func TestAccAwsWafv2RuleGroup_tags(t *testing.T) {
	var v wafv2.RuleGroup
	ruleGroupName := fmt.Sprintf("rule-group-%s", acctest.RandString(5))
	resourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsWafv2RuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsWafv2RuleGroupConfigOneTag(ruleGroupName, "Tag1", "Value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists("aws_wafv2_rule_group.test", &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag1", "Value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccAwsWafv2RuleGroupImportStateIdFunc(resourceName),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfigTwoTags(ruleGroupName, "Tag1", "Value1Updated", "Tag2", "Value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists("aws_wafv2_rule_group.test", &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag1", "Value1Updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag2", "Value2"),
				),
			},
			{
				Config: testAccAwsWafv2RuleGroupConfigOneTag(ruleGroupName, "Tag2", "Value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsWafv2RuleGroupExists("aws_wafv2_rule_group.test", &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "wafv2", regexp.MustCompile(`regional/rulegroup/.+$`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Tag2", "Value2"),
				),
			},
		},
	})
}

func testAccCheckAwsWafv2RuleGroupDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_wafv2_rule_group" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).wafv2conn
		resp, err := conn.GetRuleGroup(
			&wafv2.GetRuleGroupInput{
				Id:    aws.String(rs.Primary.ID),
				Name:  aws.String(rs.Primary.Attributes["name"]),
				Scope: aws.String(rs.Primary.Attributes["scope"]),
			})

		if err == nil {
			if *resp.RuleGroup.Id == rs.Primary.ID {
				return fmt.Errorf("WAFV2 RuleGroup %s still exists", rs.Primary.ID)
			}
		}

		// Return nil if the RuleGroup is already destroyed
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == wafv2.ErrCodeWAFNonexistentItemException {
				return nil
			}
		}

		return err
	}

	return nil
}

func testAccCheckAwsWafv2RuleGroupExists(n string, v *wafv2.RuleGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAFV2 RuleGroup ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).wafv2conn
		resp, err := conn.GetRuleGroup(&wafv2.GetRuleGroupInput{
			Id:    aws.String(rs.Primary.ID),
			Name:  aws.String(rs.Primary.Attributes["name"]),
			Scope: aws.String(rs.Primary.Attributes["scope"]),
		})

		if err != nil {
			return err
		}

		if *resp.RuleGroup.Id == rs.Primary.ID {
			*v = *resp.RuleGroup
			return nil
		}

		return fmt.Errorf("WAFV2 RuleGroup (%s) not found", rs.Primary.ID)
	}
}

func testAccAwsWafv2RuleGroupConfigBasic(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name = "%s"
  description = "%s"
  scope = "REGIONAL"

  tags = {
    Tag1 = "Value1"
    Tag2 = "Value2"
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name = "friendly-metric-name"
    sampled_requests_enabled = false
  }
}
`, name, name)
}

func testAccAwsWafv2RuleGroupConfigBasicUpdate(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name = "%s"
  description = "Updated"
  scope = "REGIONAL"

  rule {

    name = "rule-1"
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
      metric_name = "friendly-rule-metric-name"
      sampled_requests_enabled = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name = "friendly-metric-name"
    sampled_requests_enabled = false
  }
}
`, name)
}

func testAccAwsWafv2RuleGroupConfigUpdate_capacity(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 3
  name = "%s"
  description = "%s"
  scope = "REGIONAL"

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name = "friendly-metric-name"
    sampled_requests_enabled = false
  }
}
`, name, name)
}

func testAccAwsWafv2RuleGroupConfigUpdate_metricName(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name = "%s"
  description = "%s"
  scope = "REGIONAL"

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name = "updated-friendly-metric-name"
    sampled_requests_enabled = false
  }
}
`, name, name)
}

func testAccAwsWafv2RuleGroupConfigMinimal(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name = "%s"
  scope = "REGIONAL"

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name = "friendly-metric-name"
    sampled_requests_enabled = false
  }
}
`, name)
}

func testAccAwsWafv2RuleGroupConfigRuleActionAllow(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name = "%s"
  scope = "REGIONAL"

  rule {
    name = "rule-1"
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
      metric_name = "friendly-rule-metric-name"
      sampled_requests_enabled = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name = "friendly-metric-name"
    sampled_requests_enabled = false
  }
}
`, name)
}

func testAccAwsWafv2RuleGroupConfigRuleActionBlock(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name = "%s"
  scope = "REGIONAL"

  rule {
    name = "rule-1"
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
      metric_name = "friendly-rule-metric-name"
      sampled_requests_enabled = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name = "friendly-metric-name"
    sampled_requests_enabled = false
  }
}
`, name)
}

func testAccAwsWafv2RuleGroupConfigRuleActionCount(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name = "%s"
  scope = "REGIONAL"

  rule {
    name = "rule-1"
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
      metric_name = "friendly-rule-metric-name"
      sampled_requests_enabled = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name = "friendly-metric-name"
    sampled_requests_enabled = false
  }
}
`, name)
}

func testAccAwsWafv2RuleGroupConfigByteMatchStatement(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 30
  name = "%s"
  scope = "REGIONAL"

  rule {
    name = "rule-1"
    priority = 1

    action {
  	  allow {}
    }

    statement {
      byte_match_statement {
        positional_constraint = "CONTAINS"
        search_string = "word"

        field_to_match {
          all_query_arguments {}
        }

        text_transformation {
          priority = 5
          type = "NONE"
        }

        text_transformation {
          priority = 2
          type = "LOWERCASE"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name = "friendly-rule-metric-name"
      sampled_requests_enabled = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name = "friendly-metric-name"
    sampled_requests_enabled = false
  }
}
`, name)
}

func testAccAwsWafv2RuleGroupConfigByteMatchStatementUpdate(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 30
  name = "%s"
  scope = "REGIONAL"

  rule {
    name = "rule-1"
    priority = 1

    action {
  	  allow {}
    }

    statement {
      byte_match_statement {
        positional_constraint = "EXACTLY"
        search_string = "sentence"

        field_to_match {
          all_query_arguments {}
        }

        text_transformation {
          priority = 3
          type = "CMD_LINE"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name = "friendly-rule-metric-name"
      sampled_requests_enabled = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name = "friendly-metric-name"
    sampled_requests_enabled = false
  }
}
`, name)
}

func testAccAwsWafv2RuleGroupConfigByteMatchStatementFieldToMatchAllQueryArguments(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 30
  name = "%s"
  scope = "REGIONAL"

  rule {
    name = "rule-1"
    priority = 1

    action {
  	  allow {}
    }

    statement {
      byte_match_statement {
        positional_constraint = "CONTAINS"
        search_string = "word"

        field_to_match {
          all_query_arguments {}
        }

        text_transformation {
          priority = 5
          type = "NONE"
        }
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name = "friendly-rule-metric-name"
      sampled_requests_enabled = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name = "friendly-metric-name"
    sampled_requests_enabled = false
  }
}
`, name)
}

func testAccAwsWafv2RuleGroupConfigByteMatchStatementFieldToMatchBody(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
 capacity = 15
 name = "%s"
 scope = "REGIONAL"

 rule {
   name = "rule-1"
   priority = 1

   action {
 	  allow {}
   }

   statement {
     byte_match_statement {
       positional_constraint = "CONTAINS"
       search_string = "word"

       field_to_match {
         body {}
       }

       text_transformation {
         priority = 1
         type = "NONE"
       }
     }
   }

   visibility_config {
     cloudwatch_metrics_enabled = false
     metric_name = "friendly-rule-metric-name"
     sampled_requests_enabled = false
   }
 }

 visibility_config {
   cloudwatch_metrics_enabled = false
   metric_name = "friendly-metric-name"
   sampled_requests_enabled = false
 }
}
`, name)
}

func testAccAwsWafv2RuleGroupConfigByteMatchStatementFieldToMatchMethod(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
 capacity = 15
 name = "%s"
 scope = "REGIONAL"

 rule {
   name = "rule-1"
   priority = 1

   action {
 	  allow {}
   }

   statement {
     byte_match_statement {
       positional_constraint = "CONTAINS"
       search_string = "word"

       field_to_match {
         method {}
       }

       text_transformation {
         priority = 1
         type = "NONE"
       }
     }
   }

   visibility_config {
     cloudwatch_metrics_enabled = false
     metric_name = "friendly-rule-metric-name"
     sampled_requests_enabled = false
   }
 }

 visibility_config {
   cloudwatch_metrics_enabled = false
   metric_name = "friendly-metric-name"
   sampled_requests_enabled = false
 }
}
`, name)
}

func testAccAwsWafv2RuleGroupConfigByteMatchStatementFieldToMatchQueryString(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
 capacity = 15
 name = "%s"
 scope = "REGIONAL"

 rule {
   name = "rule-1"
   priority = 1

   action {
 	  allow {}
   }

   statement {
     byte_match_statement {
       positional_constraint = "CONTAINS"
       search_string = "word"

       field_to_match {
         query_string {}
       }

       text_transformation {
         priority = 1
         type = "NONE"
       }
     }
   }

   visibility_config {
     cloudwatch_metrics_enabled = false
     metric_name = "friendly-rule-metric-name"
     sampled_requests_enabled = false
   }
 }

 visibility_config {
   cloudwatch_metrics_enabled = false
   metric_name = "friendly-metric-name"
   sampled_requests_enabled = false
 }
}
`, name)
}

func testAccAwsWafv2RuleGroupConfigByteMatchStatementFieldToMatchSingleHeader(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
 capacity = 15
 name = "%s"
 scope = "REGIONAL"

 rule {
   name = "rule-1"
   priority = 1

   action {
 	  allow {}
   }

   statement {
     byte_match_statement {
       positional_constraint = "CONTAINS"
       search_string = "word"

       field_to_match {
         single_header {
           name = "a-forty-character-long-header-name-40-40"
         }
       }

       text_transformation {
         priority = 1
         type = "NONE"
       }
     }
   }

   visibility_config {
     cloudwatch_metrics_enabled = false
     metric_name = "friendly-rule-metric-name"
     sampled_requests_enabled = false
   }
 }

 visibility_config {
   cloudwatch_metrics_enabled = false
   metric_name = "friendly-metric-name"
   sampled_requests_enabled = false
 }
}
`, name)
}

func testAccAwsWafv2RuleGroupConfigByteMatchStatementFieldToMatchSingleQueryArgument(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
 capacity = 30
 name = "%s"
 scope = "REGIONAL"

 rule {
   name = "rule-1"
   priority = 1

   action {
 	  allow {}
   }

   statement {
     byte_match_statement {
       positional_constraint = "CONTAINS"
       search_string = "word"

       field_to_match {
         single_query_argument {
           name = "a-max-30-characters-long-name-"
         }
       }

       text_transformation {
         priority = 1
         type = "NONE"
       }
     }
   }

   visibility_config {
     cloudwatch_metrics_enabled = false
     metric_name = "friendly-rule-metric-name"
     sampled_requests_enabled = false
   }
 }

 visibility_config {
   cloudwatch_metrics_enabled = false
   metric_name = "friendly-metric-name"
   sampled_requests_enabled = false
 }
}
`, name)
}

func testAccAwsWafv2RuleGroupConfigByteMatchStatementFieldToMatchUriPath(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
 capacity = 15
 name = "%s"
 scope = "REGIONAL"

 rule {
   name = "rule-1"
   priority = 1

   action {
 	  allow {}
   }

   statement {
     byte_match_statement {
       positional_constraint = "CONTAINS"
       search_string = "word"

       field_to_match {
         uri_path {}
       }

       text_transformation {
         priority = 1
         type = "NONE"
       }
     }
   }

   visibility_config {
     cloudwatch_metrics_enabled = false
     metric_name = "friendly-rule-metric-name"
     sampled_requests_enabled = false
   }
 }

 visibility_config {
   cloudwatch_metrics_enabled = false
   metric_name = "friendly-metric-name"
   sampled_requests_enabled = false
 }
}
`, name)
}

func testAccAwsWafv2RuleGroupConfigGeoMatchStatement(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name = "%s"
  scope = "REGIONAL"

  rule {
    name = "rule-1"
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
      metric_name = "friendly-rule-metric-name"
      sampled_requests_enabled = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name = "friendly-metric-name"
    sampled_requests_enabled = false
  }
}
`, name)
}

func testAccAwsWafv2RuleGroupConfigGeoMatchStatementUpdate(name string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name = "%s"
  scope = "REGIONAL"

  rule {
    name = "rule-1"
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
      metric_name = "friendly-rule-metric-name"
      sampled_requests_enabled = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name = "friendly-metric-name"
    sampled_requests_enabled = false
  }
}
`, name)
}

func testAccAwsWafv2RuleGroupConfigOneTag(name, tagKey, tagValue string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name = "%s"
  description = "%s"
  scope = "REGIONAL"

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name = "friendly-metric-name"
    sampled_requests_enabled = false
  }

  tags = {
    %q = %q
  }
}
`, name, name, tagKey, tagValue)
}

func testAccAwsWafv2RuleGroupConfigTwoTags(name, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  capacity = 2
  name = "%s"
  description = "%s"
  scope = "REGIONAL"

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name = "friendly-metric-name"
    sampled_requests_enabled = false
  }

  tags = {
    %q = %q
    %q = %q
  }
}
`, name, name, tag1Key, tag1Value, tag2Key, tag2Value)
}

func testAccAwsWafv2RuleGroupImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s/%s", rs.Primary.ID, rs.Primary.Attributes["name"], rs.Primary.Attributes["scope"]), nil
	}
}
