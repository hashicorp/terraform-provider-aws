// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafregional_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafregional/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfwafregional "github.com/hashicorp/terraform-provider-aws/internal/service/wafregional"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWAFRegionalRateBasedRule_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RateBasedRule
	resourceName := "aws_wafregional_rate_based_rule.wafrule"
	wafRuleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRateBasedRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRateBasedRuleConfig_basic(wafRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRateBasedRuleExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "waf-regional", regexache.MustCompile(`ratebasedrule/.+`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, wafRuleName),
					resource.TestCheckResourceAttr(resourceName, "predicate.#", acctest.Ct1),
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

func TestAccWAFRegionalRateBasedRule_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RateBasedRule
	resourceName := "aws_wafregional_rate_based_rule.wafrule"
	wafRuleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRateBasedRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRateBasedRuleConfig_tags1(wafRuleName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRateBasedRuleExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRateBasedRuleConfig_tags2(wafRuleName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRateBasedRuleExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccRateBasedRuleConfig_tags1(wafRuleName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRateBasedRuleExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccWAFRegionalRateBasedRule_changeNameForceNew(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after awstypes.RateBasedRule
	resourceName := "aws_wafregional_rate_based_rule.wafrule"
	wafRuleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	wafRuleNewName := fmt.Sprintf("wafrulenew%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRateBasedRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRateBasedRuleConfig_basic(wafRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRateBasedRuleExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, wafRuleName),
					resource.TestCheckResourceAttr(resourceName, "predicate.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrMetricName, wafRuleName),
				),
			},
			{
				Config: testAccRateBasedRuleConfig_changeName(wafRuleNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRateBasedRuleExists(ctx, resourceName, &after),
					testAccCheckRateBasedRuleIdDiffers(&before, &after),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, wafRuleNewName),
					resource.TestCheckResourceAttr(resourceName, "predicate.#", acctest.Ct1),
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

func TestAccWAFRegionalRateBasedRule_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.RateBasedRule
	resourceName := "aws_wafregional_rate_based_rule.wafrule"
	wafRuleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRateBasedRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRateBasedRuleConfig_basic(wafRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRateBasedRuleExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfwafregional.ResourceRateBasedRule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWAFRegionalRateBasedRule_changePredicates(t *testing.T) {
	ctx := acctest.Context(t)
	var ipset awstypes.IPSet
	var byteMatchSet awstypes.ByteMatchSet

	var before, after awstypes.RateBasedRule
	resourceName := "aws_wafregional_rate_based_rule.wafrule"
	ruleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRateBasedRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRateBasedRuleConfig_basic(ruleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPSetExists(ctx, "aws_wafregional_ipset.ipset", &ipset),
					testAccCheckRateBasedRuleExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ruleName),
					resource.TestCheckResourceAttr(resourceName, "predicate.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "predicate.*", map[string]string{
						"negated":      acctest.CtFalse,
						names.AttrType: "IPMatch",
					}),
				),
			},
			{
				Config: testAccRateBasedRuleConfig_changePredicates(ruleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckByteMatchSetExists(ctx, "aws_wafregional_byte_match_set.set", &byteMatchSet),
					testAccCheckRateBasedRuleExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ruleName),
					resource.TestCheckResourceAttr(resourceName, "predicate.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "predicate.*", map[string]string{
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

func TestAccWAFRegionalRateBasedRule_changeRateLimit(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after awstypes.RateBasedRule
	resourceName := "aws_wafregional_rate_based_rule.wafrule"
	ruleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	rateLimitBefore := "2000"
	rateLimitAfter := "2001"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRateBasedRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRateBasedRuleConfig_limit(ruleName, rateLimitBefore),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRateBasedRuleExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ruleName),
					resource.TestCheckResourceAttr(resourceName, "rate_limit", rateLimitBefore),
				),
			},
			{
				Config: testAccRateBasedRuleConfig_limit(ruleName, rateLimitAfter),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRateBasedRuleExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ruleName),
					resource.TestCheckResourceAttr(resourceName, "rate_limit", rateLimitAfter),
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

func TestAccWAFRegionalRateBasedRule_noPredicates(t *testing.T) {
	ctx := acctest.Context(t)
	var rule awstypes.RateBasedRule
	resourceName := "aws_wafregional_rate_based_rule.wafrule"
	ruleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRateBasedRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRateBasedRuleConfig_noPredicates(ruleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRateBasedRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ruleName),
					resource.TestCheckResourceAttr(resourceName, "predicate.#", acctest.Ct0),
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

func testAccCheckRateBasedRuleIdDiffers(before, after *awstypes.RateBasedRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *before.RuleId == *after.RuleId {
			return fmt.Errorf("Expected different IDs, given %q for both rules", *before.RuleId)
		}
		return nil
	}
}

func testAccCheckRateBasedRuleDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_wafregional_rate_based_rule" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).WAFRegionalClient(ctx)

			_, err := tfwafregional.FindRateBasedRuleByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("WAF Regional Rate Based Rule %s still exists", rs.Primary.ID)
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

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFRegionalClient(ctx)

		output, err := tfwafregional.FindRateBasedRuleByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccRateBasedRuleConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_ipset" "ipset" {
  name = %[1]q

  ip_set_descriptor {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_wafregional_rate_based_rule" "wafrule" {
  name        = %[1]q
  metric_name = %[1]q
  rate_key    = "IP"
  rate_limit  = 2000

  predicate {
    data_id = aws_wafregional_ipset.ipset.id
    negated = false
    type    = "IPMatch"
  }
}
`, name)
}

func testAccRateBasedRuleConfig_tags1(name, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_ipset" "ipset" {
  name = %[1]q

  ip_set_descriptor {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_wafregional_rate_based_rule" "wafrule" {
  name        = %[1]q
  metric_name = %[1]q
  rate_key    = "IP"
  rate_limit  = 2000

  predicate {
    data_id = aws_wafregional_ipset.ipset.id
    negated = false
    type    = "IPMatch"
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, name, tagKey1, tagValue1)
}

func testAccRateBasedRuleConfig_tags2(name, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_ipset" "ipset" {
  name = %[1]q

  ip_set_descriptor {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_wafregional_rate_based_rule" "wafrule" {
  name        = %[1]q
  metric_name = %[1]q
  rate_key    = "IP"
  rate_limit  = 2000

  predicate {
    data_id = aws_wafregional_ipset.ipset.id
    negated = false
    type    = "IPMatch"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, name, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccRateBasedRuleConfig_changeName(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_ipset" "ipset" {
  name = %[1]q

  ip_set_descriptor {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_wafregional_rate_based_rule" "wafrule" {
  name        = %[1]q
  metric_name = %[1]q
  rate_key    = "IP"
  rate_limit  = 2000

  predicate {
    data_id = aws_wafregional_ipset.ipset.id
    negated = false
    type    = "IPMatch"
  }
}
`, name)
}

func testAccRateBasedRuleConfig_changePredicates(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_ipset" "ipset" {
  name = %[1]q

  ip_set_descriptor {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_wafregional_byte_match_set" "set" {
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

resource "aws_wafregional_rate_based_rule" "wafrule" {
  name        = %[1]q
  metric_name = %[1]q
  rate_key    = "IP"
  rate_limit  = 2000

  predicate {
    data_id = aws_wafregional_byte_match_set.set.id
    negated = true
    type    = "ByteMatch"
  }
}
`, name)
}

func testAccRateBasedRuleConfig_noPredicates(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_rate_based_rule" "wafrule" {
  name        = %[1]q
  metric_name = %[1]q
  rate_key    = "IP"
  rate_limit  = 2000
}
`, name)
}

func testAccRateBasedRuleConfig_limit(name string, limit string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_rate_based_rule" "wafrule" {
  name        = %[1]q
  metric_name = %[1]q
  rate_key    = "IP"
  rate_limit  = %[2]q
}
`, name, limit)
}
