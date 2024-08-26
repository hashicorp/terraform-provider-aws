// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkfirewall_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkfirewall"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkfirewall/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfnetworkfirewall "github.com/hashicorp/terraform-provider-aws/internal/service/networkfirewall"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNetworkFirewallFirewallPolicy_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var firewallPolicy networkfirewall.DescribeFirewallPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckFirewallPolicyExists(ctx, resourceName, &firewallPolicy),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "network-firewall", fmt.Sprintf("firewall-policy/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.policy_variables.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_default_actions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_engine_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_rule_group_reference.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_custom_action.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_default_actions.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "firewall_policy.0.stateless_default_actions.*", "aws:pass"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_fragment_default_actions.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "firewall_policy.0.stateless_fragment_default_actions.*", "aws:drop"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_rule_group_reference.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.tls_inspection_configuration_arn", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetworkFirewallFirewallPolicy_encryptionConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var firewallPolicy networkfirewall.DescribeFirewallPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyConfig_encryptionConfiguration(rName, "aws:pass"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(ctx, resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.type", "CUSTOMER_KMS"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_default_actions.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "firewall_policy.0.stateless_default_actions.*", "aws:pass"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFirewallPolicyConfig_encryptionConfigurationDisabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(ctx, resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", acctest.Ct0),
				),
			},
			{
				Config: testAccFirewallPolicyConfig_encryptionConfiguration(rName, "aws:pass"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(ctx, resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.type", "CUSTOMER_KMS"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_default_actions.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "firewall_policy.0.stateless_default_actions.*", "aws:pass"),
				),
			},
			{
				Config: testAccFirewallPolicyConfig_encryptionConfiguration(rName, "aws:forward_to_sfe"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(ctx, resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.type", "CUSTOMER_KMS"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_default_actions.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "firewall_policy.0.stateless_default_actions.*", "aws:forward_to_sfe"),
				),
			},
		},
	})
}

func TestAccNetworkFirewallFirewallPolicy_policyVariables(t *testing.T) {
	ctx := acctest.Context(t)
	var firewallPolicy networkfirewall.DescribeFirewallPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyConfig_policyVariables(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(ctx, resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.policy_variables.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.policy_variables.0.rule_variables.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firewall_policy.0.policy_variables.0.rule_variables.*", map[string]string{
						names.AttrKey:           "HOME_NET",
						"ip_set.#":              acctest.Ct1,
						"ip_set.0.definition.#": acctest.Ct2,
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "firewall_policy.0.policy_variables.0.rule_variables.*.ip_set.0.definition.*", "10.0.0.0/16"),
					resource.TestCheckTypeSetElemAttr(resourceName, "firewall_policy.0.policy_variables.0.rule_variables.*.ip_set.0.definition.*", "10.0.1.0/24"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_fragment_default_actions.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "firewall_policy.0.stateless_fragment_default_actions.*", "aws:drop"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_default_actions.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "firewall_policy.0.stateless_default_actions.*", "aws:pass"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFirewallPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(ctx, resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.policy_variables.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_fragment_default_actions.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "firewall_policy.0.stateless_fragment_default_actions.*", "aws:drop"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_default_actions.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "firewall_policy.0.stateless_default_actions.*", "aws:pass"),
				),
			},
		},
	})
}

func TestAccNetworkFirewallFirewallPolicy_statefulDefaultActions(t *testing.T) {
	ctx := acctest.Context(t)
	var firewallPolicy networkfirewall.DescribeFirewallPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyConfig_statefulDefaultActions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(ctx, resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_default_actions.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_default_actions.0", "aws:drop_established"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetworkFirewallFirewallPolicy_statefulEngineOption(t *testing.T) {
	ctx := acctest.Context(t)
	var firewallPolicy networkfirewall.DescribeFirewallPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyConfig_statefulEngineOptions(rName, "STRICT_ORDER", "DROP"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(ctx, resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_engine_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_engine_options.0.rule_order", string(awstypes.RuleOrderStrictOrder)),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_engine_options.0.stream_exception_policy", string(awstypes.StreamExceptionPolicyDrop)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetworkFirewallFirewallPolicy_updateStatefulEngineOption(t *testing.T) {
	ctx := acctest.Context(t)
	var firewallPolicy1, firewallPolicy2, firewallPolicy3 networkfirewall.DescribeFirewallPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyConfig_statefulEngineOptions(rName, "DEFAULT_ACTION_ORDER", "CONTINUE"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(ctx, resourceName, &firewallPolicy1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_engine_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_engine_options.0.rule_order", string(awstypes.RuleOrderDefaultActionOrder)),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_engine_options.0.stream_exception_policy", string(awstypes.StreamExceptionPolicyContinue)),
				),
			},
			{
				Config: testAccFirewallPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(ctx, resourceName, &firewallPolicy2),
					testAccCheckFirewallPolicyNotRecreated(&firewallPolicy1, &firewallPolicy2),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", acctest.Ct1),
				),
			},
			{
				Config: testAccFirewallPolicyConfig_statefulEngineOptions(rName, "STRICT_ORDER", "REJECT"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(ctx, resourceName, &firewallPolicy3),
					testAccCheckFirewallPolicyRecreated(&firewallPolicy2, &firewallPolicy3),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_engine_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_engine_options.0.rule_order", string(awstypes.RuleOrderStrictOrder)),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_engine_options.0.stream_exception_policy", string(awstypes.StreamExceptionPolicyReject)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetworkFirewallFirewallPolicy_statefulEngineOptionsSingle(t *testing.T) {
	ctx := acctest.Context(t)
	var firewallPolicy networkfirewall.DescribeFirewallPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyConfig_ruleOrderOnly(rName, "DEFAULT_ACTION_ORDER"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(ctx, resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_engine_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_engine_options.0.rule_order", string(awstypes.RuleOrderDefaultActionOrder)),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_engine_options.0.stream_exception_policy", ""),
				),
			},
			{
				Config: testAccFirewallPolicyConfig_streamExceptionPolicyOnly(rName, "REJECT"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(ctx, resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_engine_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_engine_options.0.rule_order", ""),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_engine_options.0.stream_exception_policy", string(awstypes.StreamExceptionPolicyReject)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetworkFirewallFirewallPolicy_statefulRuleGroupReference(t *testing.T) {
	ctx := acctest.Context(t)
	var firewallPolicy networkfirewall.DescribeFirewallPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy.test"
	ruleGroupResourceName := "aws_networkfirewall_rule_group.test.0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyConfig_statefulRuleGroupReference(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(ctx, resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_default_actions.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_engine_options.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_rule_group_reference.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_policy.0.stateful_rule_group_reference.*.resource_arn", ruleGroupResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"firewall_policy.0.stateful_rule_group_reference.0.priority"},
			},
		},
	})
}

func TestAccNetworkFirewallFirewallPolicy_statefulRuleGroupReferenceManaged(t *testing.T) {
	ctx := acctest.Context(t)
	var firewallPolicy networkfirewall.DescribeFirewallPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyConfig_statefulRuleGroupReferenceManaged(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(ctx, resourceName, &firewallPolicy),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"firewall_policy.0.stateful_rule_group_reference.0.priority"},
			},
		},
	})
}

func TestAccNetworkFirewallFirewallPolicy_updateStatefulRuleGroupReference(t *testing.T) {
	ctx := acctest.Context(t)
	var firewallPolicy networkfirewall.DescribeFirewallPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy.test"
	ruleGroupResourceName := "aws_networkfirewall_rule_group.test.0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(ctx, resourceName, &firewallPolicy),
				),
			},
			{
				Config: testAccFirewallPolicyConfig_statefulRuleGroupReference(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(ctx, resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_rule_group_reference.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_policy.0.stateful_rule_group_reference.*.resource_arn", ruleGroupResourceName, names.AttrARN),
				),
			},
			{
				Config: testAccFirewallPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(ctx, resourceName, &firewallPolicy),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetworkFirewallFirewallPolicy_multipleStatefulRuleGroupReferences(t *testing.T) {
	ctx := acctest.Context(t)
	var firewallPolicy networkfirewall.DescribeFirewallPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy.test"
	ruleGroupResourceName1 := "aws_networkfirewall_rule_group.test.0"
	ruleGroupResourceName2 := "aws_networkfirewall_rule_group.test.1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyConfig_multipleStatefulRuleGroupReferences(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(ctx, resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_rule_group_reference.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_policy.0.stateful_rule_group_reference.*.resource_arn", ruleGroupResourceName1, names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_policy.0.stateful_rule_group_reference.*.resource_arn", ruleGroupResourceName2, names.AttrARN),
				),
			},
			{
				Config: testAccFirewallPolicyConfig_singleStatefulRuleGroupReference(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(ctx, resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_rule_group_reference.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_policy.0.stateful_rule_group_reference.*.resource_arn", ruleGroupResourceName1, names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"firewall_policy.0.stateful_rule_group_reference.0.priority"},
			},
		},
	})
}

func TestAccNetworkFirewallFirewallPolicy_statefulRuleGroupPriorityReference(t *testing.T) {
	ctx := acctest.Context(t)
	var firewallPolicy networkfirewall.DescribeFirewallPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy.test"
	ruleGroupResourceName := "aws_networkfirewall_rule_group.test.0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyConfig_statefulRuleGroupPriorityReference(rName, acctest.Ct1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(ctx, resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_rule_group_reference.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_rule_group_reference.0.priority", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_policy.0.stateful_rule_group_reference.*.resource_arn", ruleGroupResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetworkFirewallFirewallPolicy_statefulRuleGroupOverrideActionReference(t *testing.T) {
	ctx := acctest.Context(t)
	var firewallPolicy networkfirewall.DescribeFirewallPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy.test"
	overrideAction := string(awstypes.OverrideActionDropToAlert)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyConfig_statefulRuleGroupReferenceManagedOverrideAction(rName, overrideAction),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(ctx, resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_rule_group_reference.0.override.0.action", overrideAction),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"firewall_policy.0.stateful_rule_group_reference.0.priority"},
			},
		},
	})
}

func TestAccNetworkFirewallFirewallPolicy_updateStatefulRuleGroupPriorityReference(t *testing.T) {
	ctx := acctest.Context(t)
	var firewallPolicy1, firewallPolicy2 networkfirewall.DescribeFirewallPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy.test"
	ruleGroupResourceName := "aws_networkfirewall_rule_group.test.0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyConfig_statefulRuleGroupPriorityReference(rName, acctest.Ct1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(ctx, resourceName, &firewallPolicy1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_rule_group_reference.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_rule_group_reference.0.priority", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_policy.0.stateful_rule_group_reference.*.resource_arn", ruleGroupResourceName, names.AttrARN),
				),
			},
			{
				Config: testAccFirewallPolicyConfig_statefulRuleGroupPriorityReference(rName, acctest.Ct2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(ctx, resourceName, &firewallPolicy2),
					testAccCheckFirewallPolicyNotRecreated(&firewallPolicy1, &firewallPolicy2),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_rule_group_reference.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_rule_group_reference.0.priority", acctest.Ct2),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_policy.0.stateful_rule_group_reference.*.resource_arn", ruleGroupResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetworkFirewallFirewallPolicy_statelessRuleGroupReference(t *testing.T) {
	ctx := acctest.Context(t)
	var firewallPolicy networkfirewall.DescribeFirewallPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy.test"
	ruleGroupResourceName := "aws_networkfirewall_rule_group.test.0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyConfig_statelessRuleGroupReference(rName, 20),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(ctx, resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_rule_group_reference.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_policy.0.stateless_rule_group_reference.*.resource_arn", ruleGroupResourceName, names.AttrARN),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firewall_policy.0.stateless_rule_group_reference.*", map[string]string{
						names.AttrPriority: "20",
					}),
				),
			},
			{
				Config: testAccFirewallPolicyConfig_statelessRuleGroupReference(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(ctx, resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_rule_group_reference.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firewall_policy.0.stateless_rule_group_reference.*", map[string]string{
						names.AttrPriority: acctest.Ct1,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetworkFirewallFirewallPolicy_updateStatelessRuleGroupReference(t *testing.T) {
	ctx := acctest.Context(t)
	var firewallPolicy networkfirewall.DescribeFirewallPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy.test"
	ruleGroupResourceName := "aws_networkfirewall_rule_group.test.0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(ctx, resourceName, &firewallPolicy),
				),
			},
			{
				Config: testAccFirewallPolicyConfig_statelessRuleGroupReference(rName, 20),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(ctx, resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_policy.0.stateless_rule_group_reference.*.resource_arn", ruleGroupResourceName, names.AttrARN),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firewall_policy.0.stateless_rule_group_reference.*", map[string]string{
						names.AttrPriority: "20",
					}),
				),
			},
			{
				Config: testAccFirewallPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(ctx, resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_rule_group_reference.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetworkFirewallFirewallPolicy_multipleStatelessRuleGroupReferences(t *testing.T) {
	ctx := acctest.Context(t)
	var firewallPolicy networkfirewall.DescribeFirewallPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy.test"
	ruleGroupResourceName1 := "aws_networkfirewall_rule_group.test.0"
	ruleGroupResourceName2 := "aws_networkfirewall_rule_group.test.1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyConfig_multipleStatelessRuleGroupReferences(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(ctx, resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_rule_group_reference.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_policy.0.stateless_rule_group_reference.*.resource_arn", ruleGroupResourceName1, names.AttrARN),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firewall_policy.0.stateless_rule_group_reference.*", map[string]string{
						names.AttrPriority: acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_policy.0.stateless_rule_group_reference.*.resource_arn", ruleGroupResourceName2, names.AttrARN),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firewall_policy.0.stateless_rule_group_reference.*", map[string]string{
						names.AttrPriority: acctest.Ct2,
					}),
				),
			},
			{
				Config: testAccFirewallPolicyConfig_singleStatelessRuleGroupReference(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(ctx, resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_rule_group_reference.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firewall_policy.0.stateless_rule_group_reference.*", map[string]string{
						names.AttrPriority: acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_policy.0.stateless_rule_group_reference.*.resource_arn", ruleGroupResourceName1, names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetworkFirewallFirewallPolicy_statelessCustomAction(t *testing.T) {
	ctx := acctest.Context(t)
	var firewallPolicy networkfirewall.DescribeFirewallPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyConfig_statelessCustomAction(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(ctx, resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_custom_action.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firewall_policy.0.stateless_custom_action.*", map[string]string{
						"action_definition.#":                                     acctest.Ct1,
						"action_definition.0.publish_metric_action.#":             acctest.Ct1,
						"action_definition.0.publish_metric_action.0.dimension.#": acctest.Ct1,
						"action_name": "CustomAction",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetworkFirewallFirewallPolicy_updateStatelessCustomAction(t *testing.T) {
	ctx := acctest.Context(t)
	var firewallPolicy1, firewallPolicy2, firewallPolicy3, firewallPolicy4 networkfirewall.DescribeFirewallPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(ctx, resourceName, &firewallPolicy1),
				),
			},
			{
				Config: testAccFirewallPolicyConfig_statelessCustomAction(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(ctx, resourceName, &firewallPolicy2),
					testAccCheckFirewallPolicyRecreated(&firewallPolicy1, &firewallPolicy2),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_custom_action.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firewall_policy.0.stateless_custom_action.*", map[string]string{
						"action_name":         "CustomAction",
						"action_definition.#": acctest.Ct1,
						"action_definition.0.publish_metric_action.#":             acctest.Ct1,
						"action_definition.0.publish_metric_action.0.dimension.#": acctest.Ct1,
					}),
				),
			},
			{
				Config: testAccFirewallPolicyConfig_updateStatelessCustomAction(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(ctx, resourceName, &firewallPolicy3),
					testAccCheckFirewallPolicyRecreated(&firewallPolicy2, &firewallPolicy3),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_custom_action.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firewall_policy.0.stateless_custom_action.*", map[string]string{
						"action_name":         "updated",
						"action_definition.#": acctest.Ct1,
						"action_definition.0.publish_metric_action.#":             acctest.Ct1,
						"action_definition.0.publish_metric_action.0.dimension.#": acctest.Ct1,
					}),
				),
			},
			{
				Config: testAccFirewallPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(ctx, resourceName, &firewallPolicy4),
					testAccCheckFirewallPolicyRecreated(&firewallPolicy3, &firewallPolicy4),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_custom_action.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetworkFirewallFirewallPolicy_multipleStatelessCustomActions(t *testing.T) {
	ctx := acctest.Context(t)
	var firewallPolicy1, firewallPolicy2 networkfirewall.DescribeFirewallPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyConfig_multipleStatelessCustomActions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(ctx, resourceName, &firewallPolicy1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_custom_action.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firewall_policy.0.stateless_custom_action.*", map[string]string{
						"action_name":         "CustomAction",
						"action_definition.#": acctest.Ct1,
						"action_definition.0.publish_metric_action.#":             acctest.Ct1,
						"action_definition.0.publish_metric_action.0.dimension.#": acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firewall_policy.0.stateless_custom_action.*", map[string]string{
						"action_name":         "CustomAction2",
						"action_definition.#": acctest.Ct1,
						"action_definition.0.publish_metric_action.#":             acctest.Ct1,
						"action_definition.0.publish_metric_action.0.dimension.#": acctest.Ct1,
					}),
				),
			},
			{
				Config: testAccFirewallPolicyConfig_statelessCustomAction(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(ctx, resourceName, &firewallPolicy2),
					testAccCheckFirewallPolicyRecreated(&firewallPolicy1, &firewallPolicy2),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_custom_action.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firewall_policy.0.stateless_custom_action.*", map[string]string{
						"action_name":         "CustomAction",
						"action_definition.#": acctest.Ct1,
						"action_definition.0.publish_metric_action.#":             acctest.Ct1,
						"action_definition.0.publish_metric_action.0.dimension.#": acctest.Ct1,
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetworkFirewallFirewallPolicy_statefulRuleGroupReferenceAndCustomAction(t *testing.T) {
	ctx := acctest.Context(t)
	var firewallPolicy networkfirewall.DescribeFirewallPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy.test"
	ruleGroupResourceName := "aws_networkfirewall_rule_group.test.0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyConfig_statefulRuleGroupReferenceAndStatelessCustomAction(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(ctx, resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_rule_group_reference.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_policy.0.stateful_rule_group_reference.*.resource_arn", ruleGroupResourceName, names.AttrARN),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firewall_policy.0.stateless_custom_action.*", map[string]string{
						"action_name":         "CustomAction",
						"action_definition.#": acctest.Ct1,
						"action_definition.0.publish_metric_action.#":             acctest.Ct1,
						"action_definition.0.publish_metric_action.0.dimension.#": acctest.Ct1,
					}),
				),
			},
			{
				Config: testAccFirewallPolicyConfig_statefulRuleGroupReference(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(ctx, resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_rule_group_reference.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_policy.0.stateful_rule_group_reference.*.resource_arn", ruleGroupResourceName, names.AttrARN),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"firewall_policy.0.stateful_rule_group_reference.0.priority"},
			},
		},
	})
}

func TestAccNetworkFirewallFirewallPolicy_tlsInspectionConfigurationARN(t *testing.T) {
	ctx := acctest.Context(t)
	var firewallPolicy networkfirewall.DescribeFirewallPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy.test"
	arn1 := acctest.SkipIfEnvVarNotSet(t, "AWS_NETWORKFIREWALL_TLS_INSPECTION_CONFIGURATION_ARN_1")
	arn2 := acctest.SkipIfEnvVarNotSet(t, "AWS_NETWORKFIREWALL_TLS_INSPECTION_CONFIGURATION_ARN_2")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyConfig_tlsInspectionConfigurationARN(rName, arn1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(ctx, resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.tls_inspection_configuration_arn", arn1),
				),
			},
			{
				Config: testAccFirewallPolicyConfig_tlsInspectionConfigurationARN(rName, arn2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(ctx, resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.tls_inspection_configuration_arn", arn2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetworkFirewallFirewallPolicy_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var firewallPolicy networkfirewall.DescribeFirewallPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(ctx, resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFirewallPolicyConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(ctx, resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccFirewallPolicyConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(ctx, resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccNetworkFirewallFirewallPolicy_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var firewallPolicy networkfirewall.DescribeFirewallPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFirewallPolicyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(ctx, resourceName, &firewallPolicy),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfnetworkfirewall.ResourceFirewallPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckFirewallPolicyDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networkfirewall_firewall_policy" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkFirewallClient(ctx)

			_, err := tfnetworkfirewall.FindFirewallPolicyByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("NetworkFirewall Firewall Policy %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckFirewallPolicyExists(ctx context.Context, n string, v *networkfirewall.DescribeFirewallPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkFirewallClient(ctx)

		output, err := tfnetworkfirewall.FindFirewallPolicyByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckFirewallPolicyNotRecreated(i, j *networkfirewall.DescribeFirewallPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(i.FirewallPolicyResponse.FirewallPolicyId), aws.ToString(j.FirewallPolicyResponse.FirewallPolicyId); before != after {
			return fmt.Errorf("NetworkFirewall Firewall Policy was recreated. got: %s, expected: %s", after, before)
		}
		return nil
	}
}

func testAccCheckFirewallPolicyRecreated(i, j *networkfirewall.DescribeFirewallPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(i.FirewallPolicyResponse.FirewallPolicyId), aws.ToString(j.FirewallPolicyResponse.FirewallPolicyId); before == after {
			return fmt.Errorf("NetworkFirewall Firewall Policy (%s) was not recreated", before)
		}
		return nil
	}
}

func testAccFirewallPolicyConfig_baseStatelessRuleGroup(rName string, count int) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_rule_group" "test" {
  count = %[1]d

  capacity = 100
  name     = "%[2]s-${count.index}"
  type     = "STATELESS"

  rule_group {
    rules_source {
      stateless_rules_and_custom_actions {
        stateless_rule {
          priority = 1

          rule_definition {
            actions = ["aws:drop"]

            match_attributes {
              destination {
                address_definition = "1.2.3.4/32"
              }

              source {
                address_definition = "124.1.1.5/32"
              }
            }
          }
        }
      }
    }
  }

  lifecycle {
    create_before_destroy = true
  }
}
`, count, rName)
}

func testAccFirewallPolicyConfig_baseStatefulRuleGroup(rName string, count int) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_rule_group" "test" {
  count = %[1]d

  capacity = 100
  name     = "%[2]s-${count.index}"
  type     = "STATEFUL"

  rule_group {
    rules_source {
      rules_source_list {
        generated_rules_type = "ALLOWLIST"
        target_types         = ["HTTP_HOST"]
        targets              = ["test.example.com"]
      }
    }
  }

  lifecycle {
    create_before_destroy = true
  }
}
`, count, rName)
}

func testAccFirewallPolicyConfig_baseStatefulRuleGroupStrictOrder(rName string, count int) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_rule_group" "test" {
  count = %[1]d

  capacity = 100
  name     = "%[2]s-${count.index}"
  type     = "STATEFUL"

  rule_group {
    rules_source {
      stateful_rule {
        action = "PASS"

        header {
          destination      = "124.1.1.24/32"
          destination_port = 53
          direction        = "ANY"
          protocol         = "TCP"
          source           = "1.2.3.4/32"
          source_port      = 53
        }

        rule_option {
          keyword  = "sid"
          settings = ["1"]
        }
      }
    }

    stateful_rule_options {
      rule_order = "STRICT_ORDER"
    }
  }

  lifecycle {
    create_before_destroy = true
  }
}
`, count, rName)
}

func testAccFirewallPolicyConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q

  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]
  }
}
`, rName)
}

func testAccFirewallPolicyConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q

  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccFirewallPolicyConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q

  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccFirewallPolicyConfig_statefulEngineOptions(rName, ruleOrder, streamExceptionPolicy string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q

  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]

    stateful_engine_options {
      rule_order              = %[2]q
      stream_exception_policy = %[3]q
    }
  }
}
`, rName, ruleOrder, streamExceptionPolicy)
}

func testAccFirewallPolicyConfig_policyVariables(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q

  firewall_policy {
    policy_variables {
      rule_variables {
        key = "HOME_NET"
        ip_set {
          definition = ["10.0.0.0/16", "10.0.1.0/24"]
        }
      }
    }
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]
  }
}
`, rName)
}

func testAccFirewallPolicyConfig_ruleOrderOnly(rName, ruleOrder string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q

  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]

    stateful_engine_options {
      rule_order = %[2]q
    }
  }
}
`, rName, ruleOrder)
}

func testAccFirewallPolicyConfig_streamExceptionPolicyOnly(rName, streamExceptionPolicy string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q

  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]

    stateful_engine_options {
      stream_exception_policy = %[2]q
    }
  }
}
`, rName, streamExceptionPolicy)
}

func testAccFirewallPolicyConfig_statefulDefaultActions(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q

  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]
    stateful_default_actions           = ["aws:drop_established"]

    stateful_engine_options {
      rule_order = "STRICT_ORDER"
    }
  }
}
`, rName)
}

func testAccFirewallPolicyConfig_statefulRuleGroupReference(rName string) string {
	return acctest.ConfigCompose(testAccFirewallPolicyConfig_baseStatefulRuleGroup(rName, 1), fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q

  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]

    stateful_rule_group_reference {
      resource_arn = aws_networkfirewall_rule_group.test[0].arn
    }
  }
}
`, rName))
}

func testAccFirewallPolicyConfig_statefulRuleGroupReferenceManaged(rName string) string {
	return acctest.ConfigCompose(testAccFirewallPolicyConfig_baseStatefulRuleGroup(rName, 1), fmt.Sprintf(`
data "aws_region" "current" {}
data "aws_partition" "current" {}

resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q

  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]

    stateful_rule_group_reference {
      resource_arn = "arn:${data.aws_partition.current.partition}:network-firewall:${data.aws_region.current.name}:aws-managed:stateful-rulegroup/MalwareDomainsActionOrder"
    }
  }
}
`, rName))
}

func testAccFirewallPolicyConfig_multipleStatefulRuleGroupReferences(rName string) string {
	return acctest.ConfigCompose(testAccFirewallPolicyConfig_baseStatefulRuleGroup(rName, 2), fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q

  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]

    stateful_rule_group_reference {
      resource_arn = aws_networkfirewall_rule_group.test[0].arn
    }

    stateful_rule_group_reference {
      resource_arn = aws_networkfirewall_rule_group.test[1].arn
    }
  }
}
`, rName))
}

func testAccFirewallPolicyConfig_statefulRuleGroupPriorityReference(rName, priority string) string {
	return acctest.ConfigCompose(testAccFirewallPolicyConfig_baseStatefulRuleGroupStrictOrder(rName, 1), fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q

  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]

    stateful_engine_options {
      rule_order = "STRICT_ORDER"
    }

    stateful_rule_group_reference {
      priority     = %[2]q
      resource_arn = aws_networkfirewall_rule_group.test[0].arn
    }
  }
}
`, rName, priority))
}

func testAccFirewallPolicyConfig_statefulRuleGroupReferenceManagedOverrideAction(rName, override_action string) string {
	return acctest.ConfigCompose(testAccFirewallPolicyConfig_baseStatefulRuleGroup(rName, 1), fmt.Sprintf(`
data "aws_region" "current" {}
data "aws_partition" "current" {}

resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q

  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]

    stateful_rule_group_reference {
      resource_arn = "arn:${data.aws_partition.current.partition}:network-firewall:${data.aws_region.current.name}:aws-managed:stateful-rulegroup/MalwareDomainsActionOrder"

      override {
        action = %[2]q
      }
    }
  }
}
`, rName, override_action))
}

func testAccFirewallPolicyConfig_singleStatefulRuleGroupReference(rName string) string {
	return acctest.ConfigCompose(testAccFirewallPolicyConfig_baseStatefulRuleGroup(rName, 2), fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q

  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]

    stateful_rule_group_reference {
      resource_arn = aws_networkfirewall_rule_group.test[0].arn
    }
  }
}
`, rName))
}

func testAccFirewallPolicyConfig_statelessRuleGroupReference(rName string, priority int) string {
	return acctest.ConfigCompose(testAccFirewallPolicyConfig_baseStatelessRuleGroup(rName, 1), fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q

  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]

    stateless_rule_group_reference {
      priority     = %[2]d
      resource_arn = aws_networkfirewall_rule_group.test[0].arn
    }
  }
}
`, rName, priority))
}

func testAccFirewallPolicyConfig_multipleStatelessRuleGroupReferences(rName string) string {
	return acctest.ConfigCompose(testAccFirewallPolicyConfig_baseStatelessRuleGroup(rName, 2), fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q

  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]

    stateless_rule_group_reference {
      priority     = 1
      resource_arn = aws_networkfirewall_rule_group.test[0].arn
    }

    stateless_rule_group_reference {
      priority     = 2
      resource_arn = aws_networkfirewall_rule_group.test[1].arn
    }
  }
}
`, rName))
}

func testAccFirewallPolicyConfig_singleStatelessRuleGroupReference(rName string) string {
	return acctest.ConfigCompose(testAccFirewallPolicyConfig_baseStatelessRuleGroup(rName, 2), fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q

  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]

    stateless_rule_group_reference {
      priority     = 1
      resource_arn = aws_networkfirewall_rule_group.test[0].arn
    }
  }
}
`, rName))
}

func testAccFirewallPolicyConfig_statelessCustomAction(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q

  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]

    stateless_custom_action {
      action_name = "CustomAction"
      action_definition {
        publish_metric_action {
          dimension {
            value = "example"
          }
        }
      }
    }
  }
}
`, rName)
}

func testAccFirewallPolicyConfig_updateStatelessCustomAction(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q

  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]

    stateless_custom_action {
      action_name = "updated"

      action_definition {
        publish_metric_action {
          dimension {
            value = "example-update"
          }
        }
      }
    }
  }
}
`, rName)
}

func testAccFirewallPolicyConfig_multipleStatelessCustomActions(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q

  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]

    stateless_custom_action {
      action_definition {
        publish_metric_action {
          dimension {
            value = "example"
          }
        }
      }

      action_name = "CustomAction"
    }

    stateless_custom_action {
      action_definition {
        publish_metric_action {
          dimension {
            value = "example-custom-action"
          }
        }
      }

      action_name = "CustomAction2"
    }
  }
}
`, rName)
}

func testAccFirewallPolicyConfig_statefulRuleGroupReferenceAndStatelessCustomAction(rName string) string {
	return acctest.ConfigCompose(testAccFirewallPolicyConfig_baseStatefulRuleGroup(rName, 1), fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q

  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]

    stateful_rule_group_reference {
      resource_arn = aws_networkfirewall_rule_group.test[0].arn
    }

    stateless_custom_action {
      action_definition {
        publish_metric_action {
          dimension {
            value = "example"
          }
        }
      }

      action_name = "CustomAction"
    }
  }
}
`, rName))
}

func testAccFirewallPolicyConfig_tlsInspectionConfigurationARN(rName, arn string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q

  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]
    tls_inspection_configuration_arn   = %[2]q
  }
}
`, rName, arn)
}

func testAccFirewallPolicyConfig_encryptionConfiguration(rName, statelessDefaultActions string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {}

resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q

  encryption_configuration {
    key_id = aws_kms_key.test.arn
    type   = "CUSTOMER_KMS"
  }

  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = [%[2]q]
  }
}
`, rName, statelessDefaultActions)
}

// The KMS key resource must stay in state while removing encryption configuration. If not
// (ie. using the _basic config), the KMS key is deleted before the firewall policy is updated,
// leaving the policy in a "misconfigured" state. This causes update to fail with:
//
// InvalidRequestException: firewall policy has KMS key misconfigured
func testAccFirewallPolicyConfig_encryptionConfigurationDisabled(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {}

resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q
  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]
  }
}
`, rName)
}
