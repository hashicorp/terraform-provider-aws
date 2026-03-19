// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package wafv2_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfwafv2 "github.com/hashicorp/terraform-provider-aws/internal/service/wafv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWAFV2WebACLRule_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrPriority, "1"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, flex.ResourceIdSeparator, "web_acl_arn", names.AttrName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "web_acl_arn",
			},
		},
	})
}

func TestAccWAFV2WebACLRule_ipSetReference(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleConfig_ipSetReference(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "statement.0.ip_set_reference_statement.0.arn", "aws_wafv2_ip_set.test", names.AttrARN),
				),
			},
		},
	})
}

func TestAccWAFV2WebACLRule_deletionOrdering(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleConfig_ipSetReference(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleExists(ctx, t, resourceName),
				),
			},
			{
				// Delete everything - verifies correct deletion order
				// Rule should be deleted before IPSet to avoid WAFAssociatedItemException
				Config: testAccWebACLRuleConfig_webACLOnly(rName),
			},
		},
	})
}

func TestAccWAFV2WebACLRule_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfwafv2.ResourceWebACLRule, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWAFV2WebACLRule_newStatementTypes(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleConfig_rateBasedStatement(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "statement.0.rate_based_statement.0.limit", "2000"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.rate_based_statement.0.aggregate_key_type", "IP"),
				),
			},
			{
				Config: testAccWebACLRuleConfig_sqliMatchStatement(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "statement.0.sqli_match_statement.#", "1"),
				),
			},
			{
				Config: testAccWebACLRuleConfig_xssMatchStatement(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "statement.0.xss_match_statement.#", "1"),
				),
			},
		},
	})
}

func TestAccWAFV2WebACLRule_managedRuleGroup(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleConfig_managedRuleGroup(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "statement.0.managed_rule_group_statement.0.name", "AWSManagedRulesCommonRuleSet"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.managed_rule_group_statement.0.vendor_name", "AWS"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.managed_rule_group_statement.0.rule_action_override.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.managed_rule_group_statement.0.rule_action_override.0.name", "SizeRestrictions_BODY"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.managed_rule_group_statement.0.rule_action_override.0.action_to_use.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.managed_rule_group_statement.0.rule_action_override.0.action_to_use.0.count.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.managed_rule_group_statement.0.rule_action_override.0.action_to_use.0.count.0.custom_request_handling.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.managed_rule_group_statement.0.rule_action_override.0.action_to_use.0.count.0.custom_request_handling.0.insert_header.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.managed_rule_group_statement.0.rule_action_override.0.action_to_use.0.count.0.custom_request_handling.0.insert_header.0.name", "X-Test-Header1"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.managed_rule_group_statement.0.rule_action_override.0.action_to_use.0.count.0.custom_request_handling.0.insert_header.0.value", "TestValue1"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.managed_rule_group_statement.0.rule_action_override.1.name", "NoUserAgent_HEADER"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.managed_rule_group_statement.0.rule_action_override.1.action_to_use.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.managed_rule_group_statement.0.rule_action_override.1.action_to_use.0.block.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.managed_rule_group_statement.0.rule_action_override.1.action_to_use.0.block.0.custom_response.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.managed_rule_group_statement.0.rule_action_override.1.action_to_use.0.block.0.custom_response.0.response_code", "403"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.managed_rule_group_statement.0.rule_action_override.1.action_to_use.0.block.0.custom_response.0.custom_response_body_key", "CustomResponseBody"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.managed_rule_group_statement.0.rule_action_override.1.action_to_use.0.block.0.custom_response.0.response_header.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.managed_rule_group_statement.0.rule_action_override.1.action_to_use.0.block.0.custom_response.0.response_header.0.name", "X-Test-Header2"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.managed_rule_group_statement.0.rule_action_override.1.action_to_use.0.block.0.custom_response.0.response_header.0.value", "TestValue2"),
				),
			},
		},
	})
}

func TestAccWAFV2WebACLRule_managedRuleGroupBotControl(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleConfig_managedRuleGroupBotControl(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "statement.0.managed_rule_group_statement.0.name", "AWSManagedRulesBotControlRuleSet"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.managed_rule_group_statement.0.vendor_name", "AWS"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.managed_rule_group_statement.0.managed_rule_group_configs.0.aws_managed_rules_bot_control_rule_set.0.inspection_level", "COMMON"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, flex.ResourceIdSeparator, "web_acl_arn", names.AttrName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "web_acl_arn",
			},
		},
	})
}

func TestAccWAFV2WebACLRule_asnMatchStatement(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleConfig_asnMatchStatement(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "statement.0.asn_match_statement.0.asn_list.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.asn_match_statement.0.asn_list.0", "12345"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.asn_match_statement.0.asn_list.1", "67890"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.asn_match_statement.0.forwarded_ip_config.0.fallback_behavior", "NO_MATCH"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.asn_match_statement.0.forwarded_ip_config.0.header_name", "X-Forwarded-For"),
				),
			},
		},
	})
}

func TestAccWAFV2WebACLRule_regexMatchStatement(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleConfig_regexMatchStatement(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "statement.0.regex_match_statement.0.regex_string", "^[a-zA-Z0-9]+$"),
				),
			},
		},
	})
}

func TestAccWAFV2WebACLRule_regexPatternSetReferenceStatement(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleConfig_regexPatternSetReferenceStatement(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "statement.0.regex_pattern_set_reference_statement.0.arn", "aws_wafv2_regex_pattern_set.test", names.AttrARN),
				),
			},
		},
	})
}

func TestAccWAFV2WebACLRule_ruleGroupReferenceStatement(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleConfig_ruleGroupReferenceStatement(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "statement.0.rule_group_reference_statement.0.arn", "aws_wafv2_rule_group.test", names.AttrARN),
				),
			},
		},
	})
}

func TestAccWAFV2WebACLRule_sizeConstraintStatement(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleConfig_sizeConstraintStatement(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrPriority, "1"),
				),
			},
		},
	})
}

func TestAccWAFV2WebACLRule_rateBasedStatement(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleConfig_rateBasedStatement(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.rate_based_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.rate_based_statement.0.limit", "2000"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.rate_based_statement.0.aggregate_key_type", "IP"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, flex.ResourceIdSeparator, "web_acl_arn", names.AttrName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "web_acl_arn",
			},
		},
	})
}

func TestAccWAFV2WebACLRule_rateBasedStatementCustomKeys(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleConfig_rateBasedStatementCustomKeys(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "statement.0.rate_based_statement.0.aggregate_key_type", "CUSTOM_KEYS"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.rate_based_statement.0.custom_keys.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.rate_based_statement.0.custom_keys.0.header.0.name", "x-api-key"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.rate_based_statement.0.custom_keys.1.uri_path.0.text_transformation.0.type", "LOWERCASE"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.rate_based_statement.0.custom_keys.2.ip.#", "1"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, flex.ResourceIdSeparator, "web_acl_arn", names.AttrName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "web_acl_arn",
			},
		},
	})
}

func TestAccWAFV2WebACLRule_labelMatchStatement(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleConfig_labelMatchStatement(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.label_match_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.label_match_statement.0.key", "test:label"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.label_match_statement.0.scope", "LABEL"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrsImportStateIdFunc(resourceName, flex.ResourceIdSeparator, "web_acl_arn", names.AttrName),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "web_acl_arn",
			},
		},
	})
}

func TestAccWAFV2WebACLRule_byteMatchStatement(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleConfig_byteMatchStatement(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "statement.0.byte_match_statement.0.search_string", "test-string"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.byte_match_statement.0.positional_constraint", "CONTAINS"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.byte_match_statement.0.field_to_match.0.uri_path.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.byte_match_statement.0.text_transformation.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.byte_match_statement.0.text_transformation.0.priority", "0"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.byte_match_statement.0.text_transformation.0.type", "LOWERCASE"),
				),
			},
		},
	})
}

func TestAccWAFV2WebACLRule_sqliMatchStatement(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleConfig_sqliMatchStatement(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "statement.0.sqli_match_statement.0.sensitivity_level", "LOW"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.sqli_match_statement.0.field_to_match.0.uri_path.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.sqli_match_statement.0.text_transformation.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.sqli_match_statement.0.text_transformation.0.priority", "0"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.sqli_match_statement.0.text_transformation.0.type", "LOWERCASE"),
				),
			},
		},
	})
}

func TestAccWAFV2WebACLRule_xssMatchStatement(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleConfig_xssMatchStatement(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "statement.0.xss_match_statement.0.field_to_match.0.uri_path.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.xss_match_statement.0.text_transformation.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.xss_match_statement.0.text_transformation.0.priority", "0"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.xss_match_statement.0.text_transformation.0.type", "LOWERCASE"),
				),
			},
		},
	})
}

func TestAccWAFV2WebACLRule_andStatement(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleConfig_andStatement(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "statement.0.and_statement.0.statement.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.and_statement.0.statement.0.geo_match_statement.0.country_codes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.and_statement.0.statement.1.byte_match_statement.0.search_string", "test"),
				),
			},
		},
	})
}

func TestAccWAFV2WebACLRule_orStatement(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleConfig_orStatement(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "statement.0.or_statement.0.statement.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.or_statement.0.statement.0.geo_match_statement.0.country_codes.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.or_statement.0.statement.1.geo_match_statement.0.country_codes.#", "1"),
				),
			},
		},
	})
}

func TestAccWAFV2WebACLRule_notStatement(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleConfig_notStatement(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "statement.0.not_statement.0.statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.not_statement.0.statement.0.geo_match_statement.0.country_codes.#", "1"),
				),
			},
		},
	})
}

func TestAccWAFV2WebACLRule_fieldToMatch(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleConfig_fieldToMatch(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleExists(ctx, t, resourceName),
					// json_body
					resource.TestCheckResourceAttr(resourceName, "statement.0.and_statement.0.statement.0.byte_match_statement.0.field_to_match.0.json_body.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.and_statement.0.statement.0.byte_match_statement.0.field_to_match.0.json_body.0.match_scope", "VALUE"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.and_statement.0.statement.0.byte_match_statement.0.field_to_match.0.json_body.0.match_pattern.0.all.#", "1"),
					// headers
					resource.TestCheckResourceAttr(resourceName, "statement.0.and_statement.0.statement.1.byte_match_statement.0.field_to_match.0.headers.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.and_statement.0.statement.1.byte_match_statement.0.field_to_match.0.headers.0.match_scope", "KEY"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.and_statement.0.statement.1.byte_match_statement.0.field_to_match.0.headers.0.match_pattern.0.included_headers.#", "2"),
					// cookies
					resource.TestCheckResourceAttr(resourceName, "statement.0.and_statement.0.statement.2.byte_match_statement.0.field_to_match.0.cookies.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.and_statement.0.statement.2.byte_match_statement.0.field_to_match.0.cookies.0.match_scope", "ALL"),
					// single_header
					resource.TestCheckResourceAttr(resourceName, "statement.0.and_statement.0.statement.3.byte_match_statement.0.field_to_match.0.single_header.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "statement.0.and_statement.0.statement.3.byte_match_statement.0.field_to_match.0.single_header.0.name", "user-agent"),
				),
			},
		},
	})
}

func TestAccWAFV2WebACLRule_multipleRules(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName0 := "aws_wafv2_web_acl_rule.test0"
	resourceName1 := "aws_wafv2_web_acl_rule.test1"
	resourceName2 := "aws_wafv2_web_acl_rule.test2"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Step 1: Create 3 rules on the same Web ACL
			{
				Config: testAccWebACLRuleConfig_multiple(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleExists(ctx, t, resourceName0),
					testAccCheckWebACLRuleExists(ctx, t, resourceName1),
					testAccCheckWebACLRuleExists(ctx, t, resourceName2),
					resource.TestCheckResourceAttr(resourceName0, names.AttrPriority, "0"),
					resource.TestCheckResourceAttr(resourceName1, names.AttrPriority, "1"),
					resource.TestCheckResourceAttr(resourceName2, names.AttrPriority, "2"),
				),
			},
			// Step 2: Update priority of one rule — others must be unaffected
			{
				Config: testAccWebACLRuleConfig_multipleUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleExists(ctx, t, resourceName0),
					testAccCheckWebACLRuleExists(ctx, t, resourceName1),
					testAccCheckWebACLRuleExists(ctx, t, resourceName2),
					resource.TestCheckResourceAttr(resourceName0, names.AttrPriority, "10"),
					resource.TestCheckResourceAttr(resourceName1, names.AttrPriority, "1"),
					resource.TestCheckResourceAttr(resourceName2, names.AttrPriority, "2"),
				),
			},
			// Step 3: Remove one rule — others must survive
			{
				Config: testAccWebACLRuleConfig_multiple(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleExists(ctx, t, resourceName0),
					testAccCheckWebACLRuleExists(ctx, t, resourceName1),
				),
			},
		},
	})
}

func TestAccWAFV2WebACLRule_migrateInlineToSeparateResource(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	webACLResourceName := "aws_wafv2_web_acl.test"
	ruleResourceName := "aws_wafv2_web_acl_rule.test"

	var wa awstypes.WebACL

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				// Step 1: Create WebACL with inline rule
				Config: testAccWebACLRuleMigrationConfig_inline(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, t, webACLResourceName, &wa),
					resource.TestCheckResourceAttr(webACLResourceName, acctest.CtRulePound, "1"),
				),
			},
			{
				// Step 2: Import existing inline rule into separate aws_wafv2_web_acl_rule resource
				Config:          testAccWebACLRuleMigrationConfig_separate(rName),
				ImportState:     true,
				ResourceName:    ruleResourceName,
				ImportStateKind: resource.ImportCommandWithID,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources[webACLResourceName]
					if !ok {
						return "", fmt.Errorf("Not found: %s", webACLResourceName)
					}
					return fmt.Sprintf("%s%stest-rule", rs.Primary.Attributes[names.AttrARN], flex.ResourceIdSeparator), nil
				},
				ImportStateVerify: false,
				ImportPlanChecks: resource.ImportPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue(ruleResourceName, tfjsonpath.New("web_acl_arn"), knownvalue.NotNull()),
						plancheck.ExpectKnownValue(ruleResourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact("test-rule")),
						plancheck.ExpectKnownValue(ruleResourceName, tfjsonpath.New(names.AttrRegion), knownvalue.StringExact(acctest.Region())),
					},
				},
			},
			{
				// Step 3: Verify migration completed successfully
				Config: testAccWebACLRuleMigrationConfig_separate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, t, webACLResourceName, &wa),
					testAccCheckWebACLRuleExists(ctx, t, ruleResourceName),
					resource.TestCheckResourceAttr(ruleResourceName, names.AttrName, "test-rule"),
					resource.TestCheckResourceAttr(ruleResourceName, names.AttrPriority, "1"),
					resource.TestCheckResourceAttrPair(ruleResourceName, "web_acl_arn", webACLResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(ruleResourceName, "action.0.count.#", "1"),
					resource.TestCheckResourceAttr(ruleResourceName, "statement.0.geo_match_statement.0.country_codes.#", "1"),
					resource.TestCheckResourceAttr(ruleResourceName, "statement.0.geo_match_statement.0.country_codes.0", "US"),
					resource.TestCheckResourceAttr(ruleResourceName, "visibility_config.0.cloudwatch_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(ruleResourceName, "visibility_config.0.metric_name", "test-rule"),
				),
			},
		},
	})
}

func testAccCheckWebACLRuleDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).WAFV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_wafv2_web_acl_rule" {
				continue
			}

			webACLARN := rs.Primary.Attributes["web_acl_arn"]
			ruleName := rs.Primary.Attributes[names.AttrName]

			// Parse ARN to get Web ACL details
			webACLID, webACLName, webACLScope, err := parseWebACLARNForTest(webACLARN)
			if err != nil {
				return err
			}

			input := wafv2.GetWebACLInput{
				Id:    aws.String(webACLID),
				Name:  aws.String(webACLName),
				Scope: awstypes.Scope(webACLScope),
			}
			output, err := conn.GetWebACL(ctx, &input)

			if err != nil {
				// Web ACL doesn't exist, rule is gone
				continue
			}

			// Check if rule still exists in Web ACL
			for _, rule := range output.WebACL.Rules {
				if aws.ToString(rule.Name) == ruleName {
					return fmt.Errorf("WAFv2 Web ACL Rule %s still exists", ruleName)
				}
			}
		}

		return nil
	}
}

func testAccCheckWebACLRuleExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).WAFV2Client(ctx)

		webACLARN := rs.Primary.Attributes["web_acl_arn"]
		ruleName := rs.Primary.Attributes[names.AttrName]

		webACLID, webACLName, webACLScope, err := parseWebACLARNForTest(webACLARN)
		if err != nil {
			return err
		}

		input := wafv2.GetWebACLInput{
			Id:    aws.String(webACLID),
			Name:  aws.String(webACLName),
			Scope: awstypes.Scope(webACLScope),
		}
		output, err := conn.GetWebACL(ctx, &input)

		if err != nil {
			return err
		}

		for _, rule := range output.WebACL.Rules {
			if aws.ToString(rule.Name) == ruleName {
				return nil
			}
		}

		return fmt.Errorf("WAFv2 Web ACL Rule %s not found in Web ACL", ruleName)
	}
}

func parseWebACLARNForTest(arnStr string) (id, name, scope string, err error) {
	// ARN format: arn:aws:wafv2:region:account:scope/webacl/name/id
	// Example: arn:aws:wafv2:us-east-1:123456789012:regional/webacl/my-acl/abc123
	parts := splitARN(arnStr)
	if len(parts) < 4 {
		return "", "", "", fmt.Errorf("invalid Web ACL ARN: %s", arnStr)
	}

	scope = parts[0]
	switch scope {
	case "regional":
		scope = "REGIONAL"
	case "global":
		scope = "CLOUDFRONT"
	}
	name = parts[2]
	id = parts[3]

	return id, name, scope, nil
}

func splitARN(arn string) []string {
	// Split on : to get resource part, then split on /
	var result []string
	colonParts := make([]string, 0)
	start := 0
	for i, c := range arn {
		if c == ':' {
			colonParts = append(colonParts, arn[start:i])
			start = i + 1
		}
	}
	colonParts = append(colonParts, arn[start:])

	if len(colonParts) < 6 {
		return result
	}

	resource := colonParts[5]
	start = 0
	for i, c := range resource {
		if c == '/' {
			result = append(result, resource[start:i])
			start = i + 1
		}
	}
	result = append(result, resource[start:])

	return result
}

func testAccWebACLRuleConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name  = %[1]q
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }

  lifecycle {
    ignore_changes = [rule]
  }
}

resource "aws_wafv2_web_acl_rule" "test" {
  name        = %[1]q
  priority    = 1
  web_acl_arn = aws_wafv2_web_acl.test.arn

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
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccWebACLRuleConfig_regexMatchStatement(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name  = %[1]q
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }

  lifecycle {
    ignore_changes = [rule]
  }
}

resource "aws_wafv2_web_acl_rule" "test" {
  name        = %[1]q
  priority    = 1
  web_acl_arn = aws_wafv2_web_acl.test.arn

  action {
    block {}
  }

  statement {
    regex_match_statement {
      regex_string = "^[a-zA-Z0-9]+$"

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
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccWebACLRuleConfig_regexPatternSetReferenceStatement(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_regex_pattern_set" "test" {
  name  = %[1]q
  scope = "REGIONAL"

  regular_expression {
    regex_string = "^[a-zA-Z0-9]+$"
  }
}

resource "aws_wafv2_web_acl" "test" {
  name  = %[1]q
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }

  lifecycle {
    ignore_changes = [rule]
  }
}

resource "aws_wafv2_web_acl_rule" "test" {
  name        = %[1]q
  priority    = 1
  web_acl_arn = aws_wafv2_web_acl.test.arn

  action {
    block {}
  }

  statement {
    regex_pattern_set_reference_statement {
      arn = aws_wafv2_regex_pattern_set.test.arn

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
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccWebACLRuleConfig_ruleGroupReferenceStatement(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  name     = %[1]q
  scope    = "REGIONAL"
  capacity = 10

  rule {
    name     = "rule1"
    priority = 1

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
      metric_name                = "rule1"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }
}

resource "aws_wafv2_web_acl" "test" {
  name  = %[1]q
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }

  lifecycle {
    ignore_changes = [rule]
  }
}

resource "aws_wafv2_web_acl_rule" "test" {
  name        = %[1]q
  priority    = 1
  web_acl_arn = aws_wafv2_web_acl.test.arn

  override_action {
    none {}
  }

  statement {
    rule_group_reference_statement {
      arn = aws_wafv2_rule_group.test.arn
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccWebACLRuleConfig_ipSetReference(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_ip_set" "test" {
  name               = %[1]q
  scope              = "REGIONAL"
  ip_address_version = "IPV4"
  addresses          = ["1.2.3.4/32"]
}

resource "aws_wafv2_web_acl" "test" {
  name  = %[1]q
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }

  lifecycle {
    ignore_changes = [rule]
  }
}

resource "aws_wafv2_web_acl_rule" "test" {
  name        = %[1]q
  priority    = 1
  web_acl_arn = aws_wafv2_web_acl.test.arn

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
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccWebACLRuleConfig_webACLOnly(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name  = %[1]q
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }

  lifecycle {
    ignore_changes = [rule]
  }
}
`, rName)
}

func testAccWebACLRuleConfig_rateBasedStatement(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name  = %[1]q
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }

  lifecycle {
    ignore_changes = [rule]
  }
}

resource "aws_wafv2_web_acl_rule" "test" {
  name        = %[1]q
  priority    = 1
  web_acl_arn = aws_wafv2_web_acl.test.arn

  action {
    block {}
  }

  statement {
    rate_based_statement {
      limit              = 2000
      aggregate_key_type = "IP"
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccWebACLRuleConfig_rateBasedStatementCustomKeys(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name  = %[1]q
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }

  lifecycle {
    ignore_changes = [rule]
  }
}

resource "aws_wafv2_web_acl_rule" "test" {
  name        = %[1]q
  priority    = 1
  web_acl_arn = aws_wafv2_web_acl.test.arn

  action {
    block {}
  }

  statement {
    rate_based_statement {
      limit              = 2000
      aggregate_key_type = "CUSTOM_KEYS"

      custom_keys {
        header {
          name = "x-api-key"
          text_transformation {
            priority = 0
            type     = "NONE"
          }
        }
      }

      custom_keys {
        uri_path {
          text_transformation {
            priority = 0
            type     = "LOWERCASE"
          }
        }
      }

      custom_keys {
        ip {}
      }
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccWebACLRuleConfig_labelMatchStatement(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name  = %[1]q
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }

  lifecycle {
    ignore_changes = [rule]
  }
}

resource "aws_wafv2_web_acl_rule" "test" {
  name        = %[1]q
  web_acl_arn = aws_wafv2_web_acl.test.arn
  priority    = 1

  action {
    block {}
  }

  statement {
    label_match_statement {
      key   = "test:label"
      scope = "LABEL"
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccWebACLRuleConfig_byteMatchStatement(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name  = %[1]q
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }

  lifecycle {
    ignore_changes = [rule]
  }
}

resource "aws_wafv2_web_acl_rule" "test" {
  name        = %[1]q
  priority    = 1
  web_acl_arn = aws_wafv2_web_acl.test.arn

  action {
    block {}
  }

  statement {
    byte_match_statement {
      search_string         = "test-string"
      positional_constraint = "CONTAINS"

      field_to_match {
        uri_path {}
      }

      text_transformation {
        priority = 0
        type     = "LOWERCASE"
      }
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccWebACLRuleConfig_sqliMatchStatement(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name  = %[1]q
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }

  lifecycle {
    ignore_changes = [rule]
  }
}

resource "aws_wafv2_web_acl_rule" "test" {
  name        = %[1]q
  priority    = 1
  web_acl_arn = aws_wafv2_web_acl.test.arn

  action {
    block {}
  }

  statement {
    sqli_match_statement {
      sensitivity_level = "LOW"

      field_to_match {
        uri_path {}
      }

      text_transformation {
        priority = 0
        type     = "LOWERCASE"
      }
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccWebACLRuleConfig_xssMatchStatement(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name  = %[1]q
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }

  lifecycle {
    ignore_changes = [rule]
  }
}

resource "aws_wafv2_web_acl_rule" "test" {
  name        = %[1]q
  priority    = 1
  web_acl_arn = aws_wafv2_web_acl.test.arn

  action {
    block {}
  }

  statement {
    xss_match_statement {
      field_to_match {
        uri_path {}
      }

      text_transformation {
        priority = 0
        type     = "LOWERCASE"
      }
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccWebACLRuleConfig_managedRuleGroup(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name  = %[1]q
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }

  custom_response_body {
    key = "CustomResponseBody"
    content = "{\"message\": \"Custom response body\"}"
    content_type = "APPLICATION_JSON"
  }

  lifecycle {
    ignore_changes = [rule]
  }
}

resource "aws_wafv2_web_acl_rule" "test" {
  name        = %[1]q
  priority    = 1
  web_acl_arn = aws_wafv2_web_acl.test.arn

  override_action {
    none {}
  }

  statement {
    managed_rule_group_statement {
      name        = "AWSManagedRulesCommonRuleSet"
      vendor_name = "AWS"

      rule_action_override {
        name = "SizeRestrictions_BODY"
        action_to_use {
          count {
            custom_request_handling {
              insert_header {
                name  = "X-Test-Header1"
                value = "TestValue1"
              }
            }
          }
        }
      }
      rule_action_override {
        name = "NoUserAgent_HEADER"
        action_to_use {
          block {
            custom_response {
              response_code            = "403"
              custom_response_body_key = "CustomResponseBody"
              response_header {
                name  = "X-Test-Header2"
                value = "TestValue2"
              }
            }
          }
        }
      }
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccWebACLRuleConfig_managedRuleGroupBotControl(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name  = %[1]q
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }

  lifecycle {
    ignore_changes = [rule]
  }
}

resource "aws_wafv2_web_acl_rule" "test" {
  name        = %[1]q
  priority    = 1
  web_acl_arn = aws_wafv2_web_acl.test.arn

  override_action {
    none {}
  }

  statement {
    managed_rule_group_statement {
      name        = "AWSManagedRulesBotControlRuleSet"
      vendor_name = "AWS"

      managed_rule_group_configs {
        aws_managed_rules_bot_control_rule_set {
          inspection_level = "COMMON"
        }
      }
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccWebACLRuleConfig_asnMatchStatement(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name  = %[1]q
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }

  lifecycle {
    ignore_changes = [rule]
  }
}

resource "aws_wafv2_web_acl_rule" "test" {
  name        = %[1]q
  web_acl_arn = aws_wafv2_web_acl.test.arn
  priority    = 1

  action {
    allow {}
  }

  statement {
    asn_match_statement {
      asn_list = [12345, 67890]

      forwarded_ip_config {
        fallback_behavior = "NO_MATCH"
        header_name       = "X-Forwarded-For"
      }
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccWebACLRuleConfig_sizeConstraintStatement(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name  = %[1]q
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }

  lifecycle {
    ignore_changes = [rule]
  }
}

resource "aws_wafv2_web_acl_rule" "test" {
  name        = %[1]q
  priority    = 1
  web_acl_arn = aws_wafv2_web_acl.test.arn

  action {
    block {}
  }

  statement {
    size_constraint_statement {
      comparison_operator = "GT"
      size                = 1000

      field_to_match {
        uri_path {}
      }

      text_transformation {
        priority = 0
        type     = "NONE"
      }
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccWebACLRuleConfig_andStatement(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name  = %[1]q
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }

  lifecycle {
    ignore_changes = [rule]
  }
}

resource "aws_wafv2_web_acl_rule" "test" {
  name        = %[1]q
  priority    = 1
  web_acl_arn = aws_wafv2_web_acl.test.arn

  action {
    block {}
  }

  statement {
    and_statement {
      statement {
        geo_match_statement {
          country_codes = ["US"]
        }
      }

      statement {
        byte_match_statement {
          search_string         = "test"
          positional_constraint = "CONTAINS"

          field_to_match {
            uri_path {}
          }

          text_transformation {
            priority = 0
            type     = "LOWERCASE"
          }
        }
      }
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccWebACLRuleConfig_orStatement(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name  = %[1]q
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }

  lifecycle {
    ignore_changes = [rule]
  }
}

resource "aws_wafv2_web_acl_rule" "test" {
  name        = %[1]q
  priority    = 1
  web_acl_arn = aws_wafv2_web_acl.test.arn

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
        }
      }
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccWebACLRuleConfig_notStatement(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name  = %[1]q
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }

  lifecycle {
    ignore_changes = [rule]
  }
}

resource "aws_wafv2_web_acl_rule" "test" {
  name        = %[1]q
  priority    = 1
  web_acl_arn = aws_wafv2_web_acl.test.arn

  action {
    block {}
  }

  statement {
    not_statement {
      statement {
        geo_match_statement {
          country_codes = ["US"]
        }
      }
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccWebACLRuleConfig_fieldToMatch(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name  = %[1]q
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }

  lifecycle {
    ignore_changes = [rule]
  }
}

resource "aws_wafv2_web_acl_rule" "test" {
  name        = %[1]q
  priority    = 1
  web_acl_arn = aws_wafv2_web_acl.test.arn

  action {
    block {}
  }

  statement {
    and_statement {
      # json_body with match_pattern all
      statement {
        byte_match_statement {
          search_string         = "bad"
          positional_constraint = "CONTAINS"

          field_to_match {
            json_body {
              match_scope = "VALUE"

              match_pattern {
                all {}
              }
            }
          }

          text_transformation {
            priority = 0
            type     = "NONE"
          }
        }
      }

      # headers with included_headers
      statement {
        byte_match_statement {
          search_string         = "malicious"
          positional_constraint = "CONTAINS"

          field_to_match {
            headers {
              match_scope       = "KEY"
              oversize_handling = "MATCH"

              match_pattern {
                included_headers = ["x-custom", "x-forwarded-for"]
              }
            }
          }

          text_transformation {
            priority = 0
            type     = "LOWERCASE"
          }
        }
      }

      # cookies with all pattern
      statement {
        byte_match_statement {
          search_string         = "session"
          positional_constraint = "STARTS_WITH"

          field_to_match {
            cookies {
              match_scope       = "ALL"
              oversize_handling = "NO_MATCH"

              match_pattern {
                all {}
              }
            }
          }

          text_transformation {
            priority = 0
            type     = "NONE"
          }
        }
      }

      # single_header
      statement {
        byte_match_statement {
          search_string         = "bot"
          positional_constraint = "CONTAINS"

          field_to_match {
            single_header {
              name = "user-agent"
            }
          }

          text_transformation {
            priority = 0
            type     = "LOWERCASE"
          }
        }
      }
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccWebACLRuleConfig_multiple(rName string, count int) string {
	var rules strings.Builder
	for i := range count {
		fmt.Fprintf(&rules, `
resource "aws_wafv2_web_acl_rule" "test%[1]d" {
  name        = "%[2]s-%[1]d"
  priority    = %[1]d
  web_acl_arn = aws_wafv2_web_acl.test.arn

  action {
    block {}
  }

  statement {
    geo_match_statement {
      country_codes = ["US"]
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "%[2]s-%[1]d"
    sampled_requests_enabled   = false
  }
}
  `, i, rName)
	}

	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name  = %[1]q
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }

  lifecycle {
    ignore_changes = [rule]
  }
}
`, rName) + rules.String()
}

func testAccWebACLRuleConfig_multipleUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name  = %[1]q
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }

  lifecycle {
    ignore_changes = [rule]
  }
}

resource "aws_wafv2_web_acl_rule" "test0" {
  name        = "%[1]s-0"
  priority    = 10
  web_acl_arn = aws_wafv2_web_acl.test.arn

  action {
    block {}
  }

  statement {
    geo_match_statement {
      country_codes = ["US"]
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "%[1]s-0"
    sampled_requests_enabled   = false
  }
}

resource "aws_wafv2_web_acl_rule" "test1" {
  name        = "%[1]s-1"
  priority    = 1
  web_acl_arn = aws_wafv2_web_acl.test.arn

  action {
    block {}
  }

  statement {
    geo_match_statement {
      country_codes = ["US"]
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "%[1]s-1"
    sampled_requests_enabled   = false
  }
}

resource "aws_wafv2_web_acl_rule" "test2" {
  name        = "%[1]s-2"
  priority    = 2
  web_acl_arn = aws_wafv2_web_acl.test.arn

  action {
    block {}
  }

  statement {
    geo_match_statement {
      country_codes = ["US"]
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "%[1]s-2"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccWebACLRuleMigrationConfig_inline(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name  = %[1]q
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  rule {
    name     = "test-rule"
    priority = 1

    action {
      count {}
    }

    statement {
      geo_match_statement {
        country_codes = ["US"]
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "test-rule"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccWebACLRuleMigrationConfig_separate(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name  = %[1]q
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }

  lifecycle {
    ignore_changes = [rule]
  }
}

resource "aws_wafv2_web_acl_rule" "test" {
  name        = "test-rule"
  priority    = 1
  web_acl_arn = aws_wafv2_web_acl.test.arn

  action {
    count {}
  }

  statement {
    geo_match_statement {
      country_codes = ["US"]
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "test-rule"
    sampled_requests_enabled   = false
  }
}
`, rName)
}
