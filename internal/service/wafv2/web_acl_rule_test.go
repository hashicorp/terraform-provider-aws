// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafv2_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfwafv2 "github.com/hashicorp/terraform-provider-aws/internal/service/wafv2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// TestParseWebACLRuleID covers parsing of composite IDs for WebACL rules
func TestParseWebACLRuleID(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		id                 string
		expectedWebACLID   string
		expectedWebACLName string
		expectedScope      string
		expectedRuleName   string
		expectError        bool
	}{
		"valid composite ID": {
			id:                 "webaclid|webaclname|REGIONAL|rulename",
			expectedWebACLID:   "webaclid",
			expectedWebACLName: "webaclname",
			expectedScope:      "REGIONAL",
			expectedRuleName:   "rulename",
			expectError:        false,
		},
		"invalid - too few parts": {
			id:          "webaclid|webaclname|REGIONAL",
			expectError: true,
		},
		"invalid - empty": {
			id:          "",
			expectError: true,
		},
		"invalid - not delimited": {
			id:          "not-a-composite-id",
			expectError: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			webACLID, webACLName, scope, ruleName, err := tfwafv2.ParseWebACLRuleID(tc.id)
			if tc.expectError {
				if err == nil {
					t.Fatalf("expected error, got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if webACLID != tc.expectedWebACLID || webACLName != tc.expectedWebACLName || scope != tc.expectedScope || ruleName != tc.expectedRuleName {
				t.Errorf("unexpected parse result: got %q %q %q %q", webACLID, webACLName, scope, ruleName)
			}
		})
	}
}

func TestAccWAFV2WebACLRule_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var webacl awstypes.WebACL
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule.test"
	webACLResourceName := "aws_wafv2_web_acl.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleExists(ctx, resourceName, &webacl),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrPriority, "1"),
					resource.TestCheckResourceAttr(resourceName, "web_acl_scope", "REGIONAL"),
					resource.TestCheckResourceAttrPair(resourceName, "web_acl_id", webACLResourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName, "web_acl_name", webACLResourceName, names.AttrName),
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

func TestAccWAFV2WebACLRule_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var webacl awstypes.WebACL
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleExists(ctx, resourceName, &webacl),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfwafv2.ResourceWebACLRule, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWAFV2WebACLRule_priorityUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleConfig_basic(rName),
				Check:  resource.TestCheckResourceAttr(resourceName, names.AttrPriority, "1"),
			},
			{
				Config: testAccWebACLRuleConfig_priority(rName, 5),
				Check:  resource.TestCheckResourceAttr(resourceName, names.AttrPriority, "5"),
			},
		},
	})
}

func TestAccWAFV2WebACLRule_actionUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleConfig_basic(rName),
				Check:  resource.TestCheckResourceAttr(resourceName, "action.0.allow", "{}"),
			},
			{
				Config: testAccWebACLRuleConfig_blockAction(rName),
				Check:  resource.TestCheckResourceAttr(resourceName, "action.0.block", "{}"),
			},
		},
	})
}

func TestAccWAFV2WebACLRule_ruleNameRequiresReplace(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleConfig_basic(rName),
				Check:  resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
			},
			{
				Config:             testAccWebACLRuleConfig_basic(rName + "-updated"),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWAFV2WebACLRule_importState(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleConfig_basic(rName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccWAFV2WebACLRule_byteMatchStatement(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleConfig_byteMatchStatement(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "statement.0.byte_match_statement.0.search_string", "test"),
				),
			},
		},
	})
}

func testAccCheckWebACLRuleDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_wafv2_web_acl_rule" {
				continue
			}

			// Parse the composite ID to get WebACL details and rule name
			webACLID, webACLName, scope, ruleName, err := tfwafv2.ParseWebACLRuleID(rs.Primary.ID)
			if err != nil {
				return err
			}

			// Get the WebACL and check if the rule still exists in it
			webACL, err := tfwafv2.FindWebACLByThreePartKey(ctx, conn, webACLID, webACLName, scope)
			if tfresource.NotFound(err) {
				// If WebACL is gone, rule is gone too
				continue
			}
			if err != nil {
				return fmt.Errorf("error checking WebACL: %w", err)
			}

			// Check if the rule still exists in the WebACL
			for _, rule := range webACL.WebACL.Rules {
				if rule.Name != nil && *rule.Name == ruleName {
					return fmt.Errorf("WAFv2 WebACL Rule %s still exists", rs.Primary.ID)
				}
			}
		}

		return nil
	}
}

func testAccCheckWebACLRuleExists(ctx context.Context, name string, webacl *awstypes.WebACL) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no WAFv2 WebACL Rule ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFV2Client(ctx)

		// Parse the composite ID to get WebACL details and rule name
		webACLID, webACLName, scope, ruleName, err := tfwafv2.ParseWebACLRuleID(rs.Primary.ID)
		if err != nil {
			return err
		}

		// Get the WebACL and check if the rule exists in it
		resp, err := tfwafv2.FindWebACLByThreePartKey(ctx, conn, webACLID, webACLName, scope)
		if err != nil {
			return fmt.Errorf("error finding WebACL: %w", err)
		}

		// Check if the rule exists in the WebACL
		for _, rule := range resp.WebACL.Rules {
			if rule.Name != nil && *rule.Name == ruleName {
				*webacl = *resp.WebACL
				return nil
			}
		}

		return fmt.Errorf("WAFv2 WebACL Rule %s not found in WebACL", ruleName)
	}
}

func testAccWebACLRuleConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name  = %[1]q
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }

  lifecycle {
    ignore_changes = [rule]
  }
}

resource "aws_wafv2_web_acl_rule" "test" {
  name          = %[1]q
  priority      = 1
  web_acl_id    = aws_wafv2_web_acl.test.id
  web_acl_name  = aws_wafv2_web_acl.test.name
  web_acl_scope = "REGIONAL"

  action {
    allow {}
  }

  statement {
    geo_match_statement {
      country_codes = ["US", "NL"]
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = true
    metric_name                = "%[1]s-rule"
    sampled_requests_enabled   = true
  }
}
`, rName)
}

func testAccWebACLRuleConfig_priority(rName string, priority int) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name  = %[1]q
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }

  lifecycle {
    ignore_changes = [rule]
  }
}

resource "aws_wafv2_web_acl_rule" "test" {
  name          = %[1]q
  priority      = %[2]d
  web_acl_id    = aws_wafv2_web_acl.test.id
  web_acl_name  = aws_wafv2_web_acl.test.name
  web_acl_scope = "REGIONAL"

  action {
    allow {}
  }

  statement {
    geo_match_statement {
      country_codes = ["US", "NL"]
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = true
    metric_name                = "%[1]s-rule"
    sampled_requests_enabled   = true
  }
}
`, rName, priority)
}

func testAccWebACLRuleConfig_blockAction(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name  = %[1]q
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }

  lifecycle {
    ignore_changes = [rule]
  }
}

resource "aws_wafv2_web_acl_rule" "test" {
  name          = %[1]q
  priority      = 1
  web_acl_id    = aws_wafv2_web_acl.test.id
  web_acl_name  = aws_wafv2_web_acl.test.name
  web_acl_scope = "REGIONAL"

  action {
    block {}
  }

  statement {
    geo_match_statement {
      country_codes = ["US", "NL"]
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = true
    metric_name                = "%[1]s-rule"
    sampled_requests_enabled   = true
  }
}
`, rName)
}

func testAccWebACLRuleConfig_byteMatchStatement(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_web_acl" "test" {
  name  = %[1]q
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }

  lifecycle {
    ignore_changes = [rule]
  }
}

resource "aws_wafv2_web_acl_rule" "test" {
  name          = %[1]q
  priority      = 1
  web_acl_id    = aws_wafv2_web_acl.test.id
  web_acl_name  = aws_wafv2_web_acl.test.name
  web_acl_scope = "REGIONAL"

  action {
    count {}
  }

  statement {
    byte_match_statement {
      search_string         = "test"
      positional_constraint = "CONTAINS"

      field_to_match {
        uri_path {}
      }

      text_transformation {
        priority = 0
        type     = "NONE"
      }
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = true
    metric_name                = "%[1]s-rule"
    sampled_requests_enabled   = true
  }
}
`, rName)
}
