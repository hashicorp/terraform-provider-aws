package wafregional_test

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/wafregional"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfwafregional "github.com/hashicorp/terraform-provider-aws/internal/service/wafregional"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_wafregional_rule", &resource.Sweeper{
		Name: "aws_wafregional_rule",
		F:    sweepRules,
		Dependencies: []string{
			"aws_wafregional_web_acl",
		},
	})
}

func sweepRules(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).WAFRegionalConn

	input := &waf.ListRulesInput{}

	for {
		output, err := conn.ListRules(input)

		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping WAF Regional Rule sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing WAF Regional Rules: %s", err)
		}

		for _, rule := range output.Rules {
			deleteInput := &waf.DeleteRuleInput{
				RuleId: rule.RuleId,
			}
			id := aws.StringValue(rule.RuleId)
			wr := tfwafregional.NewRetryer(conn, region)

			_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
				deleteInput.ChangeToken = token
				log.Printf("[INFO] Deleting WAF Regional Rule: %s", id)
				return conn.DeleteRule(deleteInput)
			})

			if tfawserr.ErrMessageContains(err, wafregional.ErrCodeWAFNonEmptyEntityException, "") {
				getRuleInput := &waf.GetRuleInput{
					RuleId: rule.RuleId,
				}

				getRuleOutput, getRuleErr := conn.GetRule(getRuleInput)

				if getRuleErr != nil {
					return fmt.Errorf("error getting WAF Regional Rule (%s): %s", id, getRuleErr)
				}

				var updates []*waf.RuleUpdate
				updateRuleInput := &waf.UpdateRuleInput{
					RuleId:  rule.RuleId,
					Updates: updates,
				}

				for _, predicate := range getRuleOutput.Rule.Predicates {
					update := &waf.RuleUpdate{
						Action:    aws.String(waf.ChangeActionDelete),
						Predicate: predicate,
					}

					updateRuleInput.Updates = append(updateRuleInput.Updates, update)
				}

				_, updateWebACLErr := wr.RetryWithToken(func(token *string) (interface{}, error) {
					updateRuleInput.ChangeToken = token
					log.Printf("[INFO] Removing Predicates from WAF Regional Rule: %s", id)
					return conn.UpdateRule(updateRuleInput)
				})

				if updateWebACLErr != nil {
					return fmt.Errorf("error removing predicates from WAF Regional Rule (%s): %s", id, updateWebACLErr)
				}

				_, err = wr.RetryWithToken(func(token *string) (interface{}, error) {
					deleteInput.ChangeToken = token
					log.Printf("[INFO] Deleting WAF Regional Rule: %s", id)
					return conn.DeleteRule(deleteInput)
				})
			}

			if err != nil {
				return fmt.Errorf("error deleting WAF Regional Rule (%s): %s", id, err)
			}
		}

		if aws.StringValue(output.NextMarker) == "" {
			break
		}

		input.NextMarker = output.NextMarker
	}

	return nil
}

func TestAccWAFRegionalRule_basic(t *testing.T) {
	var v waf.Rule
	wafRuleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	resourceName := "aws_wafregional_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(wafregional.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, wafregional.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRuleDestroy,
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
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(wafregional.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, wafregional.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRuleDestroy,
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
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(wafregional.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, wafregional.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckIPSetDestroy,
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
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(wafregional.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, wafregional.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRuleDestroy,
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
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(wafregional.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, wafregional.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRuleDestroy,
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
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(wafregional.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, wafregional.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckWAFRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig(ruleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPSetExists("aws_wafregional_ipset.ipset", &ipset),
					testAccCheckRuleExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "predicate.#", "1"),
					computeWafRegionalRulePredicate(&ipset.IPSetId, false, "IPMatch", &idx),
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
					computeWafRegionalRulePredicate(&xssMatchSet.XssMatchSetId, true, "XssMatch", &idx),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "predicate.*", map[string]string{
						"negated": "true",
						"type":    "XssMatch",
					}),
					computeWafRegionalRulePredicate(&ipset.IPSetId, true, "IPMatch", &idx),
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
func computeWafRegionalRulePredicate(dataId **string, negated bool, pType string, idx *int) resource.TestCheckFunc {
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
		if tfawserr.ErrMessageContains(err, wafregional.ErrCodeWAFNonexistentItemException, "") {
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

func testAccCheckWAFRuleDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_waf_rule" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFConn
		resp, err := conn.GetRule(
			&waf.GetRuleInput{
				RuleId: aws.String(rs.Primary.ID),
			})

		if tfawserr.ErrCodeEquals(err, waf.ErrCodeNonexistentItemException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error reading WAF Rule (%s): %w", rs.Primary.ID, err)
		}

		if resp != nil && resp.Rule != nil {
			return fmt.Errorf("WAF Rule (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}
