// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafv2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfwafv2 "github.com/hashicorp/terraform-provider-aws/internal/service/wafv2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestParseWebACLARN(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		arn           string
		expectedID    string
		expectedName  string
		expectedScope string
		expectError   bool
	}{
		"valid regional ARN": {
			arn:           "arn:aws:wafv2:us-east-1:123456789012:regional/webacl/test-web-acl/12345678-1234-1234-1234-123456789012", //lintignore:AWSAT003,AWSAT005
			expectedID:    "12345678-1234-1234-1234-123456789012",
			expectedName:  "test-web-acl",
			expectedScope: "REGIONAL",
			expectError:   false,
		},
		"valid CloudFront ARN with global region": {
			arn:           "arn:aws:wafv2:global:123456789012:global/webacl/test-web-acl/12345678-1234-1234-1234-123456789012", //lintignore:AWSAT003,AWSAT005
			expectedID:    "12345678-1234-1234-1234-123456789012",
			expectedName:  "test-web-acl",
			expectedScope: "CLOUDFRONT",
			expectError:   false,
		},
		"valid CloudFront ARN with specific region": {
			arn:           "arn:aws:wafv2:us-east-1:123456789012:global/webacl/test-web-acl/12345678-1234-1234-1234-123456789012", //lintignore:AWSAT003,AWSAT005
			expectedID:    "12345678-1234-1234-1234-123456789012",
			expectedName:  "test-web-acl",
			expectedScope: "CLOUDFRONT",
			expectError:   false,
		},
		"web ACL name with hyphens": {
			arn:           "arn:aws:wafv2:us-west-2:123456789012:regional/webacl/my-test-web-acl-name/12345678-1234-1234-1234-123456789012", //lintignore:AWSAT003,AWSAT005
			expectedID:    "12345678-1234-1234-1234-123456789012",
			expectedName:  "my-test-web-acl-name",
			expectedScope: "REGIONAL",
			expectError:   false,
		},
		"web ACL name with underscores": {
			arn:           "arn:aws:wafv2:eu-west-1:123456789012:regional/webacl/my_test_web_acl_name/12345678-1234-1234-1234-123456789012", //lintignore:AWSAT003,AWSAT005
			expectedID:    "12345678-1234-1234-1234-123456789012",
			expectedName:  "my_test_web_acl_name",
			expectedScope: "REGIONAL",
			expectError:   false,
		},
		"invalid ARN - too few parts": {
			arn:         "arn:aws:wafv2:us-east-1:123456789012", //lintignore:AWSAT003,AWSAT005
			expectError: true,
		},
		"invalid ARN - empty": {
			arn:         "",
			expectError: true,
		},
		"invalid ARN - not an ARN": {
			arn:         "not-an-arn",
			expectError: true,
		},
		"invalid resource format - too few parts": {
			arn:         "arn:aws:wafv2:us-east-1:123456789012:regional/webacl/test-web-acl", //lintignore:AWSAT003,AWSAT005
			expectError: true,
		},
		"invalid resource format - wrong resource type": {
			arn:         "arn:aws:wafv2:us-east-1:123456789012:regional/rulegroup/test-rule-group/12345678-1234-1234-1234-123456789012", //lintignore:AWSAT003,AWSAT005
			expectError: true,
		},
		"different AWS partition": {
			arn:           "arn:aws-us-gov:wafv2:us-gov-east-1:123456789012:regional/webacl/test-web-acl/12345678-1234-1234-1234-123456789012", //lintignore:AWSAT003,AWSAT005
			expectedID:    "12345678-1234-1234-1234-123456789012",
			expectedName:  "test-web-acl",
			expectedScope: "REGIONAL",
			expectError:   false,
		},
		"different AWS partition with CloudFront": {
			arn:           "arn:aws-cn:wafv2:global:123456789012:global/webacl/test-web-acl/12345678-1234-1234-1234-123456789012", //lintignore:AWSAT003,AWSAT005
			expectedID:    "12345678-1234-1234-1234-123456789012",
			expectedName:  "test-web-acl",
			expectedScope: "CLOUDFRONT",
			expectError:   false,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			id, name, scope, err := tfwafv2.ParseWebACLARN(testCase.arn)

			if testCase.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if id != testCase.expectedID {
				t.Errorf("expected ID %q, got %q", testCase.expectedID, id)
			}

			if name != testCase.expectedName {
				t.Errorf("expected name %q, got %q", testCase.expectedName, name)
			}

			if scope != testCase.expectedScope {
				t.Errorf("expected scope %q, got %q", testCase.expectedScope, scope)
			}
		})
	}
}

func TestAccWAFV2WebACLRuleGroupAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v wafv2.GetWebACLOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule_group_association.test"
	webACLResourceName := "aws_wafv2_web_acl.test"
	ruleGroupResourceName := "aws_wafv2_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleGroupAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleGroupAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleGroupAssociationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "rule_name", fmt.Sprintf("%s-association", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrPriority, "10"),
					resource.TestCheckResourceAttr(resourceName, "override_action", "none"),
					resource.TestCheckResourceAttrPair(resourceName, "web_acl_arn", webACLResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "rule_group_arn", ruleGroupResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: testAccWebACLRuleGroupAssociationImportStateIDFunc(resourceName),
			},
		},
	})
}

func TestAccWAFV2WebACLRuleGroupAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v wafv2.GetWebACLOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule_group_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleGroupAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleGroupAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleGroupAssociationExists(ctx, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfwafv2.ResourceWebACLRuleGroupAssociation, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWAFV2WebACLRuleGroupAssociation_overrideAction(t *testing.T) {
	ctx := acctest.Context(t)
	var v wafv2.GetWebACLOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule_group_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleGroupAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleGroupAssociationConfig_overrideAction(rName, "count"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleGroupAssociationExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "override_action", "count"),
				),
			},
		},
	})
}

func testAccCheckWebACLRuleGroupAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_wafv2_web_acl_rule_group_association" {
				continue
			}

			// Parse the ID using the standard flex utility
			// Format: webACLARN,ruleName,ruleGroupARN
			parts, err := flex.ExpandResourceId(rs.Primary.ID, 3, false)
			if err != nil {
				continue
			}

			webACLARN := parts[0]
			ruleName := parts[1]
			ruleGroupARN := parts[2]

			// Parse Web ACL ARN to get ID, name, and scope
			webACLID, webACLName, webACLScope, err := tfwafv2.ParseWebACLARN(webACLARN)
			if err != nil {
				continue
			}

			// Get the Web ACL
			webACL, err := tfwafv2.FindWebACLByThreePartKey(ctx, conn, webACLID, webACLName, webACLScope)
			if tfresource.NotFound(err) {
				// Web ACL is gone, so the association is definitely destroyed
				continue
			}
			if err != nil {
				return fmt.Errorf("error reading Web ACL (%s): %w", webACLARN, err)
			}

			// Check if the rule still exists in the Web ACL
			for _, rule := range webACL.WebACL.Rules {
				if aws.ToString(rule.Name) == ruleName &&
					rule.Statement != nil &&
					rule.Statement.RuleGroupReferenceStatement != nil &&
					aws.ToString(rule.Statement.RuleGroupReferenceStatement.ARN) == ruleGroupARN {
					return fmt.Errorf("WAFv2 Web ACL Rule Group Association still exists: %s", rs.Primary.ID)
				}
			}
		}

		return nil
	}
}

func testAccCheckWebACLRuleGroupAssociationExists(ctx context.Context, n string, v *wafv2.GetWebACLOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No WAFv2 WebACLRuleGroupAssociation ID is set")
		}

		// Parse the ID using the standard flex utility
		// Format: webACLARN,ruleName,ruleGroupARN
		parts, err := flex.ExpandResourceId(rs.Primary.ID, 3, false)
		if err != nil {
			return fmt.Errorf("Invalid ID format: %s", rs.Primary.ID)
		}

		webACLARN := parts[0]
		ruleName := parts[1]
		ruleGroupARN := parts[2]

		// Parse Web ACL ARN to get ID, name, and scope
		webACLID, webACLName, webACLScope, err := tfwafv2.ParseWebACLARN(webACLARN)
		if err != nil {
			return fmt.Errorf("error parsing Web ACL ARN: %w", err)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WAFV2Client(ctx)

		// Get the Web ACL
		webACL, err := tfwafv2.FindWebACLByThreePartKey(ctx, conn, webACLID, webACLName, webACLScope)
		if err != nil {
			return fmt.Errorf("error reading Web ACL (%s): %w", webACLARN, err)
		}

		// Check if the rule exists in the Web ACL with the correct configuration
		found := false
		for _, rule := range webACL.WebACL.Rules {
			if aws.ToString(rule.Name) == ruleName &&
				rule.Statement != nil &&
				rule.Statement.RuleGroupReferenceStatement != nil &&
				aws.ToString(rule.Statement.RuleGroupReferenceStatement.ARN) == ruleGroupARN {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("WAFv2 Web ACL Rule Group Association not found in Web ACL: %s", rs.Primary.ID)
		}

		*v = *webACL

		return nil
	}
}

func testAccWebACLRuleGroupAssociationImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		webACLARN := rs.Primary.Attributes["web_acl_arn"]
		ruleGroupARN := rs.Primary.Attributes["rule_group_arn"]
		ruleName := rs.Primary.Attributes["rule_name"]

		return fmt.Sprintf("%s,%s,%s", webACLARN, ruleGroupARN, ruleName), nil
	}
}

func testAccWebACLRuleGroupAssociationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  name     = %[1]q
  scope    = "REGIONAL"
  capacity = 10

  rule {
    name     = "rule-1"
    priority = 1

    action {
      count {}
    }

    statement {
      geo_match_statement {
        country_codes = ["US", "CA"]
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "rule-1"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }
}

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

resource "aws_wafv2_web_acl_rule_group_association" "test" {
  rule_name      = "%[1]s-association"
  priority       = 10
  rule_group_arn = aws_wafv2_rule_group.test.arn
  web_acl_arn    = aws_wafv2_web_acl.test.arn
}
`, rName)
}

func testAccWebACLRuleGroupAssociationConfig_overrideAction(rName, overrideAction string) string {
	return fmt.Sprintf(`
resource "aws_wafv2_rule_group" "test" {
  name     = %[1]q
  scope    = "REGIONAL"
  capacity = 10

  rule {
    name     = "rule-1"
    priority = 1

    action {
      block {}
    }

    statement {
      geo_match_statement {
        country_codes = ["US", "CA"]
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "rule-1"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }
}

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

resource "aws_wafv2_web_acl_rule_group_association" "test" {
  rule_name       = "%[1]s-association"
  priority        = 10
  rule_group_arn  = aws_wafv2_rule_group.test.arn
  web_acl_arn     = aws_wafv2_web_acl.test.arn
  override_action = %[2]q
}
`, rName, overrideAction)
}
