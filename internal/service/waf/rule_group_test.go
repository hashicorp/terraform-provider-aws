// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waf_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/waf/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfwaf "github.com/hashicorp/terraform-provider-aws/internal/service/waf"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccWAFRuleGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var rule awstypes.Rule
	var group awstypes.RuleGroup
	var idx int

	ruleName := fmt.Sprintf("tfacc%s", sdkacctest.RandString(5))
	groupName := fmt.Sprintf("tfacc%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_basic(ruleName, groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(ctx, "aws_waf_rule.test", &rule),
					testAccCheckRuleGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, groupName),
					resource.TestCheckResourceAttr(resourceName, "activated_rule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrMetricName, groupName),
					computeActivatedRuleWithRuleId(&rule, "COUNT", 50, &idx),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "activated_rule.*", map[string]string{
						"action.0.type":    "COUNT",
						names.AttrPriority: "50",
						names.AttrType:     string(awstypes.WafRuleTypeRegular),
					}),
					acctest.MatchResourceAttrGlobalARN(resourceName, names.AttrARN, "waf", regexache.MustCompile(`rulegroup/.+`)),
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

func TestAccWAFRuleGroup_changeNameForceNew(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after awstypes.RuleGroup

	ruleName := fmt.Sprintf("tfacc%s", sdkacctest.RandString(5))
	groupName := fmt.Sprintf("tfacc%s", sdkacctest.RandString(5))
	newGroupName := fmt.Sprintf("tfacc%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_basic(ruleName, groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, groupName),
					resource.TestCheckResourceAttr(resourceName, "activated_rule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrMetricName, groupName),
				),
			},
			{
				Config: testAccRuleGroupConfig_basic(ruleName, newGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, newGroupName),
					resource.TestCheckResourceAttr(resourceName, "activated_rule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrMetricName, newGroupName),
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

func TestAccWAFRuleGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.RuleGroup
	ruleName := fmt.Sprintf("tfacc%s", sdkacctest.RandString(5))
	groupName := fmt.Sprintf("tfacc%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_basic(ruleName, groupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &group),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfwaf.ResourceRuleGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWAFRuleGroup_changeActivatedRules(t *testing.T) {
	ctx := acctest.Context(t)
	var rule0, rule1, rule2, rule3 awstypes.Rule
	var groupBefore, groupAfter awstypes.RuleGroup
	var idx0, idx1, idx2, idx3 int

	groupName := fmt.Sprintf("tfacc%s", sdkacctest.RandString(5))
	ruleName1 := fmt.Sprintf("tfacc%s", sdkacctest.RandString(5))
	ruleName2 := fmt.Sprintf("tfacc%s", sdkacctest.RandString(5))
	ruleName3 := fmt.Sprintf("tfacc%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_basic(ruleName1, groupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, "aws_waf_rule.test", &rule0),
					testAccCheckRuleGroupExists(ctx, resourceName, &groupBefore),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, groupName),
					resource.TestCheckResourceAttr(resourceName, "activated_rule.#", acctest.Ct1),
					computeActivatedRuleWithRuleId(&rule0, "COUNT", 50, &idx0),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "activated_rule.*", map[string]string{
						"action.0.type":    "COUNT",
						names.AttrPriority: "50",
						names.AttrType:     string(awstypes.WafRuleTypeRegular),
					}),
				),
			},
			{
				Config: testAccRuleGroupConfig_changeActivateds(ruleName1, ruleName2, ruleName3, groupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrName, groupName),
					resource.TestCheckResourceAttr(resourceName, "activated_rule.#", acctest.Ct3),
					testAccCheckRuleGroupExists(ctx, resourceName, &groupAfter),

					testAccCheckRuleExists(ctx, "aws_waf_rule.test", &rule1),
					computeActivatedRuleWithRuleId(&rule1, "BLOCK", 10, &idx1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "activated_rule.*", map[string]string{
						"action.0.type":    "BLOCK",
						names.AttrPriority: acctest.Ct10,
						names.AttrType:     string(awstypes.WafRuleTypeRegular),
					}),

					testAccCheckRuleExists(ctx, "aws_waf_rule.test2", &rule2),
					computeActivatedRuleWithRuleId(&rule2, "COUNT", 1, &idx2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "activated_rule.*", map[string]string{
						"action.0.type":    "COUNT",
						names.AttrPriority: acctest.Ct1,
						names.AttrType:     string(awstypes.WafRuleTypeRegular),
					}),

					testAccCheckRuleExists(ctx, "aws_waf_rule.test3", &rule3),
					computeActivatedRuleWithRuleId(&rule3, "BLOCK", 15, &idx3),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "activated_rule.*", map[string]string{
						"action.0.type":    "BLOCK",
						names.AttrPriority: "15",
						names.AttrType:     string(awstypes.WafRuleTypeRegular),
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

// computeActivatedRuleWithRuleId calculates index
// which isn't static because ruleId is generated as part of the test
func computeActivatedRuleWithRuleId(rule *awstypes.Rule, actionType string, priority int, idx *int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ruleResource := tfwaf.ResourceRuleGroup().SchemaMap()["activated_rule"].Elem.(*schema.Resource)

		m := map[string]interface{}{
			names.AttrAction: []interface{}{
				map[string]interface{}{
					names.AttrType: actionType,
				},
			},
			names.AttrPriority: priority,
			"rule_id":          *rule.RuleId,
			names.AttrType:     string(awstypes.WafRuleTypeRegular),
		}

		f := schema.HashResource(ruleResource)
		*idx = f(m)

		return nil
	}
}

func TestAccWAFRuleGroup_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.RuleGroup
	groupName := fmt.Sprintf("test%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_tags1(groupName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, groupName),
					resource.TestCheckResourceAttr(resourceName, "activated_rule.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config: testAccRuleGroupConfig_tags2(groupName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, groupName),
					resource.TestCheckResourceAttr(resourceName, "activated_rule.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccRuleGroupConfig_tags1(groupName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, groupName),
					resource.TestCheckResourceAttr(resourceName, "activated_rule.#", acctest.Ct0),
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

func TestAccWAFRuleGroup_noActivatedRules(t *testing.T) {
	ctx := acctest.Context(t)
	var group awstypes.RuleGroup
	groupName := fmt.Sprintf("test%s", sdkacctest.RandString(5))
	resourceName := "aws_waf_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_noActivateds(groupName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &group),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, groupName),
					resource.TestCheckResourceAttr(resourceName, "activated_rule.#", acctest.Ct0),
				),
			},
		},
	})
}

func testAccCheckRuleGroupDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_waf_rule_group" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).WAFClient(ctx)

			_, err := tfwaf.FindRuleGroupByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("WAF Rule Group %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckRuleGroupExists(ctx context.Context, n string, v *awstypes.RuleGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFClient(ctx)

		output, err := tfwaf.FindRuleGroupByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccRuleGroupConfig_basic(ruleName, groupName string) string {
	return fmt.Sprintf(`
resource "aws_waf_rule" "test" {
  name        = %[1]q
  metric_name = %[1]q
}

resource "aws_waf_rule_group" "test" {
  name        = %[2]q
  metric_name = %[2]q

  activated_rule {
    action {
      type = "COUNT"
    }

    priority = 50
    rule_id  = aws_waf_rule.test.id
  }
}
`, ruleName, groupName)
}

func testAccRuleGroupConfig_changeActivateds(ruleName1, ruleName2, ruleName3, groupName string) string {
	return fmt.Sprintf(`
resource "aws_waf_rule" "test" {
  name        = %[1]q
  metric_name = %[1]q
}

resource "aws_waf_rule" "test2" {
  name        = %[2]q
  metric_name = %[2]q
}

resource "aws_waf_rule" "test3" {
  name        = %[3]q
  metric_name = %[3]q
}

resource "aws_waf_rule_group" "test" {
  name        = %[4]q
  metric_name = %[4]q

  activated_rule {
    action {
      type = "BLOCK"
    }

    priority = 10
    rule_id  = aws_waf_rule.test.id
  }

  activated_rule {
    action {
      type = "COUNT"
    }

    priority = 1
    rule_id  = aws_waf_rule.test2.id
  }

  activated_rule {
    action {
      type = "BLOCK"
    }

    priority = 15
    rule_id  = aws_waf_rule.test3.id
  }
}
`, ruleName1, ruleName2, ruleName3, groupName)
}

func testAccRuleGroupConfig_noActivateds(groupName string) string {
	return fmt.Sprintf(`
resource "aws_waf_rule_group" "test" {
  name        = %[1]q
  metric_name = %[1]q
}
`, groupName)
}

func testAccRuleGroupConfig_tags1(gName, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_waf_rule_group" "test" {
  name        = %[1]q
  metric_name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, gName, tag1Key, tag1Value)
}

func testAccRuleGroupConfig_tags2(gName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_waf_rule_group" "test" {
  name        = %[1]q
  metric_name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, gName, tag1Key, tag1Value, tag2Key, tag2Value)
}
