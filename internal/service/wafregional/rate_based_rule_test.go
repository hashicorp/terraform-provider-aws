package wafregional_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/wafregional"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfwafregional "github.com/hashicorp/terraform-provider-aws/internal/service/wafregional"
)

func TestAccWAFRegionalRateBasedRule_basic(t *testing.T) {
	var v waf.RateBasedRule
	resourceName := "aws_wafregional_rate_based_rule.wafrule"
	wafRuleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(wafregional.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafregional.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRateBasedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRateBasedRuleConfig_basic(wafRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRateBasedRuleExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "waf-regional", regexp.MustCompile(`ratebasedrule/.+`)),
					resource.TestCheckResourceAttr(resourceName, "name", wafRuleName),
					resource.TestCheckResourceAttr(resourceName, "predicate.#", "1"),
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

func TestAccWAFRegionalRateBasedRule_tags(t *testing.T) {
	var v waf.RateBasedRule
	resourceName := "aws_wafregional_rate_based_rule.wafrule"
	wafRuleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(wafregional.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafregional.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRateBasedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRateBasedRuleConfig_tags1(wafRuleName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRateBasedRuleExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRateBasedRuleConfig_tags2(wafRuleName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRateBasedRuleExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccRateBasedRuleConfig_tags1(wafRuleName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRateBasedRuleExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccWAFRegionalRateBasedRule_changeNameForceNew(t *testing.T) {
	var before, after waf.RateBasedRule
	resourceName := "aws_wafregional_rate_based_rule.wafrule"
	wafRuleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	wafRuleNewName := fmt.Sprintf("wafrulenew%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(wafregional.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafregional.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRateBasedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRateBasedRuleConfig_basic(wafRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRateBasedRuleExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", wafRuleName),
					resource.TestCheckResourceAttr(resourceName, "predicate.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_name", wafRuleName),
				),
			},
			{
				Config: testAccRateBasedRuleConfig_changeName(wafRuleNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRateBasedRuleExists(resourceName, &after),
					testAccCheckRateBasedRuleIdDiffers(&before, &after),
					resource.TestCheckResourceAttr(resourceName, "name", wafRuleNewName),
					resource.TestCheckResourceAttr(resourceName, "predicate.#", "1"),
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

func TestAccWAFRegionalRateBasedRule_disappears(t *testing.T) {
	var v waf.RateBasedRule
	resourceName := "aws_wafregional_rate_based_rule.wafrule"
	wafRuleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(wafregional.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafregional.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRateBasedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRateBasedRuleConfig_basic(wafRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRateBasedRuleExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfwafregional.ResourceRateBasedRule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWAFRegionalRateBasedRule_changePredicates(t *testing.T) {
	var ipset waf.IPSet
	var byteMatchSet waf.ByteMatchSet

	var before, after waf.RateBasedRule
	resourceName := "aws_wafregional_rate_based_rule.wafrule"
	ruleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(wafregional.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafregional.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRateBasedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRateBasedRuleConfig_basic(ruleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPSetExists("aws_wafregional_ipset.ipset", &ipset),
					testAccCheckRateBasedRuleExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "predicate.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "predicate.*", map[string]string{
						"negated": "false",
						"type":    "IPMatch",
					}),
				),
			},
			{
				Config: testAccRateBasedRuleConfig_changePredicates(ruleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckByteMatchSetExists("aws_wafregional_byte_match_set.set", &byteMatchSet),
					testAccCheckRateBasedRuleExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "predicate.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "predicate.*", map[string]string{
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

func TestAccWAFRegionalRateBasedRule_changeRateLimit(t *testing.T) {
	var before, after waf.RateBasedRule
	resourceName := "aws_wafregional_rate_based_rule.wafrule"
	ruleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	rateLimitBefore := "2000"
	rateLimitAfter := "2001"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(wafregional.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafregional.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRateBasedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRateBasedRuleConfig_limit(ruleName, rateLimitBefore),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRateBasedRuleExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "rate_limit", rateLimitBefore),
				),
			},
			{
				Config: testAccRateBasedRuleConfig_limit(ruleName, rateLimitAfter),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRateBasedRuleExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", ruleName),
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
	var rule waf.RateBasedRule
	resourceName := "aws_wafregional_rate_based_rule.wafrule"
	ruleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(wafregional.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafregional.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRateBasedRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRateBasedRuleConfig_noPredicates(ruleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRateBasedRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "predicate.#", "0"),
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

func testAccCheckRateBasedRuleIdDiffers(before, after *waf.RateBasedRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *before.RuleId == *after.RuleId {
			return fmt.Errorf("Expected different IDs, given %q for both rules", *before.RuleId)
		}
		return nil
	}
}

func testAccCheckRateBasedRuleDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_wafregional_rate_based_rule" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFRegionalConn
		resp, err := conn.GetRateBasedRule(
			&waf.GetRateBasedRuleInput{
				RuleId: aws.String(rs.Primary.ID),
			})

		if err == nil {
			if *resp.Rule.RuleId == rs.Primary.ID {
				return fmt.Errorf("WAF Rule %s still exists", rs.Primary.ID)
			}
		}

		// Return nil if the Rule is already destroyed
		if tfawserr.ErrCodeEquals(err, wafregional.ErrCodeWAFNonexistentItemException) {
			return nil
		}

		return err
	}

	return nil
}

func testAccCheckRateBasedRuleExists(n string, v *waf.RateBasedRule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAF Rule ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFRegionalConn
		resp, err := conn.GetRateBasedRule(&waf.GetRateBasedRuleInput{
			RuleId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if *resp.Rule.RuleId == rs.Primary.ID {
			*v = *resp.Rule
			return nil
		}

		return fmt.Errorf("WAF Regional Rule (%s) not found", rs.Primary.ID)
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
  name        = "%s"
  metric_name = "%s"
  rate_key    = "IP"
  rate_limit  = %s
}
`, name, name, limit)
}
