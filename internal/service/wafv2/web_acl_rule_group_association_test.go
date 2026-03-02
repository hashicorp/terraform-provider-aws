// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package wafv2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfwafv2 "github.com/hashicorp/terraform-provider-aws/internal/service/wafv2"
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule_group_association.test"
	webACLResourceName := "aws_wafv2_web_acl.test"
	ruleGroupResourceName := "aws_wafv2_rule_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleGroupAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleGroupAssociationConfig_RuleGroupReference_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleGroupAssociationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "rule_name", fmt.Sprintf("%s-association", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrPriority, "10"),
					resource.TestCheckResourceAttr(resourceName, "override_action", "none"),
					resource.TestCheckResourceAttrPair(resourceName, "web_acl_arn", webACLResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "rule_group_reference.0.arn", ruleGroupResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccWebACLRuleGroupAssociationImportStateIDFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: "web_acl_arn",
			},
		},
	})
}

func TestAccWAFV2WebACLRuleGroupAssociation_withVisibilityConfig(t *testing.T) {
	ctx := acctest.Context(t)
	var v wafv2.GetWebACLOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule_group_association.test"
	webACLResourceName := "aws_wafv2_web_acl.test"
	ruleGroupResourceName := "aws_wafv2_rule_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleGroupAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleGroupAssociationConfig_RuleGroupReference_withVisibilityConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleGroupAssociationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "rule_name", fmt.Sprintf("%s-association", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrPriority, "10"),
					resource.TestCheckResourceAttr(resourceName, "override_action", "none"),
					resource.TestCheckResourceAttrPair(resourceName, "web_acl_arn", webACLResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "rule_group_reference.0.arn", ruleGroupResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.cloudwatch_metrics_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.metric_name", "friendly-metric-name"),
					resource.TestCheckResourceAttr(resourceName, "visibility_config.0.sampled_requests_enabled", acctest.CtFalse),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccWebACLRuleGroupAssociationImportStateIDFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: "web_acl_arn",
			},
		},
	})
}

func TestAccWAFV2WebACLRuleGroupAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v wafv2.GetWebACLOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule_group_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleGroupAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleGroupAssociationConfig_RuleGroupReference_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleGroupAssociationExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfwafv2.ResourceWebACLRuleGroupAssociation, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccWAFV2WebACLRuleGroupAssociation_RuleGroupReference_overrideAction(t *testing.T) {
	ctx := acctest.Context(t)
	var v wafv2.GetWebACLOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule_group_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleGroupAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleGroupAssociationConfig_RuleGroupReference_overrideAction(rName, "count"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleGroupAssociationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "override_action", "count"),
				),
			},
		},
	})
}

func TestAccWAFV2WebACLRuleGroupAssociation_RuleGroupReference_ruleActionOverride(t *testing.T) {
	ctx := acctest.Context(t)
	var v wafv2.GetWebACLOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule_group_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleGroupAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleGroupAssociationConfig_RuleGroupReference_ruleActionOverride(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleGroupAssociationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "rule_group_reference.0.rule_action_override.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "rule_group_reference.0.rule_action_override.0.name", "rule-1"),
					resource.TestCheckResourceAttr(resourceName, "rule_group_reference.0.rule_action_override.0.action_to_use.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_group_reference.0.rule_action_override.0.action_to_use.0.allow.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_group_reference.0.rule_action_override.0.action_to_use.0.allow.0.custom_request_handling.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_group_reference.0.rule_action_override.0.action_to_use.0.allow.0.custom_request_handling.0.insert_header.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_group_reference.0.rule_action_override.0.action_to_use.0.allow.0.custom_request_handling.0.insert_header.0.name", "X-Custom-Header"),
					resource.TestCheckResourceAttr(resourceName, "rule_group_reference.0.rule_action_override.0.action_to_use.0.allow.0.custom_request_handling.0.insert_header.0.value", "custom-value"),
					resource.TestCheckResourceAttr(resourceName, "rule_group_reference.0.rule_action_override.1.name", "rule-2"),
					resource.TestCheckResourceAttr(resourceName, "rule_group_reference.0.rule_action_override.1.action_to_use.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_group_reference.0.rule_action_override.1.action_to_use.0.block.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_group_reference.0.rule_action_override.1.action_to_use.0.block.0.custom_response.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_group_reference.0.rule_action_override.1.action_to_use.0.block.0.custom_response.0.response_code", "403"),
					resource.TestCheckResourceAttr(resourceName, "rule_group_reference.0.rule_action_override.1.action_to_use.0.block.0.custom_response.0.response_header.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_group_reference.0.rule_action_override.1.action_to_use.0.block.0.custom_response.0.response_header.0.name", "X-Block-Reason"),
					resource.TestCheckResourceAttr(resourceName, "rule_group_reference.0.rule_action_override.1.action_to_use.0.block.0.custom_response.0.response_header.0.value", "rule-override"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccWebACLRuleGroupAssociationImportStateIDFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: "web_acl_arn",
			},
		},
	})
}

func TestAccWAFV2WebACLRuleGroupAssociation_RuleGroupReference_ruleActionOverrideUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var v wafv2.GetWebACLOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule_group_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleGroupAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleGroupAssociationConfig_RuleGroupReference_ruleActionOverrideCount(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleGroupAssociationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "rule_group_reference.0.rule_action_override.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_group_reference.0.rule_action_override.0.name", "rule-1"),
					resource.TestCheckResourceAttr(resourceName, "rule_group_reference.0.rule_action_override.0.action_to_use.0.count.#", "1"),
				),
			},
			{
				Config: testAccWebACLRuleGroupAssociationConfig_RuleGroupReference_ruleActionOverrideCaptcha(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleGroupAssociationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "rule_group_reference.0.rule_action_override.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_group_reference.0.rule_action_override.0.name", "rule-1"),
					resource.TestCheckResourceAttr(resourceName, "rule_group_reference.0.rule_action_override.0.action_to_use.0.captcha.#", "1"),
				),
			},
		},
	})
}

func TestAccWAFV2WebACLRuleGroupAssociation_RuleGroupReference_priorityUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var v wafv2.GetWebACLOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule_group_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleGroupAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleGroupAssociationConfig_RuleGroupReference_priority(rName, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleGroupAssociationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrPriority, "10"),
				),
			},
			{
				Config: testAccWebACLRuleGroupAssociationConfig_RuleGroupReference_priority(rName, 20),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleGroupAssociationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrPriority, "20"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccWebACLRuleGroupAssociationImportStateIDFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: "web_acl_arn",
			},
		},
	})
}

func TestAccWAFV2WebACLRuleGroupAssociation_RuleGroupReference_overrideActionUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var v wafv2.GetWebACLOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule_group_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleGroupAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleGroupAssociationConfig_RuleGroupReference_overrideAction(rName, "none"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleGroupAssociationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "override_action", "none"),
				),
			},
			{
				Config: testAccWebACLRuleGroupAssociationConfig_RuleGroupReference_overrideAction(rName, "count"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleGroupAssociationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "override_action", "count"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccWebACLRuleGroupAssociationImportStateIDFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: "web_acl_arn",
			},
		},
	})
}

func TestAccWAFV2WebACLRuleGroupAssociation_RuleGroupReference_ruleNameRequiresReplace(t *testing.T) {
	ctx := acctest.Context(t)
	var v wafv2.GetWebACLOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule_group_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleGroupAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleGroupAssociationConfig_RuleGroupReference_ruleName(rName, "original-rule"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleGroupAssociationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "rule_name", "original-rule"),
				),
			},
			{
				Config: testAccWebACLRuleGroupAssociationConfig_RuleGroupReference_ruleName(rName, "updated-rule"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleGroupAssociationExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "rule_name", "updated-rule"),
				),
			},
		},
	})
}

func TestAccWAFV2WebACLRuleGroupAssociation_RuleGroupReference_webACLARNRequiresReplace(t *testing.T) {
	ctx := acctest.Context(t)
	var v wafv2.GetWebACLOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule_group_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleGroupAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleGroupAssociationConfig_RuleGroupReference_webACL(rName, "first"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleGroupAssociationExists(ctx, t, resourceName, &v),
				),
			},
			{
				Config: testAccWebACLRuleGroupAssociationConfig_RuleGroupReference_webACL(rName, "second"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleGroupAssociationExists(ctx, t, resourceName, &v),
				),
			},
		},
	})
}

func TestAccWAFV2WebACLRuleGroupAssociation_ManagedRuleGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var webACL wafv2.GetWebACLOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule_group_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleGroupAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleGroupAssociationConfig_ManagedRuleGroup_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleGroupAssociationExists(ctx, t, resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "rule_name", "test-rule"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPriority, "1"),
					resource.TestCheckResourceAttr(resourceName, "override_action", "none"),
					resource.TestCheckResourceAttr(resourceName, "managed_rule_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "managed_rule_group.0.name", "AWSManagedRulesCommonRuleSet"),
					resource.TestCheckResourceAttr(resourceName, "managed_rule_group.0.vendor_name", "AWS"),
					resource.TestCheckResourceAttr(resourceName, "rule_group_reference.#", "0"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccWebACLRuleGroupAssociationManagedRuleGroupImportStateIDFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: "web_acl_arn",
			},
		},
	})
}

func TestAccWAFV2WebACLRuleGroupAssociation_ManagedRuleGroup_withVersion(t *testing.T) {
	ctx := acctest.Context(t)
	var webACL wafv2.GetWebACLOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule_group_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleGroupAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleGroupAssociationConfig_ManagedRuleGroup_withVersion(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleGroupAssociationExists(ctx, t, resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "rule_name", "test-rule"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPriority, "1"),
					resource.TestCheckResourceAttr(resourceName, "override_action", "none"),
					resource.TestCheckResourceAttr(resourceName, "managed_rule_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "managed_rule_group.0.name", "AWSManagedRulesCommonRuleSet"),
					resource.TestCheckResourceAttr(resourceName, "managed_rule_group.0.vendor_name", "AWS"),
					resource.TestCheckResourceAttr(resourceName, "managed_rule_group.0.version", "Version_1.0"),
					resource.TestCheckResourceAttr(resourceName, "rule_group_reference.#", "0"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccWebACLRuleGroupAssociationManagedRuleGroupImportStateIDFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: "web_acl_arn",
			},
		},
	})
}

func TestAccWAFV2WebACLRuleGroupAssociation_ManagedRuleGroup_ruleActionOverride(t *testing.T) {
	ctx := acctest.Context(t)
	var webACL wafv2.GetWebACLOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule_group_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleGroupAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleGroupAssociationConfig_ManagedRuleGroup_ruleActionOverride(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleGroupAssociationExists(ctx, t, resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "rule_name", "test-rule"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPriority, "1"),
					resource.TestCheckResourceAttr(resourceName, "override_action", "none"),
					resource.TestCheckResourceAttr(resourceName, "managed_rule_group.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "managed_rule_group.0.name", "AWSManagedRulesCommonRuleSet"),
					resource.TestCheckResourceAttr(resourceName, "managed_rule_group.0.vendor_name", "AWS"),
					resource.TestCheckResourceAttr(resourceName, "managed_rule_group.0.rule_action_override.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "managed_rule_group.0.rule_action_override.0.name", "GenericRFI_BODY"),
					resource.TestCheckResourceAttr(resourceName, "managed_rule_group.0.rule_action_override.0.action_to_use.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "managed_rule_group.0.rule_action_override.0.action_to_use.0.count.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_group_reference.#", "0"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccWebACLRuleGroupAssociationManagedRuleGroupImportStateIDFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: "web_acl_arn",
			},
		},
	})
}

func TestAccWAFV2WebACLRuleGroupAssociation_ManagedRuleGroup_ManagedRuleGroupConfig_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var webACL wafv2.GetWebACLOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule_group_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleGroupAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleGroupAssociationConfig_ManagedRuleGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleGroupAssociationExists(ctx, t, resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "managed_rule_group.0.managed_rule_group_configs.0.aws_managed_rules_bot_control_rule_set.0.inspection_level", "COMMON"),
				),
			},
			{
				Config: testAccWebACLRuleGroupAssociationConfig_ManagedRuleGroupConfigUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckWebACLRuleGroupAssociationExists(ctx, t, resourceName, &webACL),
					resource.TestCheckResourceAttr(resourceName, "managed_rule_group.0.managed_rule_group_configs.0.aws_managed_rules_bot_control_rule_set.0.inspection_level", "TARGETED"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccWebACLRuleGroupAssociationManagedRuleGroupImportStateIDFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: "web_acl_arn",
			},
		},
	})
}

func TestAccWAFV2WebACLRuleGroupAssociation_ManagedRuleGroup_ManagedRuleGroupConfig_ACFPRuleSet(t *testing.T) {
	ctx := acctest.Context(t)
	var webACL wafv2.GetWebACLOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule_group_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleGroupAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleGroupAssociationConfig_ManagedRuleGroupConfig_acfpRuleSet(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebACLRuleGroupAssociationExists(ctx, t, resourceName, &webACL),
					// Check top-level attributes
					resource.TestCheckResourceAttr(resourceName, "rule_name", "test-rule"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPriority, "1"),
					resource.TestCheckResourceAttr(resourceName, "managed_rule_group.0.name", "AWSManagedRulesACFPRuleSet"),
					resource.TestCheckResourceAttr(resourceName, "managed_rule_group.0.vendor_name", "AWS"),

					// Verify Managed Rule Group Config
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "managed_rule_group.0.managed_rule_group_configs.*", map[string]string{
						"aws_managed_rules_acfp_rule_set.0.creation_path":                                            "/creation",
						"aws_managed_rules_acfp_rule_set.0.registration_page_path":                                   "/registration",
						"aws_managed_rules_acfp_rule_set.0.request_inspection.0.email_field.0.identifier":            "/email",
						"aws_managed_rules_acfp_rule_set.0.request_inspection.0.password_field.0.identifier":         "/password",
						"aws_managed_rules_acfp_rule_set.0.request_inspection.0.phone_number_fields.0.identifiers.0": "/phone1",
						"aws_managed_rules_acfp_rule_set.0.request_inspection.0.phone_number_fields.0.identifiers.1": "/phone2",
						"aws_managed_rules_acfp_rule_set.0.request_inspection.0.address_fields.0.identifiers.0":      "home",
						"aws_managed_rules_acfp_rule_set.0.request_inspection.0.address_fields.0.identifiers.1":      "work",
						"aws_managed_rules_acfp_rule_set.0.request_inspection.0.payload_type":                        "JSON",
						"aws_managed_rules_acfp_rule_set.0.request_inspection.0.username_field.0.identifier":         "/username",
					}),
				),
			},
			{
				Config: testAccWebACLRuleGroupAssociationConfig_ManagedRuleGroupConfig_acfpRuleSetUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebACLRuleGroupAssociationExists(ctx, t, resourceName, &webACL),
					// Check top-level attributes
					resource.TestCheckResourceAttr(resourceName, "rule_name", "test-rule"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPriority, "1"),
					resource.TestCheckResourceAttr(resourceName, "managed_rule_group.0.name", "AWSManagedRulesACFPRuleSet"),
					resource.TestCheckResourceAttr(resourceName, "managed_rule_group.0.vendor_name", "AWS"),

					// Verify Managed Rule Group Config
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "managed_rule_group.0.managed_rule_group_configs.*", map[string]string{
						"aws_managed_rules_acfp_rule_set.0.enable_regex_in_path":                                     acctest.CtTrue,
						"aws_managed_rules_acfp_rule_set.0.creation_path":                                            "/creation",
						"aws_managed_rules_acfp_rule_set.0.registration_page_path":                                   "/registration",
						"aws_managed_rules_acfp_rule_set.0.request_inspection.0.email_field.0.identifier":            "/email",
						"aws_managed_rules_acfp_rule_set.0.request_inspection.0.password_field.0.identifier":         "/pass",
						"aws_managed_rules_acfp_rule_set.0.request_inspection.0.phone_number_fields.0.identifiers.0": "/phone3",
						"aws_managed_rules_acfp_rule_set.0.request_inspection.0.address_fields.0.identifiers.0":      "mobile",
						"aws_managed_rules_acfp_rule_set.0.request_inspection.0.payload_type":                        "JSON",
						"aws_managed_rules_acfp_rule_set.0.request_inspection.0.username_field.0.identifier":         "/user",
					}),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccWebACLRuleGroupAssociationManagedRuleGroupImportStateIDFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: "web_acl_arn",
			},
		},
	})
}

func TestAccWAFV2WebACLRuleGroupAssociation_ManagedRuleGroup_ManagedRuleGroupConfig_AntiDDoSRuleSet(t *testing.T) {
	ctx := acctest.Context(t)
	var webACL wafv2.GetWebACLOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule_group_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleGroupAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleGroupAssociationConfig_ManagedRuleGroupConfig_antiDDoSRuleSet(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebACLRuleGroupAssociationExists(ctx, t, resourceName, &webACL),
					// Check top-level attributes
					resource.TestCheckResourceAttr(resourceName, "rule_name", "test-rule"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPriority, "1"),
					resource.TestCheckResourceAttr(resourceName, "managed_rule_group.0.name", "AWSManagedRulesAntiDDoSRuleSet"),
					resource.TestCheckResourceAttr(resourceName, "managed_rule_group.0.vendor_name", "AWS"),

					// Verify Managed Rule Group Config
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "managed_rule_group.0.managed_rule_group_configs.*", map[string]string{
						"aws_managed_rules_anti_ddos_rule_set.0.client_side_action_config.0.challenge.0.usage_of_action":                              "ENABLED",
						"aws_managed_rules_anti_ddos_rule_set.0.client_side_action_config.0.challenge.0.exempt_uri_regular_expression.#":              "2",
						"aws_managed_rules_anti_ddos_rule_set.0.client_side_action_config.0.challenge.0.exempt_uri_regular_expression.0.regex_string": "\\/api\\/",
						"aws_managed_rules_anti_ddos_rule_set.0.client_side_action_config.0.challenge.0.exempt_uri_regular_expression.1.regex_string": "jpg",
						"aws_managed_rules_anti_ddos_rule_set.0.client_side_action_config.0.challenge.0.sensitivity":                                  "MEDIUM",
						"aws_managed_rules_anti_ddos_rule_set.0.sensitivity_to_block":                                                                 "HIGH",
					}),
				),
			},
			{
				Config: testAccWebACLRuleGroupAssociationConfig_ManagedRuleGroupConfig_antiDDoSRuleSetWithDefaultSensitivity(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebACLRuleGroupAssociationExists(ctx, t, resourceName, &webACL),
					// Check top-level attributes
					resource.TestCheckResourceAttr(resourceName, "rule_name", "test-rule"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPriority, "1"),
					resource.TestCheckResourceAttr(resourceName, "managed_rule_group.0.name", "AWSManagedRulesAntiDDoSRuleSet"),
					resource.TestCheckResourceAttr(resourceName, "managed_rule_group.0.vendor_name", "AWS"),

					// Verify Managed Rule Group Config
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "managed_rule_group.0.managed_rule_group_configs.*", map[string]string{
						"aws_managed_rules_anti_ddos_rule_set.0.client_side_action_config.0.challenge.0.usage_of_action":                              "ENABLED",
						"aws_managed_rules_anti_ddos_rule_set.0.client_side_action_config.0.challenge.0.exempt_uri_regular_expression.#":              "2",
						"aws_managed_rules_anti_ddos_rule_set.0.client_side_action_config.0.challenge.0.exempt_uri_regular_expression.0.regex_string": "\\/api\\/",
						"aws_managed_rules_anti_ddos_rule_set.0.client_side_action_config.0.challenge.0.exempt_uri_regular_expression.1.regex_string": "jpg",
						"aws_managed_rules_anti_ddos_rule_set.0.client_side_action_config.0.challenge.0.sensitivity":                                  "HIGH",
					}),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccWebACLRuleGroupAssociationManagedRuleGroupImportStateIDFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: "web_acl_arn",
			},
		},
	})
}

func TestAccWAFV2WebACLRuleGroupAssociation_ManagedRuleGroup_ManagedRuleGroupConfig_ATPRuleSet(t *testing.T) {
	ctx := acctest.Context(t)
	var webACL wafv2.GetWebACLOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule_group_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleGroupAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleGroupAssociationConfig_ManagedRuleGroupConfig_atpRuleSet(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebACLRuleGroupAssociationExists(ctx, t, resourceName, &webACL),
					// Check top-level attributes
					resource.TestCheckResourceAttr(resourceName, "rule_name", "test-rule"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPriority, "1"),
					resource.TestCheckResourceAttr(resourceName, "managed_rule_group.0.name", "AWSManagedRulesATPRuleSet"),
					resource.TestCheckResourceAttr(resourceName, "managed_rule_group.0.vendor_name", "AWS"),

					// Verify Managed Rule Group Config
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "managed_rule_group.0.managed_rule_group_configs.*", map[string]string{
						"aws_managed_rules_atp_rule_set.0.login_path":                                       "/api/1/signin",
						"aws_managed_rules_atp_rule_set.0.request_inspection.#":                             "1",
						"aws_managed_rules_atp_rule_set.0.request_inspection.0.password_field.#":            "1",
						"aws_managed_rules_atp_rule_set.0.request_inspection.0.password_field.0.identifier": "/password",
						"aws_managed_rules_atp_rule_set.0.request_inspection.0.username_field.#":            "1",
						"aws_managed_rules_atp_rule_set.0.request_inspection.0.username_field.0.identifier": "/username",
						"aws_managed_rules_atp_rule_set.0.request_inspection.0.payload_type":                "JSON",
						"aws_managed_rules_atp_rule_set.0.response_inspection.#":                            "0",
					}),
				),
			},
			{
				Config: testAccWebACLRuleGroupAssociationConfig_ManagedRuleGroupConfig_atpRuleSetUpdate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebACLRuleGroupAssociationExists(ctx, t, resourceName, &webACL),
					// Check top-level attributes
					resource.TestCheckResourceAttr(resourceName, "rule_name", "test-rule"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPriority, "1"),
					resource.TestCheckResourceAttr(resourceName, "managed_rule_group.0.name", "AWSManagedRulesATPRuleSet"),
					resource.TestCheckResourceAttr(resourceName, "managed_rule_group.0.vendor_name", "AWS"),

					// Verify Managed Rule Group Config
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "managed_rule_group.0.managed_rule_group_configs.*", map[string]string{
						"aws_managed_rules_atp_rule_set.0.enable_regex_in_path":                             acctest.CtTrue,
						"aws_managed_rules_atp_rule_set.0.login_path":                                       "/api/2/signin",
						"aws_managed_rules_atp_rule_set.0.request_inspection.#":                             "1",
						"aws_managed_rules_atp_rule_set.0.request_inspection.0.password_field.#":            "1",
						"aws_managed_rules_atp_rule_set.0.request_inspection.0.password_field.0.identifier": "/pass",
						"aws_managed_rules_atp_rule_set.0.request_inspection.0.username_field.#":            "1",
						"aws_managed_rules_atp_rule_set.0.request_inspection.0.username_field.0.identifier": "/user",
						"aws_managed_rules_atp_rule_set.0.request_inspection.0.payload_type":                "JSON",
						"aws_managed_rules_atp_rule_set.0.response_inspection.#":                            "0",
					}),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccWebACLRuleGroupAssociationManagedRuleGroupImportStateIDFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: "web_acl_arn",
			},
		},
	})
}

func TestAccWAFV2WebACLRuleGroupAssociation_ManagedRuleGroup_ManagedRuleGroupConfig_BotControl(t *testing.T) {
	ctx := acctest.Context(t)
	var webACL wafv2.GetWebACLOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_wafv2_web_acl_rule_group_association.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.WAFV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckWebACLRuleGroupAssociationDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccWebACLRuleGroupAssociationConfig_ManagedRuleGroupConfig_botControl(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckWebACLRuleGroupAssociationExists(ctx, t, resourceName, &webACL),
					// Check top-level attributes
					resource.TestCheckResourceAttr(resourceName, "rule_name", "test-rule"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPriority, "1"),
					resource.TestCheckResourceAttr(resourceName, "managed_rule_group.0.name", "AWSManagedRulesBotControlRuleSet"),
					resource.TestCheckResourceAttr(resourceName, "managed_rule_group.0.vendor_name", "AWS"),

					// Verify Managed Rule Group Config
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "managed_rule_group.0.managed_rule_group_configs.*", map[string]string{
						"aws_managed_rules_bot_control_rule_set.0.inspection_level":        "TARGETED",
						"aws_managed_rules_bot_control_rule_set.0.enable_machine_learning": acctest.CtTrue,
					}),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccWebACLRuleGroupAssociationManagedRuleGroupImportStateIDFunc(resourceName),
				ImportStateVerifyIdentifierAttribute: "web_acl_arn",
			},
		},
	})
}

func testAccCheckWebACLRuleGroupAssociationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).WAFV2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_wafv2_web_acl_rule_group_association" {
				continue
			}

			// Use resource attributes directly instead of parsing ID
			webACLARN := rs.Primary.Attributes["web_acl_arn"]
			ruleName := rs.Primary.Attributes["rule_name"]

			// Determine rule group type and identifier from attributes
			var ruleGroupType, ruleGroupIdentifier string
			if rs.Primary.Attributes["rule_group_reference.0.arn"] != "" {
				ruleGroupType = "custom"
				ruleGroupIdentifier = rs.Primary.Attributes["rule_group_reference.0.arn"]
			} else if rs.Primary.Attributes["managed_rule_group.0.name"] != "" {
				ruleGroupType = "managed"
				vendorName := rs.Primary.Attributes["managed_rule_group.0.vendor_name"]
				ruleGroupName := rs.Primary.Attributes["managed_rule_group.0.name"]
				version := rs.Primary.Attributes["managed_rule_group.0.version"]
				ruleGroupIdentifier = fmt.Sprintf("%s:%s", vendorName, ruleGroupName)
				if version != "" {
					ruleGroupIdentifier += ":" + version
				}
			} else {
				continue // Skip if no rule group configuration found
			}

			// Parse Web ACL ARN to get ID, name, and scope
			webACLID, webACLName, webACLScope, err := tfwafv2.ParseWebACLARN(webACLARN)
			if err != nil {
				continue
			}

			// Get the Web ACL
			webACL, err := tfwafv2.FindWebACLByThreePartKey(ctx, conn, webACLID, webACLName, webACLScope)
			if retry.NotFound(err) {
				// Web ACL is gone, so the association is definitely destroyed
				continue
			}
			if err != nil {
				return fmt.Errorf("error reading Web ACL (%s): %w", webACLARN, err)
			}

			// Check if the rule still exists in the Web ACL
			for _, rule := range webACL.WebACL.Rules {
				if aws.ToString(rule.Name) != ruleName || rule.Statement == nil {
					continue
				}

				// Check if this rule matches our rule group type and identifier
				var matchesRuleGroup bool
				if ruleGroupType == "custom" && rule.Statement.RuleGroupReferenceStatement != nil {
					// For custom rule groups, the identifier is the ARN
					if aws.ToString(rule.Statement.RuleGroupReferenceStatement.ARN) == ruleGroupIdentifier {
						matchesRuleGroup = true
					}
				} else if ruleGroupType == "managed" && rule.Statement.ManagedRuleGroupStatement != nil {
					// For managed rule groups, construct identifier and compare
					managedStmt := rule.Statement.ManagedRuleGroupStatement
					managedIdentifier := fmt.Sprintf("%s:%s", aws.ToString(managedStmt.VendorName), aws.ToString(managedStmt.Name))
					if managedStmt.Version != nil && aws.ToString(managedStmt.Version) != "" {
						managedIdentifier += ":" + aws.ToString(managedStmt.Version)
					}
					if managedIdentifier == ruleGroupIdentifier {
						matchesRuleGroup = true
					}
				}

				if matchesRuleGroup {
					return fmt.Errorf("WAFv2 Web ACL Rule Group Association still exists in Web ACL %s for rule %s", webACLARN, ruleName)
				}
			}
		}

		return nil
	}
}

func testAccCheckWebACLRuleGroupAssociationExists(ctx context.Context, t *testing.T, n string, v *wafv2.GetWebACLOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		// Use resource attributes directly instead of parsing ID
		webACLARN := rs.Primary.Attributes["web_acl_arn"]
		ruleName := rs.Primary.Attributes["rule_name"]

		if webACLARN == "" || ruleName == "" {
			return fmt.Errorf("Missing required attributes: web_acl_arn=%s, rule_name=%s", webACLARN, ruleName)
		}

		// Determine rule group type and identifier from attributes
		var ruleGroupType, ruleGroupIdentifier string
		if rs.Primary.Attributes["rule_group_reference.0.arn"] != "" {
			ruleGroupType = "custom"
			ruleGroupIdentifier = rs.Primary.Attributes["rule_group_reference.0.arn"]
		} else if rs.Primary.Attributes["managed_rule_group.0.name"] != "" {
			ruleGroupType = "managed"
			vendorName := rs.Primary.Attributes["managed_rule_group.0.vendor_name"]
			ruleGroupName := rs.Primary.Attributes["managed_rule_group.0.name"]
			version := rs.Primary.Attributes["managed_rule_group.0.version"]
			ruleGroupIdentifier = fmt.Sprintf("%s:%s", vendorName, ruleGroupName)
			if version != "" {
				ruleGroupIdentifier += ":" + version
			}
		} else {
			return fmt.Errorf("No rule group configuration found in state")
		}

		// Parse Web ACL ARN to get ID, name, and scope
		webACLID, webACLName, webACLScope, err := tfwafv2.ParseWebACLARN(webACLARN)
		if err != nil {
			return fmt.Errorf("error parsing Web ACL ARN: %w", err)
		}

		conn := acctest.ProviderMeta(ctx, t).WAFV2Client(ctx)

		// Get the Web ACL
		webACL, err := tfwafv2.FindWebACLByThreePartKey(ctx, conn, webACLID, webACLName, webACLScope)
		if err != nil {
			return fmt.Errorf("error reading Web ACL (%s): %w", webACLARN, err)
		}

		// Check if the rule exists in the Web ACL with the correct configuration
		found := false
		for _, rule := range webACL.WebACL.Rules {
			if aws.ToString(rule.Name) != ruleName || rule.Statement == nil {
				continue
			}

			// Check if this rule matches our rule group type and identifier
			var matchesRuleGroup bool
			if ruleGroupType == "custom" && rule.Statement.RuleGroupReferenceStatement != nil {
				// For custom rule groups, the identifier is the ARN
				if aws.ToString(rule.Statement.RuleGroupReferenceStatement.ARN) == ruleGroupIdentifier {
					matchesRuleGroup = true
				}
			} else if ruleGroupType == "managed" && rule.Statement.ManagedRuleGroupStatement != nil {
				// For managed rule groups, construct identifier and compare
				managedStmt := rule.Statement.ManagedRuleGroupStatement
				managedIdentifier := fmt.Sprintf("%s:%s", aws.ToString(managedStmt.VendorName), aws.ToString(managedStmt.Name))
				if managedStmt.Version != nil && aws.ToString(managedStmt.Version) != "" {
					managedIdentifier += ":" + aws.ToString(managedStmt.Version)
				}
				if managedIdentifier == ruleGroupIdentifier {
					matchesRuleGroup = true
				}
			}

			if matchesRuleGroup {
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("WAFv2 Web ACL Rule Group Association not found in Web ACL %s for rule %s", webACLARN, ruleName)
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
		ruleGroupARN := rs.Primary.Attributes["rule_group_reference.0.arn"]
		ruleName := rs.Primary.Attributes["rule_name"]

		// Format: webACLARN,ruleName,ruleGroupType,ruleGroupIdentifier
		return fmt.Sprintf("%s,%s,%s,%s", webACLARN, ruleName, "custom", ruleGroupARN), nil
	}
}

func testAccWebACLRuleGroupAssociationManagedRuleGroupImportStateIDFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		webACLARN := rs.Primary.Attributes["web_acl_arn"]
		vendorName := rs.Primary.Attributes["managed_rule_group.0.vendor_name"]
		ruleGroupName := rs.Primary.Attributes["managed_rule_group.0.name"]
		version := rs.Primary.Attributes["managed_rule_group.0.version"]
		ruleName := rs.Primary.Attributes["rule_name"]

		// Build managed rule group identifier: vendorName:ruleGroupName[:version]
		ruleGroupIdentifier := fmt.Sprintf("%s:%s", vendorName, ruleGroupName)
		if version != "" {
			ruleGroupIdentifier += ":" + version
		}

		// Format: webACLARN,ruleName,ruleGroupType,ruleGroupIdentifier
		return fmt.Sprintf("%s,%s,%s,%s", webACLARN, ruleName, "managed", ruleGroupIdentifier), nil
	}
}

func testAccWebACLRuleGroupAssociationConfig_RuleGroupReference_basic(rName string) string {
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
  rule_name   = "%[1]s-association"
  priority    = 10
  web_acl_arn = aws_wafv2_web_acl.test.arn

  rule_group_reference {
    arn = aws_wafv2_rule_group.test.arn
  }
}
`, rName)
}

func testAccWebACLRuleGroupAssociationConfig_RuleGroupReference_withVisibilityConfig(rName string) string {
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
  rule_name   = "%[1]s-association"
  priority    = 10
  web_acl_arn = aws_wafv2_web_acl.test.arn

  rule_group_reference {
    arn = aws_wafv2_rule_group.test.arn
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "friendly-metric-name"
    sampled_requests_enabled   = false
  }
}
`, rName)
}

func testAccWebACLRuleGroupAssociationConfig_RuleGroupReference_overrideAction(rName, overrideAction string) string {
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
  web_acl_arn     = aws_wafv2_web_acl.test.arn
  override_action = %[2]q

  rule_group_reference {
    arn = aws_wafv2_rule_group.test.arn
  }
}
`, rName, overrideAction)
}

func testAccWebACLRuleGroupAssociationConfig_RuleGroupReference_ruleActionOverride(rName string) string {
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

  rule {
    name     = "rule-2"
    priority = 2

    action {
      allow {}
    }

    statement {
      ip_set_reference_statement {
        arn = aws_wafv2_ip_set.test.arn
      }
    }

    visibility_config {
      cloudwatch_metrics_enabled = false
      metric_name                = "rule-2"
      sampled_requests_enabled   = false
    }
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = %[1]q
    sampled_requests_enabled   = false
  }
}

resource "aws_wafv2_ip_set" "test" {
  name  = %[1]q
  scope = "REGIONAL"

  ip_address_version = "IPV4"
  addresses          = ["192.0.2.0/24"]
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
  rule_name   = "%[1]s-association"
  priority    = 10
  web_acl_arn = aws_wafv2_web_acl.test.arn

  rule_group_reference {
    arn = aws_wafv2_rule_group.test.arn

    rule_action_override {
      name = "rule-1"
      action_to_use {
        allow {
          custom_request_handling {
            insert_header {
              name  = "X-Custom-Header"
              value = "custom-value"
            }
          }
        }
      }
    }

    rule_action_override {
      name = "rule-2"
      action_to_use {
        block {
          custom_response {
            response_code = 403
            response_header {
              name  = "X-Block-Reason"
              value = "rule-override"
            }
          }
        }
      }
    }
  }
}
`, rName)
}

func testAccWebACLRuleGroupAssociationConfig_RuleGroupReference_ruleActionOverrideCount(rName string) string {
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
  rule_name   = "%[1]s-association"
  priority    = 10
  web_acl_arn = aws_wafv2_web_acl.test.arn

  rule_group_reference {
    arn = aws_wafv2_rule_group.test.arn

    rule_action_override {
      name = "rule-1"
      action_to_use {
        count {
          custom_request_handling {
            insert_header {
              name  = "X-Count-Header"
              value = "counted"
            }
          }
        }
      }
    }
  }
}
`, rName)
}

func testAccWebACLRuleGroupAssociationConfig_RuleGroupReference_ruleActionOverrideCaptcha(rName string) string {
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
  rule_name = "%[1]s-association"
  priority  = 10
  rule_group_reference {
    arn = aws_wafv2_rule_group.test.arn

    rule_action_override {
      name = "rule-1"
      action_to_use {
        captcha {
          custom_request_handling {
            insert_header {
              name  = "X-Captcha-Header"
              value = "captcha-required"
            }
          }
        }
      }
    }
  }
  web_acl_arn = aws_wafv2_web_acl.test.arn
}
`, rName)
}

func testAccWebACLRuleGroupAssociationConfig_RuleGroupReference_priority(rName string, priority int) string {
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
        country_codes = ["US"]
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
  rule_name = "%[1]s-association"
  priority  = %[2]d
  rule_group_reference {
    arn = aws_wafv2_rule_group.test.arn
  }
  web_acl_arn     = aws_wafv2_web_acl.test.arn
  override_action = "none"
}
`, rName, priority)
}

func testAccWebACLRuleGroupAssociationConfig_RuleGroupReference_ruleName(rName, ruleName string) string {
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
        country_codes = ["US"]
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
  rule_name = %[2]q
  priority  = 10
  rule_group_reference {
    arn = aws_wafv2_rule_group.test.arn
  }
  web_acl_arn     = aws_wafv2_web_acl.test.arn
  override_action = "none"
}
`, rName, ruleName)
}

func testAccWebACLRuleGroupAssociationConfig_RuleGroupReference_webACL(rName, webACLSuffix string) string {
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
        country_codes = ["US"]
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
  name  = "%[1]s-%[2]s"
  scope = "REGIONAL"

  default_action {
    allow {}
  }

  visibility_config {
    cloudwatch_metrics_enabled = false
    metric_name                = "%[1]s-%[2]s"
    sampled_requests_enabled   = false
  }

  lifecycle {
    ignore_changes = [rule]
  }
}

resource "aws_wafv2_web_acl_rule_group_association" "test" {
  rule_name = "%[1]s-association"
  priority  = 10
  rule_group_reference {
    arn = aws_wafv2_rule_group.test.arn
  }
  web_acl_arn     = aws_wafv2_web_acl.test.arn
  override_action = "none"
}
`, rName, webACLSuffix)
}

func testAccWebACLRuleGroupAssociationConfig_ManagedRuleGroup_basic(rName string) string {
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

resource "aws_wafv2_web_acl_rule_group_association" "test" {
  rule_name   = "test-rule"
  priority    = 1
  web_acl_arn = aws_wafv2_web_acl.test.arn

  managed_rule_group {
    name        = "AWSManagedRulesCommonRuleSet"
    vendor_name = "AWS"
  }

  override_action = "none"
}
`, rName)
}

func testAccWebACLRuleGroupAssociationConfig_ManagedRuleGroup_withVersion(rName string) string {
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

resource "aws_wafv2_web_acl_rule_group_association" "test" {
  rule_name   = "test-rule"
  priority    = 1
  web_acl_arn = aws_wafv2_web_acl.test.arn

  managed_rule_group {
    name        = "AWSManagedRulesCommonRuleSet"
    vendor_name = "AWS"
    version     = "Version_1.0"
  }

  override_action = "none"
}
`, rName)
}

func testAccWebACLRuleGroupAssociationConfig_ManagedRuleGroup_ruleActionOverride(rName string) string {
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

resource "aws_wafv2_web_acl_rule_group_association" "test" {
  rule_name   = "test-rule"
  priority    = 1
  web_acl_arn = aws_wafv2_web_acl.test.arn

  managed_rule_group {
    name        = "AWSManagedRulesCommonRuleSet"
    vendor_name = "AWS"

    rule_action_override {
      name = "GenericRFI_BODY"
      action_to_use {
        count {}
      }
    }
  }

  override_action = "none"
}
`, rName)
}

func testAccWebACLRuleGroupAssociationConfig_ManagedRuleGroupConfig(rName string) string {
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

resource "aws_wafv2_web_acl_rule_group_association" "test" {
  rule_name   = "test-rule"
  priority    = 1
  web_acl_arn = aws_wafv2_web_acl.test.arn

  managed_rule_group {
    name        = "AWSManagedRulesBotControlRuleSet"
    vendor_name = "AWS"

    managed_rule_group_configs {
      aws_managed_rules_bot_control_rule_set {
        inspection_level = "COMMON"
      }
    }
  }

  override_action = "none"
}
`, rName)
}

func testAccWebACLRuleGroupAssociationConfig_ManagedRuleGroupConfigUpdate(rName string) string {
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

resource "aws_wafv2_web_acl_rule_group_association" "test" {
  rule_name   = "test-rule"
  priority    = 1
  web_acl_arn = aws_wafv2_web_acl.test.arn

  managed_rule_group {
    name        = "AWSManagedRulesBotControlRuleSet"
    vendor_name = "AWS"

    managed_rule_group_configs {
      aws_managed_rules_bot_control_rule_set {
        inspection_level = "TARGETED"
      }
    }
  }

  override_action = "none"
}
`, rName)
}

func testAccWebACLRuleGroupAssociationConfig_ManagedRuleGroupConfig_acfpRuleSet(rName string) string {
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

resource "aws_wafv2_web_acl_rule_group_association" "test" {
  rule_name   = "test-rule"
  priority    = 1
  web_acl_arn = aws_wafv2_web_acl.test.arn

  managed_rule_group {

    name        = "AWSManagedRulesACFPRuleSet"
    vendor_name = "AWS"

    managed_rule_group_configs {
      aws_managed_rules_acfp_rule_set {
        creation_path          = "/creation"
        registration_page_path = "/registration"
        request_inspection {
          email_field {
            identifier = "/email"
          }
          password_field {
            identifier = "/password"
          }
          phone_number_fields {
            identifiers = ["/phone1", "/phone2"]
          }
          address_fields {
            identifiers = ["home", "work"]
          }
          payload_type = "JSON"
          username_field {
            identifier = "/username"
          }
        }
      }
    }
  }
}
`, rName)
}

func testAccWebACLRuleGroupAssociationConfig_ManagedRuleGroupConfig_acfpRuleSetUpdate(rName string) string {
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

resource "aws_wafv2_web_acl_rule_group_association" "test" {
  rule_name   = "test-rule"
  priority    = 1
  web_acl_arn = aws_wafv2_web_acl.test.arn

  managed_rule_group {
    name        = "AWSManagedRulesACFPRuleSet"
    vendor_name = "AWS"

    managed_rule_group_configs {
      aws_managed_rules_acfp_rule_set {
        enable_regex_in_path   = true
        creation_path          = "/creation"
        registration_page_path = "/registration"

        request_inspection {
          email_field {
            identifier = "/email"
          }
          password_field {
            identifier = "/pass"
          }
          phone_number_fields {
            identifiers = ["/phone3"]
          }
          address_fields {
            identifiers = ["mobile"]
          }
          payload_type = "JSON"
          username_field {
            identifier = "/user"
          }
        }
      }
    }
  }

  override_action = "none"
}
`, rName)
}

func testAccWebACLRuleGroupAssociationConfig_ManagedRuleGroupConfig_antiDDoSRuleSet(rName string) string {
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

resource "aws_wafv2_web_acl_rule_group_association" "test" {
  rule_name   = "test-rule"
  priority    = 1
  web_acl_arn = aws_wafv2_web_acl.test.arn

  managed_rule_group {

    name        = "AWSManagedRulesAntiDDoSRuleSet"
    vendor_name = "AWS"

    managed_rule_group_configs {
      aws_managed_rules_anti_ddos_rule_set {
        client_side_action_config {
          challenge {
            usage_of_action = "ENABLED"
            exempt_uri_regular_expression {
              regex_string = "\\/api\\/"
            }
            exempt_uri_regular_expression {
              regex_string = "jpg"
            }
            sensitivity = "MEDIUM"
          }
        }
        sensitivity_to_block = "HIGH"
      }
    }

  }

}
`, rName)
}

func testAccWebACLRuleGroupAssociationConfig_ManagedRuleGroupConfig_antiDDoSRuleSetWithDefaultSensitivity(rName string) string {
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

resource "aws_wafv2_web_acl_rule_group_association" "test" {
  rule_name   = "test-rule"
  priority    = 1
  web_acl_arn = aws_wafv2_web_acl.test.arn

  managed_rule_group {
    name        = "AWSManagedRulesAntiDDoSRuleSet"
    vendor_name = "AWS"

    managed_rule_group_configs {
      aws_managed_rules_anti_ddos_rule_set {
        client_side_action_config {
          challenge {
            usage_of_action = "ENABLED"
            exempt_uri_regular_expression {
              regex_string = "\\/api\\/"
            }
            exempt_uri_regular_expression {
              regex_string = "jpg"
            }
          }
        }
      }
    }
  }
}
`, rName)
}

func testAccWebACLRuleGroupAssociationConfig_ManagedRuleGroupConfig_atpRuleSet(rName string) string {
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

resource "aws_wafv2_web_acl_rule_group_association" "test" {
  rule_name   = "test-rule"
  priority    = 1
  web_acl_arn = aws_wafv2_web_acl.test.arn

  managed_rule_group {
    name        = "AWSManagedRulesATPRuleSet"
    vendor_name = "AWS"

    managed_rule_group_configs {
      aws_managed_rules_atp_rule_set {
        login_path = "/api/1/signin"

        request_inspection {
          password_field {
            identifier = "/password"
          }

          payload_type = "JSON"

          username_field {
            identifier = "/username"
          }
        }
      }
    }
  }

  override_action = "none"
}
`, rName)
}

func testAccWebACLRuleGroupAssociationConfig_ManagedRuleGroupConfig_atpRuleSetUpdate(rName string) string {
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

resource "aws_wafv2_web_acl_rule_group_association" "test" {
  rule_name   = "test-rule"
  priority    = 1
  web_acl_arn = aws_wafv2_web_acl.test.arn

  managed_rule_group {
    name        = "AWSManagedRulesATPRuleSet"
    vendor_name = "AWS"

    managed_rule_group_configs {
      aws_managed_rules_atp_rule_set {
        enable_regex_in_path = true
        login_path           = "/api/2/signin"

        request_inspection {
          password_field {
            identifier = "/pass"
          }
          payload_type = "JSON"
          username_field {
            identifier = "/user"
          }
        }
      }
    }
  }

  override_action = "none"
}
`, rName)
}

func testAccWebACLRuleGroupAssociationConfig_ManagedRuleGroupConfig_botControl(rName string) string {
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

resource "aws_wafv2_web_acl_rule_group_association" "test" {
  rule_name   = "test-rule"
  priority    = 1
  web_acl_arn = aws_wafv2_web_acl.test.arn

  managed_rule_group {
    name        = "AWSManagedRulesBotControlRuleSet"
    vendor_name = "AWS"

    managed_rule_group_configs {
      aws_managed_rules_bot_control_rule_set {
        inspection_level        = "TARGETED"
        enable_machine_learning = true
      }
    }
  }

  override_action = "none"
}
`, rName)
}
