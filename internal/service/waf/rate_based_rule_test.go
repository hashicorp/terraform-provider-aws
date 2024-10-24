// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waf_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/waf/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfwaf "github.com/hashicorp/terraform-provider-aws/internal/service/waf"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Serialize to avoid resource limits
func TestAccWAFRateBasedRule_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"resource": {
			acctest.CtBasic:      testAccWAFRateBasedRule_basic,
			"changeNameForceNew": testAccWAFRateBasedRule_changeNameForceNew,
			acctest.CtDisappears: testAccWAFRateBasedRule_disappears,
			"changePredicates":   testAccWAFRateBasedRule_changePredicates,
			"changeRateLimit":    testAccWAFRateBasedRule_changeRateLimit,
			"noPredicates":       testAccWAFRateBasedRule_noPredicates,
			"Tags":               testAccWAFRateBasedRule_tags,
		},
		"data_source": {
			acctest.CtBasic: testAccWAFRateBasedRuleDataSource_basic,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}

func testAccWAFRateBasedRule_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RateBasedRule
	wafRuleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_rate_based_rule.wafrule"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRateBasedRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRateBasedRuleConfig_basic(wafRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRateBasedRuleExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "waf", regexache.MustCompile(`ratebasedrule/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, wafRuleName),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrMetricName, wafRuleName),
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
	var before, after awstypes.RateBasedRule
	wafRuleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	wafRuleNewName := fmt.Sprintf("wafrulenew%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_rate_based_rule.wafrule"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRateBasedRuleConfig_basic(wafRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRateBasedRuleExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, wafRuleName),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrMetricName, wafRuleName),
				),
			},
			{
				Config: testAccRateBasedRuleConfig_changeName(wafRuleNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRateBasedRuleExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, wafRuleNewName),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrMetricName, wafRuleNewName),
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
	var v awstypes.RateBasedRule
	wafRuleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_rate_based_rule.wafrule"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
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
	var ipset awstypes.IPSet
	var byteMatchSet awstypes.ByteMatchSet

	var before, after awstypes.RateBasedRule
	ruleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_rate_based_rule.wafrule"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRateBasedRuleConfig_basic(ruleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPSetExists(ctx, "aws_waf_ipset.ipset", &ipset),
					testAccCheckRateBasedRuleExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ruleName),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "predicates.*", map[string]string{
						"negated":      acctest.CtFalse,
						names.AttrType: "IPMatch",
					}),
				),
			},
			{
				Config: testAccRateBasedRuleConfig_changePredicates(ruleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckByteMatchSetExists(ctx, "aws_waf_byte_match_set.set", &byteMatchSet),
					testAccCheckRateBasedRuleExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ruleName),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "predicates.*", map[string]string{
						"negated":      acctest.CtTrue,
						names.AttrType: "ByteMatch",
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
	var ipset awstypes.IPSet
	var before, after awstypes.RateBasedRule
	ruleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_rate_based_rule.wafrule"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRateBasedRuleConfig_changeLimit(ruleName, 4000),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPSetExists(ctx, "aws_waf_ipset.ipset", &ipset),
					testAccCheckRateBasedRuleExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ruleName),
					resource.TestCheckResourceAttr(resourceName, "rate_limit", "4000"),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "predicates.*", map[string]string{
						"negated":      acctest.CtFalse,
						names.AttrType: "IPMatch",
					}),
				),
			},
			{
				Config: testAccRateBasedRuleConfig_changeLimit(ruleName, 3000),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPSetExists(ctx, "aws_waf_ipset.ipset", &ipset),
					testAccCheckRateBasedRuleExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ruleName),
					resource.TestCheckResourceAttr(resourceName, "rate_limit", "3000"),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "predicates.*", map[string]string{
						"negated":      acctest.CtFalse,
						names.AttrType: "IPMatch",
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
	var rule awstypes.RateBasedRule
	ruleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_rate_based_rule.wafrule"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRateBasedRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRateBasedRuleConfig_noPredicates(ruleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRateBasedRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ruleName),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", acctest.Ct0),
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
	var rule awstypes.RateBasedRule
	ruleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_rate_based_rule.wafrule"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRateBasedRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRateBasedRuleConfig_tags1(ruleName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRateBasedRuleExists(ctx, resourceName, &rule),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "waf", regexache.MustCompile(`ratebasedrule/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ruleName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config: testAccRateBasedRuleConfig_tags2(ruleName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRateBasedRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ruleName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccRateBasedRuleConfig_tags1(ruleName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRateBasedRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ruleName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
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

			conn := acctest.Provider.Meta().(*conns.AWSClient).WAFClient(ctx)

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

func testAccCheckRateBasedRuleExists(ctx context.Context, n string, v *awstypes.RateBasedRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFClient(ctx)

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
  name = %[1]q

  ip_set_descriptors {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_waf_rate_based_rule" "wafrule" {
  depends_on  = [aws_waf_ipset.ipset]
  name        = %[1]q
  metric_name = %[1]q
  rate_key    = "IP"
  rate_limit  = 2000

  predicates {
    data_id = aws_waf_ipset.ipset.id
    negated = false
    type    = "IPMatch"
  }
}
`, name)
}

func testAccRateBasedRuleConfig_changeLimit(name string, rateLimit int) string {
	return fmt.Sprintf(`
resource "aws_waf_ipset" "ipset" {
  name = %[1]q

  ip_set_descriptors {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_waf_rate_based_rule" "wafrule" {
  depends_on  = [aws_waf_ipset.ipset]
  name        = %[1]q
  metric_name = %[1]q
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
  name = %[1]q

  ip_set_descriptors {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_waf_rate_based_rule" "wafrule" {
  depends_on  = [aws_waf_ipset.ipset]
  name        = %[1]q
  metric_name = %[1]q
  rate_key    = "IP"
  rate_limit  = 2000

  predicates {
    data_id = aws_waf_ipset.ipset.id
    negated = false
    type    = "IPMatch"
  }
}
`, name)
}

func testAccRateBasedRuleConfig_changePredicates(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_ipset" "ipset" {
  name = %[1]q

  ip_set_descriptors {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_waf_byte_match_set" "set" {
  name = %[1]q

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
  name        = %[1]q
  metric_name = %[1]q
  rate_key    = "IP"
  rate_limit  = 2000

  predicates {
    data_id = aws_waf_byte_match_set.set.id
    negated = true
    type    = "ByteMatch"
  }
}
`, name)
}

func testAccRateBasedRuleConfig_noPredicates(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_rate_based_rule" "wafrule" {
  name        = %[1]q
  metric_name = %[1]q
  rate_key    = "IP"
  rate_limit  = 2000
}
`, name)
}

func testAccRateBasedRuleConfig_tags1(name, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_waf_rate_based_rule" "wafrule" {
  name        = %[1]q
  metric_name = %[1]q
  rate_key    = "IP"
  rate_limit  = 2000

  tags = {
    %[2]q = %[3]q
  }
}
`, name, tag1Key, tag1Value)
}

func testAccRateBasedRuleConfig_tags2(name, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_waf_rate_based_rule" "wafrule" {
  name        = %[1]q
  metric_name = %[1]q
  rate_key    = "IP"
  rate_limit  = 2000

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, name, tag1Key, tag1Value, tag2Key, tag2Value)
}
