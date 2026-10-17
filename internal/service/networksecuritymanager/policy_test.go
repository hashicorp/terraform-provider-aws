// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networksecuritymanager_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networksecuritymanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networksecuritymanager/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfnetworksecuritymanager "github.com/hashicorp/terraform-provider-aws/internal/service/networksecuritymanager"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetPolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_policy.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "associated_template_and_rule.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "associated_template_and_rule.0.rule_arn", "aws_networksecuritymanager_rule.a", names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "associated_template_and_rule.0.template_arn"),
					resource.TestCheckResourceAttr(resourceName, "policy_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "policy_configuration.0.remediation_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "policy_configuration.0.resources_clean_up", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "policy_configuration.0.waf_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "policy_configuration.0.waf_config.0.conflict_resolution", "MERGE_WHERE_APPLICABLE"),
					resource.TestCheckResourceAttr(resourceName, "policy_configuration.0.waf_config.0.existing_customer_web_acl_resolution", "NO_REMEDIATION"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("network-security-manager", regexache.MustCompile(`policy:[0-9a-z]+$`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("firewall_type"), tfknownvalue.StringExact(awstypes.PolicyFirewallTypeWaf)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrPriority), knownvalue.Int32Exact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("is_published"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(awstypes.EntityStatusActive)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("has_published_version"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact("1")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_id"), knownvalue.StringRegexp(regexache.MustCompile(`^[0-9a-z]+$`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("updated_at"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{})),
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrARN: tfknownvalue.RegionalARNRegexp("network-security-manager", regexache.MustCompile(`policy:[0-9a-z]+$`)),
					}),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func testAccPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetPolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_policy.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfnetworksecuritymanager.ResourcePolicy, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

// Every optional argument set on creation; the policy is created as a draft
// with a mixed list of rules and a template.
func testAccPolicy_full(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetPolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_policy.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_full(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "associated_template_and_rule.#", "3"),
					resource.TestCheckResourceAttrPair(resourceName, "associated_template_and_rule.0.rule_arn", "aws_networksecuritymanager_rule.a", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "associated_template_and_rule.1.template_arn", "aws_networksecuritymanager_template.t1", names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "associated_template_and_rule.1.rule_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "associated_template_and_rule.2.rule_arn", "aws_networksecuritymanager_rule.b", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "policy_configuration.0.remediation_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "policy_configuration.0.resources_clean_up", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "policy_configuration.0.waf_config.0.existing_customer_web_acl_resolution", "RETROFIT"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("Managed by Terraform")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrPriority), knownvalue.Int32Exact(5)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("is_published"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(awstypes.EntityStatusDraft)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("has_published_version"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact("1")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

// A Shield Advanced policy has no templates or rules and no WAF configuration.
func testAccPolicy_shieldAdvanced(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetPolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_policy.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_shieldAdvanced(rName, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "associated_template_and_rule.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "policy_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "policy_configuration.0.remediation_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "policy_configuration.0.resources_clean_up", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "policy_configuration.0.waf_config.#", "0"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("firewall_type"), tfknownvalue.StringExact(awstypes.PolicyFirewallTypeShieldAdvanced)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(awstypes.EntityStatusActive)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact("1")),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccPolicyConfig_shieldAdvanced(rName, true, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "policy_configuration.0.remediation_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "policy_configuration.0.resources_clean_up", acctest.CtTrue),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact("2")),
				},
			},
			{
				Config: testAccPolicyConfig_shieldAdvancedDescription(rName, "shield description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("shield description")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact("3")),
				},
			},
		},
	})
}

// Every value of the WAF configuration, and each transition between them, is
// an in-place update.
func testAccPolicy_wafConfig(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetPolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_policy.test"

	step := func(remediationEnabled, resourcesCleanUp bool, existingResolution awstypes.ExistingCustomerWebACLResolution, version string) resource.TestStep {
		return resource.TestStep{
			Config: testAccPolicyConfig_wafConfig(rName, remediationEnabled, resourcesCleanUp, string(existingResolution)),
			Check: resource.ComposeAggregateTestCheckFunc(
				testAccCheckPolicyExists(ctx, t, resourceName, &v),
			),
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
				},
			},
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_configuration").AtSliceIndex(0).AtMapKey("remediation_enabled"), knownvalue.Bool(remediationEnabled)),
				statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_configuration").AtSliceIndex(0).AtMapKey("resources_clean_up"), knownvalue.Bool(resourcesCleanUp)),
				statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_configuration").AtSliceIndex(0).AtMapKey("waf_config").AtSliceIndex(0).AtMapKey("conflict_resolution"), tfknownvalue.StringExact(awstypes.WAFConflictResolutionOptionsMergeWhereApplicable)),
				statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_configuration").AtSliceIndex(0).AtMapKey("waf_config").AtSliceIndex(0).AtMapKey("existing_customer_web_acl_resolution"), tfknownvalue.StringExact(existingResolution)),
				statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact(version)),
			},
		}
	}

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_wafConfig(rName, false, false, string(awstypes.ExistingCustomerWebACLResolutionNoRemediation)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact("1")),
				},
			},
			step(true, false, awstypes.ExistingCustomerWebACLResolutionNoRemediation, "2"),
			step(true, true, awstypes.ExistingCustomerWebACLResolutionRetrofit, "3"),
			step(false, true, awstypes.ExistingCustomerWebACLResolutionOverrideAssociation, "4"),
			step(false, false, awstypes.ExistingCustomerWebACLResolutionRetrofit, "5"),
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			step(true, true, awstypes.ExistingCustomerWebACLResolutionNoRemediation, "6"),
		},
	})
}

// priority is updated in place and is unique across the account's policies.
func testAccPolicy_priority(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetPolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_policy.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_twoPolicies(rName, 1, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrPriority), knownvalue.Int32Exact(1)),
					statecheck.ExpectKnownValue("aws_networksecuritymanager_policy.other", tfjsonpath.New(names.AttrPriority), knownvalue.Int32Exact(2)),
				},
			},
			{
				// The same priority as another policy in the account.
				Config:      testAccPolicyConfig_twoPolicies(rName, 1, 1),
				ExpectError: regexache.MustCompile(`ConflictException: Priority 1 is already used by another policy`),
			},
			{
				Config: testAccPolicyConfig_twoPolicies(rName, 2147483647, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectResourceAction("aws_networksecuritymanager_policy.other", plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrPriority), knownvalue.Int32Exact(2147483647)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact("2")),
				},
			},
			{
				// The priority the first policy just gave up.
				Config: testAccPolicyConfig_twoPolicies(rName, 2147483647, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
						plancheck.ExpectResourceAction("aws_networksecuritymanager_policy.other", plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue("aws_networksecuritymanager_policy.other", tfjsonpath.New(names.AttrPriority), knownvalue.Int32Exact(1)),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

// The list of templates and rules is ordered; adding, removing and reordering
// entries are in-place updates.
func testAccPolicy_associations(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetPolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_policy.test"

	ruleA := testAccPolicyAssociationRule("aws_networksecuritymanager_rule.a")
	ruleB := testAccPolicyAssociationRule("aws_networksecuritymanager_rule.b")
	ruleC := testAccPolicyAssociationRule("aws_networksecuritymanager_rule.c")
	template1 := testAccPolicyAssociationTemplate("aws_networksecuritymanager_template.t1")
	template2 := testAccPolicyAssociationTemplate("aws_networksecuritymanager_template.t2")

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				// Rules only.
				Config: testAccPolicyConfig_associations(rName, ruleA+ruleB),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "associated_template_and_rule.#", "2"),
					resource.TestCheckResourceAttrPair(resourceName, "associated_template_and_rule.0.rule_arn", "aws_networksecuritymanager_rule.a", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "associated_template_and_rule.1.rule_arn", "aws_networksecuritymanager_rule.b", names.AttrARN),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact("1")),
				},
			},
			{
				// Templates only: one, then two.
				Config: testAccPolicyConfig_associations(rName, template1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "associated_template_and_rule.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "associated_template_and_rule.0.template_arn", "aws_networksecuritymanager_template.t1", names.AttrARN),
					resource.TestCheckNoResourceAttr(resourceName, "associated_template_and_rule.0.rule_arn"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact("2")),
				},
			},
			{
				Config: testAccPolicyConfig_associations(rName, template1+template2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "associated_template_and_rule.#", "2"),
					resource.TestCheckResourceAttrPair(resourceName, "associated_template_and_rule.0.template_arn", "aws_networksecuritymanager_template.t1", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "associated_template_and_rule.1.template_arn", "aws_networksecuritymanager_template.t2", names.AttrARN),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact("3")),
				},
			},
			{
				// Mixed: rules around the templates, an INSPECTION rule included.
				Config: testAccPolicyConfig_associations(rName, ruleA+template1+ruleC+template2+ruleB),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "associated_template_and_rule.#", "5"),
					resource.TestCheckResourceAttrPair(resourceName, "associated_template_and_rule.0.rule_arn", "aws_networksecuritymanager_rule.a", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "associated_template_and_rule.1.template_arn", "aws_networksecuritymanager_template.t1", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "associated_template_and_rule.2.rule_arn", "aws_networksecuritymanager_rule.c", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "associated_template_and_rule.3.template_arn", "aws_networksecuritymanager_template.t2", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "associated_template_and_rule.4.rule_arn", "aws_networksecuritymanager_rule.b", names.AttrARN),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact("4")),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				// The same five entries in a different order is an update.
				Config: testAccPolicyConfig_associations(rName, template2+ruleB+template1+ruleA+ruleC),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "associated_template_and_rule.#", "5"),
					resource.TestCheckResourceAttrPair(resourceName, "associated_template_and_rule.0.template_arn", "aws_networksecuritymanager_template.t2", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "associated_template_and_rule.1.rule_arn", "aws_networksecuritymanager_rule.b", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "associated_template_and_rule.2.template_arn", "aws_networksecuritymanager_template.t1", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "associated_template_and_rule.3.rule_arn", "aws_networksecuritymanager_rule.a", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "associated_template_and_rule.4.rule_arn", "aws_networksecuritymanager_rule.c", names.AttrARN),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact("5")),
				},
			},
			{
				// Down to one rule.
				Config: testAccPolicyConfig_associations(rName, ruleC),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "associated_template_and_rule.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "associated_template_and_rule.0.rule_arn", "aws_networksecuritymanager_rule.c", names.AttrARN),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact("6")),
				},
			},
			{
				// The same configuration again is not an update.
				Config: testAccPolicyConfig_associations(rName, ruleC),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact("6")),
				},
			},
		},
	})
}

// The list takes up to 100 entries.
func testAccPolicy_associationsMax(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetPolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_policy.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_associationsCount(rName, 98),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "associated_template_and_rule.#", "100"),
					resource.TestCheckResourceAttrPair(resourceName, "associated_template_and_rule.0.template_arn", "aws_networksecuritymanager_template.t1", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "associated_template_and_rule.1.template_arn", "aws_networksecuritymanager_template.t2", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "associated_template_and_rule.2.rule_arn", "aws_networksecuritymanager_rule.many.0", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "associated_template_and_rule.99.rule_arn", "aws_networksecuritymanager_rule.many.97", names.AttrARN),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config:      testAccPolicyConfig_associationsCount(rName, 99),
				ExpectError: regexache.MustCompile(`list must contain at most 100\s+elements`),
			},
		},
	})
}

// is_published moves the policy between DRAFT and ACTIVE; a draft saved on a
// published policy is a pending draft on top of the published version.
func testAccPolicy_publish(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetPolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_policy.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_isPublished(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("network-security-manager", regexache.MustCompile(`policy:[0-9a-z]+$`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("is_published"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(awstypes.EntityStatusDraft)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("has_published_version"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact("1")),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccPolicyConfig_isPublished(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("is_published"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(awstypes.EntityStatusActive)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("has_published_version"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact("1")),
				},
			},
			{
				Config: testAccPolicyConfig_isPublished(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("is_published"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(awstypes.EntityStatusDraft)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("has_published_version"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact("2")),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccPolicyConfig_isPublished(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("is_published"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(awstypes.EntityStatusActive)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("has_published_version"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact("2")),
				},
			},
			{
				// Destroyed with a pending draft.
				Config: testAccPolicyConfig_isPublished(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &v),
				),
			},
		},
	})
}

func testAccPolicy_description(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetPolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_policy.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_description(rName, "description 1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("description 1")),
				},
			},
			{
				Config: testAccPolicyConfig_description(rName, "description 2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("description 2")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact("2")),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				Config: testAccPolicyConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact("3")),
				},
			},
		},
	})
}

// An update token read on the last refresh is stale after any change outside
// Terraform; the resource reads the current token right before updating.
func testAccPolicy_updateToken(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetPolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_policy.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_description(rName, "description 1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &v),
				),
			},
			{
				PreConfig: func() {
					testAccPolicyUpdateOutOfBand(ctx, t, &v, "changed outside Terraform")
				},
				Config: testAccPolicyConfig_descriptionAndPriority(rName, "description 1", 3),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("description 1")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrPriority), knownvalue.Int32Exact(3)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact("3")),
				},
			},
			{
				PreConfig: func() {
					// A draft saved outside Terraform on a published policy
					// is published by the next apply.
					testAccPolicyUpdateOutOfBand(ctx, t, &v, "draft saved outside Terraform")
				},
				Config: testAccPolicyConfig_descriptionAndPriority(rName, "description 1", 3),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("description 1")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("is_published"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(awstypes.EntityStatusActive)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("has_published_version"), knownvalue.Bool(false)),
				},
			},
		},
	})
}

// name and firewall_type cannot be changed in place.
func testAccPolicy_replace(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2, v3 networksecuritymanager.GetPolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_policy.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_name(rName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &v1),
				),
			},
			{
				Config: testAccPolicyConfig_name(rName, rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &v2),
					testAccCheckPolicyRecreated(&v1, &v2),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rNameUpdated)),
				},
			},
			{
				// WAF → SHIELD_ADVANCED, with the list and WAF configuration
				// dropped as the firewall type requires.
				Config: testAccPolicyConfig_shieldAdvancedName(rName, rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &v3),
					testAccCheckPolicyRecreated(&v2, &v3),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("firewall_type"), tfknownvalue.StringExact(awstypes.PolicyFirewallTypeShieldAdvanced)),
				},
			},
		},
	})
}

// A policy can be re-created right after it was deleted.
func testAccPolicy_recreate(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 networksecuritymanager.GetPolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_policy.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &v1),
				),
			},
			{
				Taint:  []string{resourceName},
				Config: testAccPolicyConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &v2),
					testAccCheckPolicyRecreated(&v1, &v2),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
		},
	})
}

// A rule or template referenced by a published policy cannot be deleted, so
// replacing one needs create_before_destroy.
func testAccPolicy_referenceReplace(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetPolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_policy.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_referenceNames(rName, rName, rName, false, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "associated_template_and_rule.0.rule_arn", "aws_networksecuritymanager_rule.a", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "associated_template_and_rule.1.template_arn", "aws_networksecuritymanager_template.t1", names.AttrARN),
				),
			},
			{
				// The rule is renamed with the default lifecycle.
				Config:      testAccPolicyConfig_referenceNames(rName, rNameUpdated, rName, false, true),
				ExpectError: regexache.MustCompile(`ConflictException: The resource is currently in use`),
			},
			{
				Config: testAccPolicyConfig_referenceNames(rName, rNameUpdated, rName, true, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "associated_template_and_rule.0.rule_arn", "aws_networksecuritymanager_rule.a", names.AttrARN),
					resource.TestCheckResourceAttr("aws_networksecuritymanager_rule.a", names.AttrName, rNameUpdated),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("aws_networksecuritymanager_rule.a", plancheck.ResourceActionCreateBeforeDestroy),
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact("2")),
				},
			},
			{
				// The template is renamed with the default lifecycle.
				Config:      testAccPolicyConfig_referenceNames(rName, rNameUpdated, rNameUpdated, true, false),
				ExpectError: regexache.MustCompile(`ConflictException: The resource is currently in use`),
			},
			{
				Config: testAccPolicyConfig_referenceNames(rName, rNameUpdated, rNameUpdated, true, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "associated_template_and_rule.1.template_arn", "aws_networksecuritymanager_template.t1", names.AttrARN),
					resource.TestCheckResourceAttr("aws_networksecuritymanager_template.t1", names.AttrName, rNameUpdated),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("aws_networksecuritymanager_template.t1", plancheck.ResourceActionCreateBeforeDestroy),
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact("3")),
				},
			},
		},
	})
}

// A policy that exists only as a draft does not protect the rules and
// templates it references: they can be replaced in the default order.
func testAccPolicy_draftReferences(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetPolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_policy.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyConfig_draftReferenceNames(rName, rName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(awstypes.EntityStatusDraft)),
				},
			},
			{
				// The template is replaced (destroy, then create) and the
				// draft follows it.
				Config: testAccPolicyConfig_draftReferenceNames(rName, rName, rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "associated_template_and_rule.1.template_arn", "aws_networksecuritymanager_template.t1", names.AttrARN),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("aws_networksecuritymanager_template.t1", plancheck.ResourceActionReplace),
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(awstypes.EntityStatusDraft)),
					// An update of a draft that was never published keeps version 1.
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact("1")),
				},
			},
			{
				// The rule is replaced (destroy, then create) the same way.
				Config: testAccPolicyConfig_draftReferenceNames(rName, rNameUpdated, rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "associated_template_and_rule.0.rule_arn", "aws_networksecuritymanager_rule.a", names.AttrARN),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("aws_networksecuritymanager_rule.a", plancheck.ResourceActionReplace),
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(awstypes.EntityStatusDraft)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact("1")),
				},
			},
		},
	})
}

// Argument validation at plan time, and reference validation by the API.
func testAccPolicy_validation(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	fakeARN := func(kind string, n int) string {
		return fmt.Sprintf("arn:%s:network-security-manager:%s:123456789012:%s:tfacc%021d", acctest.Partition(), acctest.Region(), kind, n)
	}
	fakeAssociation := func(kind string, n int) string {
		return fmt.Sprintf("\n  associated_template_and_rule {\n    %s_arn = %q\n  }\n", kind, fakeARN(kind, n))
	}
	fakeAssociations := func(kind string, count int) string {
		var b strings.Builder
		for n := range count {
			b.WriteString(fakeAssociation(kind, n))
		}
		return b.String()
	}
	fakeRule := fakeAssociation(names.AttrRule, 1)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccPolicyConfig_literal("-"+rName, "WAF", 1, testAccPolicyWAFConfig, fakeRule),
				ExpectError: regexache.MustCompile(`must start with an alphanumeric character`),
			},
			{
				Config:      testAccPolicyConfig_literal(strings.Repeat("a", 129), "WAF", 1, testAccPolicyWAFConfig, fakeRule),
				ExpectError: regexache.MustCompile(`string length must be between 1 and 128`),
			},
			{
				Config:      testAccPolicyConfig_description(rName, strings.Repeat("a", 257)),
				ExpectError: regexache.MustCompile(`string length must be at most 256`),
			},
			{
				Config:      testAccPolicyConfig_description(rName, "not # allowed"),
				ExpectError: regexache.MustCompile(`must contain only alphanumeric characters`),
			},
			{
				Config:      testAccPolicyConfig_literal(rName, "BOGUS", 1, testAccPolicyWAFConfig, fakeRule),
				ExpectError: regexache.MustCompile(`Invalid String Enum Value`),
			},
			{
				Config:      testAccPolicyConfig_literal(rName, "WAF", 0, testAccPolicyWAFConfig, fakeRule),
				ExpectError: regexache.MustCompile(`value must be at least 1`),
			},
			{
				// WAF without a list.
				Config:      testAccPolicyConfig_literal(rName, "WAF", 1, testAccPolicyWAFConfig, ""),
				ExpectError: regexache.MustCompile(`"associated_template_and_rule" must be specified when\s+"firewall_type" is "WAF"`),
			},
			{
				// WAF without waf_config.
				Config:      testAccPolicyConfig_literal(rName, "WAF", 1, "", fakeRule),
				ExpectError: regexache.MustCompile(`"policy_configuration\[0\].waf_config" must be specified when\s+"firewall_type" is "WAF"`),
			},
			{
				// SHIELD_ADVANCED with a list.
				Config:      testAccPolicyConfig_literal(rName, "SHIELD_ADVANCED", 1, "", fakeRule),
				ExpectError: regexache.MustCompile(`"associated_template_and_rule" cannot be specified when\s+"firewall_type" is "SHIELD_ADVANCED"`),
			},
			{
				// SHIELD_ADVANCED with waf_config.
				Config:      testAccPolicyConfig_literal(rName, "SHIELD_ADVANCED", 1, testAccPolicyWAFConfig, ""),
				ExpectError: regexache.MustCompile(`"policy_configuration\[0\].waf_config" cannot be specified when\s+"firewall_type" is "SHIELD_ADVANCED"`),
			},
			{
				// No policy_configuration.
				Config:      testAccPolicyConfig_noConfiguration(rName),
				ExpectError: regexache.MustCompile(`Block policy_configuration must have a configuration value`),
			},
			{
				// Three templates.
				Config:      testAccPolicyConfig_literal(rName, "WAF", 1, testAccPolicyWAFConfig, fakeAssociations("template", 3)),
				ExpectError: regexache.MustCompile(`must reference at most 2\s+templates`),
			},
			{
				// 101 entries.
				Config:      testAccPolicyConfig_literal(rName, "WAF", 1, testAccPolicyWAFConfig, fakeAssociations(names.AttrRule, 101)),
				ExpectError: regexache.MustCompile(`list must contain at most 100\s+elements`),
			},
			{
				// The same rule twice.
				Config:      testAccPolicyConfig_literal(rName, "WAF", 1, testAccPolicyWAFConfig, fakeRule+fakeRule),
				ExpectError: regexache.MustCompile(`must not reference the same\s+template or rule\s+more than once`),
			},
			{
				// Both ARNs in one entry.
				Config:      testAccPolicyConfig_literal(rName, "WAF", 1, testAccPolicyWAFConfig, fmt.Sprintf("\n  associated_template_and_rule {\n    rule_arn     = %q\n    template_arn = %q\n  }\n", fakeARN(names.AttrRule, 1), fakeARN("template", 1))),
				ExpectError: regexache.MustCompile(`one \(and only one\) of`),
			},
			{
				// Neither ARN in an entry.
				Config:      testAccPolicyConfig_literal(rName, "WAF", 1, testAccPolicyWAFConfig, "\n  associated_template_and_rule {\n  }\n"),
				ExpectError: regexache.MustCompile(`one \(and only one\) of`),
			},
			{
				Config:      testAccPolicyConfig_literal(rName, "WAF", 1, testAccPolicyWAFConfig, "\n  associated_template_and_rule {\n    rule_arn = \"not-an-arn\"\n  }\n"),
				ExpectError: regexache.MustCompile(`cannot be parsed as an ARN`),
			},
			{
				// A rule in another account is an authorization failure, not a
				// validation one.
				Config:      testAccPolicyConfig_literal(rName, "WAF", 1, testAccPolicyWAFConfig, fakeRule),
				ExpectError: regexache.MustCompile(`UnauthorizedException|AccessDeniedException`),
			},
			{
				// A rule that does not exist.
				Config:      testAccPolicyConfig_missingReference(rName, names.AttrRule),
				ExpectError: regexache.MustCompile(`ValidationException: One or more resources referenced by this request`),
			},
			{
				// A template that does not exist.
				Config:      testAccPolicyConfig_missingReference(rName, "template"),
				ExpectError: regexache.MustCompile(`ValidationException: One or more resources referenced by this request`),
			},
			{
				// A draft rule cannot be referenced.
				Config:      testAccPolicyConfig_draftRule(rName),
				ExpectError: regexache.MustCompile(`ValidationException: One or more resources referenced by this request`),
			},
			{
				// A draft template cannot be referenced.
				Config:      testAccPolicyConfig_draftTemplate(rName),
				ExpectError: regexache.MustCompile(`ValidationException: One or more resources referenced by this request`),
			},
			{
				// A DRAFT-qualified ARN.
				Config:      testAccPolicyConfig_draftQualifiedRule(rName),
				ExpectError: regexache.MustCompile(`ValidationException: DRAFT-qualified ARNs not allowed`),
			},
			{
				// A template ARN given as a rule.
				Config:      testAccPolicyConfig_templateAsRule(rName),
				ExpectError: regexache.MustCompile(`ValidationException: Invalid entity type in ARN`),
			},
		},
	})
}

func testAccCheckPolicyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).NetworkSecurityManagerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networksecuritymanager_policy" {
				continue
			}

			_, err := tfnetworksecuritymanager.FindPolicyByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Network Security Manager Policy %s still exists", rs.Primary.Attributes[names.AttrARN])
		}

		return nil
	}
}

func testAccCheckPolicyExists(ctx context.Context, t *testing.T, n string, v *networksecuritymanager.GetPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).NetworkSecurityManagerClient(ctx)

		output, err := tfnetworksecuritymanager.FindPolicyByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckPolicyRecreated(before, after *networksecuritymanager.GetPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.PolicyId), aws.ToString(after.PolicyId); before == after {
			return fmt.Errorf("Network Security Manager Policy (%s) not recreated", before)
		}

		return nil
	}
}

// testAccPolicyUpdateOutOfBand changes the policy's description through the
// API, invalidating the update token Terraform last saw.
func testAccPolicyUpdateOutOfBand(ctx context.Context, t *testing.T, v *networksecuritymanager.GetPolicyOutput, description string) {
	conn := acctest.ProviderMeta(ctx, t).NetworkSecurityManagerClient(ctx)

	input := networksecuritymanager.UpdatePolicyInput{
		PolicyIdentifier:  v.PolicyArn,
		UpdateToken:       v.UpdateToken,
		IsPublished:       aws.Bool(v.Status != awstypes.EntityStatusDraft),
		PolicyDescription: aws.String(description),
	}
	if strings.HasPrefix(description, "draft") {
		input.IsPublished = aws.Bool(false)
	}

	if _, err := conn.UpdatePolicy(ctx, &input); err != nil {
		t.Fatalf("updating Network Security Manager Policy (%s) outside Terraform: %s", aws.ToString(v.PolicyArn), err)
	}
}

const testAccPolicyWAFConfig = `
    waf_config {
      conflict_resolution                  = "MERGE_WHERE_APPLICABLE"
      existing_customer_web_acl_resolution = "NO_REMEDIATION"
    }
`

func testAccPolicyAssociationRule(address string) string {
	return fmt.Sprintf("\n  associated_template_and_rule {\n    rule_arn = %s.arn\n  }\n", address)
}

func testAccPolicyAssociationTemplate(address string) string {
	return fmt.Sprintf("\n  associated_template_and_rule {\n    template_arn = %s.arn\n  }\n", address)
}

// Three published rules and two published templates, one rule each.
func testAccPolicyConfig_base(rName string) string {
	return acctest.ConfigCompose(testAccTemplateConfig_base(rName), fmt.Sprintf(`
resource "aws_networksecuritymanager_template" "t1" {
  name          = "%[1]s-t1"
  firewall_type = "WAF"
  rule_arns     = [aws_networksecuritymanager_rule.a.arn]
}

resource "aws_networksecuritymanager_template" "t2" {
  name          = "%[1]s-t2"
  firewall_type = "WAF"
  rule_arns     = [aws_networksecuritymanager_rule.b.arn]
}
`, rName))
}

func testAccPolicyConfig_basic(rName string) string {
	return testAccPolicyConfig_associations(rName, testAccPolicyAssociationRule("aws_networksecuritymanager_rule.a"))
}

func testAccPolicyConfig_associations(rName, associations string) string {
	return acctest.ConfigCompose(testAccPolicyConfig_base(rName), fmt.Sprintf(`
resource "aws_networksecuritymanager_policy" "test" {
  name          = %[1]q
  firewall_type = "WAF"
  priority      = 1

  policy_configuration {
    remediation_enabled = false
    resources_clean_up  = false

    waf_config {
      conflict_resolution                  = "MERGE_WHERE_APPLICABLE"
      existing_customer_web_acl_resolution = "NO_REMEDIATION"
    }
  }
%[2]s
}
`, rName, associations))
}

func testAccPolicyConfig_full(rName string) string {
	return acctest.ConfigCompose(testAccPolicyConfig_base(rName), fmt.Sprintf(`
resource "aws_networksecuritymanager_policy" "test" {
  name          = %[1]q
  description   = "Managed by Terraform"
  firewall_type = "WAF"
  is_published  = false
  priority      = 5

  policy_configuration {
    remediation_enabled = true
    resources_clean_up  = true

    waf_config {
      conflict_resolution                  = "MERGE_WHERE_APPLICABLE"
      existing_customer_web_acl_resolution = "RETROFIT"
    }
  }

  associated_template_and_rule {
    rule_arn = aws_networksecuritymanager_rule.a.arn
  }

  associated_template_and_rule {
    template_arn = aws_networksecuritymanager_template.t1.arn
  }

  associated_template_and_rule {
    rule_arn = aws_networksecuritymanager_rule.b.arn
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, acctest.CtKey1, acctest.CtValue1))
}

func testAccPolicyConfig_shieldAdvanced(rName string, remediationEnabled, resourcesCleanUp bool) string {
	return fmt.Sprintf(`
resource "aws_networksecuritymanager_policy" "test" {
  name          = %[1]q
  firewall_type = "SHIELD_ADVANCED"
  priority      = 1

  policy_configuration {
    remediation_enabled = %[2]t
    resources_clean_up  = %[3]t
  }
}
`, rName, remediationEnabled, resourcesCleanUp)
}

func testAccPolicyConfig_shieldAdvancedDescription(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_networksecuritymanager_policy" "test" {
  name          = %[1]q
  description   = %[2]q
  firewall_type = "SHIELD_ADVANCED"
  priority      = 1

  policy_configuration {
    remediation_enabled = true
    resources_clean_up  = true
  }
}
`, rName, description)
}

// A Shield Advanced policy in the configuration otherwise used by the WAF
// policy named policyName.
func testAccPolicyConfig_shieldAdvancedName(rName, policyName string) string {
	return acctest.ConfigCompose(testAccPolicyConfig_base(rName), fmt.Sprintf(`
resource "aws_networksecuritymanager_policy" "test" {
  name          = %[1]q
  firewall_type = "SHIELD_ADVANCED"
  priority      = 1

  policy_configuration {
    remediation_enabled = false
    resources_clean_up  = false
  }
}
`, policyName))
}

func testAccPolicyConfig_wafConfig(rName string, remediationEnabled, resourcesCleanUp bool, existingResolution string) string {
	return acctest.ConfigCompose(testAccPolicyConfig_base(rName), fmt.Sprintf(`
resource "aws_networksecuritymanager_policy" "test" {
  name          = %[1]q
  firewall_type = "WAF"
  priority      = 1

  policy_configuration {
    remediation_enabled = %[2]t
    resources_clean_up  = %[3]t

    waf_config {
      conflict_resolution                  = "MERGE_WHERE_APPLICABLE"
      existing_customer_web_acl_resolution = %[4]q
    }
  }

  associated_template_and_rule {
    rule_arn = aws_networksecuritymanager_rule.a.arn
  }
}
`, rName, remediationEnabled, resourcesCleanUp, existingResolution))
}

func testAccPolicyConfig_twoPolicies(rName string, priority, otherPriority int32) string {
	return acctest.ConfigCompose(testAccPolicyConfig_base(rName), fmt.Sprintf(`
resource "aws_networksecuritymanager_policy" "test" {
  name          = %[1]q
  firewall_type = "WAF"
  priority      = %[2]d

  policy_configuration {
    remediation_enabled = false
    resources_clean_up  = false

    waf_config {
      conflict_resolution                  = "MERGE_WHERE_APPLICABLE"
      existing_customer_web_acl_resolution = "NO_REMEDIATION"
    }
  }

  associated_template_and_rule {
    rule_arn = aws_networksecuritymanager_rule.a.arn
  }
}

resource "aws_networksecuritymanager_policy" "other" {
  name          = "%[1]s-other"
  firewall_type = "SHIELD_ADVANCED"
  priority      = %[3]d

  policy_configuration {
    remediation_enabled = false
    resources_clean_up  = false
  }
}
`, rName, priority, otherPriority))
}

// Two templates followed by ruleCount rules created with count.
func testAccPolicyConfig_associationsCount(rName string, ruleCount int) string {
	return acctest.ConfigCompose(testAccPolicyConfig_base(rName), fmt.Sprintf(`
resource "aws_networksecuritymanager_rule" "many" {
  count = %[2]d

  name          = "%[1]s-${count.index}"
  firewall_type = "WAF"
  rule_type     = "CONFIGURATION"
  configuration = jsonencode({
    VisibilityConfig = {
      SampledRequestsEnabled   = true
      CloudWatchMetricsEnabled = true
      MetricName               = "tf-acc-test-${count.index}"
    }
  })
}

resource "aws_networksecuritymanager_policy" "test" {
  name          = %[1]q
  firewall_type = "WAF"
  priority      = 1

  policy_configuration {
    remediation_enabled = false
    resources_clean_up  = false

    waf_config {
      conflict_resolution                  = "MERGE_WHERE_APPLICABLE"
      existing_customer_web_acl_resolution = "NO_REMEDIATION"
    }
  }

  associated_template_and_rule {
    template_arn = aws_networksecuritymanager_template.t1.arn
  }

  associated_template_and_rule {
    template_arn = aws_networksecuritymanager_template.t2.arn
  }

  dynamic "associated_template_and_rule" {
    for_each = aws_networksecuritymanager_rule.many[*].arn

    content {
      rule_arn = associated_template_and_rule.value
    }
  }
}
`, rName, ruleCount))
}

func testAccPolicyConfig_isPublished(rName string, isPublished bool) string {
	return acctest.ConfigCompose(testAccPolicyConfig_base(rName), fmt.Sprintf(`
resource "aws_networksecuritymanager_policy" "test" {
  name          = %[1]q
  firewall_type = "WAF"
  is_published  = %[2]t
  priority      = 1

  policy_configuration {
    remediation_enabled = false
    resources_clean_up  = false

    waf_config {
      conflict_resolution                  = "MERGE_WHERE_APPLICABLE"
      existing_customer_web_acl_resolution = "NO_REMEDIATION"
    }
  }

  associated_template_and_rule {
    rule_arn = aws_networksecuritymanager_rule.a.arn
  }
}
`, rName, isPublished))
}

func testAccPolicyConfig_description(rName, description string) string {
	return testAccPolicyConfig_descriptionAndPriority(rName, description, 1)
}

func testAccPolicyConfig_descriptionAndPriority(rName, description string, priority int32) string {
	return acctest.ConfigCompose(testAccPolicyConfig_base(rName), fmt.Sprintf(`
resource "aws_networksecuritymanager_policy" "test" {
  name          = %[1]q
  description   = %[2]q
  firewall_type = "WAF"
  priority      = %[3]d

  policy_configuration {
    remediation_enabled = false
    resources_clean_up  = false

    waf_config {
      conflict_resolution                  = "MERGE_WHERE_APPLICABLE"
      existing_customer_web_acl_resolution = "NO_REMEDIATION"
    }
  }

  associated_template_and_rule {
    rule_arn = aws_networksecuritymanager_rule.a.arn
  }
}
`, rName, description, priority))
}

func testAccPolicyConfig_name(rName, policyName string) string {
	return acctest.ConfigCompose(testAccPolicyConfig_base(rName), fmt.Sprintf(`
resource "aws_networksecuritymanager_policy" "test" {
  name          = %[1]q
  firewall_type = "WAF"
  priority      = 1

  policy_configuration {
    remediation_enabled = false
    resources_clean_up  = false

    waf_config {
      conflict_resolution                  = "MERGE_WHERE_APPLICABLE"
      existing_customer_web_acl_resolution = "NO_REMEDIATION"
    }
  }

  associated_template_and_rule {
    rule_arn = aws_networksecuritymanager_rule.a.arn
  }
}
`, policyName))
}

// A published policy referencing one rule directly and one through a
// template, with the lifecycle of each reference under test.
func testAccPolicyConfig_referenceNames(rName, ruleName, templateName string, ruleCreateBeforeDestroy, templateCreateBeforeDestroy bool) string {
	return fmt.Sprintf(`
resource "aws_networksecuritymanager_rule" "a" {
  name          = %[2]q
  firewall_type = "WAF"
  rule_type     = "CONFIGURATION"
  configuration = jsonencode({
    DefaultAction = {
      Allow = {}
    }
  })

  lifecycle {
    create_before_destroy = %[4]t
  }
}

resource "aws_networksecuritymanager_rule" "b" {
  name          = "%[1]s-b"
  firewall_type = "WAF"
  rule_type     = "CONFIGURATION"
  configuration = jsonencode({
    VisibilityConfig = {
      SampledRequestsEnabled   = true
      CloudWatchMetricsEnabled = true
      MetricName               = "tf-acc-test"
    }
  })
}

resource "aws_networksecuritymanager_template" "t1" {
  name          = %[3]q
  firewall_type = "WAF"
  rule_arns     = [aws_networksecuritymanager_rule.b.arn]

  lifecycle {
    create_before_destroy = %[5]t
  }
}

resource "aws_networksecuritymanager_policy" "test" {
  name          = %[1]q
  firewall_type = "WAF"
  priority      = 1

  policy_configuration {
    remediation_enabled = false
    resources_clean_up  = false

    waf_config {
      conflict_resolution                  = "MERGE_WHERE_APPLICABLE"
      existing_customer_web_acl_resolution = "NO_REMEDIATION"
    }
  }

  associated_template_and_rule {
    rule_arn = aws_networksecuritymanager_rule.a.arn
  }

  associated_template_and_rule {
    template_arn = aws_networksecuritymanager_template.t1.arn
  }
}
`, rName, ruleName, templateName, ruleCreateBeforeDestroy, templateCreateBeforeDestroy)
}

// A draft-only policy referencing one rule directly and one template.
func testAccPolicyConfig_draftReferenceNames(rName, ruleName, templateName string) string {
	return fmt.Sprintf(`
resource "aws_networksecuritymanager_rule" "a" {
  name          = %[2]q
  firewall_type = "WAF"
  rule_type     = "CONFIGURATION"
  configuration = jsonencode({
    DefaultAction = {
      Allow = {}
    }
  })
}

resource "aws_networksecuritymanager_rule" "b" {
  name          = "%[1]s-b"
  firewall_type = "WAF"
  rule_type     = "CONFIGURATION"
  configuration = jsonencode({
    VisibilityConfig = {
      SampledRequestsEnabled   = true
      CloudWatchMetricsEnabled = true
      MetricName               = "tf-acc-test"
    }
  })
}

resource "aws_networksecuritymanager_template" "t1" {
  name          = %[3]q
  firewall_type = "WAF"
  rule_arns     = [aws_networksecuritymanager_rule.b.arn]
}

resource "aws_networksecuritymanager_policy" "test" {
  name          = %[1]q
  firewall_type = "WAF"
  is_published  = false
  priority      = 1

  policy_configuration {
    remediation_enabled = false
    resources_clean_up  = false

    waf_config {
      conflict_resolution                  = "MERGE_WHERE_APPLICABLE"
      existing_customer_web_acl_resolution = "NO_REMEDIATION"
    }
  }

  associated_template_and_rule {
    rule_arn = aws_networksecuritymanager_rule.a.arn
  }

  associated_template_and_rule {
    template_arn = aws_networksecuritymanager_template.t1.arn
  }
}
`, rName, ruleName, templateName)
}

// A policy given as literals, with no rule or template resource.
func testAccPolicyConfig_literal(policyName, firewallType string, priority int32, wafConfig, associations string) string {
	return fmt.Sprintf(`
resource "aws_networksecuritymanager_policy" "test" {
  name          = %[1]q
  firewall_type = %[2]q
  priority      = %[3]d

  policy_configuration {
    remediation_enabled = false
    resources_clean_up  = false
%[4]s
  }
%[5]s
}
`, policyName, firewallType, priority, wafConfig, associations)
}

func testAccPolicyConfig_noConfiguration(rName string) string {
	return fmt.Sprintf(`
resource "aws_networksecuritymanager_policy" "test" {
  name          = %[1]q
  firewall_type = "SHIELD_ADVANCED"
  priority      = 1
}
`, rName)
}

func testAccPolicyConfig_missingReference(rName, kind string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {}

resource "aws_networksecuritymanager_policy" "test" {
  name          = %[1]q
  firewall_type = "WAF"
  priority      = 1

  policy_configuration {
    remediation_enabled = false
    resources_clean_up  = false

    waf_config {
      conflict_resolution                  = "MERGE_WHERE_APPLICABLE"
      existing_customer_web_acl_resolution = "NO_REMEDIATION"
    }
  }

  associated_template_and_rule {
    %[2]s_arn = "arn:${data.aws_partition.current.partition}:network-security-manager:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:%[2]s:tfacc000000000000000000001"
  }
}
`, rName, kind)
}

func testAccPolicyConfig_draftRule(rName string) string {
	return fmt.Sprintf(`
resource "aws_networksecuritymanager_rule" "draft" {
  name          = %[1]q
  firewall_type = "WAF"
  rule_type     = "CONFIGURATION"
  is_published  = false
  configuration = jsonencode({
    DefaultAction = {
      Allow = {}
    }
  })
}

resource "aws_networksecuritymanager_policy" "test" {
  name          = %[1]q
  firewall_type = "WAF"
  priority      = 1

  policy_configuration {
    remediation_enabled = false
    resources_clean_up  = false

    waf_config {
      conflict_resolution                  = "MERGE_WHERE_APPLICABLE"
      existing_customer_web_acl_resolution = "NO_REMEDIATION"
    }
  }

  associated_template_and_rule {
    rule_arn = aws_networksecuritymanager_rule.draft.arn
  }
}
`, rName)
}

func testAccPolicyConfig_draftTemplate(rName string) string {
	return acctest.ConfigCompose(testAccTemplateConfig_base(rName), fmt.Sprintf(`
resource "aws_networksecuritymanager_template" "draft" {
  name          = %[1]q
  firewall_type = "WAF"
  is_published  = false
  rule_arns     = [aws_networksecuritymanager_rule.a.arn]
}

resource "aws_networksecuritymanager_policy" "test" {
  name          = %[1]q
  firewall_type = "WAF"
  priority      = 1

  policy_configuration {
    remediation_enabled = false
    resources_clean_up  = false

    waf_config {
      conflict_resolution                  = "MERGE_WHERE_APPLICABLE"
      existing_customer_web_acl_resolution = "NO_REMEDIATION"
    }
  }

  associated_template_and_rule {
    template_arn = aws_networksecuritymanager_template.draft.arn
  }
}
`, rName))
}

func testAccPolicyConfig_draftQualifiedRule(rName string) string {
	return acctest.ConfigCompose(testAccTemplateConfig_base(rName), fmt.Sprintf(`
resource "aws_networksecuritymanager_policy" "test" {
  name          = %[1]q
  firewall_type = "WAF"
  priority      = 1

  policy_configuration {
    remediation_enabled = false
    resources_clean_up  = false

    waf_config {
      conflict_resolution                  = "MERGE_WHERE_APPLICABLE"
      existing_customer_web_acl_resolution = "NO_REMEDIATION"
    }
  }

  associated_template_and_rule {
    rule_arn = "${aws_networksecuritymanager_rule.a.arn}:DRAFT"
  }
}
`, rName))
}

func testAccPolicyConfig_templateAsRule(rName string) string {
	return acctest.ConfigCompose(testAccPolicyConfig_base(rName), fmt.Sprintf(`
resource "aws_networksecuritymanager_policy" "test" {
  name          = %[1]q
  firewall_type = "WAF"
  priority      = 1

  policy_configuration {
    remediation_enabled = false
    resources_clean_up  = false

    waf_config {
      conflict_resolution                  = "MERGE_WHERE_APPLICABLE"
      existing_customer_web_acl_resolution = "NO_REMEDIATION"
    }
  }

  associated_template_and_rule {
    rule_arn = aws_networksecuritymanager_template.t1.arn
  }
}
`, rName))
}
