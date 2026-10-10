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

func testAccTemplate_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetTemplateOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_template.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTemplateConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTemplateExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "rule_arns.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "rule_arns.0", "aws_networksecuritymanager_rule.a", names.AttrARN),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("network-security-manager", regexache.MustCompile(`template:[0-9a-z]+$`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("firewall_type"), tfknownvalue.StringExact(awstypes.TemplateFirewallTypeWaf)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("is_published"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(awstypes.EntityStatusActive)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("has_published_version"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact("1")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("template_id"), knownvalue.StringRegexp(regexache.MustCompile(`^[0-9a-z]+$`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("updated_at"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTagsAll), knownvalue.MapExact(map[string]knownvalue.Check{})),
					statecheck.ExpectIdentity(resourceName, map[string]knownvalue.Check{
						names.AttrARN: tfknownvalue.RegionalARNRegexp("network-security-manager", regexache.MustCompile(`template:[0-9a-z]+$`)),
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

func testAccTemplate_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetTemplateOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_template.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTemplateConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTemplateExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfnetworksecuritymanager.ResourceTemplate, resourceName),
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

// Every optional argument set on creation; the template is created as a draft
// and destroyed while it exists only as a draft.
func testAccTemplate_full(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetTemplateOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_template.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTemplateConfig_full(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTemplateExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "rule_arns.#", "2"),
					resource.TestCheckResourceAttrPair(resourceName, "rule_arns.0", "aws_networksecuritymanager_rule.a", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "rule_arns.1", "aws_networksecuritymanager_rule.b", names.AttrARN),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("network-security-manager", regexache.MustCompile(`template:[0-9a-z]+$`))),
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

// rule_arns is an ordered list: rules are added, reordered and removed in
// place, and a change of order alone is an update that creates a new version.
func testAccTemplate_rules(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetTemplateOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_template.test"

	rulesStep := func(ruleARNs string, version string, checks ...resource.TestCheckFunc) resource.TestStep {
		return resource.TestStep{
			Config: testAccTemplateConfig_rules(rName, ruleARNs),
			Check: resource.ComposeAggregateTestCheckFunc(
				append([]resource.TestCheckFunc{testAccCheckTemplateExists(ctx, t, resourceName, &v)}, checks...)...,
			),
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
				},
			},
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(awstypes.EntityStatusActive)),
				statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact(version)),
			},
		}
	}
	importStep := resource.TestStep{
		ResourceName:                         resourceName,
		ImportState:                          true,
		ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
		ImportStateVerify:                    true,
		ImportStateVerifyIdentifierAttribute: names.AttrARN,
	}

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTemplateConfig_rules(rName, "aws_networksecuritymanager_rule.a.arn"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTemplateExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "rule_arns.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "rule_arns.0", "aws_networksecuritymanager_rule.a", names.AttrARN),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact("1")),
				},
			},
			// Add a second CONFIGURATION rule.
			rulesStep("aws_networksecuritymanager_rule.a.arn, aws_networksecuritymanager_rule.b.arn", "2",
				resource.TestCheckResourceAttr(resourceName, "rule_arns.#", "2"),
				resource.TestCheckResourceAttrPair(resourceName, "rule_arns.0", "aws_networksecuritymanager_rule.a", names.AttrARN),
				resource.TestCheckResourceAttrPair(resourceName, "rule_arns.1", "aws_networksecuritymanager_rule.b", names.AttrARN),
			),
			importStep,
			// Add an INSPECTION rule: CONFIGURATION and INSPECTION rules mix.
			rulesStep("aws_networksecuritymanager_rule.a.arn, aws_networksecuritymanager_rule.b.arn, aws_networksecuritymanager_rule.c.arn", "3",
				resource.TestCheckResourceAttr(resourceName, "rule_arns.#", "3"),
				resource.TestCheckResourceAttrPair(resourceName, "rule_arns.0", "aws_networksecuritymanager_rule.a", names.AttrARN),
				resource.TestCheckResourceAttrPair(resourceName, "rule_arns.1", "aws_networksecuritymanager_rule.b", names.AttrARN),
				resource.TestCheckResourceAttrPair(resourceName, "rule_arns.2", "aws_networksecuritymanager_rule.c", names.AttrARN),
			),
			// Same rules, different order: the order is kept by the API.
			rulesStep("aws_networksecuritymanager_rule.c.arn, aws_networksecuritymanager_rule.a.arn, aws_networksecuritymanager_rule.b.arn", "4",
				resource.TestCheckResourceAttr(resourceName, "rule_arns.#", "3"),
				resource.TestCheckResourceAttrPair(resourceName, "rule_arns.0", "aws_networksecuritymanager_rule.c", names.AttrARN),
				resource.TestCheckResourceAttrPair(resourceName, "rule_arns.1", "aws_networksecuritymanager_rule.a", names.AttrARN),
				resource.TestCheckResourceAttrPair(resourceName, "rule_arns.2", "aws_networksecuritymanager_rule.b", names.AttrARN),
			),
			importStep,
			// Remove two rules.
			rulesStep("aws_networksecuritymanager_rule.b.arn", "5",
				resource.TestCheckResourceAttr(resourceName, "rule_arns.#", "1"),
				resource.TestCheckResourceAttrPair(resourceName, "rule_arns.0", "aws_networksecuritymanager_rule.b", names.AttrARN),
			),
			// Replace the only rule with another.
			rulesStep("aws_networksecuritymanager_rule.c.arn", "6",
				resource.TestCheckResourceAttr(resourceName, "rule_arns.#", "1"),
				resource.TestCheckResourceAttrPair(resourceName, "rule_arns.0", "aws_networksecuritymanager_rule.c", names.AttrARN),
			),
		},
	})
}

// is_published drives the DRAFT/ACTIVE lifecycle: a draft is published in
// place, and saving a published template as a draft creates a draft on top of
// the published version (has_published_version) that is published again later.
func testAccTemplate_publish(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetTemplateOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_template.test"

	publishStep := func(isPublished bool, status awstypes.EntityStatus, hasPublishedVersion bool, version string) resource.TestStep {
		return resource.TestStep{
			Config: testAccTemplateConfig_isPublished(rName, isPublished),
			Check: resource.ComposeAggregateTestCheckFunc(
				testAccCheckTemplateExists(ctx, t, resourceName, &v),
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
				statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("network-security-manager", regexache.MustCompile(`template:[0-9a-z]+$`))),
			},
		}
	}
	importStep := resource.TestStep{
		ResourceName:                         resourceName,
		ImportState:                          true,
		ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
		ImportStateVerify:                    true,
		ImportStateVerifyIdentifierAttribute: names.AttrARN,
	}

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTemplateConfig_isPublished(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTemplateExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("is_published"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrStatus), tfknownvalue.StringExact(awstypes.EntityStatusDraft)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("has_published_version"), knownvalue.Bool(false)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact("1")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), tfknownvalue.RegionalARNRegexp("network-security-manager", regexache.MustCompile(`template:[0-9a-z]+$`))),
				},
			},
			importStep,
			// Publishing a draft-only template keeps its version.
			publishStep(true, awstypes.EntityStatusActive, false, "1"),
			importStep,
			// A draft on top of the published template.
			publishStep(false, awstypes.EntityStatusDraft, true, "2"),
			importStep,
			publishStep(true, awstypes.EntityStatusActive, false, "2"),
			// Destroyed with a pending draft.
			publishStep(false, awstypes.EntityStatusDraft, true, "3"),
		},
	})
}

func testAccTemplate_description(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetTemplateOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_template.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTemplateConfig_description(rName, "description 1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTemplateExists(ctx, t, resourceName, &v),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("description 1")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact("1")),
				},
			},
			{
				Config: testAccTemplateConfig_description(rName, "description 2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTemplateExists(ctx, t, resourceName, &v),
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
				Config: testAccTemplateConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTemplateExists(ctx, t, resourceName, &v),
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

// The template is modified outside Terraform between two applies; the update
// token is read back, so the second apply succeeds and reverts the change.
func testAccTemplate_updateToken(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetTemplateOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_template.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTemplateConfig_description(rName, "description 1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTemplateExists(ctx, t, resourceName, &v),
				),
			},
			{
				PreConfig: func() {
					testAccTemplateUpdateOutOfBand(ctx, t, &v, "changed outside Terraform")
				},
				Config: testAccTemplateConfig_descriptionAndRules(rName, "description 1", "aws_networksecuritymanager_rule.a.arn, aws_networksecuritymanager_rule.b.arn"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTemplateExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "rule_arns.#", "2"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("description 1")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrVersion), knownvalue.StringExact("3")),
				},
			},
			{
				PreConfig: func() {
					// A draft saved outside Terraform on a published template
					// is published by the next apply.
					testAccTemplateUpdateOutOfBand(ctx, t, &v, "draft saved outside Terraform")
				},
				Config: testAccTemplateConfig_descriptionAndRules(rName, "description 1", "aws_networksecuritymanager_rule.a.arn, aws_networksecuritymanager_rule.b.arn"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTemplateExists(ctx, t, resourceName, &v),
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

// name cannot be changed in place.
func testAccTemplate_replace(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 networksecuritymanager.GetTemplateOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_template.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTemplateConfig_name(rName, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTemplateExists(ctx, t, resourceName, &v1),
				),
			},
			{
				Config: testAccTemplateConfig_name(rName, rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTemplateExists(ctx, t, resourceName, &v2),
					testAccCheckTemplateRecreated(&v1, &v2),
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
		},
	})
}

// Delete followed by an immediate re-creation under the same name.
func testAccTemplate_recreate(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 networksecuritymanager.GetTemplateOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_template.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTemplateConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTemplateExists(ctx, t, resourceName, &v1),
				),
			},
			{
				Taint:  []string{resourceName},
				Config: testAccTemplateConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTemplateExists(ctx, t, resourceName, &v2),
					testAccCheckTemplateRecreated(&v1, &v2),
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

// A rule referenced by a published template cannot be deleted. Replacing such
// a rule therefore fails in the default order (destroy, then create) and
// requires create_before_destroy on the rule.
func testAccTemplate_ruleReplace(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetTemplateOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_template.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccTemplateConfig_ruleName(rName, rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTemplateExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "rule_arns.0", "aws_networksecuritymanager_rule.a", names.AttrARN),
				),
			},
			{
				Config:      testAccTemplateConfig_ruleName(rName, rNameUpdated, false),
				ExpectError: regexache.MustCompile(`ConflictException: The resource is currently in use`),
			},
			{
				Config: testAccTemplateConfig_ruleName(rName, rNameUpdated, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTemplateExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "rule_arns.0", "aws_networksecuritymanager_rule.a", names.AttrARN),
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
		},
	})
}

// Argument validation at plan time, and rule reference validation by the API.
func testAccTemplate_validation(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	fakeARN := func(n int) string {
		return fmt.Sprintf("%q", fmt.Sprintf("arn:%s:network-security-manager:%s:123456789012:rule:tfacc%021d", acctest.Partition(), acctest.Region(), n))
	}
	fakeARNs := func(count int) string {
		arns := make([]string, 0, count)
		for n := range count {
			arns = append(arns, fakeARN(n))
		}
		return strings.Join(arns, ", ")
	}

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				// The rule resource validates its name the same way, so the
				// template is named on its own to exercise its own validator.
				Config:      testAccTemplateConfig_ruleARNs("-"+rName, fakeARN(1)),
				ExpectError: regexache.MustCompile(`must start with an alphanumeric character`),
			},
			{
				Config:      testAccTemplateConfig_ruleARNs(strings.Repeat("a", 129), fakeARN(1)),
				ExpectError: regexache.MustCompile(`string length must be between 1 and 128`),
			},
			{
				Config:      testAccTemplateConfig_description(rName, strings.Repeat("a", 257)),
				ExpectError: regexache.MustCompile(`string length must be at most 256`),
			},
			{
				Config:      testAccTemplateConfig_description(rName, "not # allowed"),
				ExpectError: regexache.MustCompile(`must contain only alphanumeric characters`),
			},
			{
				Config:      testAccTemplateConfig_firewallType(rName, "SHIELD_ADVANCED"),
				ExpectError: regexache.MustCompile(`Invalid String Enum Value`),
			},
			{
				Config:      testAccTemplateConfig_ruleARNs(rName, ""),
				ExpectError: regexache.MustCompile(`list must contain at least 1 elements and at most 50\s+elements`),
			},
			{
				Config:      testAccTemplateConfig_ruleARNs(rName, fakeARNs(51)),
				ExpectError: regexache.MustCompile(`list must contain at least 1 elements and at most 50\s+elements`),
			},
			{
				Config:      testAccTemplateConfig_ruleARNs(rName, fakeARN(1)+", "+fakeARN(1)),
				ExpectError: regexache.MustCompile(`This attribute contains duplicate values`),
			},
			{
				Config:      testAccTemplateConfig_ruleARNs(rName, `"not-an-arn"`),
				ExpectError: regexache.MustCompile(`value must be a valid ARN`),
			},
			{
				// A rule in another account is an authorization failure, not a
				// validation one.
				Config:      testAccTemplateConfig_ruleARNs(rName, fakeARN(1)),
				ExpectError: regexache.MustCompile(`UnauthorizedException|AccessDeniedException`),
			},
			{
				// A rule that does not exist.
				Config:      testAccTemplateConfig_missingRule(rName),
				ExpectError: regexache.MustCompile(`ValidationException: One or more resources referenced by this request`),
			},
			{
				// A draft rule cannot be referenced.
				Config:      testAccTemplateConfig_draftRule(rName),
				ExpectError: regexache.MustCompile(`ValidationException: One or more resources referenced by this request`),
			},
		},
	})
}

func testAccCheckTemplateDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).NetworkSecurityManagerClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networksecuritymanager_template" {
				continue
			}

			_, err := tfnetworksecuritymanager.FindTemplateByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Network Security Manager Template %s still exists", rs.Primary.Attributes[names.AttrARN])
		}

		return nil
	}
}

func testAccCheckTemplateExists(ctx context.Context, t *testing.T, n string, v *networksecuritymanager.GetTemplateOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).NetworkSecurityManagerClient(ctx)

		output, err := tfnetworksecuritymanager.FindTemplateByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckTemplateRecreated(before, after *networksecuritymanager.GetTemplateOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.TemplateId), aws.ToString(after.TemplateId); before == after {
			return fmt.Errorf("Network Security Manager Template (%s) not recreated", before)
		}

		return nil
	}
}

// testAccTemplateUpdateOutOfBand changes the template's description through
// the API, invalidating the update token Terraform last saw.
func testAccTemplateUpdateOutOfBand(ctx context.Context, t *testing.T, v *networksecuritymanager.GetTemplateOutput, description string) {
	conn := acctest.ProviderMeta(ctx, t).NetworkSecurityManagerClient(ctx)

	input := networksecuritymanager.UpdateTemplateInput{
		TemplateIdentifier:  v.TemplateArn,
		UpdateToken:         v.UpdateToken,
		IsPublished:         aws.Bool(v.Status != awstypes.EntityStatusDraft),
		TemplateDescription: aws.String(description),
	}
	if strings.HasPrefix(description, "draft") {
		input.IsPublished = aws.Bool(false)
	}

	if _, err := conn.UpdateTemplate(ctx, &input); err != nil {
		t.Fatalf("updating Network Security Manager Template (%s) outside Terraform: %s", aws.ToString(v.TemplateArn), err)
	}
}

// Three published rules: two CONFIGURATION rules and one INSPECTION rule.
func testAccTemplateConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_networksecuritymanager_rule" "a" {
  name          = "%[1]s-a"
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

resource "aws_networksecuritymanager_rule" "c" {
  name          = "%[1]s-c"
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

func testAccTemplateConfig_basic(rName string) string {
	return testAccTemplateConfig_rules(rName, "aws_networksecuritymanager_rule.a.arn")
}

func testAccTemplateConfig_rules(rName, ruleARNs string) string {
	return acctest.ConfigCompose(testAccTemplateConfig_base(rName), fmt.Sprintf(`
resource "aws_networksecuritymanager_template" "test" {
  name          = %[1]q
  firewall_type = "WAF"
  rule_arns     = [%[2]s]
}
`, rName, ruleARNs))
}

func testAccTemplateConfig_full(rName string) string {
	return acctest.ConfigCompose(testAccTemplateConfig_base(rName), fmt.Sprintf(`
resource "aws_networksecuritymanager_template" "test" {
  name          = %[1]q
  description   = "Managed by Terraform"
  firewall_type = "WAF"
  is_published  = false
  rule_arns     = [aws_networksecuritymanager_rule.a.arn, aws_networksecuritymanager_rule.b.arn]

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, acctest.CtKey1, acctest.CtValue1))
}

func testAccTemplateConfig_isPublished(rName string, isPublished bool) string {
	return acctest.ConfigCompose(testAccTemplateConfig_base(rName), fmt.Sprintf(`
resource "aws_networksecuritymanager_template" "test" {
  name          = %[1]q
  firewall_type = "WAF"
  is_published  = %[2]t
  rule_arns     = [aws_networksecuritymanager_rule.a.arn]
}
`, rName, isPublished))
}

func testAccTemplateConfig_description(rName, description string) string {
	return testAccTemplateConfig_descriptionAndRules(rName, description, "aws_networksecuritymanager_rule.a.arn")
}

func testAccTemplateConfig_descriptionAndRules(rName, description, ruleARNs string) string {
	return acctest.ConfigCompose(testAccTemplateConfig_base(rName), fmt.Sprintf(`
resource "aws_networksecuritymanager_template" "test" {
  name          = %[1]q
  description   = %[2]q
  firewall_type = "WAF"
  rule_arns     = [%[3]s]
}
`, rName, description, ruleARNs))
}

func testAccTemplateConfig_firewallType(rName, firewallType string) string {
	return acctest.ConfigCompose(testAccTemplateConfig_base(rName), fmt.Sprintf(`
resource "aws_networksecuritymanager_template" "test" {
  name          = %[1]q
  firewall_type = %[2]q
  rule_arns     = [aws_networksecuritymanager_rule.a.arn]
}
`, rName, firewallType))
}

// The rules are given as literals, with no rule resource.
func testAccTemplateConfig_ruleARNs(rName, ruleARNs string) string {
	return fmt.Sprintf(`
resource "aws_networksecuritymanager_template" "test" {
  name          = %[1]q
  firewall_type = "WAF"
  rule_arns     = [%[2]s]
}
`, rName, ruleARNs)
}

func testAccTemplateConfig_missingRule(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}
data "aws_region" "current" {}

resource "aws_networksecuritymanager_template" "test" {
  name          = %[1]q
  firewall_type = "WAF"
  rule_arns     = ["arn:${data.aws_partition.current.partition}:network-security-manager:${data.aws_region.current.region}:${data.aws_caller_identity.current.account_id}:rule:tfacc000000000000000000001"]
}
`, rName)
}

func testAccTemplateConfig_name(rName, templateName string) string {
	return acctest.ConfigCompose(testAccTemplateConfig_base(rName), fmt.Sprintf(`
resource "aws_networksecuritymanager_template" "test" {
  name          = %[1]q
  firewall_type = "WAF"
  rule_arns     = [aws_networksecuritymanager_rule.a.arn]
}
`, templateName))
}

func testAccTemplateConfig_ruleName(rName, ruleName string, createBeforeDestroy bool) string {
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
    create_before_destroy = %[3]t
  }
}

resource "aws_networksecuritymanager_template" "test" {
  name          = %[1]q
  firewall_type = "WAF"
  rule_arns     = [aws_networksecuritymanager_rule.a.arn]
}
`, rName, ruleName, createBeforeDestroy)
}

func testAccTemplateConfig_draftRule(rName string) string {
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

resource "aws_networksecuritymanager_template" "test" {
  name          = %[1]q
  firewall_type = "WAF"
  rule_arns     = [aws_networksecuritymanager_rule.draft.arn]
}
`, rName)
}
