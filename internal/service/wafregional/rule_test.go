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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfwafregional "github.com/hashicorp/terraform-provider-aws/internal/service/wafregional"
)

func TestAccWAFRegionalRule_basic(t *testing.T) {
	var v waf.Rule
	wafRuleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	resourceName := "aws_wafregional_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(wafregional.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafregional.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig(wafRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "waf-regional", regexp.MustCompile(`rule/.+`)),
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

func TestAccWAFRegionalRule_tags(t *testing.T) {
	var v waf.Rule
	wafRuleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	resourceName := "aws_wafregional_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(wafregional.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafregional.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleTags1Config(wafRuleName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(resourceName, &v),
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
				Config: testAccRuleTags2Config(wafRuleName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccRuleTags1Config(wafRuleName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccWAFRegionalRule_changeNameForceNew(t *testing.T) {
	var before, after waf.Rule
	wafRuleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	wafRuleNewName := fmt.Sprintf("wafrulenew%s", sdkacctest.RandString(5))
	resourceName := "aws_wafregional_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(wafregional.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafregional.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckIPSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig(wafRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", wafRuleName),
					resource.TestCheckResourceAttr(resourceName, "predicate.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_name", wafRuleName),
				),
			},
			{
				Config: testAccRuleChangeNameConfig(wafRuleNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(resourceName, &after),
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

func TestAccWAFRegionalRule_disappears(t *testing.T) {
	var v waf.Rule
	wafRuleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	resourceName := "aws_wafregional_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(wafregional.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafregional.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig(wafRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(resourceName, &v),
					testAccCheckRuleDisappears(&v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWAFRegionalRule_noPredicates(t *testing.T) {
	var v waf.Rule
	wafRuleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	resourceName := "aws_wafregional_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(wafregional.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafregional.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRule_noPredicates(wafRuleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", wafRuleName),
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

func TestAccWAFRegionalRule_changePredicates(t *testing.T) {
	var ipset waf.IPSet
	var xssMatchSet waf.XssMatchSet

	var before, after waf.Rule
	var idx int
	ruleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	resourceName := "aws_wafregional_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(wafregional.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, wafregional.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig(ruleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPSetExists("aws_wafregional_ipset.ipset", &ipset),
					testAccCheckRuleExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "predicate.#", "1"),
					computeRulePredicate(&ipset.IPSetId, false, "IPMatch", &idx),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "predicate.*", map[string]string{
						"negated": "false",
						"type":    "IPMatch",
					}),
				),
			},
			{
				Config: testAccRule_changePredicates(ruleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXSSMatchSetExists("aws_wafregional_xss_match_set.xss_match_set", &xssMatchSet),
					testAccCheckRuleExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "predicate.#", "2"),
					computeRulePredicate(&xssMatchSet.XssMatchSetId, true, "XssMatch", &idx),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "predicate.*", map[string]string{
						"negated": "true",
						"type":    "XssMatch",
					}),
					computeRulePredicate(&ipset.IPSetId, true, "IPMatch", &idx),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "predicate.*", map[string]string{
						"negated": "true",
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

// Calculates the index which isn't static because dataId is generated as part of the test
func computeRulePredicate(dataId **string, negated bool, pType string, idx *int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		predicateResource := tfwafregional.ResourceRule().Schema["predicate"].Elem.(*schema.Resource)
		m := map[string]interface{}{
			"data_id": **dataId,
			"negated": negated,
			"type":    pType,
		}

		f := schema.HashResource(predicateResource)
		*idx = f(m)

		return nil
	}
}

func testAccCheckRuleDisappears(v *waf.Rule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFRegionalConn
		region := acctest.Provider.Meta().(*conns.AWSClient).Region

		wr := tfwafregional.NewRetryer(conn, region)
		_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
			req := &waf.UpdateRuleInput{
				ChangeToken: token,
				RuleId:      v.RuleId,
			}

			for _, predicate := range v.Predicates {
				predicate := &waf.RuleUpdate{
					Action: aws.String("DELETE"),
					Predicate: &waf.Predicate{
						Negated: predicate.Negated,
						Type:    predicate.Type,
						DataId:  predicate.DataId,
					},
				}
				req.Updates = append(req.Updates, predicate)
			}

			return conn.UpdateRule(req)
		})
		if err != nil {
			return fmt.Errorf("Error Updating WAF Rule: %s", err)
		}

		_, err = wr.RetryWithToken(func(token *string) (interface{}, error) {
			opts := &waf.DeleteRuleInput{
				ChangeToken: token,
				RuleId:      v.RuleId,
			}
			return conn.DeleteRule(opts)
		})
		if err != nil {
			return fmt.Errorf("Error Deleting WAF Rule: %s", err)
		}
		return nil
	}
}

func testAccCheckRuleDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_wafregional_rule" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFRegionalConn
		resp, err := conn.GetRule(
			&waf.GetRuleInput{
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

func testAccCheckRuleExists(n string, v *waf.Rule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAF Rule ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFRegionalConn
		resp, err := conn.GetRule(&waf.GetRuleInput{
			RuleId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if *resp.Rule.RuleId == rs.Primary.ID {
			*v = *resp.Rule
			return nil
		}

		return fmt.Errorf("WAF Rule (%s) not found", rs.Primary.ID)
	}
}

func testAccRuleConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_ipset" "ipset" {
  name = %[1]q

  ip_set_descriptor {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_wafregional_rule" "wafrule" {
  name        = %[1]q
  metric_name = %[1]q

  predicate {
    data_id = aws_wafregional_ipset.ipset.id
    negated = false
    type    = "IPMatch"
  }
}
`, name)
}

func testAccRuleTags1Config(name, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_ipset" "ipset" {
  name = %[1]q

  ip_set_descriptor {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_wafregional_rule" "wafrule" {
  name        = %[1]q
  metric_name = %[1]q

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

func testAccRuleTags2Config(name, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_ipset" "ipset" {
  name = %[1]q

  ip_set_descriptor {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_wafregional_rule" "wafrule" {
  name        = %[1]q
  metric_name = %[1]q

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

func testAccRuleChangeNameConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_ipset" "ipset" {
  name = %[1]q

  ip_set_descriptor {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_wafregional_rule" "wafrule" {
  name        = %[1]q
  metric_name = %[1]q

  predicate {
    data_id = aws_wafregional_ipset.ipset.id
    negated = false
    type    = "IPMatch"
  }
}
`, name)
}

func testAccRule_noPredicates(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_rule" "wafrule" {
  name        = %[1]q
  metric_name = %[1]q
}
`, name)
}

func testAccRule_changePredicates(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_ipset" "ipset" {
  name = %[1]q

  ip_set_descriptor {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_wafregional_xss_match_set" "xss_match_set" {
  name = %[1]q

  xss_match_tuple {
    text_transformation = "NONE"

    field_to_match {
      type = "URI"
    }
  }
}

resource "aws_wafregional_rule" "wafrule" {
  name        = %[1]q
  metric_name = %[1]q

  predicate {
    data_id = aws_wafregional_xss_match_set.xss_match_set.id
    negated = true
    type    = "XssMatch"
  }

  predicate {
    data_id = aws_wafregional_ipset.ipset.id
    negated = true
    type    = "IPMatch"
  }
}
`, name)
}
