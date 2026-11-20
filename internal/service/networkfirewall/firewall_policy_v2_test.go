// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkfirewall_test

import (
	"context"
	"fmt"
	"testing"

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
	keyResourceName := "aws_kms_key.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallPolicyV2Destroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyV2Config_encryptionConfiguration(rName),
				Check:  testAccCheckFirewallPolicyV2Exists(ctx, t, resourceName, &v),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrEncryptionConfiguration).AtSliceIndex(0).AtMapKey(names.AttrType), tfknownvalue.StringExact(awstypes.EncryptionTypeCustomerKms)),
					statecheck.CompareValuePairs(resourceName, tfjsonpath.New(names.AttrEncryptionConfiguration).AtSliceIndex(0).AtMapKey(names.AttrKeyID), keyResourceName, tfjsonpath.New(names.AttrARN), compare.ValuesSame()),
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

func TestAccNetworkFirewallFirewallPolicyV2_full(t *testing.T) {
	ctx := acctest.Context(t)
	var v networkfirewall.DescribeFirewallPolicyV2Output
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy_v2.test"
	policyPath := tfjsonpath.New("firewall_policy").AtSliceIndex(0)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallPolicyV2Destroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyV2Config_full(rName),
				Check:  testAccCheckFirewallPolicyV2Exists(ctx, t, resourceName, &v),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, policyPath.AtMapKey("firewall_default_action"), tfknownvalue.StringExact(awstypes.PolicyDefaultActionDrop)),
					statecheck.ExpectKnownValue(resourceName, policyPath.AtMapKey("policy_mode"), tfknownvalue.StringExact(awstypes.PolicyModeEnforce)),
					statecheck.ExpectKnownValue(resourceName, policyPath.AtMapKey("firewall_rule_group_entry"), knownvalue.ListSizeExact(2)),
					statecheck.ExpectKnownValue(resourceName, policyPath.AtMapKey("firewall_rule_group_entry").AtSliceIndex(0).AtMapKey("inline_rule_group").AtSliceIndex(0).AtMapKey("rules_source").AtSliceIndex(0).AtMapKey("firewall_rule").AtSliceIndex(0), knownvalue.ObjectPartial(map[string]knownvalue.Check{
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
					statecheck.ExpectKnownValue(resourceName, policyPath.AtMapKey("ips_rule_group_entry"), knownvalue.ListSizeExact(2)),
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
					statecheck.ExpectKnownValue(resourceName, policyPath.AtMapKey("firewall_default_action"), tfknownvalue.StringExact(awstypes.PolicyDefaultActionAlert)),
					statecheck.ExpectKnownValue(resourceName, policyPath.AtMapKey("policy_mode"), tfknownvalue.StringExact(awstypes.PolicyModeMonitor)),
					statecheck.ExpectKnownValue(resourceName, policyPath.AtMapKey("firewall_rule_group_entry"), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, policyPath.AtMapKey("ips_rule_group_entry"), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, policyPath.AtMapKey("policy_variables"), knownvalue.ListSizeExact(0)),
					statecheck.ExpectKnownValue(resourceName, policyPath.AtMapKey("stateful_engine_options").AtSliceIndex(0).AtMapKey("stream_exception_policy"), tfknownvalue.StringExact(awstypes.StreamExceptionPolicyV2Reject)),
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

func testAccFirewallPolicyV2Config_encryptionConfiguration(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_networkfirewall_firewall_policy_v2" "test" {
  name = %[1]q

  encryption_configuration {
    key_id = aws_kms_key.test.arn
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
