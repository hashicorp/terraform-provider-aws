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

func TestAccWAFRegionalWebACL_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.WebACL
	wafAclName := fmt.Sprintf("wafacl%s", sdkacctest.RandString(5))
	resourceName := "aws_wafregional_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_basic(wafAclName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "waf-regional", regexache.MustCompile(`webacl/.+`)),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, wafAclName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrMetricName, wafAclName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", acctest.Ct0),
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

func TestAccWAFRegionalWebACL_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.WebACL
	wafAclName := fmt.Sprintf("wafacl%s", sdkacctest.RandString(5))
	resourceName := "aws_wafregional_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_tags1(wafAclName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
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
				Config: testAccWebACLConfig_tags2(wafAclName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccWebACLConfig_tags1(wafAclName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccWAFRegionalWebACL_createRateBased(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.WebACL
	wafAclName := fmt.Sprintf("wafacl%s", sdkacctest.RandString(5))
	resourceName := "aws_wafregional_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_rateBased(wafAclName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, wafAclName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrMetricName, wafAclName),
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

func TestAccWAFRegionalWebACL_createGroup(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.WebACL
	wafAclName := fmt.Sprintf("wafacl%s", sdkacctest.RandString(5))
	resourceName := "aws_wafregional_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_group(wafAclName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, wafAclName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrMetricName, wafAclName),
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

func TestAccWAFRegionalWebACL_changeNameForceNew(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after awstypes.WebACL
	wafAclName := fmt.Sprintf("wafacl%s", sdkacctest.RandString(5))
	wafAclNewName := fmt.Sprintf("wafacl%s", sdkacctest.RandString(5))
	resourceName := "aws_wafregional_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_basic(wafAclName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, wafAclName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrMetricName, wafAclName),
				),
			},
			{
				Config: testAccWebACLConfig_changeName(wafAclNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, wafAclNewName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrMetricName, wafAclNewName),
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

func TestAccWAFRegionalWebACL_changeDefaultAction(t *testing.T) {
	ctx := acctest.Context(t)
	var before, after awstypes.WebACL
	wafAclName := fmt.Sprintf("wafacl%s", sdkacctest.RandString(5))
	wafAclNewName := fmt.Sprintf("wafacl%s", sdkacctest.RandString(5))
	resourceName := "aws_wafregional_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_basic(wafAclName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, wafAclName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrMetricName, wafAclName),
				),
			},
			{
				Config: testAccWebACLConfig_changeDefaultAction(wafAclNewName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "BLOCK"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, wafAclNewName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrMetricName, wafAclNewName),
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

func TestAccWAFRegionalWebACL_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.WebACL
	wafAclName := fmt.Sprintf("wafacl%s", sdkacctest.RandString(5))
	resourceName := "aws_wafregional_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_basic(wafAclName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfwafregional.ResourceWebACL(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWAFRegionalWebACL_noRules(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.WebACL
	wafAclName := fmt.Sprintf("wafacl%s", sdkacctest.RandString(5))
	resourceName := "aws_wafregional_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_webACLNos(wafAclName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, wafAclName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct0),
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

func TestAccWAFRegionalWebACL_changeRules(t *testing.T) {
	ctx := acctest.Context(t)
	var v awstypes.WebACL
	var r awstypes.Rule
	var idx int
	wafAclName := fmt.Sprintf("wafacl%s", sdkacctest.RandString(5))
	resourceName := "aws_wafregional_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_basic(wafAclName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleExists(ctx, "aws_wafregional_rule.test", &r),
					testAccCheckWebACLExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, wafAclName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct1),
					computeWebACLRuleIndex(&r.RuleId, 1, "REGULAR", "BLOCK", &idx),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.*", map[string]string{
						names.AttrPriority: acctest.Ct1,
					}),
				),
			},
			{
				Config: testAccWebACLConfig_changeRules(wafAclName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "default_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "default_action.0.type", "ALLOW"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, wafAclName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtRulePound, acctest.Ct2),
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

func TestAccWAFRegionalWebACL_logging(t *testing.T) {
	ctx := acctest.Context(t)
	var webACL1, webACL2, webACL3 awstypes.WebACL
	rName := fmt.Sprintf("wafacl%s", sdkacctest.RandString(5))
	resourceName := "aws_wafregional_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.WAFRegionalEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFRegionalServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLConfig_loggingConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &webACL1),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.redacted_fields.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.redacted_fields.0.field_to_match.#", acctest.Ct2),
				),
			},
			// Test logging configuration update
			{
				Config: testAccWebACLConfig_loggingConfigurationUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &webACL2),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.redacted_fields.#", acctest.Ct0),
				),
			},
			// Test logging configuration removal
			{
				Config: testAccRuleConfig_webACLNos(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLExists(ctx, resourceName, &webACL3),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", acctest.Ct0),
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

// Calculates the index which isn't static because ruleId is generated as part of the test
func computeWebACLRuleIndex(ruleId **string, priority int, ruleType string, actionType string, idx *int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ruleResource := tfwafregional.ResourceWebACL().SchemaMap()[names.AttrRule].Elem.(*schema.Resource)
		actionMap := map[string]interface{}{
			names.AttrType: actionType,
		}
		m := map[string]interface{}{
			"rule_id":          **ruleId,
			names.AttrType:     ruleType,
			names.AttrPriority: priority,
			names.AttrAction:   []interface{}{actionMap},
			"override_action":  []interface{}{},
		}

		f := schema.HashResource(ruleResource)
		*idx = f(m)

		return nil
	}
}

func testAccCheckWebACLDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_wafregional_web_acl" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).WAFRegionalClient(ctx)

			_, err := tfwafregional.FindWebACLByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("WAF Regional Web ACL %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckWebACLExists(ctx context.Context, n string, v *awstypes.WebACL) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFRegionalClient(ctx)

		output, err := tfwafregional.FindWebACLByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccWebACLConfig_basic(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_rule" "test" {
  name        = %[1]q
  metric_name = %[1]q
}

resource "aws_wafregional_web_acl" "test" {
  name        = %[1]q
  metric_name = %[1]q

  default_action {
    type = "ALLOW"
  }

  rule {
    action {
      type = "BLOCK"
    }

    priority = 1
    rule_id  = aws_wafregional_rule.test.id
  }
}
`, name)
}

func testAccWebACLConfig_tags1(name, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_rule" "test" {
  name        = %[1]q
  metric_name = %[1]q
}

resource "aws_wafregional_web_acl" "test" {
  name        = %[1]q
  metric_name = %[1]q

  default_action {
    type = "ALLOW"
  }

  rule {
    action {
      type = "BLOCK"
    }

    priority = 1
    rule_id  = aws_wafregional_rule.test.id
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, name, tagKey1, tagValue1)
}

func testAccWebACLConfig_tags2(name, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_rule" "test" {
  name        = %[1]q
  metric_name = %[1]q
}

resource "aws_wafregional_web_acl" "test" {
  name        = %[1]q
  metric_name = %[1]q

  default_action {
    type = "ALLOW"
  }

  rule {
    action {
      type = "BLOCK"
    }

    priority = 1
    rule_id  = aws_wafregional_rule.test.id
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, name, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccWebACLConfig_rateBased(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_rate_based_rule" "test" {
  name        = %[1]q
  metric_name = %[1]q

  rate_key   = "IP"
  rate_limit = 2000
}

resource "aws_wafregional_web_acl" "test" {
  name        = %[1]q
  metric_name = %[1]q

  default_action {
    type = "ALLOW"
  }

  rule {
    action {
      type = "BLOCK"
    }

    priority = 1
    type     = "RATE_BASED"
    rule_id  = aws_wafregional_rate_based_rule.test.id
  }
}
`, name)
}

func testAccWebACLConfig_group(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_rule_group" "test" {
  name        = %[1]q
  metric_name = %[1]q
}

resource "aws_wafregional_web_acl" "test" {
  name        = %[1]q
  metric_name = %[1]q

  default_action {
    type = "ALLOW"
  }

  rule {
    override_action {
      type = "NONE"
    }

    priority = 1
    type     = "GROUP"
    rule_id  = aws_wafregional_rule_group.test.id
  }
}
`, name)
}

func testAccWebACLConfig_changeName(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_rule" "test" {
  name        = %[1]q
  metric_name = %[1]q
}

resource "aws_wafregional_web_acl" "test" {
  name        = %[1]q
  metric_name = %[1]q

  default_action {
    type = "ALLOW"
  }

  rule {
    action {
      type = "BLOCK"
    }

    priority = 1
    rule_id  = aws_wafregional_rule.test.id
  }
}
`, name)
}

func testAccWebACLConfig_changeDefaultAction(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_rule" "test" {
  name        = %[1]q
  metric_name = %[1]q
}

resource "aws_wafregional_web_acl" "test" {
  name        = %[1]q
  metric_name = %[1]q

  default_action {
    type = "BLOCK"
  }

  rule {
    action {
      type = "BLOCK"
    }

    priority = 1
    rule_id  = aws_wafregional_rule.test.id
  }
}
`, name)
}

func testAccRuleConfig_webACLNos(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_web_acl" "test" {
  name        = %[1]q
  metric_name = %[1]q

  default_action {
    type = "ALLOW"
  }
}
`, name)
}

func testAccWebACLConfig_changeRules(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_rule" "test" {
  name        = %[1]q
  metric_name = %[1]q
}

resource "aws_wafregional_web_acl" "test" {
  name        = %[1]q
  metric_name = %[1]q

  default_action {
    type = "ALLOW"
  }

  rule {
    action {
      type = "ALLOW"
    }

    priority = 3
    rule_id  = aws_wafregional_rule.test.id
  }

  rule {
    action {
      type = "BLOCK"
    }

    priority = 99
    rule_id  = aws_wafregional_rule.test.id
  }
}
`, name)
}

func testAccWebACLConfig_loggingConfiguration(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_web_acl" "test" {
  name        = %[1]q
  metric_name = %[1]q

  default_action {
    type = "ALLOW"
  }

  logging_configuration {
    log_destination = aws_kinesis_firehose_delivery_stream.test.arn

    redacted_fields {
      field_to_match {
        type = "URI"
      }

      field_to_match {
        data = "referer"
        type = "HEADER"
      }
    }
  }
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "firehose.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF

}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  # the name must begin with aws-waf-logs-
  name        = "aws-waf-logs-%[1]s"
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = aws_iam_role.test.arn
    bucket_arn = aws_s3_bucket.test.arn
  }
}
`, rName)
}

func testAccWebACLConfig_loggingConfigurationUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_web_acl" "test" {
  name        = %[1]q
  metric_name = %[1]q

  default_action {
    type = "ALLOW"
  }

  logging_configuration {
    log_destination = aws_kinesis_firehose_delivery_stream.test.arn
  }
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "firehose.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF

}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  # the name must begin with aws-waf-logs-
  name        = "aws-waf-logs-%[1]s"
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = aws_iam_role.test.arn
    bucket_arn = aws_s3_bucket.test.arn
  }
}
`, rName)
}
