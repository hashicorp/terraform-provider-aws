// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networksecuritymanager_test

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strconv"
	"testing"

	"github.com/YakDriver/regexache"
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

var checkPolicyARN knownvalue.Check = tfknownvalue.RegionalARNRegexp("network-security-manager", regexache.MustCompile(`policy:[0-9a-z]+$`))

// Policy priorities must be unique within an account. To allow parallel tests, the test configurations
// derive the priority from the random resource name: parseint(substr(sha1(var.rName), 0, 6), 16) + 1.
func testAccPolicyPriority(rName string) int32 {
	sum := sha1.Sum([]byte(rName))
	v, _ := strconv.ParseInt(hex.EncodeToString(sum[:])[:6], 16, 32)

	return int32(v) + 1
}

func TestAccNetworkSecurityManagerPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckPolicy(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Policy/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), checkPolicyARN),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("firewall_type"), tfknownvalue.StringExact(awstypes.PolicyFirewallTypeShieldAdvanced)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("is_published"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"remediation_enabled": knownvalue.Bool(false),
							"resources_clean_up":  knownvalue.Bool(true),
							"waf_config":          knownvalue.ListSizeExact(0),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrPriority), knownvalue.Int32Exact(testAccPolicyPriority(rName))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("template_or_rule"), knownvalue.ListSizeExact(0)),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Policy/basic/"),
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

func TestAccNetworkSecurityManagerPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckPolicy(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Policy/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfnetworksecuritymanager.ResourcePolicy, resourceName),
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

// A WAF policy that references rules created by the test. The update step changes
// description, priority, configuration and the rule list in place.
func TestAccNetworkSecurityManagerPolicy_waf(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_policy.test"
	rule1ResourceName := "aws_networksecuritymanager_rule.test[0]"
	rule2ResourceName := "aws_networksecuritymanager_rule.test[1]"
	priority := int(testAccPolicyPriority(rName))

	variables := func(description string, priority, ruleCount int, remediationEnabled bool, resolution awstypes.ExistingCustomerWebACLResolution) config.Variables {
		return config.Variables{
			acctest.CtRName:                        config.StringVariable(rName),
			"description":                          config.StringVariable(description),
			"priority":                             config.IntegerVariable(priority),
			"rule_count":                           config.IntegerVariable(ruleCount),
			"remediation_enabled":                  config.BoolVariable(remediationEnabled),
			"existing_customer_web_acl_resolution": config.StringVariable(string(resolution)),
		}
	}

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckPolicy(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Policy/waf/"),
				ConfigVariables: variables("initial", priority, 1, false, awstypes.ExistingCustomerWebACLResolutionNoRemediation),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("initial")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("firewall_type"), tfknownvalue.StringExact(awstypes.PolicyFirewallTypeWaf)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"remediation_enabled": knownvalue.Bool(false),
							"resources_clean_up":  knownvalue.Bool(true),
							"waf_config": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"conflict_resolution":                  tfknownvalue.StringExact(awstypes.WAFConflictResolutionOptionsMergeWhereApplicable),
									"existing_customer_web_acl_resolution": tfknownvalue.StringExact(awstypes.ExistingCustomerWebACLResolutionNoRemediation),
								}),
							}),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrPriority), knownvalue.Int32Exact(int32(priority))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("template_or_rule"), knownvalue.ListSizeExact(1)),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("template_or_rule").AtSliceIndex(0).AtMapKey("rule_identifier"), rule1ResourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("template_or_rule").AtSliceIndex(0).AtMapKey("template_identifier"), knownvalue.Null()),
				},
			},
			{
				ConfigDirectory:                      config.StaticDirectory("testdata/Policy/waf/"),
				ConfigVariables:                      variables("initial", priority, 1, false, awstypes.ExistingCustomerWebACLResolutionNoRemediation),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Policy/waf/"),
				ConfigVariables: variables("updated", priority+1, 2, true, awstypes.ExistingCustomerWebACLResolutionRetrofit),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName),
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
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("updated")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration).AtSliceIndex(0).AtMapKey("remediation_enabled"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration).AtSliceIndex(0).AtMapKey("waf_config").AtSliceIndex(0).AtMapKey("existing_customer_web_acl_resolution"), tfknownvalue.StringExact(awstypes.ExistingCustomerWebACLResolutionRetrofit)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrPriority), knownvalue.Int32Exact(int32(priority+1))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("template_or_rule"), knownvalue.ListSizeExact(2)),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("template_or_rule").AtSliceIndex(0).AtMapKey("rule_identifier"), rule1ResourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New("template_or_rule").AtSliceIndex(1).AtMapKey("rule_identifier"), rule2ResourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
		},
	})
}

func TestAccNetworkSecurityManagerPolicy_descriptionRemoved(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckPolicy(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Policy/description/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
					"description":   config.StringVariable("initial"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("initial")),
				},
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Policy/basic/"),
				ConfigVariables: config.Variables{
					acctest.CtRName: config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName),
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
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
				},
			},
		},
	})
}

// is_published drives the DRAFT/ACTIVE lifecycle. The policy is created as a draft,
// published in place, and finally given a pending draft on top of the published
// version, so that destroy removes both.
func TestAccNetworkSecurityManagerPolicy_isPublished(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_policy.test"

	step := func(isPublished, remediationEnabled bool, action plancheck.ResourceActionType) resource.TestStep {
		return resource.TestStep{
			ConfigDirectory: config.StaticDirectory("testdata/Policy/is_published/"),
			ConfigVariables: config.Variables{
				acctest.CtRName:       config.StringVariable(rName),
				"is_published":        config.BoolVariable(isPublished),
				"remediation_enabled": config.BoolVariable(remediationEnabled),
			},
			Check: resource.ComposeAggregateTestCheckFunc(
				testAccCheckPolicyExists(ctx, t, resourceName),
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
				statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), checkPolicyARN),
				statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("is_published"), knownvalue.Bool(isPublished)),
				statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrConfiguration).AtSliceIndex(0).AtMapKey("remediation_enabled"), knownvalue.Bool(remediationEnabled)),
			},
		}
	}

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckPolicy(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			// Created as a draft.
			step(false, false, plancheck.ResourceActionCreate),
			{
				ConfigDirectory: config.StaticDirectory("testdata/Policy/is_published/"),
				ConfigVariables: config.Variables{
					acctest.CtRName:       config.StringVariable(rName),
					"is_published":        config.BoolVariable(false),
					"remediation_enabled": config.BoolVariable(false),
				},
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			// DRAFT -> ACTIVE: the draft itself is published.
			step(true, false, plancheck.ResourceActionUpdate),
			// ACTIVE -> DRAFT: a draft is saved on top of the published version.
			step(false, true, plancheck.ResourceActionUpdate),
		},
	})
}

func testAccCheckPolicyExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		ctx := acctest.NewTestResourceContext(ctx, rs.Type, rs.Primary.Attributes[names.AttrRegion])
		conn := acctest.ProviderMeta(ctx, t).NetworkSecurityManagerClient(ctx)

		_, err := tfnetworksecuritymanager.FindPolicyByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

		return err
	}
}

func testAccCheckPolicyDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networksecuritymanager_policy" {
				continue
			}

			ctx := acctest.NewTestResourceContext(ctx, rs.Type, rs.Primary.Attributes[names.AttrRegion])
			conn := acctest.ProviderMeta(ctx, t).NetworkSecurityManagerClient(ctx)
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

func testAccPreCheckPolicy(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).NetworkSecurityManagerClient(ctx)

	var input networksecuritymanager.ListPoliciesInput
	_, err := conn.ListPolicies(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}
