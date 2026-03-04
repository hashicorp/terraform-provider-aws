// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package wafv2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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
				),
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
					resource.TestCheckResourceAttr(resourceName, "statement.0.rate_based_statement.0.limit", "1000"),
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

func TestAccWAFV2WebACLRule_regexPatternSetReference(t *testing.T) {
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
				Config: testAccWebACLRuleConfig_regexPatternSetReference(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleExists(ctx, t, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "statement.0.regex_pattern_set_reference_statement.0.arn", "aws_wafv2_regex_pattern_set.test", names.AttrARN),
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
      limit              = 1000
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

func testAccWebACLRuleConfig_captchaCustomRequestHandling(rName string) string {
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
    captcha {
      custom_request_handling {
        insert_header {
          name  = "x-custom-header"
          value = "custom-value"
        }
      }
    }
  }

  statement {
    geo_match_statement {
      country_codes = ["US"]
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

func testAccWebACLRuleConfig_regexPatternSetReference(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_regex_pattern_set" "test" {
  name  = %[1]q
  scope = "REGIONAL"

  regular_expression {
    regex_string = "test.*pattern"
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
