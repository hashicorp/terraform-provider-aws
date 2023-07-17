// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waf_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/waf"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfwaf "github.com/hashicorp/terraform-provider-aws/internal/service/waf"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// Serialize to avoid resource limits
func TestAccWAFRateBasedRule_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"resource": {
			"basic":              testAccWAFRateBasedRule_basic,
			"changeNameForceNew": testAccWAFRateBasedRule_changeNameForceNew,
			"disappears":         testAccWAFRateBasedRule_disappears,
			"changePredicates":   testAccWAFRateBasedRule_changePredicates,
			"changeRateLimit":    testAccWAFRateBasedRule_changeRateLimit,
			"noPredicates":       testAccWAFRateBasedRule_noPredicates,
			"Tags":               testAccWAFRateBasedRule_tags,
		},
		"data_source": {
			"basic": testAccWAFRateBasedRuleDataSource_basic,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}

func testAccWAFRateBasedRule_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v waf.RateBasedRule
	wafRuleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_rate_based_rule.wafrule"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, waf.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRateBasedRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRateBasedRuleConfig_basic(wafRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRateBasedRuleExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "waf", regexp.MustCompile(`ratebasedrule/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", wafRuleName),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_name", wafRuleName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccWAFRateBasedRule_changeNameForceNew(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after waf.RateBasedRule
	wafRuleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	wafRuleNewName := fmt.Sprintf("wafrulenew%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_rate_based_rule.wafrule"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, waf.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRateBasedRuleConfig_basic(wafRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRateBasedRuleExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", wafRuleName),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_name", wafRuleName),
				),
			},
			{
				Config: testAccRateBasedRuleConfig_changeName(wafRuleNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRateBasedRuleExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", wafRuleNewName),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_name", wafRuleNewName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccWAFRateBasedRule_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v waf.RateBasedRule
	wafRuleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_rate_based_rule.wafrule"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, waf.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRateBasedRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRateBasedRuleConfig_basic(wafRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRateBasedRuleExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfwaf.ResourceRateBasedRule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccWAFRateBasedRule_changePredicates(t *testing.T) {
	ctx := acctest.Context(t)
	var ipset waf.IPSet
	var byteMatchSet waf.ByteMatchSet

	var before, after waf.RateBasedRule
	ruleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_rate_based_rule.wafrule"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, waf.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRateBasedRuleConfig_basic(ruleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPSetExists(ctx, "aws_waf_ipset.ipset", &ipset),
					testAccCheckRateBasedRuleExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "predicates.*", map[string]string{
						"negated": "false",
						"type":    "IPMatch",
					}),
				),
			},
			{
				Config: testAccRateBasedRuleConfig_changePredicates(ruleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckByteMatchSetExists(ctx, "aws_waf_byte_match_set.set", &byteMatchSet),
					testAccCheckRateBasedRuleExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "predicates.*", map[string]string{
						"negated": "true",
						"type":    "ByteMatch",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/9659
func testAccWAFRateBasedRule_changeRateLimit(t *testing.T) {
	ctx := acctest.Context(t)
	var ipset waf.IPSet
	var before, after waf.RateBasedRule
	ruleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_rate_based_rule.wafrule"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, waf.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRateBasedRuleConfig_changeLimit(ruleName, 4000),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPSetExists(ctx, "aws_waf_ipset.ipset", &ipset),
					testAccCheckRateBasedRuleExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "rate_limit", "4000"),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "predicates.*", map[string]string{
						"negated": "false",
						"type":    "IPMatch",
					}),
				),
			},
			{
				Config: testAccRateBasedRuleConfig_changeLimit(ruleName, 3000),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPSetExists(ctx, "aws_waf_ipset.ipset", &ipset),
					testAccCheckRateBasedRuleExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "rate_limit", "3000"),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "predicates.*", map[string]string{
						"negated": "false",
						"type":    "IPMatch",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccWAFRateBasedRule_noPredicates(t *testing.T) {
	ctx := acctest.Context(t)
	var rule waf.RateBasedRule
	ruleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_rate_based_rule.wafrule"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, waf.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRateBasedRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRateBasedRuleConfig_noPredicates(ruleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRateBasedRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccWAFRateBasedRule_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var rule waf.RateBasedRule
	ruleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_rate_based_rule.wafrule"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, waf.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRateBasedRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRateBasedRuleConfig_tags1(ruleName, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRateBasedRuleExists(ctx, resourceName, &rule),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "waf", regexp.MustCompile(`ratebasedrule/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccRateBasedRuleConfig_tags2(ruleName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRateBasedRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccRateBasedRuleConfig_tags1(ruleName, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRateBasedRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckRateBasedRuleDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_waf_rate_based_rule" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).WAFConn(ctx)

			_, err := tfwaf.FindRateBasedRuleByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("WAF Rate Based Rule %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckRateBasedRuleExists(ctx context.Context, n string, v *waf.RateBasedRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAF Rate Based Rule ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFConn(ctx)

		output, err := tfwaf.FindRateBasedRuleByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccRateBasedRuleConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_ipset" "ipset" {
  name = "%s"

  ip_set_descriptors {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_waf_rate_based_rule" "wafrule" {
  depends_on  = [aws_waf_ipset.ipset]
  name        = "%s"
  metric_name = "%s"
  rate_key    = "IP"
  rate_limit  = 2000

  predicates {
    data_id = aws_waf_ipset.ipset.id
    negated = false
    type    = "IPMatch"
  }
}
`, name, name, name)
}

func testAccRateBasedRuleConfig_changeLimit(name string, rateLimit int) string {
	return fmt.Sprintf(`
resource "aws_waf_ipset" "ipset" {
  name = "%s"

  ip_set_descriptors {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_waf_rate_based_rule" "wafrule" {
  depends_on  = [aws_waf_ipset.ipset]
  name        = "%[1]s"
  metric_name = "%[1]s"
  rate_key    = "IP"
  rate_limit  = %[2]d

  predicates {
    data_id = aws_waf_ipset.ipset.id
    negated = false
    type    = "IPMatch"
  }
}
`, name, rateLimit)
}

func testAccRateBasedRuleConfig_changeName(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_ipset" "ipset" {
  name = "%s"

  ip_set_descriptors {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_waf_rate_based_rule" "wafrule" {
  depends_on  = [aws_waf_ipset.ipset]
  name        = "%s"
  metric_name = "%s"
  rate_key    = "IP"
  rate_limit  = 2000

  predicates {
    data_id = aws_waf_ipset.ipset.id
    negated = false
    type    = "IPMatch"
  }
}
`, name, name, name)
}

func testAccRateBasedRuleConfig_changePredicates(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_ipset" "ipset" {
  name = "%s"

  ip_set_descriptors {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_waf_byte_match_set" "set" {
  name = "%s"

  byte_match_tuples {
    text_transformation   = "NONE"
    target_string         = "badrefer1"
    positional_constraint = "CONTAINS"

    field_to_match {
      type = "HEADER"
      data = "referer"
    }
  }
}

resource "aws_waf_rate_based_rule" "wafrule" {
  name        = "%s"
  metric_name = "%s"
  rate_key    = "IP"
  rate_limit  = 2000

  predicates {
    data_id = aws_waf_byte_match_set.set.id
    negated = true
    type    = "ByteMatch"
  }
}
`, name, name, name, name)
}

func testAccRateBasedRuleConfig_noPredicates(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_rate_based_rule" "wafrule" {
  name        = "%s"
  metric_name = "%s"
  rate_key    = "IP"
  rate_limit  = 2000
}
`, name, name)
}

func testAccRateBasedRuleConfig_tags1(name, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_waf_rate_based_rule" "wafrule" {
  name        = "%s"
  metric_name = "%s"
  rate_key    = "IP"
  rate_limit  = 2000

  tags = {
    %q = %q
  }
}
`, name, name, tag1Key, tag1Value)
}

func testAccRateBasedRuleConfig_tags2(name, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_waf_rate_based_rule" "wafrule" {
  name        = "%s"
  metric_name = "%s"
  rate_key    = "IP"
  rate_limit  = 2000

  tags = {
    %q = %q
    %q = %q
  }
}
`, name, name, tag1Key, tag1Value, tag2Key, tag2Value)
}
