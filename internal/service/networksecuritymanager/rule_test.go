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

func testAccRule_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetRuleOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_rule.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("network-security-manager", regexache.MustCompile(`rule:[0-9a-z]+$`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("firewall_type"), tfknownvalue.StringExact(awstypes.RuleFirewallTypeWaf)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("rule_type"), tfknownvalue.StringExact(awstypes.RuleTypeConfiguration)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration), knownvalue.StringExact(`{"DefaultAction":{"Allow":{}}}`)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("is_published"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(awstypes.EntityStatusActive)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("has_published_version"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact("1")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("rule_id"), knownvalue.StringRegexp(regexache.MustCompile(`^[0-9a-z]+$`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("updated_at"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{})),
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrARN: tfknownvalue.RegionalARNRegexp("network-security-manager", regexache.MustCompile(`rule:[0-9a-z]+$`)),
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

func testAccRule_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetRuleOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_rule.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfnetworksecuritymanager.ResourceRule, resourceName),
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

// Every optional argument set on creation; the rule is created as a draft and
// destroyed while it exists only as a draft.
func testAccRule_full(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetRuleOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_rule.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_full(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("network-security-manager", regexache.MustCompile(`rule:[0-9a-z]+$`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("Managed by Terraform")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("is_published"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(awstypes.EntityStatusDraft)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("has_published_version"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact("1")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{
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

// A CONFIGURATION rule holds a single web ACL setting, named by the root key
// of the document; the setting is changed in place. Numeric values are
// returned by the API as strings and must not produce a diff.
func testAccRule_configurationTypes(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetRuleOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_rule.test"

	configurationStep := func(configuration string, want knownvalue.Check, version string) resource.TestStep {
		return resource.TestStep{
			Config: testAccRuleConfig_configuration(rName, awstypes.RuleTypeConfiguration, configuration),
			Check: resource.ComposeAggregateTestCheckFunc(
				testAccCheckRuleExists(ctx, t, resourceName, &v),
			),
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
				},
			},
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration), want),
				statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(awstypes.EntityStatusActive)),
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
		CheckDestroy:             testAccCheckRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_configuration(rName, awstypes.RuleTypeConfiguration, `{"DefaultAction":{"Block":{"CustomResponse":{"ResponseCode":403}}}}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					// The configured document is kept although the API returns "403".
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration), knownvalue.StringExact(`{"DefaultAction":{"Block":{"CustomResponse":{"ResponseCode":403}}}}`)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact("1")),
				},
			},
			configurationStep(
				`{"VisibilityConfig":{"SampledRequestsEnabled":true,"CloudWatchMetricsEnabled":true,"MetricName":"tf-acc-test"}}`,
				tfknownvalue.JSONNoDiff(`{"VisibilityConfig":{"SampledRequestsEnabled":true,"CloudWatchMetricsEnabled":true,"MetricName":"tf-acc-test"}}`),
				"2",
			),
			configurationStep(
				`{"CustomResponseBodies":{"forbidden":{"ContentType":"TEXT_PLAIN","Content":"Forbidden"}}}`,
				tfknownvalue.JSONNoDiff(`{"CustomResponseBodies":{"forbidden":{"ContentType":"TEXT_PLAIN","Content":"Forbidden"}}}`),
				"3",
			),
			// The configured number is kept although the API returns "300".
			configurationStep(
				`{"CaptchaConfig":{"ImmunityTimeProperty":{"ImmunityTime":300}}}`,
				knownvalue.StringExact(`{"CaptchaConfig":{"ImmunityTimeProperty":{"ImmunityTime":300}}}`),
				"4",
			),
			configurationStep(
				`{"TokenDomains":["example.com","example.net"]}`,
				knownvalue.StringExact(`{"TokenDomains":["example.com","example.net"]}`),
				"5",
			),
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

// An INSPECTION rule holds exactly one Firewall Manager rule group, either
// pre-processing or post-processing.
func testAccRule_inspection(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetRuleOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_rule.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_inspectionPreProcess(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("rule_type"), tfknownvalue.StringExact(awstypes.RuleTypeInspection)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(awstypes.EntityStatusActive)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration), tfknownvalue.JSONNoDiff(`{"PreProcessFirewallManagerRuleGroups":[{"Name":"AWSManagedRulesCommonRuleSet","FirewallManagerStatement":{"ManagedRuleGroupStatement":{"VendorName":"AWS","Name":"AWSManagedRulesCommonRuleSet"}},"OverrideAction":{"None":{}},"VisibilityConfig":{"SampledRequestsEnabled":true,"CloudWatchMetricsEnabled":true,"MetricName":"AWSManagedRulesCommonRuleSet"}}]}`)),
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
				Config: testAccRuleConfig_inspectionPostProcess(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact("2")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration), tfknownvalue.JSONNoDiff(`{"PostProcessFirewallManagerRuleGroups":[{"Name":"AWSManagedRulesKnownBadInputsRuleSet","FirewallManagerStatement":{"ManagedRuleGroupStatement":{"VendorName":"AWS","Name":"AWSManagedRulesKnownBadInputsRuleSet","RuleActionOverrides":[{"Name":"Log4JRCE","ActionToUse":{"Count":{}}}]}},"OverrideAction":{"Count":{}},"VisibilityConfig":{"SampledRequestsEnabled":false,"CloudWatchMetricsEnabled":false,"MetricName":"AWSManagedRulesKnownBadInputsRuleSet"}}]}`)),
				},
			},
		},
	})
}

// is_published drives the DRAFT/ACTIVE lifecycle: a draft is published in
// place, and saving a published rule as a draft creates a draft on top of the
// published version (has_published_version) that is published again later.
func testAccRule_publish(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetRuleOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_rule.test"

	publishStep := func(isPublished bool, status awstypes.EntityStatus, hasPublishedVersion bool, version string) resource.TestStep {
		return resource.TestStep{
			Config: testAccRuleConfig_isPublished(rName, isPublished),
			Check: resource.ComposeAggregateTestCheckFunc(
				testAccCheckRuleExists(ctx, t, resourceName, &v),
			),
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
				},
			},
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("is_published"), knownvalue.Bool(isPublished)),
				statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(status)),
				statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("has_published_version"), knownvalue.Bool(hasPublishedVersion)),
				statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact(version)),
				statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("network-security-manager", regexache.MustCompile(`rule:[0-9a-z]+$`))),
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
		CheckDestroy:             testAccCheckRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_isPublished(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
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
			// DRAFT -> ACTIVE: the draft itself is published.
			publishStep(true, awstypes.EntityStatusActive, false, "1"),
			// ACTIVE -> DRAFT: a draft is created on top of the published version.
			publishStep(false, awstypes.EntityStatusDraft, true, "2"),
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			// DRAFT -> ACTIVE: the pending draft replaces the published version.
			publishStep(true, awstypes.EntityStatusActive, false, "2"),
			// ACTIVE -> DRAFT again, so that destroy removes a rule with a pending draft.
			publishStep(false, awstypes.EntityStatusDraft, true, "3"),
		},
	})
}

func testAccRule_description(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetRuleOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_rule.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_description(rName, "description 1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("description 1")),
				},
			},
			{
				Config: testAccRuleConfig_description(rName, "description 2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, t, resourceName, &v),
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
				Config: testAccRuleConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, t, resourceName, &v),
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

// A configuration that differs only in whitespace or key order is planned as
// an in-place update (the string differs) but is not written back: the API is
// not called and the version does not change, and the following plan is empty.
func testAccRule_normalization(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetRuleOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_rule.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_configuration(rName, awstypes.RuleTypeConfiguration, `{"CaptchaConfig":{"ImmunityTimeProperty":{"ImmunityTime":300}}}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, t, resourceName, &v),
				),
			},
			{
				Config: testAccRuleConfig_configurationRaw(rName, "{ \"CaptchaConfig\" : {\n  \"ImmunityTimeProperty\" : { \"ImmunityTime\" : 300 }\n} }\n"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact("1")),
				},
			},
			{
				Config: testAccRuleConfig_configuration(rName, awstypes.RuleTypeConfiguration, `{"VisibilityConfig":{"SampledRequestsEnabled":true,"CloudWatchMetricsEnabled":true,"MetricName":"tf-acc-test"}}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, t, resourceName, &v),
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
				Config: testAccRuleConfig_configurationRaw(rName, "{\"VisibilityConfig\": {\"MetricName\": \"tf-acc-test\", \"CloudWatchMetricsEnabled\": true, \"SampledRequestsEnabled\": true}}"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, t, resourceName, &v),
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
		},
	})
}

// The rule is modified outside Terraform between two applies; the update token
// is read back, so the second apply succeeds and reverts the change.
func testAccRule_updateToken(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetRuleOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_rule.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_description(rName, "description 1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, t, resourceName, &v),
				),
			},
			{
				PreConfig: func() {
					testAccRuleUpdateOutOfBand(ctx, t, &v, "changed outside Terraform", true)
				},
				Config: testAccRuleConfig_descriptionAndConfiguration(rName, "description 1", `{"DefaultAction":{"Block":{}}}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("description 1")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration), knownvalue.StringExact(`{"DefaultAction":{"Block":{}}}`)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact("3")),
				},
			},
			{
				PreConfig: func() {
					// A draft saved outside Terraform on a published rule is
					// published by the next apply.
					testAccRuleUpdateOutOfBand(ctx, t, &v, "draft saved outside Terraform", false)
				},
				Config: testAccRuleConfig_descriptionAndConfiguration(rName, "description 1", `{"DefaultAction":{"Block":{}}}`),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, t, resourceName, &v),
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

// name and rule_type cannot be changed in place.
func testAccRule_replace(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2, v3 networksecuritymanager.GetRuleOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_rule.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, t, resourceName, &v1),
				),
			},
			{
				Config: testAccRuleConfig_basic(rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, t, resourceName, &v2),
					testAccCheckRuleRecreated(&v1, &v2),
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
				Config: testAccRuleConfig_inspectionPreProcess(rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, t, resourceName, &v3),
					testAccCheckRuleRecreated(&v2, &v3),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("rule_type"), tfknownvalue.StringExact(awstypes.RuleTypeInspection)),
				},
			},
		},
	})
}

// Delete followed by an immediate re-creation under the same name.
func testAccRule_recreate(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 networksecuritymanager.GetRuleOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_rule.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, t, resourceName, &v1),
				),
			},
			{
				Taint:  []string{resourceName},
				Config: testAccRuleConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckRuleExists(ctx, t, resourceName, &v2),
					testAccCheckRuleRecreated(&v1, &v2),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(awstypes.EntityStatusActive)),
				},
			},
		},
	})
}

// Argument validation at plan time, and configuration validation by the API.
func testAccRule_validation(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccRuleConfig_basic("-" + rName),
				ExpectError: regexache.MustCompile(`must start with an alphanumeric character`),
			},
			{
				Config:      testAccRuleConfig_basic(strings.Repeat("a", 129)),
				ExpectError: regexache.MustCompile(`string length must be between 1 and 128`),
			},
			{
				Config:      testAccRuleConfig_description(rName, strings.Repeat("a", 257)),
				ExpectError: regexache.MustCompile(`string length must be at most 256`),
			},
			{
				Config:      testAccRuleConfig_description(rName, "not # allowed"),
				ExpectError: regexache.MustCompile(`must contain only alphanumeric characters`),
			},
			{
				Config:      testAccRuleConfig_types(rName, "SHIELD_ADVANCED", string(awstypes.RuleTypeConfiguration)),
				ExpectError: regexache.MustCompile(`Invalid String Enum Value`),
			},
			{
				Config:      testAccRuleConfig_types(rName, string(awstypes.RuleFirewallTypeWaf), "BOGUS"),
				ExpectError: regexache.MustCompile(`Invalid String Enum Value`),
			},
			{
				Config:      testAccRuleConfig_configurationRaw(rName, "not json"),
				ExpectError: regexache.MustCompile(`Invalid JSON String Value`),
			},
			{
				Config:      testAccRuleConfig_configuration(rName, awstypes.RuleTypeConfiguration, `{"DefaultAction":{"Allow":{}},"VisibilityConfig":{"SampledRequestsEnabled":true,"CloudWatchMetricsEnabled":true,"MetricName":"tf-acc-test"}}`),
				ExpectError: regexache.MustCompile(`must contain exactly one root field`),
			},
			{
				Config:      testAccRuleConfig_configuration(rName, awstypes.RuleTypeInspection, `{"DefaultAction":{"Allow":{}}}`),
				ExpectError: regexache.MustCompile(`RuleType INSPECTION requires\s+PreProcessFirewallManagerRuleGroups`),
			},
			{
				Config:      testAccRuleConfig_configuration(rName, awstypes.RuleTypeConfiguration, `{"VisibilityConfig":{"SampledRequestsEnabled":true}}`),
				ExpectError: regexache.MustCompile(`Rule configuration validation failed`),
			},
		},
	})
}

func testAccCheckRuleDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).NetworkSecurityManagerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networksecuritymanager_rule" {
				continue
			}

			_, err := tfnetworksecuritymanager.FindRuleByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Network Security Manager Rule %s still exists", rs.Primary.Attributes[names.AttrARN])
		}

		return nil
	}
}

func testAccCheckRuleExists(ctx context.Context, t *testing.T, n string, v *networksecuritymanager.GetRuleOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).NetworkSecurityManagerClient(ctx)

		output, err := tfnetworksecuritymanager.FindRuleByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckRuleRecreated(before, after *networksecuritymanager.GetRuleOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.RuleId), aws.ToString(after.RuleId); before == after {
			return fmt.Errorf("Network Security Manager Rule (%s) not recreated", before)
		}

		return nil
	}
}

// testAccRuleUpdateOutOfBand changes the rule's description through the API,
// invalidating the update token Terraform last saw. isPublished selects whether
// the change is published or saved as a draft.
func testAccRuleUpdateOutOfBand(ctx context.Context, t *testing.T, v *networksecuritymanager.GetRuleOutput, description string, isPublished bool) {
	conn := acctest.ProviderMeta(ctx, t).NetworkSecurityManagerClient(ctx)

	input := networksecuritymanager.UpdateRuleInput{
		RuleIdentifier:  v.RuleArn,
		UpdateToken:     v.UpdateToken,
		IsPublished:     aws.Bool(isPublished),
		RuleDescription: aws.String(description),
	}

	if _, err := conn.UpdateRule(ctx, &input); err != nil {
		t.Fatalf("updating Network Security Manager Rule (%s) outside Terraform: %s", aws.ToString(v.RuleArn), err)
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).NetworkSecurityManagerClient(ctx)

	input := networksecuritymanager.ListRulesInput{
		MaxResults: aws.Int32(1),
	}
	_, err := conn.ListRules(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccRuleConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_networksecuritymanager_rule" "test" {
  name          = %[1]q
  firewall_type = "WAF"
  rule_type     = "CONFIGURATION"
  configuration = jsonencode({
    DefaultAction = {
      Allow = {}
    }
  })
}
`, rName)
}

func testAccRuleConfig_full(rName string) string {
	return fmt.Sprintf(`
resource "aws_networksecuritymanager_rule" "test" {
  name          = %[1]q
  description   = "Managed by Terraform"
  firewall_type = "WAF"
  rule_type     = "CONFIGURATION"
  is_published  = false
  configuration = jsonencode({
    DefaultAction = {
      Allow = {}
    }
  })

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, acctest.CtKey1, acctest.CtValue1)
}

func testAccRuleConfig_configuration(rName string, ruleType awstypes.RuleType, configuration string) string {
	return fmt.Sprintf(`
resource "aws_networksecuritymanager_rule" "test" {
  name          = %[1]q
  firewall_type = "WAF"
  rule_type     = %[2]q
  configuration = jsonencode(%[3]s)
}
`, rName, ruleType, configuration)
}

func testAccRuleConfig_configurationRaw(rName string, configuration string) string {
	return fmt.Sprintf(`
resource "aws_networksecuritymanager_rule" "test" {
  name          = %[1]q
  firewall_type = "WAF"
  rule_type     = "CONFIGURATION"
  configuration = %[2]q
}
`, rName, configuration)
}

func testAccRuleConfig_isPublished(rName string, isPublished bool) string {
	return fmt.Sprintf(`
resource "aws_networksecuritymanager_rule" "test" {
  name          = %[1]q
  firewall_type = "WAF"
  rule_type     = "CONFIGURATION"
  is_published  = %[2]t
  configuration = jsonencode({
    DefaultAction = {
      Allow = {}
    }
  })
}
`, rName, isPublished)
}

func testAccRuleConfig_description(rName, description string) string {
	return testAccRuleConfig_descriptionAndConfiguration(rName, description, `{"DefaultAction":{"Allow":{}}}`)
}

func testAccRuleConfig_descriptionAndConfiguration(rName, description, configuration string) string {
	return fmt.Sprintf(`
resource "aws_networksecuritymanager_rule" "test" {
  name          = %[1]q
  description   = %[2]q
  firewall_type = "WAF"
  rule_type     = "CONFIGURATION"
  configuration = jsonencode(%[3]s)
}
`, rName, description, configuration)
}

func testAccRuleConfig_types(rName, firewallType, ruleType string) string {
	return fmt.Sprintf(`
resource "aws_networksecuritymanager_rule" "test" {
  name          = %[1]q
  firewall_type = %[2]q
  rule_type     = %[3]q
  configuration = jsonencode({
    DefaultAction = {
      Allow = {}
    }
  })
}
`, rName, firewallType, ruleType)
}

func testAccRuleConfig_inspectionPreProcess(rName string) string {
	return fmt.Sprintf(`
resource "aws_networksecuritymanager_rule" "test" {
  name          = %[1]q
  firewall_type = "WAF"
  rule_type     = "INSPECTION"
  configuration = jsonencode({
    PreProcessFirewallManagerRuleGroups = [{
      Name = "AWSManagedRulesCommonRuleSet"
      FirewallManagerStatement = {
        ManagedRuleGroupStatement = {
          VendorName = "AWS"
          Name       = "AWSManagedRulesCommonRuleSet"
        }
      }
      OverrideAction = {
        None = {}
      }
      VisibilityConfig = {
        SampledRequestsEnabled   = true
        CloudWatchMetricsEnabled = true
        MetricName               = "AWSManagedRulesCommonRuleSet"
      }
    }]
  })
}
`, rName)
}

func testAccRuleConfig_inspectionPostProcess(rName string) string {
	return fmt.Sprintf(`
resource "aws_networksecuritymanager_rule" "test" {
  name          = %[1]q
  firewall_type = "WAF"
  rule_type     = "INSPECTION"
  configuration = jsonencode({
    PostProcessFirewallManagerRuleGroups = [{
      Name = "AWSManagedRulesKnownBadInputsRuleSet"
      FirewallManagerStatement = {
        ManagedRuleGroupStatement = {
          VendorName = "AWS"
          Name       = "AWSManagedRulesKnownBadInputsRuleSet"
          RuleActionOverrides = [{
            Name = "Log4JRCE"
            ActionToUse = {
              Count = {}
            }
          }]
        }
      }
      OverrideAction = {
        Count = {}
      }
      VisibilityConfig = {
        SampledRequestsEnabled   = false
        CloudWatchMetricsEnabled = false
        MetricName               = "AWSManagedRulesKnownBadInputsRuleSet"
      }
    }]
  })
}
`, rName)
}
