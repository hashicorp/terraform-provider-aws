package waf_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfwaf "github.com/hashicorp/terraform-provider-aws/internal/service/waf"
)

func TestAccWAFRule_basic(t *testing.T) {
	var v waf.Rule
	wafRuleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_rule.wafrule"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, waf.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_basic(wafRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", wafRuleName),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_name", wafRuleName),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "waf", regexp.MustCompile(`rule/.+`)),
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

func TestAccWAFRule_changeNameForceNew(t *testing.T) {
	var before, after waf.Rule
	wafRuleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	wafRuleNewName := fmt.Sprintf("wafrulenew%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, waf.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckIPSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_basic(wafRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", wafRuleName),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_name", wafRuleName),
				),
			},
			{
				Config: testAccRuleConfig_changeName(wafRuleNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", wafRuleNewName),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_name", wafRuleNewName),
				),
			},
		},
	})
}

func TestAccWAFRule_disappears(t *testing.T) {
	var v waf.Rule
	wafRuleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, waf.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_basic(wafRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfwaf.ResourceRule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWAFRule_changePredicates(t *testing.T) {
	var ipset waf.IPSet
	var byteMatchSet waf.ByteMatchSet

	var before, after waf.Rule
	ruleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, waf.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_basic(ruleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPSetExists("aws_waf_ipset.ipset", &ipset),
					testAccCheckRuleExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "predicates.*", map[string]string{
						"negated": "false",
						"type":    "IPMatch",
					}),
				),
			},
			{
				Config: testAccRuleConfig_changePredicates(ruleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckByteMatchSetExists("aws_waf_byte_match_set.set", &byteMatchSet),
					testAccCheckRuleExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "predicates.*", map[string]string{
						"negated": "true",
						"type":    "ByteMatch",
					}),
				),
			},
		},
	})
}

func TestAccWAFRule_geoMatchSetPredicate(t *testing.T) {
	var geoMatchSet waf.GeoMatchSet

	var v waf.Rule
	ruleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, waf.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_geoMatchSetPredicate(ruleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGeoMatchSetExists("aws_waf_geo_match_set.geo_match_set", &geoMatchSet),
					testAccCheckRuleExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "predicates.*", map[string]string{
						"negated": "true",
						"type":    "GeoMatch",
					}),
				),
			},
		},
	})
}

// TestAccWAFRule_webACL validates the resource's
// retry behavior when removed from a WebACL
func TestAccWAFRule_webACL(t *testing.T) {
	var rule waf.Rule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_waf_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, waf.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_referencedByWebACL(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(resourceName, &rule),
				),
			},
			{
				Config: testAccRuleConfig_webACLNos(rName),
			},
		},
	})
}

func TestAccWAFRule_noPredicates(t *testing.T) {
	var rule waf.Rule
	ruleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, waf.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_noPredicates(ruleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", "0"),
				),
			},
		},
	})
}

func TestAccWAFRule_tags(t *testing.T) {
	var rule waf.Rule
	ruleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, waf.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckWebACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_tags1(ruleName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccRuleConfig_tags2(ruleName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "name", ruleName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccRuleConfig_tags1(ruleName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_name", ruleName),
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

func testAccCheckRuleDestroy(s *terraform.State) error {
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

func testAccCheckRuleExists(n string, v *waf.Rule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAF Rule ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFConn
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

func testAccPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).WAFConn

	input := &waf.ListRulesInput{}

	_, err := conn.ListRules(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccRuleConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_ipset" "ipset" {
  name = %[1]q

  ip_set_descriptors {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_waf_rule" "wafrule" {
  depends_on  = [aws_waf_ipset.ipset]
  name        = %[1]q
  metric_name = %[1]q

  predicates {
    data_id = aws_waf_ipset.ipset.id
    negated = false
    type    = "IPMatch"
  }
}
`, name)
}

func testAccRuleConfig_changeName(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_ipset" "ipset" {
  name = %[1]q

  ip_set_descriptors {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_waf_rule" "wafrule" {
  depends_on  = [aws_waf_ipset.ipset]
  name        = %[1]q
  metric_name = %[1]q

  predicates {
    data_id = aws_waf_ipset.ipset.id
    negated = false
    type    = "IPMatch"
  }
}
`, name)
}

func testAccRuleConfig_changePredicates(name string) string {
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

resource "aws_waf_rule" "wafrule" {
  name        = %[1]q
  metric_name = %[1]q

  predicates {
    data_id = aws_waf_byte_match_set.set.id
    negated = true
    type    = "ByteMatch"
  }
}
`, name)
}

func testAccRuleConfig_noPredicates(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_rule" "wafrule" {
  name        = %[1]q
  metric_name = %[1]q
}
`, name)
}

func testAccRuleConfig_geoMatchSetPredicate(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_geo_match_set" "geo_match_set" {
  name = %[1]q

  geo_match_constraint {
    type  = "Country"
    value = "US"
  }
}

resource "aws_waf_rule" "wafrule" {
  name        = %[1]q
  metric_name = %[1]q

  predicates {
    data_id = aws_waf_geo_match_set.geo_match_set.id
    negated = true
    type    = "GeoMatch"
  }
}
`, name)
}

func testAccRuleConfig_tags1(rName, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_waf_ipset" "ipset" {
  name = %[1]q

  ip_set_descriptors {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_waf_rule" "wafrule" {
  depends_on  = [aws_waf_ipset.ipset]
  name        = %[1]q
  metric_name = %[1]q

  predicates {
    data_id = aws_waf_ipset.ipset.id
    negated = false
    type    = "IPMatch"
  }

  tags = {
    %q = %q
  }
}
`, rName, tag1Key, tag1Value)
}

func testAccRuleConfig_tags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_waf_ipset" "ipset" {
  name = %[1]q

  ip_set_descriptors {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_waf_rule" "wafrule" {
  depends_on  = [aws_waf_ipset.ipset]
  name        = %[1]q
  metric_name = %[1]q

  predicates {
    data_id = aws_waf_ipset.ipset.id
    negated = false
    type    = "IPMatch"
  }

  tags = {
    %q = %q
    %q = %q
  }
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value)
}

func testAccRuleConfig_referencedByWebACL(rName string) string {
	return fmt.Sprintf(`
resource "aws_waf_ipset" "test" {
  name = %[1]q

  ip_set_descriptors {
    type  = "IPV4"
    value = "192.0.7.0/24"
  }
}

resource "aws_waf_rule" "test" {
  metric_name = "testrulemetric"
  name        = %[1]q

  predicates {
    data_id = aws_waf_ipset.test.id
    negated = false
    type    = "IPMatch"
  }
}

resource "aws_waf_web_acl" "test" {
  metric_name = "testwebaclmetric"
  name        = %[1]q

  default_action {
    type = "ALLOW"
  }

  rules {
    priority = 1
    rule_id  = aws_waf_rule.test.id

    action {
      type = "BLOCK"
    }
  }
}
`, rName)
}

func testAccRuleConfig_webACLNos(rName string) string {
	return fmt.Sprintf(`
resource "aws_waf_web_acl" "test" {
  metric_name = "testwebaclmetric"
  name        = %[1]q

  default_action {
    type = "ALLOW"
  }
}
`, rName)
}
