// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waf_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/waf"
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

func TestAccWAFRule_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Rule
	wafRuleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_rule.wafrule"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_basic(wafRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, wafRuleName),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrMetricName, wafRuleName),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "waf", regexache.MustCompile(`rule/.+`)),
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
	ctx := acctest.Context(t)
	var before, after awstypes.Rule
	wafRuleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	wafRuleNewName := fmt.Sprintf("wafrulenew%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_basic(wafRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, wafRuleName),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrMetricName, wafRuleName),
				),
			},
			{
				Config: testAccRuleConfig_changeName(wafRuleNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, wafRuleNewName),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrMetricName, wafRuleNewName),
				),
			},
		},
	})
}

func TestAccWAFRule_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Rule
	wafRuleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_basic(wafRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfwaf.ResourceRule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWAFRule_changePredicates(t *testing.T) {
	ctx := acctest.Context(t)
	var ipset awstypes.IPSet
	var byteMatchSet awstypes.ByteMatchSet

	var before, after awstypes.Rule
	ruleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_basic(ruleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPSetExists(ctx, "aws_waf_ipset.ipset", &ipset),
					testAccCheckRuleExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ruleName),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "predicates.*", map[string]string{
						"negated":      acctest.CtFalse,
						names.AttrType: "IPMatch",
					}),
				),
			},
			{
				Config: testAccRuleConfig_changePredicates(ruleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckByteMatchSetExists(ctx, "aws_waf_byte_match_set.set", &byteMatchSet),
					testAccCheckRuleExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ruleName),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "predicates.*", map[string]string{
						"negated":      acctest.CtTrue,
						names.AttrType: "ByteMatch",
					}),
				),
			},
		},
	})
}

func TestAccWAFRule_geoMatchSetPredicate(t *testing.T) {
	ctx := acctest.Context(t)
	var geoMatchSet awstypes.GeoMatchSet

	var v awstypes.Rule
	ruleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_geoMatchSetPredicate(ruleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGeoMatchSetExists(ctx, "aws_waf_geo_match_set.geo_match_set", &geoMatchSet),
					testAccCheckRuleExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ruleName),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "predicates.*", map[string]string{
						"negated":      acctest.CtTrue,
						names.AttrType: "GeoMatch",
					}),
				),
			},
		},
	})
}

// TestAccWAFRule_webACL validates the resource's
// retry behavior when removed from a WebACL
func TestAccWAFRule_webACL(t *testing.T) {
	ctx := acctest.Context(t)
	var rule awstypes.Rule
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_waf_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_referencedByWebACL(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &rule),
				),
			},
			{
				Config: testAccRuleConfig_webACLNos(rName),
			},
		},
	})
}

func TestAccWAFRule_noPredicates(t *testing.T) {
	ctx := acctest.Context(t)
	var rule awstypes.Rule
	ruleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_noPredicates(ruleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ruleName),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccWAFRule_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var rule awstypes.Rule
	ruleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_tags1(ruleName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrMetricName, ruleName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ruleName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config: testAccRuleConfig_tags2(ruleName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrMetricName, ruleName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ruleName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccRuleConfig_tags1(ruleName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &rule),
					resource.TestCheckResourceAttr(resourceName, "predicates.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrMetricName, ruleName),
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

func testAccCheckRuleDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_waf_rule" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).WAFClient(ctx)

			_, err := tfwaf.FindRuleByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("WAF Rule %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckRuleExists(ctx context.Context, n string, v *awstypes.Rule) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFClient(ctx)

		output, err := tfwaf.FindRuleByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).WAFClient(ctx)

	input := &waf.ListRulesInput{}

	_, err := conn.ListRules(ctx, input)

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
    %[2]q = %[3]q
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
    %[2]q = %[3]q
    %[4]q = %[5]q
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
  lifecycle {
    create_before_destroy = true
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
  lifecycle {
    create_before_destroy = true
  }
}
`, rName)
}
