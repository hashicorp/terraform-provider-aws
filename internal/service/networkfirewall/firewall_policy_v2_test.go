// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkfirewall_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/networkfirewall"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkfirewall/types"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfnetworkfirewall "github.com/hashicorp/terraform-provider-aws/internal/service/networkfirewall"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func checkFirewallPolicyV2ARN(name string) knownvalue.Check {
	return tfknownvalue.RegionalARNExact("network-firewall", "firewall-policy-v2/"+name)
}

func checkFirewallPolicyV2ARNAlternateRegion(name string) knownvalue.Check {
	return tfknownvalue.RegionalARNAlternateRegionExact("network-firewall", "firewall-policy-v2/"+name)
}

func TestAccNetworkFirewallFirewallPolicyV2_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v networkfirewall.DescribeFirewallPolicyV2Output
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy_v2.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallPolicyV2Destroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyV2Config_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFirewallPolicyV2Exists(ctx, t, resourceName, &v),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrARN), checkFirewallPolicyV2ARN(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrName), knownvalue.StringExact(rName)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEncryptionConfiguration), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrKeyID: knownvalue.StringExact("AWS_OWNED_KMS_KEY"),
							names.AttrType:  tfknownvalue.StringExact(awstypes.EncryptionTypeAwsOwnedKmsKey),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("firewall_policy"), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("update_token"), knownvalue.NotNull()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
			},
		},
	})
}

func TestAccNetworkFirewallFirewallPolicyV2_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v networkfirewall.DescribeFirewallPolicyV2Output
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy_v2.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallPolicyV2Destroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyV2Config_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFirewallPolicyV2Exists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfnetworkfirewall.ResourceFirewallPolicyV2, resourceName),
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

func TestAccNetworkFirewallFirewallPolicyV2_description(t *testing.T) {
	ctx := acctest.Context(t)
	var v networkfirewall.DescribeFirewallPolicyV2Output
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy_v2.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallPolicyV2Destroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyV2Config_description(rName, "initial"),
				Check:  testAccCheckFirewallPolicyV2Exists(ctx, t, resourceName, &v),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("initial")),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
			},
			{
				Config: testAccFirewallPolicyV2Config_description(rName, "updated"),
				Check:  testAccCheckFirewallPolicyV2Exists(ctx, t, resourceName, &v),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.StringExact("updated")),
				},
			},
			{
				Config: testAccFirewallPolicyV2Config_basic(rName),
				Check:  testAccCheckFirewallPolicyV2Exists(ctx, t, resourceName, &v),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrDescription), knownvalue.Null()),
				},
			},
		},
	})
}

func TestAccNetworkFirewallFirewallPolicyV2_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var v networkfirewall.DescribeFirewallPolicyV2Output
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy_v2.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallPolicyV2Destroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyV2Config_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check:  testAccCheckFirewallPolicyV2Exists(ctx, t, resourceName, &v),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
			},
			{
				Config: testAccFirewallPolicyV2Config_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check:  testAccCheckFirewallPolicyV2Exists(ctx, t, resourceName, &v),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
			{
				Config: testAccFirewallPolicyV2Config_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check:  testAccCheckFirewallPolicyV2Exists(ctx, t, resourceName, &v),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
		},
	})
}

func TestAccNetworkFirewallFirewallPolicyV2_encryptionConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var v networkfirewall.DescribeFirewallPolicyV2Output
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy_v2.test"
	encryptionPath := tfjsonpath.New(names.AttrEncryptionConfiguration).AtSliceIndex(0)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallPolicyV2Destroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyV2Config_encryptionConfigurationCustomerKMS(rName, 0),
				Check:  testAccCheckFirewallPolicyV2Exists(ctx, t, resourceName, &v),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, encryptionPath.AtMapKey(names.AttrType), tfknownvalue.StringExact(awstypes.EncryptionTypeCustomerKms)),
					statecheck.CompareValuePairs(resourceName, encryptionPath.AtMapKey(names.AttrKeyID), "aws_kms_key.test[0]", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
			},
			{
				Config: testAccFirewallPolicyV2Config_encryptionConfigurationCustomerKMS(rName, 1),
				Check:  testAccCheckFirewallPolicyV2Exists(ctx, t, resourceName, &v),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, encryptionPath.AtMapKey(names.AttrType), tfknownvalue.StringExact(awstypes.EncryptionTypeCustomerKms)),
					statecheck.CompareValuePairs(resourceName, encryptionPath.AtMapKey(names.AttrKeyID), "aws_kms_key.test[1]", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
			// Removing the block from configuration while a customer managed key is in use.
			// encryption_configuration is Optional+Computed with UseStateForUnknown, so the
			// prior value is expected to be kept rather than reverting to the AWS owned key.
			{
				Config: testAccFirewallPolicyV2Config_encryptionConfigurationRemoved(rName),
				Check:  testAccCheckFirewallPolicyV2Exists(ctx, t, resourceName, &v),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, encryptionPath.AtMapKey(names.AttrType), tfknownvalue.StringExact(awstypes.EncryptionTypeCustomerKms)),
					statecheck.CompareValuePairs(resourceName, encryptionPath.AtMapKey(names.AttrKeyID), "aws_kms_key.test[1]", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
				},
			},
			{
				Config: testAccFirewallPolicyV2Config_encryptionConfigurationOwnedKey(rName),
				Check:  testAccCheckFirewallPolicyV2Exists(ctx, t, resourceName, &v),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEncryptionConfiguration), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrKeyID: knownvalue.StringExact("AWS_OWNED_KMS_KEY"),
							names.AttrType:  tfknownvalue.StringExact(awstypes.EncryptionTypeAwsOwnedKmsKey),
						}),
					})),
				},
			},
			{
				Config: testAccFirewallPolicyV2Config_encryptionConfigurationRemoved(rName),
				Check:  testAccCheckFirewallPolicyV2Exists(ctx, t, resourceName, &v),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, encryptionPath.AtMapKey(names.AttrType), tfknownvalue.StringExact(awstypes.EncryptionTypeAwsOwnedKmsKey)),
				},
			},
		},
	})
}

func TestAccNetworkFirewallFirewallPolicyV2_encryptionConfigurationKeyID(t *testing.T) {
	ctx := acctest.Context(t)
	var v networkfirewall.DescribeFirewallPolicyV2Output
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy_v2.test"
	encryptionPath := tfjsonpath.New(names.AttrEncryptionConfiguration).AtSliceIndex(0)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallPolicyV2Destroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyV2Config_encryptionConfigurationKeyID(rName),
				Check:  testAccCheckFirewallPolicyV2Exists(ctx, t, resourceName, &v),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, encryptionPath.AtMapKey(names.AttrType), tfknownvalue.StringExact(awstypes.EncryptionTypeCustomerKms)),
					statecheck.CompareValuePairs(resourceName, encryptionPath.AtMapKey(names.AttrKeyID), "aws_kms_key.test", tfjsonpath.New(names.AttrKeyID), compare.ValuesSame()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				// Import reads the key ARN returned by the API; the configured key ID form cannot be recovered.
				ImportStateVerifyIgnore: []string{"encryption_configuration.0.key_id"},
			},
		},
	})
}

func TestAccNetworkFirewallFirewallPolicyV2_full(t *testing.T) {
	ctx := acctest.Context(t)
	var v networkfirewall.DescribeFirewallPolicyV2Output
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy_v2.test"
	policyPath := tfjsonpath.New("firewall_policy").AtSliceIndex(0)
	firewallRulePath := policyPath.AtMapKey("firewall_rule_group_entry").AtSliceIndex(0).AtMapKey("inline_rule_group").AtSliceIndex(0).AtMapKey("rules_source").AtSliceIndex(0).AtMapKey("firewall_rule").AtSliceIndex(0)
	compareUpdateToken := statecheck.CompareValue(compare.ValuesDiffer())

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallPolicyV2Destroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyV2Config_full(rName),
				Check:  testAccCheckFirewallPolicyV2Exists(ctx, t, resourceName, &v),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					compareUpdateToken.AddStateValue(resourceName, tfjsonpath.New("update_token")),
					statecheck.ExpectKnownValue(resourceName, policyPath.AtMapKey("firewall_default_action"), tfknownvalue.StringExact(awstypes.PolicyDefaultActionDrop)),
					statecheck.ExpectKnownValue(resourceName, policyPath.AtMapKey("policy_mode"), tfknownvalue.StringExact(awstypes.PolicyModeEnforce)),
					statecheck.ExpectKnownValue(resourceName, policyPath.AtMapKey("firewall_rule_group_entry"), knownvalue.ListSizeExact(2)),
					statecheck.ExpectKnownValue(resourceName, firewallRulePath, knownvalue.ObjectPartial(map[string]knownvalue.Check{
						names.AttrAction: tfknownvalue.StringExact(awstypes.FirewallRuleActionAccept),
						"action_scope":   tfknownvalue.StringExact(awstypes.ActionScopeFlow),
						"hook":           tfknownvalue.StringExact(awstypes.HookKeyTcpFlowStart),
						"dest_ports":     knownvalue.StringExact("443"),
						"source_ips":     knownvalue.StringExact("$HOME_NET"),
					})),
					statecheck.ExpectKnownValue(resourceName, policyPath.AtMapKey("firewall_rule_group_entry").AtSliceIndex(0).AtMapKey("inline_rule_group").AtSliceIndex(0).AtMapKey("rule_variables").AtSliceIndex(0).AtMapKey("ip_set"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrKey: knownvalue.StringExact("HOME_NET"),
							"definition":  knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("10.0.0.0/16")}),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, policyPath.AtMapKey("firewall_rule_group_entry").AtSliceIndex(0).AtMapKey("inline_rule_group").AtSliceIndex(0).AtMapKey("rule_variables").AtSliceIndex(0).AtMapKey("context_set"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrKey: knownvalue.StringExact("MY_VPCS"),
							"definition":  knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("vpc-0123456789abcdef0")}),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, policyPath.AtMapKey("ips_rule_group_entry"), knownvalue.ListSizeExact(2)),
					statecheck.ExpectKnownValue(resourceName, policyPath.AtMapKey("ips_rule_group_entry").AtSliceIndex(0).AtMapKey("inline_rule_group").AtSliceIndex(0).AtMapKey("rule_variables").AtSliceIndex(0).AtMapKey("context_set"), knownvalue.SetSizeExact(1)),
					statecheck.CompareValuePairs(resourceName, policyPath.AtMapKey("ips_rule_group_entry").AtSliceIndex(0).AtMapKey("inline_rule_group").AtSliceIndex(0).AtMapKey("reference_sets").AtSliceIndex(0).AtMapKey("ip_set_reference").AtSliceIndex(0).AtMapKey("reference_arn"), "aws_ec2_managed_prefix_list.test", tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, policyPath.AtMapKey("ips_rule_group_entry").AtSliceIndex(1).AtMapKey("override").AtSliceIndex(0).AtMapKey(names.AttrAction), tfknownvalue.StringExact(awstypes.OverrideActionDropToAlert)),
					statecheck.ExpectKnownValue(resourceName, policyPath.AtMapKey("policy_variables").AtSliceIndex(0).AtMapKey("rule_variable"), knownvalue.SetSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, policyPath.AtMapKey("stateful_engine_options").AtSliceIndex(0).AtMapKey("flow_timeouts").AtSliceIndex(0).AtMapKey("tcp_idle_timeout_seconds"), knownvalue.Int32Exact(350)),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
			},
			{
				Config: testAccFirewallPolicyV2Config_fullUpdated(rName),
				Check:  testAccCheckFirewallPolicyV2Exists(ctx, t, resourceName, &v),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					compareUpdateToken.AddStateValue(resourceName, tfjsonpath.New("update_token")),
					statecheck.ExpectKnownValue(resourceName, policyPath.AtMapKey("firewall_default_action"), tfknownvalue.StringExact(awstypes.PolicyDefaultActionAlert)),
					statecheck.ExpectKnownValue(resourceName, policyPath.AtMapKey("policy_mode"), tfknownvalue.StringExact(awstypes.PolicyModeMonitor)),
					statecheck.ExpectKnownValue(resourceName, policyPath.AtMapKey("firewall_rule_group_entry"), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, firewallRulePath, knownvalue.ObjectPartial(map[string]knownvalue.Check{
						names.AttrAction: tfknownvalue.StringExact(awstypes.FirewallRuleActionDrop),
						"dest_ports":     knownvalue.StringExact("23"),
						"source_ips":     knownvalue.StringExact("any"),
						"rule_metadata":  knownvalue.ListSizeExact(0),
						"rule_option": knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"keyword":  knownvalue.StringExact("sid"),
								"settings": knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("1000002")}),
							}),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, policyPath.AtMapKey("firewall_rule_group_entry").AtSliceIndex(0).AtMapKey("inline_rule_group").AtSliceIndex(0).AtMapKey("rule_variables"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, policyPath.AtMapKey("ips_rule_group_entry"), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, policyPath.AtMapKey("ips_rule_group_entry").AtSliceIndex(0).AtMapKey("override"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, policyPath.AtMapKey("policy_variables"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, policyPath.AtMapKey("stateful_engine_options").AtSliceIndex(0).AtMapKey("stream_exception_policy"), tfknownvalue.StringExact(awstypes.StreamExceptionPolicyV2Reject)),
					statecheck.ExpectKnownValue(resourceName, policyPath.AtMapKey("stateful_engine_options").AtSliceIndex(0).AtMapKey("flow_timeouts"), knownvalue.ListSizeExact(0)),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
			},
			{
				Config: testAccFirewallPolicyV2Config_fullCleared(rName),
				Check:  testAccCheckFirewallPolicyV2Exists(ctx, t, resourceName, &v),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					compareUpdateToken.AddStateValue(resourceName, tfjsonpath.New("update_token")),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("firewall_policy"), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, policyPath.AtMapKey("firewall_default_action"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, policyPath.AtMapKey("policy_mode"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, policyPath.AtMapKey("tls_inspection_configuration_arn"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, policyPath.AtMapKey("firewall_rule_group_entry"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, policyPath.AtMapKey("ips_rule_group_entry"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, policyPath.AtMapKey("policy_variables"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, policyPath.AtMapKey("route53_resolved_ip_validation"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, policyPath.AtMapKey("stateful_engine_options"), knownvalue.ListSizeExact(0)),
				},
			},
		},
	})
}

func TestAccNetworkFirewallFirewallPolicyV2_route53ResolvedIPValidation(t *testing.T) {
	ctx := acctest.Context(t)
	var v networkfirewall.DescribeFirewallPolicyV2Output
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy_v2.test"
	validationPath := tfjsonpath.New("firewall_policy").AtSliceIndex(0).AtMapKey("route53_resolved_ip_validation").AtSliceIndex(0)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallPolicyV2Destroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyV2Config_route53ResolvedIPValidation(rName),
				Check:  testAccCheckFirewallPolicyV2Exists(ctx, t, resourceName, &v),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, validationPath.AtMapKey(names.AttrAction), tfknownvalue.StringExact(awstypes.Route53ResolvedIPValidationActionAlert)),
					statecheck.ExpectKnownValue(resourceName, validationPath.AtMapKey("dns_expiry_exception_policy"), tfknownvalue.StringExact(awstypes.DNSExpiryExceptionPolicyContinue)),
					statecheck.CompareValuePairs(resourceName, validationPath.AtMapKey("dns_firewall_rule_group_id"), "aws_route53_resolver_firewall_rule_group.test", tfjsonpath.New(names.AttrID), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, validationPath.AtMapKey("source_ip_range"), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, validationPath.AtMapKey("bypass_destination_validation_ip_range"), knownvalue.ListSizeExact(1)),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
			},
		},
	})
}

func TestAccNetworkFirewallFirewallPolicyV2_route53ResolvedIPValidationUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var v networkfirewall.DescribeFirewallPolicyV2Output
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy_v2.test"
	policyPath := tfjsonpath.New("firewall_policy").AtSliceIndex(0)
	validationPath := policyPath.AtMapKey("route53_resolved_ip_validation").AtSliceIndex(0)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID, names.Route53ResolverServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallPolicyV2Destroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyV2Config_route53ResolvedIPValidationUpdate(rName, 0, "ALERT", "CONTINUE", "10.0.0.0/16", "169.254.169.254/32"),
				Check:  testAccCheckFirewallPolicyV2Exists(ctx, t, resourceName, &v),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, validationPath.AtMapKey(names.AttrAction), tfknownvalue.StringExact(awstypes.Route53ResolvedIPValidationActionAlert)),
					statecheck.ExpectKnownValue(resourceName, validationPath.AtMapKey("dns_expiry_exception_policy"), tfknownvalue.StringExact(awstypes.DNSExpiryExceptionPolicyContinue)),
					statecheck.CompareValuePairs(resourceName, validationPath.AtMapKey("dns_firewall_rule_group_id"), "aws_route53_resolver_firewall_rule_group.test[0]", tfjsonpath.New(names.AttrID), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, validationPath.AtMapKey("source_ip_range"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{"address_definition": knownvalue.StringExact("10.0.0.0/16")}),
					})),
					statecheck.ExpectKnownValue(resourceName, validationPath.AtMapKey("bypass_destination_validation_ip_range"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{"address_definition": knownvalue.StringExact("169.254.169.254/32")}),
					})),
				},
			},
			{
				Config: testAccFirewallPolicyV2Config_route53ResolvedIPValidationUpdate(rName, 0, "REJECT", "DROP", "10.1.0.0/16", "192.0.2.0/24"),
				Check:  testAccCheckFirewallPolicyV2Exists(ctx, t, resourceName, &v),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, validationPath.AtMapKey(names.AttrAction), tfknownvalue.StringExact(awstypes.Route53ResolvedIPValidationActionReject)),
					statecheck.ExpectKnownValue(resourceName, validationPath.AtMapKey("dns_expiry_exception_policy"), tfknownvalue.StringExact(awstypes.DNSExpiryExceptionPolicyDrop)),
					statecheck.CompareValuePairs(resourceName, validationPath.AtMapKey("dns_firewall_rule_group_id"), "aws_route53_resolver_firewall_rule_group.test[0]", tfjsonpath.New(names.AttrID), compare.ValuesSame()),
					statecheck.ExpectKnownValue(resourceName, validationPath.AtMapKey("source_ip_range"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{"address_definition": knownvalue.StringExact("10.1.0.0/16")}),
					})),
					statecheck.ExpectKnownValue(resourceName, validationPath.AtMapKey("bypass_destination_validation_ip_range"), knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{"address_definition": knownvalue.StringExact("192.0.2.0/24")}),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
			},
			{
				Config: testAccFirewallPolicyV2Config_route53ResolvedIPValidationUpdate(rName, 1, "REJECT", "DROP", "10.1.0.0/16", "192.0.2.0/24"),
				Check:  testAccCheckFirewallPolicyV2Exists(ctx, t, resourceName, &v),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(resourceName, validationPath.AtMapKey("dns_firewall_rule_group_id"), "aws_route53_resolver_firewall_rule_group.test[1]", tfjsonpath.New(names.AttrID), compare.ValuesSame()),
				},
			},
			{
				Config: testAccFirewallPolicyV2Config_route53ResolvedIPValidationRemoved(rName),
				Check:  testAccCheckFirewallPolicyV2Exists(ctx, t, resourceName, &v),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, policyPath.AtMapKey("route53_resolved_ip_validation"), knownvalue.ListSizeExact(0)),
				},
			},
		},
	})
}

// Tests which depend on TLS inspection configuration ARN(s) provided
// via environment variables.
func TestAccNetworkFirewallFirewallPolicyV2_tlsInspectionConfigurationARN_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		"tlsInspectionConfigurationARN": testAccFirewallPolicyV2_tlsInspectionConfigurationARN,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccFirewallPolicyV2_tlsInspectionConfigurationARN(t *testing.T) {
	ctx := acctest.Context(t)
	var v networkfirewall.DescribeFirewallPolicyV2Output
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy_v2.test"
	tlsPath := tfjsonpath.New("firewall_policy").AtSliceIndex(0).AtMapKey("tls_inspection_configuration_arn")
	arn1 := acctest.SkipIfEnvVarNotSet(t, "AWS_NETWORKFIREWALL_TLS_INSPECTION_CONFIGURATION_ARN_1")
	arn2 := acctest.SkipIfEnvVarNotSet(t, "AWS_NETWORKFIREWALL_TLS_INSPECTION_CONFIGURATION_ARN_2")

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallPolicyV2Destroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyV2Config_tlsInspectionConfigurationARN(rName, arn1),
				Check:  testAccCheckFirewallPolicyV2Exists(ctx, t, resourceName, &v),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tlsPath, knownvalue.StringExact(arn1)),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
			},
			{
				Config: testAccFirewallPolicyV2Config_tlsInspectionConfigurationARN(rName, arn2),
				Check:  testAccCheckFirewallPolicyV2Exists(ctx, t, resourceName, &v),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tlsPath, knownvalue.StringExact(arn2)),
				},
			},
			// The API rejects adding or removing a TLS inspection configuration on update, so the policy is replaced.
			{
				Config: testAccFirewallPolicyV2Config_basic(rName),
				Check:  testAccCheckFirewallPolicyV2Exists(ctx, t, resourceName, &v),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tlsPath, knownvalue.Null()),
				},
			},
			{
				Config: testAccFirewallPolicyV2Config_tlsInspectionConfigurationARN(rName, arn1),
				Check:  testAccCheckFirewallPolicyV2Exists(ctx, t, resourceName, &v),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tlsPath, knownvalue.StringExact(arn1)),
				},
			},
		},
	})
}

func TestAccNetworkFirewallFirewallPolicyV2_firewallRuleOptions(t *testing.T) {
	ctx := acctest.Context(t)
	var v networkfirewall.DescribeFirewallPolicyV2Output
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy_v2.test"
	policyPath := tfjsonpath.New("firewall_policy").AtSliceIndex(0)
	// tfjsonpath.Path appends share backing arrays, so build each base path fresh before extending it more than once.
	firewallGroupPath := func() tfjsonpath.Path {
		return policyPath.AtMapKey("firewall_rule_group_entry").AtSliceIndex(0).AtMapKey("inline_rule_group").AtSliceIndex(0)
	}
	firewallRulePath := firewallGroupPath().AtMapKey("rules_source").AtSliceIndex(0).AtMapKey("firewall_rule").AtSliceIndex(0)
	ipsVariablesPath := func() tfjsonpath.Path {
		return policyPath.AtMapKey("ips_rule_group_entry").AtSliceIndex(0).AtMapKey("inline_rule_group").AtSliceIndex(0).AtMapKey("rule_variables").AtSliceIndex(0)
	}

	hookStep := func(action awstypes.FirewallRuleAction, actionScope awstypes.ActionScope, hook awstypes.HookKey, protocol awstypes.FirewallRuleProtocol, destPorts string) resource.TestStep {
		return resource.TestStep{
			Config: testAccFirewallPolicyV2Config_firewallRuleHook(rName, string(action), string(actionScope), string(hook), string(protocol), destPorts),
			Check:  testAccCheckFirewallPolicyV2Exists(ctx, t, resourceName, &v),
			ConfigPlanChecks: resource.ConfigPlanChecks{
				PreApply: []plancheck.PlanCheck{
					plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
				},
			},
			ConfigStateChecks: []statecheck.StateCheck{
				statecheck.ExpectKnownValue(resourceName, firewallRulePath, knownvalue.ObjectExact(map[string]knownvalue.Check{
					"accept_prior_hooks": knownvalue.Null(),
					names.AttrAction:     tfknownvalue.StringExact(action),
					"action_scope":       tfknownvalue.StringExact(actionScope),
					"alert_enabled":      knownvalue.Null(),
					"bypass_ips_enabled": knownvalue.Null(),
					"dest_ips":           knownvalue.StringExact("any"),
					"dest_ports":         knownvalue.StringExact(destPorts),
					"hook":               tfknownvalue.StringExact(hook),
					names.AttrProtocol:   tfknownvalue.StringExact(protocol),
					"rule_metadata":      knownvalue.ListSizeExact(0),
					"rule_option": knownvalue.ListExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							"keyword":  knownvalue.StringExact("sid"),
							"settings": knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("1000102")}),
						}),
					}),
					"source_ips":   knownvalue.StringExact("any"),
					"source_ports": knownvalue.StringExact("any"),
				})),
			},
		}
	}

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallPolicyV2Destroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyV2Config_firewallRuleOptions(rName),
				Check:  testAccCheckFirewallPolicyV2Exists(ctx, t, resourceName, &v),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, firewallRulePath, knownvalue.ObjectExact(map[string]knownvalue.Check{
						"accept_prior_hooks": knownvalue.Bool(true),
						names.AttrAction:     tfknownvalue.StringExact(awstypes.FirewallRuleActionAccept),
						"action_scope":       tfknownvalue.StringExact(awstypes.ActionScopeFlow),
						"alert_enabled":      knownvalue.Bool(true),
						"bypass_ips_enabled": knownvalue.Bool(true),
						"dest_ips":           knownvalue.StringExact("any"),
						"dest_ports":         knownvalue.StringExact("443"),
						"hook":               tfknownvalue.StringExact(awstypes.HookKeyTlsClientHello),
						names.AttrProtocol:   tfknownvalue.StringExact(awstypes.FirewallRuleProtocolTls),
						"rule_metadata": knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"rule_description": knownvalue.StringExact("allow tls client hello"),
								"rule_enabled":     knownvalue.Bool(true),
							}),
						}),
						"rule_option": knownvalue.ListExact([]knownvalue.Check{
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"keyword":  knownvalue.StringExact("sid"),
								"settings": knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("1000101")}),
							}),
							knownvalue.ObjectExact(map[string]knownvalue.Check{
								"keyword":  knownvalue.StringExact("rev"),
								"settings": knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("1")}),
							}),
						}),
						"source_ips":   knownvalue.StringExact("$HOME_NET"),
						"source_ports": knownvalue.StringExact("any"),
					})),
					statecheck.ExpectKnownValue(resourceName, firewallGroupPath().AtMapKey("rule_variables").AtSliceIndex(0).AtMapKey("port_set"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrKey: knownvalue.StringExact("WEB_PORTS"),
							"definition":  knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("80"), knownvalue.StringExact("443")}),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, ipsVariablesPath().AtMapKey("ip_set"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrKey: knownvalue.StringExact("IPS_NET"),
							"definition":  knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("10.2.0.0/16")}),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, ipsVariablesPath().AtMapKey("port_set"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrKey: knownvalue.StringExact("IPS_PORTS"),
							"definition":  knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("22"), knownvalue.StringExact("3389")}),
						}),
					})),
					statecheck.ExpectKnownValue(resourceName, policyPath.AtMapKey("policy_variables").AtSliceIndex(0).AtMapKey("rule_variable"), knownvalue.SetExact([]knownvalue.Check{
						knownvalue.ObjectExact(map[string]knownvalue.Check{
							names.AttrKey: knownvalue.StringExact("HOME_NET"),
							"definition":  knownvalue.ListExact([]knownvalue.Check{knownvalue.StringExact("10.0.0.0/16"), knownvalue.StringExact("10.1.0.0/16")}),
						}),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
			},
			// accept:hook at a packet hook.
			hookStep(awstypes.FirewallRuleActionAccept, awstypes.ActionScopeHook, awstypes.HookKeyTcpAll, awstypes.FirewallRuleProtocolTcp, "443"),
			// drop:flow at a UDP packet hook.
			hookStep(awstypes.FirewallRuleActionDrop, awstypes.ActionScopeFlow, awstypes.HookKeyUdpAll, awstypes.FirewallRuleProtocolUdp, "53"),
			// drop:packet is only valid at a packet hook.
			hookStep(awstypes.FirewallRuleActionDrop, awstypes.ActionScopePacket, awstypes.HookKeyTcpAll, awstypes.FirewallRuleProtocolTcp, "23"),
			// accept:tx at an application-layer hook.
			hookStep(awstypes.FirewallRuleActionAccept, awstypes.ActionScopeTx, awstypes.HookKeyHttp1RequestHeaders, awstypes.FirewallRuleProtocolHttp, "80"),
		},
	})
}

func TestAccNetworkFirewallFirewallPolicyV2_ruleGroupEntryOrder(t *testing.T) {
	ctx := acctest.Context(t)
	var v networkfirewall.DescribeFirewallPolicyV2Output
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy_v2.test"
	policyPath := tfjsonpath.New("firewall_policy").AtSliceIndex(0)
	// tfjsonpath.Path appends share backing arrays, so build each base path fresh before extending it more than once.
	firewallEntriesPath := func() tfjsonpath.Path { return policyPath.AtMapKey("firewall_rule_group_entry") }
	ipsEntriesPath := func() tfjsonpath.Path { return policyPath.AtMapKey("ips_rule_group_entry") }
	firewallRulesPath := func() tfjsonpath.Path {
		return firewallEntriesPath().AtSliceIndex(0).AtMapKey("inline_rule_group").AtSliceIndex(0).AtMapKey("rules_source").AtSliceIndex(0).AtMapKey("firewall_rule")
	}

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallPolicyV2Destroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyV2Config_ruleGroupEntryOrder(rName),
				Check:  testAccCheckFirewallPolicyV2Exists(ctx, t, resourceName, &v),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, firewallEntriesPath(), knownvalue.ListSizeExact(2)),
					statecheck.ExpectKnownValue(resourceName, firewallEntriesPath().AtSliceIndex(0).AtMapKey(names.AttrPriority), knownvalue.Int32Exact(200)),
					statecheck.ExpectKnownValue(resourceName, firewallEntriesPath().AtSliceIndex(1).AtMapKey(names.AttrPriority), knownvalue.Int32Exact(100)),
					statecheck.ExpectKnownValue(resourceName, firewallRulesPath(), knownvalue.ListSizeExact(2)),
					statecheck.ExpectKnownValue(resourceName, firewallRulesPath().AtSliceIndex(0).AtMapKey("dest_ports"), knownvalue.StringExact("443")),
					statecheck.ExpectKnownValue(resourceName, firewallRulesPath().AtSliceIndex(1).AtMapKey("dest_ports"), knownvalue.StringExact("80")),
					statecheck.ExpectKnownValue(resourceName, ipsEntriesPath(), knownvalue.ListSizeExact(2)),
					statecheck.ExpectKnownValue(resourceName, ipsEntriesPath().AtSliceIndex(0).AtMapKey("inline_rule_group").AtSliceIndex(0).AtMapKey("rules_source").AtSliceIndex(0).AtMapKey("rules_string"), knownvalue.StringRegexp(regexache.MustCompile(`sid:2000201;`))),
					statecheck.ExpectKnownValue(resourceName, ipsEntriesPath().AtSliceIndex(1).AtMapKey("inline_rule_group").AtSliceIndex(0).AtMapKey("rules_source").AtSliceIndex(0).AtMapKey("rules_string"), knownvalue.StringRegexp(regexache.MustCompile(`sid:2000202;`))),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
			},
		},
	})
}

func TestAccNetworkFirewallFirewallPolicyV2_invalidConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	const ruleGroupARN = "arn:aws:network-firewall:us-west-2:123456789012:stateful-rulegroup/example" //lintignore:AWSAT003,AWSAT005
	const inlineFirewallRuleGroup = `
      inline_rule_group {
        rules_source {
          rules_string = "accept:flow tcp:flow_start any any -> any 443 (sid:1000301; rev:1;)"
        }
      }`
	const inlineIPSRuleGroup = `
      inline_rule_group {
        rules_source {
          rules_string = "alert tcp any any -> any 22 (msg:\"ssh\"; sid:2000301; rev:1;)"
        }
      }`
	const firewallRule = `
          firewall_rule {
            action       = "ACCEPT"
            action_scope = "FLOW"
            hook         = "TCP_FLOW_START"
            protocol     = "TCP"
          }`

	invalidStep := func(policy string, expectError string) resource.TestStep {
		return resource.TestStep{
			Config:      testAccFirewallPolicyV2Config_invalid(rName, policy),
			PlanOnly:    true,
			ExpectError: regexache.MustCompile(expectError),
		}
	}

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallPolicyV2Destroy(ctx, t),
		Steps: []resource.TestStep{
			// Firewall entry with both inline_rule_group and resource_arn.
			invalidStep(fmt.Sprintf(`
    firewall_rule_group_entry {
      priority     = 100
      resource_arn = %q
%s
    }`, ruleGroupARN, inlineFirewallRuleGroup), `Invalid Attribute Combination`),
			// Firewall entry with neither inline_rule_group nor resource_arn.
			invalidStep(`
    firewall_rule_group_entry {
      priority = 100
    }`, `Invalid Attribute Combination`),
			// IPS entry with both inline_rule_group and resource_arn.
			invalidStep(fmt.Sprintf(`
    ips_rule_group_entry {
      resource_arn = %q
%s
    }`, ruleGroupARN, inlineIPSRuleGroup), `Invalid Attribute Combination`),
			// IPS entry with neither inline_rule_group nor resource_arn.
			invalidStep(`
    ips_rule_group_entry {}`, `Invalid Attribute Combination`),
			// Firewall entry without priority.
			invalidStep(fmt.Sprintf(`
    firewall_rule_group_entry {
%s
    }`, inlineFirewallRuleGroup), `The argument "priority" is required`),
			// Two firewall entries with the same priority.
			invalidStep(fmt.Sprintf(`
    firewall_rule_group_entry {
      priority = 100
%[1]s
    }

    firewall_rule_group_entry {
      priority     = 100
      resource_arn = %[2]q
    }`, inlineFirewallRuleGroup, ruleGroupARN), `Duplicate Priority`),
			// rules_source with both rules_string and firewall_rule.
			invalidStep(fmt.Sprintf(`
    firewall_rule_group_entry {
      priority = 100

      inline_rule_group {
        rules_source {
          rules_string = "accept:flow tcp:flow_start any any -> any 443 (sid:1000302; rev:1;)"
%s
        }
      }
    }`, firewallRule), `Invalid Attribute Combination`),
		},
	})
}

func testAccCheckFirewallPolicyV2Destroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).NetworkFirewallClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networkfirewall_firewall_policy_v2" {
				continue
			}

			_, err := tfnetworkfirewall.FindFirewallPolicyV2ByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("NetworkFirewall Firewall Policy V2 %s still exists", rs.Primary.Attributes[names.AttrARN])
		}

		return nil
	}
}

func testAccCheckFirewallPolicyV2Exists(ctx context.Context, t *testing.T, n string, v *networkfirewall.DescribeFirewallPolicyV2Output) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).NetworkFirewallClient(ctx)

		out, err := tfnetworkfirewall.FindFirewallPolicyV2ByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])
		if err != nil {
			return err
		}

		*v = *out

		return nil
	}
}

func testAccFirewallPolicyV2Config_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy_v2" "test" {
  name = %[1]q

  firewall_policy {}
}
`, rName)
}

func testAccFirewallPolicyV2Config_description(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy_v2" "test" {
  name        = %[1]q
  description = %[2]q

  firewall_policy {}
}
`, rName, description)
}

func testAccFirewallPolicyV2Config_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy_v2" "test" {
  name = %[1]q

  firewall_policy {}

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccFirewallPolicyV2Config_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy_v2" "test" {
  name = %[1]q

  firewall_policy {}

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccFirewallPolicyV2Config_baseKMSKeys(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  count = 2

  description             = "%[1]s-${count.index}"
  deletion_window_in_days = 7
  enable_key_rotation     = true
}
`, rName)
}

func testAccFirewallPolicyV2Config_encryptionConfigurationCustomerKMS(rName string, keyIndex int) string {
	return acctest.ConfigCompose(testAccFirewallPolicyV2Config_baseKMSKeys(rName), fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy_v2" "test" {
  name = %[1]q

  encryption_configuration {
    key_id = aws_kms_key.test[%[2]d].arn
    type   = "CUSTOMER_KMS"
  }

  firewall_policy {}
}
`, rName, keyIndex))
}

func testAccFirewallPolicyV2Config_encryptionConfigurationOwnedKey(rName string) string {
	return acctest.ConfigCompose(testAccFirewallPolicyV2Config_baseKMSKeys(rName), fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy_v2" "test" {
  name = %[1]q

  encryption_configuration {
    key_id = "AWS_OWNED_KMS_KEY"
    type   = "AWS_OWNED_KMS_KEY"
  }

  firewall_policy {}
}
`, rName))
}

func testAccFirewallPolicyV2Config_encryptionConfigurationRemoved(rName string) string {
	return acctest.ConfigCompose(testAccFirewallPolicyV2Config_baseKMSKeys(rName), fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy_v2" "test" {
  name = %[1]q

  firewall_policy {}
}
`, rName))
}

func testAccFirewallPolicyV2Config_encryptionConfigurationKeyID(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_networkfirewall_firewall_policy_v2" "test" {
  name = %[1]q

  encryption_configuration {
    key_id = aws_kms_key.test.key_id
    type   = "CUSTOMER_KMS"
  }

  firewall_policy {}
}
`, rName)
}

func testAccFirewallPolicyV2Config_full(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_region" "current" {}

resource "aws_ec2_managed_prefix_list" "test" {
  name           = %[1]q
  address_family = "IPv4"
  max_entries    = 1

  entry {
    cidr = "10.0.0.0/16"
  }
}

resource "aws_networkfirewall_firewall_policy_v2" "test" {
  name = %[1]q

  firewall_policy {
    firewall_default_action = "DROP"
    policy_mode             = "ENFORCE"

    firewall_rule_group_entry {
      priority = 100

      inline_rule_group {
        rules_source {
          firewall_rule {
            action       = "ACCEPT"
            action_scope = "FLOW"
            hook         = "TCP_FLOW_START"
            protocol     = "TCP"
            source_ips   = "$HOME_NET"
            source_ports = "any"
            dest_ips     = "any"
            dest_ports   = "443"

            rule_metadata {
              rule_description = "allow https"
              rule_enabled     = true
            }

            rule_option {
              keyword  = "sid"
              settings = ["1000001"]
            }
          }
        }

        rule_variables {
          ip_set {
            key        = "HOME_NET"
            definition = ["10.0.0.0/16"]
          }
          port_set {
            key        = "WEB_PORTS"
            definition = ["80", "443"]
          }
          context_set {
            key        = "MY_VPCS"
            definition = ["vpc-0123456789abcdef0"]
          }
        }
      }
    }

    firewall_rule_group_entry {
      priority = 200

      inline_rule_group {
        rules_source {
          rules_string = "accept:flow tcp:flow_start any any -> any 80 (sid:3000001; rev:1;)"
        }
      }
    }

    ips_rule_group_entry {
      inline_rule_group {
        rules_source {
          rules_string = "alert tcp any any -> any 22 (msg:\"ssh\"; sid:2000001; rev:1;)"
        }

        reference_sets {
          ip_set_reference {
            key           = "TRUSTED_NET"
            reference_arn = aws_ec2_managed_prefix_list.test.arn
          }
        }

        rule_variables {
          context_set {
            key        = "IPS_VPCS"
            definition = ["vpc-0123456789abcdef0", "vpc-0123456789abcdef1"]
          }
        }
      }
    }

    ips_rule_group_entry {
      resource_arn = "arn:${data.aws_partition.current.partition}:network-firewall:${data.aws_region.current.region}:aws-managed:ips-rulegroup/ThreatSignaturesPhishing"

      override {
        action = "DROP_TO_ALERT"
      }
    }

    policy_variables {
      rule_variable {
        key        = "HOME_NET"
        definition = ["10.0.0.0/16", "10.1.0.0/16"]
      }
    }

    stateful_engine_options {
      stream_exception_policy = "DROP"

      flow_timeouts {
        tcp_idle_timeout_seconds = 350
      }
    }
  }
}
`, rName)
}

func testAccFirewallPolicyV2Config_fullUpdated(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}
data "aws_region" "current" {}

# Retained so that it isn't deleted while the previous policy version still references it.
resource "aws_ec2_managed_prefix_list" "test" {
  name           = %[1]q
  address_family = "IPv4"
  max_entries    = 1

  entry {
    cidr = "10.0.0.0/16"
  }
}

resource "aws_networkfirewall_firewall_policy_v2" "test" {
  name = %[1]q

  firewall_policy {
    firewall_default_action = "ALERT"
    policy_mode             = "MONITOR"

    firewall_rule_group_entry {
      priority = 100

      inline_rule_group {
        rules_source {
          firewall_rule {
            action       = "DROP"
            action_scope = "FLOW"
            hook         = "TCP_FLOW_START"
            protocol     = "TCP"
            source_ips   = "any"
            source_ports = "any"
            dest_ips     = "any"
            dest_ports   = "23"

            rule_option {
              keyword  = "sid"
              settings = ["1000002"]
            }
          }
        }
      }
    }

    ips_rule_group_entry {
      resource_arn = "arn:${data.aws_partition.current.partition}:network-firewall:${data.aws_region.current.region}:aws-managed:ips-rulegroup/ThreatSignaturesPhishing"
    }

    stateful_engine_options {
      stream_exception_policy = "REJECT"
    }
  }
}
`, rName)
}

func testAccFirewallPolicyV2Config_route53ResolvedIPValidation(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53_resolver_firewall_rule_group" "test" {
  name = %[1]q
}

resource "aws_networkfirewall_firewall_policy_v2" "test" {
  name = %[1]q

  firewall_policy {
    route53_resolved_ip_validation {
      action                      = "ALERT"
      dns_expiry_exception_policy = "CONTINUE"
      dns_firewall_rule_group_id  = aws_route53_resolver_firewall_rule_group.test.id

      source_ip_range {
        address_definition = "10.0.0.0/16"
      }

      bypass_destination_validation_ip_range {
        address_definition = "169.254.169.254/32"
      }
    }
  }
}
`, rName)
}

func testAccFirewallPolicyV2Config_fullCleared(rName string) string {
	return fmt.Sprintf(`
# Retained so that it isn't deleted while the previous policy version still references it.
resource "aws_ec2_managed_prefix_list" "test" {
  name           = %[1]q
  address_family = "IPv4"
  max_entries    = 1

  entry {
    cidr = "10.0.0.0/16"
  }
}

resource "aws_networkfirewall_firewall_policy_v2" "test" {
  name = %[1]q

  firewall_policy {}
}
`, rName)
}

func testAccFirewallPolicyV2Config_baseRoute53ResolverFirewallRuleGroups(rName string) string {
	return fmt.Sprintf(`
# Both rule groups are always present: the DNS Firewall rule group swap is asynchronous,
# so neither can be deleted while the policy may still reference it.
resource "aws_route53_resolver_firewall_rule_group" "test" {
  count = 2

  name = "%[1]s-${count.index}"
}
`, rName)
}

func testAccFirewallPolicyV2Config_route53ResolvedIPValidationUpdate(rName string, ruleGroupIndex int, action, dnsExpiryExceptionPolicy, sourceIPRange, bypassIPRange string) string {
	return acctest.ConfigCompose(testAccFirewallPolicyV2Config_baseRoute53ResolverFirewallRuleGroups(rName), fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy_v2" "test" {
  name = %[1]q

  firewall_policy {
    route53_resolved_ip_validation {
      action                      = %[3]q
      dns_expiry_exception_policy = %[4]q
      dns_firewall_rule_group_id  = aws_route53_resolver_firewall_rule_group.test[%[2]d].id

      source_ip_range {
        address_definition = %[5]q
      }

      bypass_destination_validation_ip_range {
        address_definition = %[6]q
      }
    }
  }
}
`, rName, ruleGroupIndex, action, dnsExpiryExceptionPolicy, sourceIPRange, bypassIPRange))
}

func testAccFirewallPolicyV2Config_route53ResolvedIPValidationRemoved(rName string) string {
	return acctest.ConfigCompose(testAccFirewallPolicyV2Config_baseRoute53ResolverFirewallRuleGroups(rName), fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy_v2" "test" {
  name = %[1]q

  firewall_policy {}
}
`, rName))
}

func testAccFirewallPolicyV2Config_tlsInspectionConfigurationARN(rName, arn string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy_v2" "test" {
  name = %[1]q

  firewall_policy {
    tls_inspection_configuration_arn = %[2]q
  }
}
`, rName, arn)
}

func testAccFirewallPolicyV2Config_firewallRuleOptions(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy_v2" "test" {
  name = %[1]q

  firewall_policy {
    firewall_rule_group_entry {
      priority = 100

      inline_rule_group {
        rules_source {
          firewall_rule {
            action             = "ACCEPT"
            action_scope       = "FLOW"
            hook               = "TLS_CLIENT_HELLO"
            protocol           = "TLS"
            accept_prior_hooks = true
            alert_enabled      = true
            bypass_ips_enabled = true
            source_ips         = "$HOME_NET"
            source_ports       = "any"
            dest_ips           = "any"
            dest_ports         = "443"

            rule_metadata {
              rule_description = "allow tls client hello"
              rule_enabled     = true
            }

            rule_option {
              keyword  = "sid"
              settings = ["1000101"]
            }

            rule_option {
              keyword  = "rev"
              settings = ["1"]
            }
          }
        }

        rule_variables {
          ip_set {
            key        = "HOME_NET"
            definition = ["10.0.0.0/16"]
          }

          port_set {
            key        = "WEB_PORTS"
            definition = ["80", "443"]
          }
        }
      }
    }

    ips_rule_group_entry {
      inline_rule_group {
        rules_source {
          rules_string = "alert tcp $IPS_NET any -> any $IPS_PORTS (msg:\"ips\"; sid:2000101; rev:1;)"
        }

        rule_variables {
          ip_set {
            key        = "IPS_NET"
            definition = ["10.2.0.0/16"]
          }

          port_set {
            key        = "IPS_PORTS"
            definition = ["22", "3389"]
          }
        }
      }
    }

    policy_variables {
      rule_variable {
        key        = "HOME_NET"
        definition = ["10.0.0.0/16", "10.1.0.0/16"]
      }
    }
  }
}
`, rName)
}

func testAccFirewallPolicyV2Config_firewallRuleHook(rName, action, actionScope, hook, protocol, destPorts string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy_v2" "test" {
  name = %[1]q

  firewall_policy {
    firewall_rule_group_entry {
      priority = 100

      inline_rule_group {
        rules_source {
          firewall_rule {
            action       = %[2]q
            action_scope = %[3]q
            hook         = %[4]q
            protocol     = %[5]q
            source_ips   = "any"
            source_ports = "any"
            dest_ips     = "any"
            dest_ports   = %[6]q

            rule_option {
              keyword  = "sid"
              settings = ["1000102"]
            }
          }
        }
      }
    }
  }
}
`, rName, action, actionScope, hook, protocol, destPorts)
}

func testAccFirewallPolicyV2Config_ruleGroupEntryOrder(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy_v2" "test" {
  name = %[1]q

  firewall_policy {
    firewall_rule_group_entry {
      priority = 200

      inline_rule_group {
        rules_source {
          firewall_rule {
            action       = "ACCEPT"
            action_scope = "FLOW"
            hook         = "TCP_FLOW_START"
            protocol     = "TCP"
            source_ips   = "any"
            source_ports = "any"
            dest_ips     = "any"
            dest_ports   = "443"

            rule_option {
              keyword  = "sid"
              settings = ["1000201"]
            }
          }

          firewall_rule {
            action       = "ACCEPT"
            action_scope = "FLOW"
            hook         = "TCP_FLOW_START"
            protocol     = "TCP"
            source_ips   = "any"
            source_ports = "any"
            dest_ips     = "any"
            dest_ports   = "80"

            rule_option {
              keyword  = "sid"
              settings = ["1000202"]
            }
          }
        }
      }
    }

    firewall_rule_group_entry {
      priority = 100

      inline_rule_group {
        rules_source {
          rules_string = "drop:flow tcp:flow_start any any -> any 23 (sid:1000203; rev:1;)"
        }
      }
    }

    ips_rule_group_entry {
      inline_rule_group {
        rules_source {
          rules_string = "alert tcp any any -> any 22 (msg:\"ssh\"; sid:2000201; rev:1;)"
        }
      }
    }

    ips_rule_group_entry {
      inline_rule_group {
        rules_source {
          rules_string = "alert tcp any any -> any 3389 (msg:\"rdp\"; sid:2000202; rev:1;)"
        }
      }
    }
  }
}
`, rName)
}

func testAccFirewallPolicyV2Config_invalid(rName, policy string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy_v2" "test" {
  name = %[1]q

  firewall_policy {
%[2]s
  }
}
`, rName, policy)
}
