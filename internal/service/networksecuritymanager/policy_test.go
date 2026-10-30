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

const (
	// Until aws_networksecuritymanager_rule exists, WAF policy tests reference pre-existing published rules.
	envVarRuleARN1 = "NETWORKSECURITYMANAGER_RULE_ARN_1"
	envVarRuleARN2 = "NETWORKSECURITYMANAGER_RULE_ARN_2"
)

var checkPolicyARN knownvalue.Check = tfknownvalue.RegionalARNRegexp("network-security-manager", regexache.MustCompile(`policy:.+`))

// Policy priorities must be unique within an account. To allow parallel tests, the test configurations
// derive the priority from the random resource name: parseint(substr(sha1(var.rName), 0, 6), 16) + 1.
func testAccPolicyPriority(rName string) int32 {
	sum := sha1.Sum([]byte(rName))
	v, _ := strconv.ParseInt(hex.EncodeToString(sum[:])[:6], 16, 32)

	return int32(v) + 1
}

func TestAccNetworkSecurityManagerPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetPolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
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
					testAccCheckPolicyExists(ctx, t, resourceName, &v),
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
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_configuration"), knownvalue.ListExact([]knownvalue.Check{
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
	var v networksecuritymanager.GetPolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
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
					testAccCheckPolicyExists(ctx, t, resourceName, &v),
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

func TestAccNetworkSecurityManagerPolicy_waf(t *testing.T) {
	ctx := acctest.Context(t)
	var v networksecuritymanager.GetPolicyOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networksecuritymanager_policy.test"
	ruleARN1 := acctest.SkipIfEnvVarNotSet(t, envVarRuleARN1)
	ruleARN2 := acctest.SkipIfEnvVarNotSet(t, envVarRuleARN2)
	priority := int(testAccPolicyPriority(rName))

	variables := func(description string, priority int, remediationEnabled bool, resolution awstypes.ExistingCustomerWebACLResolution, ruleARNs ...string) config.Variables {
		arns := make([]config.Variable, 0, len(ruleARNs))
		for _, arn := range ruleARNs {
			arns = append(arns, config.StringVariable(arn))
		}

		return config.Variables{
			acctest.CtRName:                        config.StringVariable(rName),
			"description":                          config.StringVariable(description),
			"priority":                             config.IntegerVariable(priority),
			"rule_arns":                            config.ListVariable(arns...),
			"remediation_enabled":                  config.BoolVariable(remediationEnabled),
			"existing_customer_web_acl_resolution": config.StringVariable(string(resolution)),
		}
	}

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkSecurityManagerServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPolicyDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.StaticDirectory("testdata/Policy/waf/"),
				ConfigVariables: variables("initial", priority, false, awstypes.ExistingCustomerWebACLResolutionNoRemediation, ruleARN1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("initial")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("firewall_type"), tfknownvalue.StringExact(awstypes.PolicyFirewallTypeWaf)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_configuration"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"remediation_enabled": knownvalue.Bool(false),
							"resources_clean_up":  knownvalue.Bool(false),
							"waf_config": knownvalue.ListExact([]knownvalue.Check{
								knownvalue.ObjectExact(map[string]knownvalue.Check{
									"conflict_resolution":                  tfknownvalue.StringExact(awstypes.WAFConflictResolutionOptionsMergeWhereApplicable),
									"existing_customer_web_acl_resolution": tfknownvalue.StringExact(awstypes.ExistingCustomerWebACLResolutionNoRemediation),
								}),
							}),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrPriority), knownvalue.Int32Exact(int32(priority))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("template_or_rule"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"rule_arn":     knownvalue.StringExact(ruleARN1),
							"template_arn": knownvalue.Null(),
						}),
					})),
				},
			},
			{
				ConfigDirectory:                      config.StaticDirectory("testdata/Policy/waf/"),
				ConfigVariables:                      variables("initial", priority, false, awstypes.ExistingCustomerWebACLResolutionNoRemediation, ruleARN1),
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
			{
				ConfigDirectory: config.StaticDirectory("testdata/Policy/waf/"),
				ConfigVariables: variables("updated", priority+1, true, awstypes.ExistingCustomerWebACLResolutionRetrofit, ruleARN1, ruleARN2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckPolicyExists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("updated")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_configuration").AtSliceIndex(0).AtMapKey("remediation_enabled"), knownvalue.Bool(true)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("policy_configuration").AtSliceIndex(0).AtMapKey("waf_config").AtSliceIndex(0).AtMapKey("existing_customer_web_acl_resolution"), tfknownvalue.StringExact(awstypes.ExistingCustomerWebACLResolutionRetrofit)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrPriority), knownvalue.Int32Exact(int32(priority+1))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("template_or_rule"), knownvalue.ListSizeExact(2)),
				},
			},
		},
	})
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

func testAccPreCheck(ctx context.Context, t *testing.T) {
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
