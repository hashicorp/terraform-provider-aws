// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafregional_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafregional/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfwafregional "github.com/hashicorp/terraform-provider-aws/internal/service/wafregional"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWAFRegionalRule_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Rule
	wafRuleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	resourceName := "aws_wafregional_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_basic(wafRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "waf-regional", regexache.MustCompile(`rule/.+`)),
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

func TestAccWAFRegionalRule_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Rule
	wafRuleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	resourceName := "aws_wafregional_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_tags1(wafRuleName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &v),
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
				Config: testAccRuleConfig_tags2(wafRuleName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccRuleConfig_tags1(wafRuleName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccWAFRegionalRule_changeNameForceNew(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after awstypes.Rule
	wafRuleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	wafRuleNewName := fmt.Sprintf("wafrulenew%s", sdkacctest.RandString(5))
	resourceName := "aws_wafregional_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIPSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_basic(wafRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, wafRuleName),
					resource.TestCheckResourceAttr(resourceName, "predicate.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrMetricName, wafRuleName),
				),
			},
			{
				Config: testAccRuleConfig_changeName(wafRuleNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &after),
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

func TestAccWAFRegionalRule_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Rule
	wafRuleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	resourceName := "aws_wafregional_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_basic(wafRuleName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfwafregional.ResourceRule(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWAFRegionalRule_noPredicates(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.Rule
	wafRuleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	resourceName := "aws_wafregional_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_noPredicates(wafRuleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, wafRuleName),
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

func TestAccWAFRegionalRule_changePredicates(t *testing.T) {
	ctx := acctest.Context(t)
	var ipset awstypes.IPSet
	var xssMatchSet awstypes.XssMatchSet

	var before, after awstypes.Rule
	var idx int
	ruleName := fmt.Sprintf("wafrule%s", sdkacctest.RandString(5))
	resourceName := "aws_wafregional_rule.wafrule"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_basic(ruleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPSetExists(ctx, "aws_wafregional_ipset.ipset", &ipset),
					testAccCheckRuleExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ruleName),
					resource.TestCheckResourceAttr(resourceName, "predicate.#", acctest.Ct1),
					computeRulePredicate(&ipset.IPSetId, false, "IPMatch", &idx),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "predicate.*", map[string]string{
						"negated":      acctest.CtFalse,
						names.AttrType: "IPMatch",
					}),
				),
			},
			{
				Config: testAccRuleConfig_changePredicates(ruleName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckXSSMatchSetExists(ctx, "aws_wafregional_xss_match_set.xss_match_set", &xssMatchSet),
					testAccCheckRuleExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, ruleName),
					resource.TestCheckResourceAttr(resourceName, "predicate.#", acctest.Ct2),
					computeRulePredicate(&xssMatchSet.XssMatchSetId, true, "XssMatch", &idx),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "predicate.*", map[string]string{
						"negated":      acctest.CtTrue,
						names.AttrType: "XssMatch",
					}),
					computeRulePredicate(&ipset.IPSetId, true, "IPMatch", &idx),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "predicate.*", map[string]string{
						"negated":      acctest.CtTrue,
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

// Calculates the index which isn't static because dataId is generated as part of the test
func computeRulePredicate(dataId **string, negated bool, pType string, idx *int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		predicateResource := tfwafregional.ResourceRule().SchemaMap()["predicate"].Elem.(*schema.Resource)
		m := map[string]interface{}{
			"data_id":      **dataId,
			"negated":      negated,
			names.AttrType: pType,
		}

		f := schema.HashResource(predicateResource)
		*idx = f(m)

		return nil
	}
}

func testAccCheckRuleDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_wafregional_rule" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).WAFRegionalClient(ctx)

			_, err := tfwafregional.FindRuleByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("WAF Regional Rule %s still exists", rs.Primary.ID)
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

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFRegionalClient(ctx)

		output, err := tfwafregional.FindRuleByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccRuleConfig_basic(name string) string {
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

func testAccRuleConfig_tags1(name, tagKey1, tagValue1 string) string {
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

func testAccRuleConfig_tags2(name, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
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

func testAccRuleConfig_changeName(name string) string {
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

func testAccRuleConfig_noPredicates(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_rule" "wafrule" {
  name        = %[1]q
  metric_name = %[1]q
}
`, name)
}

func testAccRuleConfig_changePredicates(name string) string {
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
