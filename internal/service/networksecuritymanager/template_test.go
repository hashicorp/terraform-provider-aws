// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networksecuritymanager_test

import (
	"context"
	"fmt"
	"maps"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networksecuritymanager"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networksecuritymanager/types"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/config"
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

var checkTemplateARN knownvalue.Check = tfknownvalue.RegionalARNRegexp("network-security-manager", regexache.MustCompile(`template:[0-9a-z]+$`))

func TestAccNetworkSecurityManagerTemplate_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_template.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTemplate(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Template/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTemplateExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), checkTemplateARN),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("firewall_type"), tfknownvalue.StringExact(awstypes.TemplateFirewallTypeWaf)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("is_published"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("rule_arns").AtSliceIndex(0), "aws_networksecuritymanager_rule.test", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("rule_arns"), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Template/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccNetworkSecurityManagerTemplate_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_template.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTemplate(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Template/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTemplateExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfnetworksecuritymanager.ResourceTemplate, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

// rule_arns is an ordered list. Rules are added, reordered and removed in place,
// and a change of order alone is an update.
func TestAccNetworkSecurityManagerTemplate_rules(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_template.test"

	variables := func(rules ...string) config.Variables {
		return config.Variables{
			acctest.CtRName: config.StringVariable(rName),
			"rules":         testAccTemplateRulesVariable(rules...),
		}
	}
	step := func(action plancheck.ResourceActionType, rules ...string) resource.TestStep {
		return resource.TestStep{
			ConfigDirectory: config.StaticDirectory("testdata/Template/update/"),
			ConfigVariables: variables(rules...),
			Check: resource.ComposeAggregateTestCheckFunc(
				testAccCheckTemplateExists(ctx, t, resourceName),
			),
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(resourceName, action),
				},
				PostApplyPostRefresh: []plancheck.PlanCheck{
					plancheck.ExpectEmptyPlan(),
				},
			},
			ConfigStateChecks: testAccTemplateRuleARNsChecks(resourceName, rules...),
		}
	}
	importStep := func(rules ...string) resource.TestStep {
		return resource.TestStep{
			ConfigDirectory:                      config.StaticDirectory("testdata/Template/update/"),
			ConfigVariables:                      variables(rules...),
			ResourceName:                         resourceName,
			ImportState:                          true,
			ImportStateVerify:                    true,
			ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
			ImportStateVerifyIdentifierAttribute: names.AttrARN,
		}
	}

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTemplate(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			step(plancheck.ResourceActionCreate, "a"),
			// Add a second CONFIGURATION rule.
			step(plancheck.ResourceActionUpdate, "a", "b"),
			importStep("a", "b"),
			// Add an INSPECTION rule: CONFIGURATION and INSPECTION rules mix.
			step(plancheck.ResourceActionUpdate, "a", "b", "c"),
			// Same rules, different order: the service keeps the order.
			step(plancheck.ResourceActionUpdate, "c", "a", "b"),
			importStep("c", "a", "b"),
			// Remove two rules.
			step(plancheck.ResourceActionUpdate, "b"),
			// Replace the only rule with another.
			step(plancheck.ResourceActionUpdate, "c"),
		},
	})
}

func TestAccNetworkSecurityManagerTemplate_description(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_template.test"

	variables := func(description config.Variable) config.Variables {
		return config.Variables{
			acctest.CtRName: config.StringVariable(rName),
			"description":   description,
			"rules":         testAccTemplateRulesVariable("a"),
		}
	}
	step := func(description config.Variable, action plancheck.ResourceActionType, expected knownvalue.Check) resource.TestStep {
		return resource.TestStep{
			ConfigDirectory: config.StaticDirectory("testdata/Template/update/"),
			ConfigVariables: variables(description),
			Check: resource.ComposeAggregateTestCheckFunc(
				testAccCheckTemplateExists(ctx, t, resourceName),
			),
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(resourceName, action),
				},
				PostApplyPostRefresh: []plancheck.PlanCheck{
					plancheck.ExpectEmptyPlan(),
				},
			},
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), expected),
			},
		}
	}

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTemplate(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			step(config.StringVariable("description 1"), plancheck.ResourceActionCreate, knownvalue.StringExact("description 1")),
			step(config.StringVariable("description 2"), plancheck.ResourceActionUpdate, knownvalue.StringExact("description 2")),
			{
				ConfigDirectory:                      config.StaticDirectory("testdata/Template/update/"),
				ConfigVariables:                      variables(config.StringVariable("description 2")),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			// Removing the description from the configuration clears it.
			step(nil, plancheck.ResourceActionUpdate, knownvalue.Null()),
		},
	})
}

// is_published drives the DRAFT/ACTIVE lifecycle. The template is created as a
// draft, published in place, and given a pending draft on top of the published
// version, which is published again; destroy then removes a pending draft too.
func TestAccNetworkSecurityManagerTemplate_isPublished(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_template.test"

	variables := func(isPublished bool, rules ...string) config.Variables {
		return config.Variables{
			acctest.CtRName: config.StringVariable(rName),
			"is_published":  config.BoolVariable(isPublished),
			"rules":         testAccTemplateRulesVariable(rules...),
		}
	}
	step := func(isPublished bool, action plancheck.ResourceActionType, rules ...string) resource.TestStep {
		return resource.TestStep{
			ConfigDirectory: config.StaticDirectory("testdata/Template/update/"),
			ConfigVariables: variables(isPublished, rules...),
			Check: resource.ComposeAggregateTestCheckFunc(
				testAccCheckTemplateExists(ctx, t, resourceName),
				testAccCheckTemplateStatus(ctx, t, resourceName, isPublished),
			),
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(resourceName, action),
				},
				PostApplyPostRefresh: []plancheck.PlanCheck{
					plancheck.ExpectEmptyPlan(),
				},
			},
			ConfigStateChecks: append(testAccTemplateRuleARNsChecks(resourceName, rules...),
				statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), checkTemplateARN),
				statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("is_published"), knownvalue.Bool(isPublished)),
			),
		}
	}

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTemplate(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Created as a draft.
			step(false, plancheck.ResourceActionCreate, "a"),
			{
				ConfigDirectory:                      config.StaticDirectory("testdata/Template/update/"),
				ConfigVariables:                      variables(false, "a"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			// DRAFT -> ACTIVE: the draft itself is published.
			step(true, plancheck.ResourceActionUpdate, "a"),
			// ACTIVE -> DRAFT: a draft is saved on top of the published version.
			step(false, plancheck.ResourceActionUpdate, "a", "b"),
			{
				ConfigDirectory:                      config.StaticDirectory("testdata/Template/update/"),
				ConfigVariables:                      variables(false, "a", "b"),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			step(true, plancheck.ResourceActionUpdate, "a", "b"),
			// Destroyed with a pending draft.
			step(false, plancheck.ResourceActionUpdate, "b"),
		},
	})
}

// The template is changed outside Terraform between applies. The update token is
// read back before each update, so the next apply succeeds and reverts the
// change; a draft saved outside Terraform is published back when the
// configuration sets is_published = true.
func TestAccNetworkSecurityManagerTemplate_updateToken(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetTemplateOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_template.test"

	variables := func(rules ...string) config.Variables {
		return config.Variables{
			acctest.CtRName: config.StringVariable(rName),
			"description":   config.StringVariable("description 1"),
			"rules":         testAccTemplateRulesVariable(rules...),
		}
	}
	publishedVariables := variables("a", "b")
	publishedVariables["is_published"] = config.BoolVariable(true)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTemplate(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Template/update/"),
				ConfigVariables: variables("a"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTemplateExistsGet(ctx, t, resourceName, &v),
				),
			},
			{
				PreConfig: func() {
					testAccTemplateUpdateOutOfBand(ctx, t, &v, "changed outside Terraform", true)
				},
				ConfigDirectory: config.StaticDirectory("testdata/Template/update/"),
				ConfigVariables: variables("a", "b"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTemplateExistsGet(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: append(testAccTemplateRuleARNsChecks(resourceName, "a", "b"),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("description 1")),
				),
			},
			{
				PreConfig: func() {
					testAccTemplateUpdateOutOfBand(ctx, t, &v, "draft saved outside Terraform", false)
				},
				// is_published is Optional+Computed, so only an explicit true
				// publishes a draft that was saved outside Terraform.
				ConfigDirectory: config.StaticDirectory("testdata/Template/update/"),
				ConfigVariables: publishedVariables,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTemplateExistsGet(ctx, t, resourceName, &v),
					testAccCheckTemplateStatus(ctx, t, resourceName, true),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("description 1")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("is_published"), knownvalue.Bool(true)),
				},
			},
		},
	})
}

// name cannot be changed in place.
func TestAccNetworkSecurityManagerTemplate_nameReplace(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 networksecuritymanager.GetTemplateOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_template.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTemplate(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Template/update/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"rules":         testAccTemplateRulesVariable("a"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTemplateExistsGet(ctx, t, resourceName, &v1),
				),
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Template/update/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"template_name": config.StringVariable(rNameUpdated),
					"rules":         testAccTemplateRulesVariable("a"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTemplateExistsGet(ctx, t, resourceName, &v2),
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

// A rule that a published template references cannot be deleted. Replacing such a
// rule therefore fails in the default order (destroy, then create) and needs
// create_before_destroy on the rule.
func TestAccNetworkSecurityManagerTemplate_ruleReplace(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	rNameUpdated := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_template.test"
	ruleResourceName := "aws_networksecuritymanager_rule.a"

	variables := func(ruleName string) config.Variables {
		return config.Variables{
			acctest.CtRName: config.StringVariable(rName),
			"rule_name":     config.StringVariable(ruleName),
		}
	}

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTemplate(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Template/rule_replace/"),
				ConfigVariables: variables(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTemplateExists(ctx, t, resourceName),
				),
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Template/rule_replace/"),
				ConfigVariables: variables(rNameUpdated),
				ExpectError:     regexache.MustCompile(`ConflictException: The resource is currently in use`),
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Template/rule_replace_cbd/"),
				ConfigVariables: variables(rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckTemplateExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(ruleResourceName, plancheck.ResourceActionCreateBeforeDestroy),
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(ruleResourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rNameUpdated)),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("rule_arns").AtSliceIndex(0), ruleResourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
		},
	})
}

// Argument validation at plan time, and rule reference validation by the service.
func TestAccNetworkSecurityManagerTemplate_validation(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	ruleARN := func(accountID string, n int) string {
		return fmt.Sprintf("arn:%s:network-security-manager:%s:%s:rule:tfacc%021d", acctest.Partition(), acctest.Region(), accountID, n)
	}
	otherAccountRuleARNs := func(count int) config.Variable {
		arns := make([]config.Variable, 0, count)
		for n := range count {
			arns = append(arns, config.StringVariable(ruleARN(acctest.Ct12Digit, n)))
		}
		return config.ListVariable(arns...)
	}
	step := func(variables config.Variables, planOnly bool, expectError string) resource.TestStep {
		v := config.Variables{
			acctest.CtRName: config.StringVariable(rName),
			"rule_arns":     otherAccountRuleARNs(1),
		}
		maps.Copy(v, variables)
		return resource.TestStep{
			ConfigDirectory: config.StaticDirectory("testdata/Template/validation/"),
			ConfigVariables: v,
			PlanOnly:        planOnly,
			ExpectError:     regexache.MustCompile(expectError),
		}
	}
	const (
		invalidLength = `Invalid Attribute Value Length`
		invalidMatch  = `Invalid Attribute Value Match`
		invalidSize   = `list must contain at least 1 elements and at most 50\s+elements`
	)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckTemplate(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTemplateDestroy(ctx, t),
		Steps: []resource.TestStep{
			step(config.Variables{acctest.CtRName: config.StringVariable("-" + rName)}, true, invalidMatch),
			step(config.Variables{acctest.CtRName: config.StringVariable(strings.Repeat("a", 129))}, true, invalidLength),
			step(config.Variables{"description": config.StringVariable(strings.Repeat("a", 257))}, true, invalidLength),
			step(config.Variables{"description": config.StringVariable("not # allowed")}, true, invalidMatch),
			step(config.Variables{"firewall_type": config.StringVariable("SHIELD_ADVANCED")}, true, `Invalid String Enum Value`),
			step(config.Variables{"no_rules": config.BoolVariable(true)}, true, invalidSize),
			step(config.Variables{"rule_arns": otherAccountRuleARNs(51)}, true, invalidSize),
			step(config.Variables{"rule_arns": config.ListVariable(config.StringVariable(ruleARN(acctest.Ct12Digit, 1)), config.StringVariable(ruleARN(acctest.Ct12Digit, 1)))}, true, `This attribute contains duplicate values`),
			step(config.Variables{"rule_arns": config.ListVariable(config.StringVariable("not-an-arn"))}, true, `value must be a valid ARN`),
			// A rule in another account is an authorization failure, not a validation one.
			step(config.Variables{}, false, `UnauthorizedException|AccessDeniedException`),
			// A rule that does not exist.
			{
				ConfigDirectory: config.StaticDirectory("testdata/Template/missing_rule/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ExpectError: regexache.MustCompile(`ValidationException: One or more resources referenced by this request`),
			},
			// A draft rule cannot be referenced.
			{
				ConfigDirectory: config.StaticDirectory("testdata/Template/draft_rule/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				ExpectError: regexache.MustCompile(`ValidationException: One or more resources referenced by this request`),
			},
		},
	})
}

func testAccTemplateRulesVariable(rules ...string) config.Variable {
	v := make([]config.Variable, 0, len(rules))
	for _, rule := range rules {
		v = append(v, config.StringVariable(rule))
	}
	return config.ListVariable(v...)
}

// testAccTemplateRuleARNsChecks checks that rule_arns holds the given rules of
// testdata/Template/update, in order.
func testAccTemplateRuleARNsChecks(resourceName string, rules ...string) []statecheck.StateCheck {
	checks := []statecheck.StateCheck{
		statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("rule_arns"), knownvalue.ListSizeExact(len(rules))),
	}
	for i, rule := range rules {
		checks = append(checks, statecheck.CompareValuePairs(resourceName, tfjsonpath.New("rule_arns").AtSliceIndex(i), "aws_networksecuritymanager_rule."+rule, tfjsonpath.New(names.AttrARN), compare.ValuesSame()))
	}
	return checks
}

func testAccCheckTemplateExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return testAccCheckTemplateExistsGet(ctx, t, n, nil)
}

// testAccCheckTemplateExistsGet also stores the template as the service returns it.
func testAccCheckTemplateExistsGet(ctx context.Context, t *testing.T, n string, v *networksecuritymanager.GetTemplateOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		ctx := acctest.NewTestResourceContext(ctx, rs.Type, rs.Primary.Attributes[names.AttrRegion])
		conn := acctest.ProviderMeta(ctx, t).NetworkSecurityManagerClient(ctx)

		out, err := tfnetworksecuritymanager.FindTemplateByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])
		if err != nil {
			return err
		}

		if v != nil {
			*v = *out
		}

		return nil
	}
}

// testAccCheckTemplateStatus checks the status the service reports: a published
// template is ACTIVE, and an unpublished one has a pending DRAFT.
func testAccCheckTemplateStatus(ctx context.Context, t *testing.T, n string, isPublished bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var v networksecuritymanager.GetTemplateOutput
		if err := testAccCheckTemplateExistsGet(ctx, t, n, &v)(s); err != nil {
			return err
		}

		expected := awstypes.EntityStatusActive
		if !isPublished {
			expected = awstypes.EntityStatusDraft
		}
		if v.Status != expected {
			return fmt.Errorf("Network Security Manager Template %s status is %s, expected %s", aws.ToString(v.TemplateArn), v.Status, expected)
		}

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

func testAccCheckTemplateDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networksecuritymanager_template" {
				continue
			}

			ctx := acctest.NewTestResourceContext(ctx, rs.Type, rs.Primary.Attributes[names.AttrRegion])
			conn := acctest.ProviderMeta(ctx, t).NetworkSecurityManagerClient(ctx)
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

// testAccTemplateUpdateOutOfBand changes the template's description through the
// API, which invalidates the update token Terraform last saw.
func testAccTemplateUpdateOutOfBand(ctx context.Context, t *testing.T, v *networksecuritymanager.GetTemplateOutput, description string, isPublished bool) {
	t.Helper()

	conn := acctest.ProviderMeta(ctx, t).NetworkSecurityManagerClient(ctx)

	input := networksecuritymanager.UpdateTemplateInput{
		IsPublished:         aws.Bool(isPublished),
		TemplateDescription: aws.String(description),
		TemplateIdentifier:  v.TemplateArn,
		UpdateToken:         v.UpdateToken,
	}
	if _, err := conn.UpdateTemplate(ctx, &input); err != nil {
		t.Fatalf("updating Network Security Manager Template (%s) outside Terraform: %s", aws.ToString(v.TemplateArn), err)
	}
}

func testAccPreCheckTemplate(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).NetworkSecurityManagerClient(ctx)

	var input networksecuritymanager.ListTemplatesInput
	_, err := conn.ListTemplates(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}
