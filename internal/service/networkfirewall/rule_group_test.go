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

func TestAccNetworkFirewallRuleGroup_Basic_rulesSourceList(t *testing.T) {
	ctx := acctest.Context(t)
	var ruleGroup networkfirewall.DescribeRuleGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_basicSourceList(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &ruleGroup),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "network-firewall", fmt.Sprintf("stateful-rulegroup/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "100"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(awstypes.RuleGroupTypeStateful)),
					resource.TestCheckResourceAttr(resourceName, "rule_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.reference_sets.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.rules_source_list.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.rules_source_list.0.generated_rules_type", string(awstypes.GeneratedRulesTypeAllowlist)),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.rules_source_list.0.target_types.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_group.0.rules_source.0.rules_source_list.0.target_types.*", string(awstypes.TargetTypeHttpHost)),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.rules_source_list.0.targets.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_group.0.rules_source.0.rules_source_list.0.targets.*", "test.example.com"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rule_variables.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.stateful_rule_options.#", acctest.Ct0),
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

func TestAccNetworkFirewallRuleGroup_Basic_referenceSets(t *testing.T) {
	ctx := acctest.Context(t)
	var ruleGroup networkfirewall.DescribeRuleGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_referenceSets(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &ruleGroup),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "network-firewall", fmt.Sprintf("stateful-rulegroup/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "100"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(awstypes.RuleGroupTypeStateful)),
					resource.TestCheckResourceAttr(resourceName, "rule_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.reference_sets.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.reference_sets.0.ip_set_references.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.reference_sets.0.ip_set_references.0.ip_set_reference.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.reference_sets.0.ip_set_references.1.ip_set_reference.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.reference_sets.0.ip_set_references.2.ip_set_reference.#", acctest.Ct1),
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

func TestAccNetworkFirewallRuleGroup_Basic_updateReferenceSets(t *testing.T) {
	ctx := acctest.Context(t)
	var ruleGroup networkfirewall.DescribeRuleGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_referenceSets(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &ruleGroup),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "network-firewall", fmt.Sprintf("stateful-rulegroup/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "100"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(awstypes.RuleGroupTypeStateful)),
					resource.TestCheckResourceAttr(resourceName, "rule_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.reference_sets.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.reference_sets.0.ip_set_references.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.reference_sets.0.ip_set_references.0.ip_set_reference.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.reference_sets.0.ip_set_references.1.ip_set_reference.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.reference_sets.0.ip_set_references.2.ip_set_reference.#", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRuleGroupConfig_referenceSets1(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &ruleGroup),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "network-firewall", fmt.Sprintf("stateful-rulegroup/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "100"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(awstypes.RuleGroupTypeStateful)),
					resource.TestCheckResourceAttr(resourceName, "rule_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.reference_sets.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.reference_sets.0.ip_set_references.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.reference_sets.0.ip_set_references.0.ip_set_reference.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.reference_sets.0.ip_set_references.1.ip_set_reference.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.reference_sets.0.ip_set_references.2.ip_set_reference.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccNetworkFirewallRuleGroup_Basic_statefulRule(t *testing.T) {
	ctx := acctest.Context(t)
	var ruleGroup networkfirewall.DescribeRuleGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_basicStateful(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &ruleGroup),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "network-firewall", fmt.Sprintf("stateful-rulegroup/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "100"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(awstypes.RuleGroupTypeStateful)),
					resource.TestCheckResourceAttr(resourceName, "rule_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.action", string(awstypes.StatefulActionPass)),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.header.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.header.0.destination", "124.1.1.24/32"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.header.0.destination_port", "53"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.header.0.direction", string(awstypes.StatefulRuleDirectionAny)),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.header.0.protocol", string(awstypes.StatefulRuleProtocolTcp)),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.header.0.source", "1.2.3.4/32"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.header.0.source_port", "53"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.rule_option.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.rule_option.*", map[string]string{
						"keyword":    "sid",
						"settings.#": acctest.Ct1,
						"settings.0": acctest.Ct1,
					}),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.stateful_rule_options.#", acctest.Ct0),
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

func TestAccNetworkFirewallRuleGroup_Basic_statelessRule(t *testing.T) {
	ctx := acctest.Context(t)
	var ruleGroup networkfirewall.DescribeRuleGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_basicStateless(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &ruleGroup),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "network-firewall", fmt.Sprintf("stateless-rulegroup/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "100"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(awstypes.RuleGroupTypeStateless)),
					resource.TestCheckResourceAttr(resourceName, "rule_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateless_rules_and_custom_actions.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule_group.0.rules_source.0.stateless_rules_and_custom_actions.0.stateless_rule.*", map[string]string{
						names.AttrPriority:                                   acctest.Ct1,
						"rule_definition.#":                                  acctest.Ct1,
						"rule_definition.0.actions.#":                        acctest.Ct1,
						"rule_definition.0.match_attributes.#":               acctest.Ct1,
						"rule_definition.0.match_attributes.0.destination.#": acctest.Ct1,
						"rule_definition.0.match_attributes.0.source.#":      acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_group.0.rules_source.0.stateless_rules_and_custom_actions.0.stateless_rule.*.rule_definition.0.actions.*", "aws:drop"),
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

func TestAccNetworkFirewallRuleGroup_Basic_rules(t *testing.T) {
	ctx := acctest.Context(t)
	var ruleGroup networkfirewall.DescribeRuleGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_rule_group.test"
	rules := `#test comment
alert http any any -> any any (http_response_line; content:"403 Forbidden"; sid:1;)`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_basic(rName, rules),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &ruleGroup),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "network-firewall", fmt.Sprintf("stateful-rulegroup/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "100"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(awstypes.RuleGroupTypeStateful)),
					resource.TestCheckResourceAttr(resourceName, "rules", rules),
					resource.TestCheckResourceAttr(resourceName, "rule_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.rules_string", rules),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"rules"}, // argument not returned in RuleGroup API response
			},
		},
	})
}

func TestAccNetworkFirewallRuleGroup_statefulRuleOptions(t *testing.T) {
	ctx := acctest.Context(t)
	var ruleGroup networkfirewall.DescribeRuleGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_rule_group.test"
	rules := `alert http any any -> any any (http_response_line; content:"403 Forbidden"; sid:1;)`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_statefulOptions(rName, rules, "STRICT_ORDER"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &ruleGroup),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.stateful_rule_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.stateful_rule_options.0.rule_order", string(awstypes.RuleOrderStrictOrder)),
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

func TestAccNetworkFirewallRuleGroup_updateStatefulRuleOptions(t *testing.T) {
	ctx := acctest.Context(t)
	var ruleGroup1, ruleGroup2, ruleGroup3 networkfirewall.DescribeRuleGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_rule_group.test"
	rules := `alert http any any -> any any (http_response_line; content:"403 Forbidden"; sid:1;)`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_statefulOptions(rName, rules, "STRICT_ORDER"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &ruleGroup1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.stateful_rule_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.stateful_rule_options.0.rule_order", string(awstypes.RuleOrderStrictOrder)),
				),
			},
			{
				Config: testAccRuleGroupConfig_statefulOptions(rName, rules, "DEFAULT_ACTION_ORDER"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &ruleGroup2),
					testAccCheckRuleGroupRecreated(&ruleGroup1, &ruleGroup2),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.stateful_rule_options.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.stateful_rule_options.0.rule_order", string(awstypes.RuleOrderDefaultActionOrder)),
				),
			},
			{
				Config: testAccRuleGroupConfig_sourceString(rName, rules),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &ruleGroup3),
					testAccCheckRuleGroupNotRecreated(&ruleGroup2, &ruleGroup3),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.stateful_rule_options.#", acctest.Ct0),
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

func TestAccNetworkFirewallRuleGroup_statelessRuleWithCustomAction(t *testing.T) {
	ctx := acctest.Context(t)
	var ruleGroup networkfirewall.DescribeRuleGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_statelessCustomAction(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &ruleGroup),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "network-firewall", fmt.Sprintf("stateless-rulegroup/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "100"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(awstypes.RuleGroupTypeStateless)),
					resource.TestCheckResourceAttr(resourceName, "rule_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateless_rules_and_custom_actions.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule_group.0.rules_source.0.stateless_rules_and_custom_actions.0.stateless_rule.*", map[string]string{
						names.AttrPriority:                                   acctest.Ct1,
						"rule_definition.#":                                  acctest.Ct1,
						"rule_definition.0.actions.#":                        acctest.Ct2,
						"rule_definition.0.match_attributes.#":               acctest.Ct1,
						"rule_definition.0.match_attributes.0.destination.#": acctest.Ct1,
						"rule_definition.0.match_attributes.0.source.#":      acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_group.0.rules_source.0.stateless_rules_and_custom_actions.0.stateless_rule.*.rule_definition.0.actions.*", "aws:pass"),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_group.0.rules_source.0.stateless_rules_and_custom_actions.0.stateless_rule.*.rule_definition.0.actions.*", "example"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule_group.0.rules_source.0.stateless_rules_and_custom_actions.0.custom_action.*", map[string]string{
						"action_name":         "example",
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

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/19414
func TestAccNetworkFirewallRuleGroup_updateRules(t *testing.T) {
	ctx := acctest.Context(t)
	var ruleGroup networkfirewall.DescribeRuleGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_rule_group.test"

	rules := `pass tls $HOME_NET any -> $EXTERNAL_NET 443 (tls.sni; content:"OLD.example.com"; msg:"FQDN test"; sid:1;)`
	updatedRules := `pass tls $HOME_NET any -> $EXTERNAL_NET 443 (tls.sni; content:"NEW.example.com"; msg:"FQDN test"; sid:1;)`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_basic(rName, rules),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &ruleGroup),
				),
			},
			{
				Config: testAccRuleGroupConfig_basic(rName, updatedRules),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &ruleGroup),
					resource.TestCheckResourceAttr(resourceName, "rules", updatedRules),
					resource.TestCheckResourceAttr(resourceName, "rule_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.rules_string", updatedRules),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.#", acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"rules"}, // argument not returned in RuleGroup API response
			},
		},
	})
}

func TestAccNetworkFirewallRuleGroup_updateRulesSourceList(t *testing.T) {
	ctx := acctest.Context(t)
	var ruleGroup networkfirewall.DescribeRuleGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_basicSourceList(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &ruleGroup),
				),
			},
			{
				Config: testAccRuleGroupConfig_updateSourceList(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &ruleGroup),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "network-firewall", fmt.Sprintf("stateful-rulegroup/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "capacity", "100"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(awstypes.RuleGroupTypeStateful)),
					resource.TestCheckResourceAttr(resourceName, "rule_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.rules_source_list.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.rules_source_list.0.generated_rules_type", string(awstypes.GeneratedRulesTypeDenylist)),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.rules_source_list.0.target_types.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_group.0.rules_source.0.rules_source_list.0.target_types.*", string(awstypes.TargetTypeHttpHost)),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_group.0.rules_source.0.rules_source_list.0.target_types.*", string(awstypes.TargetTypeTlsSni)),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.rules_source_list.0.targets.#", acctest.Ct2),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_group.0.rules_source.0.rules_source_list.0.targets.*", "test.example.com"),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_group.0.rules_source.0.rules_source_list.0.targets.*", "test2.example.com"),
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

func TestAccNetworkFirewallRuleGroup_rulesSourceAndRuleVariables(t *testing.T) {
	ctx := acctest.Context(t)
	var ruleGroup networkfirewall.DescribeRuleGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_sourceListVariables(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &ruleGroup),
					resource.TestCheckResourceAttr(resourceName, "rule_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rule_variables.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rule_variables.0.ip_sets.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule_group.0.rule_variables.0.ip_sets.*", map[string]string{
						names.AttrKey:           "example",
						"ip_set.#":              acctest.Ct1,
						"ip_set.0.definition.#": acctest.Ct2,
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_group.0.rule_variables.0.ip_sets.*.ip_set.0.definition.*", "10.0.0.0/16"),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_group.0.rule_variables.0.ip_sets.*.ip_set.0.definition.*", "10.0.1.0/24"),
				),
			},
			{
				Config: testAccRuleGroupConfig_sourceListUpdateVariables(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &ruleGroup),
					resource.TestCheckResourceAttr(resourceName, "rule_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rule_variables.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rule_variables.0.ip_sets.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rule_variables.0.port_sets.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule_group.0.rule_variables.0.ip_sets.*", map[string]string{
						names.AttrKey:           "example",
						"ip_set.#":              acctest.Ct1,
						"ip_set.0.definition.#": acctest.Ct3,
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_group.0.rule_variables.0.ip_sets.*.ip_set.0.definition.*", "10.0.0.0/16"),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_group.0.rule_variables.0.ip_sets.*.ip_set.0.definition.*", "10.0.1.0/24"),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_group.0.rule_variables.0.ip_sets.*.ip_set.0.definition.*", "192.168.0.0/16"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule_group.0.rule_variables.0.ip_sets.*", map[string]string{
						names.AttrKey:           "example2",
						"ip_set.#":              acctest.Ct1,
						"ip_set.0.definition.#": acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_group.0.rule_variables.0.ip_sets.*.ip_set.0.definition.*", "1.2.3.4/32"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule_group.0.rule_variables.0.port_sets.*", map[string]string{
						names.AttrKey:             "example",
						"port_set.#":              acctest.Ct1,
						"port_set.0.definition.#": acctest.Ct2,
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_group.0.rule_variables.0.port_sets.*.port_set.0.definition.*", "443"),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_group.0.rule_variables.0.port_sets.*.port_set.0.definition.*", "80"),
				),
			},
			{
				Config: testAccRuleGroupConfig_basicSourceList(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &ruleGroup),
					resource.TestCheckResourceAttr(resourceName, "rule_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rule_variables.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.#", acctest.Ct1),
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

// TestAccNetworkFirewallRuleGroup_updateStatefulRule validates
// in-place updates to a single stateful_rule configuration block
func TestAccNetworkFirewallRuleGroup_updateStatefulRule(t *testing.T) {
	ctx := acctest.Context(t)
	var ruleGroup networkfirewall.DescribeRuleGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_basicStateful(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &ruleGroup),
				),
			},
			{
				Config: testAccRuleGroupConfig_updateStateful(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &ruleGroup),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.action", string(awstypes.StatefulActionDrop)),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.header.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.header.0.destination", "1.2.3.4/32"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.header.0.destination_port", "1001"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.header.0.direction", string(awstypes.StatefulRuleDirectionForward)),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.header.0.protocol", string(awstypes.StatefulRuleProtocolAny)),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.header.0.source", "124.1.1.24/32"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.header.0.source_port", "1001"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.rule_option.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.rule_option.*", map[string]string{
						"keyword":    "sid",
						"settings.#": acctest.Ct1,
						"settings.0": acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.rule_option.*", map[string]string{
						"keyword":    "rev",
						"settings.#": acctest.Ct1,
						"settings.0": acctest.Ct2,
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

// TestAccNetworkFirewallRuleGroup_updateMultipleStatefulRules validates
// in-place updates to stateful_rule configuration blocks
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/16868
func TestAccNetworkFirewallRuleGroup_updateMultipleStatefulRules(t *testing.T) {
	ctx := acctest.Context(t)
	var ruleGroup networkfirewall.DescribeRuleGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_basicStateful(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &ruleGroup),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.#", acctest.Ct1),
				),
			},
			{
				Config: testAccRuleGroupConfig_multipleStateful(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &ruleGroup),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.action", string(awstypes.StatefulActionPass)),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.header.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.header.0.destination", "124.1.1.24/32"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.header.0.destination_port", "53"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.header.0.direction", string(awstypes.StatefulRuleDirectionAny)),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.header.0.protocol", string(awstypes.StatefulRuleProtocolTcp)),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.header.0.source", "1.2.3.4/32"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.header.0.source_port", "53"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.rule_option.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.1.action", string(awstypes.StatefulActionAlert)),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.1.header.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.1.header.0.destination", string(awstypes.StatefulRuleDirectionAny)),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.1.header.0.destination_port", string(awstypes.StatefulRuleDirectionAny)),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.1.header.0.direction", string(awstypes.StatefulRuleDirectionAny)),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.1.header.0.protocol", string(awstypes.StatefulRuleProtocolAny)),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.1.header.0.source", string(awstypes.StatefulRuleDirectionAny)),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.1.header.0.source_port", string(awstypes.StatefulRuleDirectionAny)),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.1.rule_option.#", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRuleGroupConfig_updateStateful(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &ruleGroup),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.action", string(awstypes.StatefulActionDrop)),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.header.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.header.0.destination", "1.2.3.4/32"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.header.0.destination_port", "1001"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.header.0.direction", string(awstypes.StatefulRuleDirectionForward)),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.header.0.protocol", string(awstypes.StatefulRuleProtocolAny)),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.header.0.source", "124.1.1.24/32"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.header.0.source_port", "1001"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.rule_option.#", acctest.Ct2),
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

// TestAccNetworkFirewallRuleGroup_StatefulRule_action validates in-place
// updates to the "action" argument within 1 stateful_rule configuration block
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/16868
func TestAccNetworkFirewallRuleGroup_StatefulRule_action(t *testing.T) {
	ctx := acctest.Context(t)
	var ruleGroup networkfirewall.DescribeRuleGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_statefulAction(rName, string(awstypes.StatefulActionAlert)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &ruleGroup),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.action", string(awstypes.StatefulActionAlert)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRuleGroupConfig_statefulAction(rName, string(awstypes.StatefulActionPass)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &ruleGroup),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.action", string(awstypes.StatefulActionPass)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRuleGroupConfig_statefulAction(rName, string(awstypes.StatefulActionDrop)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &ruleGroup),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.action", string(awstypes.StatefulActionDrop)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRuleGroupConfig_statefulAction(rName, string(awstypes.StatefulActionReject)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &ruleGroup),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.action", string(awstypes.StatefulActionReject)),
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

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/16470
func TestAccNetworkFirewallRuleGroup_StatefulRule_header(t *testing.T) {
	ctx := acctest.Context(t)
	var ruleGroup networkfirewall.DescribeRuleGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_statefulHeader(rName, "1990", "1994"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &ruleGroup),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.action", string(awstypes.StatefulActionPass)),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.header.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.header.0.destination", "ANY"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.header.0.destination_port", "1990"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.header.0.direction", string(awstypes.StatefulRuleDirectionAny)),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.header.0.protocol", string(awstypes.StatefulRuleProtocolTcp)),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.header.0.source", "ANY"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.header.0.source_port", "1994"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.rule_option.#", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRuleGroupConfig_statefulHeader(rName, "ANY", "ANY"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &ruleGroup),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.action", string(awstypes.StatefulActionPass)),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.header.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.header.0.destination", "ANY"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.header.0.destination_port", "ANY"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.header.0.direction", string(awstypes.StatefulRuleDirectionAny)),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.header.0.protocol", string(awstypes.StatefulRuleProtocolTcp)),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.header.0.source", "ANY"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.header.0.source_port", "ANY"),
					resource.TestCheckResourceAttr(resourceName, "rule_group.0.rules_source.0.stateful_rule.0.rule_option.#", acctest.Ct1),
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

func TestAccNetworkFirewallRuleGroup_updateStatelessRule(t *testing.T) {
	ctx := acctest.Context(t)
	var ruleGroup networkfirewall.DescribeRuleGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_basicStateless(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &ruleGroup),
				),
			},
			{
				Config: testAccRuleGroupConfig_updateStateless(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &ruleGroup),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule_group.0.rules_source.0.stateless_rules_and_custom_actions.0.stateless_rule.*", map[string]string{
						names.AttrPriority:                                        acctest.Ct10,
						"rule_definition.#":                                       acctest.Ct1,
						"rule_definition.0.actions.#":                             acctest.Ct1,
						"rule_definition.0.match_attributes.#":                    acctest.Ct1,
						"rule_definition.0.match_attributes.0.destination.#":      acctest.Ct1,
						"rule_definition.0.match_attributes.0.destination_port.#": acctest.Ct1,
						"rule_definition.0.match_attributes.0.protocols.#":        acctest.Ct1,
						"rule_definition.0.match_attributes.0.source.#":           acctest.Ct1,
						"rule_definition.0.match_attributes.0.source_port.#":      acctest.Ct1,
						"rule_definition.0.match_attributes.0.tcp_flag.#":         acctest.Ct1,
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_group.0.rules_source.0.stateless_rules_and_custom_actions.0.stateless_rule.*.rule_definition.0.actions.*", "aws:pass"),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_group.0.rules_source.0.stateless_rules_and_custom_actions.0.stateless_rule.*.rule_definition.0.match_attributes.0.protocols.*", "6"),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_group.0.rules_source.0.stateless_rules_and_custom_actions.0.stateless_rule.*.rule_definition.0.match_attributes.0.tcp_flag.*.flags.*", "SYN"),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_group.0.rules_source.0.stateless_rules_and_custom_actions.0.stateless_rule.*.rule_definition.0.match_attributes.0.tcp_flag.*.masks.*", "SYN"),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_group.0.rules_source.0.stateless_rules_and_custom_actions.0.stateless_rule.*.rule_definition.0.match_attributes.0.tcp_flag.*.masks.*", "ACK"),
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

func TestAccNetworkFirewallRuleGroup_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var ruleGroup networkfirewall.DescribeRuleGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &ruleGroup),
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
				Config: testAccRuleGroupConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &ruleGroup),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccRuleGroupConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccNetworkFirewallRuleGroup_encryptionConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	var ruleGroup networkfirewall.DescribeRuleGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_encryptionConfiguration(rName, "ALLOWLIST"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &ruleGroup),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.type", "CUSTOMER_KMS"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRuleGroupConfig_encryptionConfigurationDisabled(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &ruleGroup),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", acctest.Ct0),
				),
			},
			{
				Config: testAccRuleGroupConfig_encryptionConfiguration(rName, "ALLOWLIST"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &ruleGroup),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.type", "CUSTOMER_KMS"),
				),
			},
			{
				Config: testAccRuleGroupConfig_encryptionConfiguration(rName, "DENYLIST"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &ruleGroup),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "encryption_configuration.0.type", "CUSTOMER_KMS"),
				),
			},
		},
	})
}

func TestAccNetworkFirewallRuleGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var ruleGroup networkfirewall.DescribeRuleGroupOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRuleGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRuleGroupConfig_basicSourceList(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRuleGroupExists(ctx, resourceName, &ruleGroup),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfnetworkfirewall.ResourceRuleGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckRuleGroupDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networkfirewall_rule_group" {
				continue
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkFirewallClient(ctx)

			_, err := tfnetworkfirewall.FindRuleGroupByARN(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("NetworkFirewall Rule Group %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckRuleGroupExists(ctx context.Context, n string, v *networkfirewall.DescribeRuleGroupOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkFirewallClient(ctx)

		output, err := tfnetworkfirewall.FindRuleGroupByARN(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckRuleGroupNotRecreated(i, j *networkfirewall.DescribeRuleGroupOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(i.RuleGroupResponse.RuleGroupId), aws.ToString(j.RuleGroupResponse.RuleGroupId); before != after {
			return fmt.Errorf("NetworkFirewall Rule Group was recreated. got: %s, expected: %s", after, before)
		}
		return nil
	}
}

func testAccCheckRuleGroupRecreated(i, j *networkfirewall.DescribeRuleGroupOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(i.RuleGroupResponse.RuleGroupId), aws.ToString(j.RuleGroupResponse.RuleGroupId); before == after {
			return fmt.Errorf("NetworkFirewall Rule Group (%s) was not recreated", before)
		}
		return nil
	}
}

func testAccRuleGroupConfig_basicSourceList(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_rule_group" "test" {
  capacity = 100
  name     = %[1]q
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
}
`, rName)
}

func testAccRuleGroupConfig_referenceSets(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "example1" {
  name           = "All VPC CIDR-s"
  address_family = "IPv4"
  max_entries    = 5
}

resource "aws_ec2_managed_prefix_list" "example2" {
  name           = "SOME VPC CIDR-s"
  address_family = "IPv4"
  max_entries    = 5
}

resource "aws_ec2_managed_prefix_list" "example3" {
  name           = "FEW VPC CIDR-s"
  address_family = "IPv4"
  max_entries    = 5
}

resource "aws_networkfirewall_rule_group" "test" {
  capacity = 100
  name     = %[1]q
  type     = "STATEFUL"

  rule_group {
    reference_sets {
      ip_set_references {
        key = "example1"
        ip_set_reference {
          reference_arn = aws_ec2_managed_prefix_list.example1.arn
        }
      }

      ip_set_references {
        key = "example2"
        ip_set_reference {
          reference_arn = aws_ec2_managed_prefix_list.example2.arn
        }
      }

      ip_set_references {
        key = "example3"
        ip_set_reference {
          reference_arn = aws_ec2_managed_prefix_list.example3.arn
        }
      }
    }

    rules_source {
      rules_source_list {
        generated_rules_type = "ALLOWLIST"
        target_types         = ["HTTP_HOST"]
        targets              = ["test.example.com"]
      }
    }
  }
}
`, rName)
}

func testAccRuleGroupConfig_referenceSets1(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_managed_prefix_list" "example1" {
  name           = "All VPC CIDR-s"
  address_family = "IPv4"
  max_entries    = 5
}

resource "aws_ec2_managed_prefix_list" "example2" {
  name           = "SOME VPC CIDR-s"
  address_family = "IPv4"
  max_entries    = 5
}

resource "aws_ec2_managed_prefix_list" "example3" {
  name           = "FEW VPC CIDR-s"
  address_family = "IPv4"
  max_entries    = 5
}

resource "aws_networkfirewall_rule_group" "test" {
  capacity = 100
  name     = %[1]q
  type     = "STATEFUL"

  rule_group {
    reference_sets {
      ip_set_references {
        key = "example11"
        ip_set_reference {
          reference_arn = aws_ec2_managed_prefix_list.example1.arn
        }
      }

      ip_set_references {
        key = "example21"
        ip_set_reference {
          reference_arn = aws_ec2_managed_prefix_list.example2.arn
        }
      }

      ip_set_references {
        key = "example31"
        ip_set_reference {
          reference_arn = aws_ec2_managed_prefix_list.example3.arn
        }
      }
    }

    rules_source {
      rules_source_list {
        generated_rules_type = "ALLOWLIST"
        target_types         = ["HTTP_HOST"]
        targets              = ["test.example.com"]
      }
    }
  }
}
`, rName)
}

func testAccRuleGroupConfig_sourceListVariables(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_rule_group" "test" {
  capacity = 100
  name     = %[1]q
  type     = "STATEFUL"

  rule_group {
    rule_variables {
      ip_sets {
        key = "example"
        ip_set {
          definition = ["10.0.0.0/16", "10.0.1.0/24"]
        }
      }
    }

    rules_source {
      rules_source_list {
        generated_rules_type = "ALLOWLIST"
        target_types         = ["HTTP_HOST"]
        targets              = ["test.example.com"]
      }
    }
  }
}
`, rName)
}

func testAccRuleGroupConfig_sourceListUpdateVariables(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_rule_group" "test" {
  capacity = 100
  name     = %[1]q
  type     = "STATEFUL"

  rule_group {
    rule_variables {
      ip_sets {
        key = "example"
        ip_set {
          definition = ["10.0.0.0/16", "10.0.1.0/24", "192.168.0.0/16"]
        }
      }

      ip_sets {
        key = "example2"
        ip_set {
          definition = ["1.2.3.4/32"]
        }
      }

      port_sets {
        key = "example"
        port_set {
          definition = ["443", "80"]
        }
      }
    }

    rules_source {
      rules_source_list {
        generated_rules_type = "ALLOWLIST"
        target_types         = ["HTTP_HOST"]
        targets              = ["test.example.com"]
      }
    }
  }
}
`, rName)
}

func testAccRuleGroupConfig_updateSourceList(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_rule_group" "test" {
  capacity = 100
  name     = %[1]q
  type     = "STATEFUL"

  rule_group {
    rules_source {
      rules_source_list {
        generated_rules_type = "DENYLIST"
        target_types         = ["HTTP_HOST", "TLS_SNI"]
        targets              = ["test.example.com", "test2.example.com"]
      }
    }
  }
}
`, rName)
}

func testAccRuleGroupConfig_basicStateful(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_rule_group" "test" {
  capacity    = 100
  name        = %[1]q
  description = %[1]q
  type        = "STATEFUL"

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
  }
}
`, rName)
}

func testAccRuleGroupConfig_statefulAction(rName, action string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_rule_group" "test" {
  capacity    = 100
  name        = %[1]q
  description = %[1]q
  type        = "STATEFUL"

  rule_group {
    rules_source {
      stateful_rule {
        action = %[2]q

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
  }
}
`, rName, action)
}

func testAccRuleGroupConfig_statefulHeader(rName, dstPort, srcPort string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_rule_group" "test" {
  capacity    = 100
  name        = %[1]q
  description = %[1]q
  type        = "STATEFUL"

  rule_group {
    rules_source {
      stateful_rule {
        action = "PASS"

        header {
          destination      = "ANY"
          destination_port = %[2]q
          direction        = "ANY"
          protocol         = "TCP"
          source           = "ANY"
          source_port      = %[3]q
        }

        rule_option {
          keyword  = "sid"
          settings = ["1"]
        }
      }
    }
  }
}
`, rName, dstPort, srcPort)
}

func testAccRuleGroupConfig_updateStateful(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_rule_group" "test" {
  capacity = 100
  name     = %[1]q
  type     = "STATEFUL"

  rule_group {
    rules_source {
      stateful_rule {
        action = "DROP"

        header {
          destination      = "1.2.3.4/32"
          destination_port = 1001
          direction        = "FORWARD"
          protocol         = "IP"
          source           = "124.1.1.24/32"
          source_port      = 1001
        }

        rule_option {
          keyword  = "sid"
          settings = ["1"]
        }

        rule_option {
          keyword  = "rev"
          settings = ["2"]
        }
      }
    }
  }
}
`, rName)
}

func testAccRuleGroupConfig_multipleStateful(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_rule_group" "test" {
  capacity = 100
  name     = %[1]q
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

      stateful_rule {
        action = "ALERT"

        header {
          destination      = "ANY"
          destination_port = "ANY"
          direction        = "ANY"
          protocol         = "IP"
          source           = "ANY"
          source_port      = "ANY"
        }

        rule_option {
          keyword  = "sid"
          settings = ["2"]
        }
      }
    }
  }
}
`, rName)
}

func testAccRuleGroupConfig_basicStateless(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_rule_group" "test" {
  capacity = 100
  name     = %[1]q
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
}
`, rName)
}

func testAccRuleGroupConfig_updateStateless(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_rule_group" "test" {
  capacity = 100
  name     = %[1]q
  type     = "STATELESS"

  rule_group {
    rules_source {
      stateless_rules_and_custom_actions {
        stateless_rule {
          priority = 10

          rule_definition {
            actions = ["aws:pass"]

            match_attributes {
              destination {
                address_definition = "1.2.3.4/32"
              }

              destination_port {
                from_port = 53
                to_port   = 53
              }

              protocols = [6]

              source {
                address_definition = "124.1.1.5/32"
              }

              source_port {
                from_port = 53
                to_port   = 53
              }

              tcp_flag {
                flags = ["SYN"]
                masks = ["SYN", "ACK"]
              }
            }
          }
        }
      }
    }
  }
}
`, rName)
}

func testAccRuleGroupConfig_basic(rName, rules string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_rule_group" "test" {
  capacity = 100
  name     = %[1]q
  type     = "STATEFUL"
  rules    = %[2]q
}
`, rName, rules)
}

func testAccRuleGroupConfig_sourceString(rName, rules string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_rule_group" "test" {
  capacity = 100
  name     = %[1]q
  type     = "STATEFUL"

  rule_group {
    rules_source {
      rules_string = %[2]q
    }
  }
}
`, rName, rules)
}

func testAccRuleGroupConfig_statefulOptions(rName, rules, ruleOrder string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_rule_group" "test" {
  capacity = 100
  name     = %[1]q
  type     = "STATEFUL"

  rule_group {
    rules_source {
      rules_string = %[2]q
    }

    stateful_rule_options {
      rule_order = %[3]q
    }
  }
}
`, rName, rules, ruleOrder)
}

func testAccRuleGroupConfig_statelessCustomAction(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_rule_group" "test" {
  capacity = 100
  name     = %[1]q
  type     = "STATELESS"

  rule_group {
    rules_source {
      stateless_rules_and_custom_actions {
        custom_action {
          action_name = "example"

          action_definition {
            publish_metric_action {
              dimension {
                value = "2"
              }
            }
          }
        }

        stateless_rule {
          priority = 1

          rule_definition {
            actions = ["aws:pass", "example"]

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
}
`, rName)
}

func testAccRuleGroupConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_rule_group" "test" {
  capacity = 100
  name     = %[1]q
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

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccRuleGroupConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_rule_group" "test" {
  capacity = 100
  name     = %[1]q
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

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccRuleGroupConfig_encryptionConfiguration(rName, generatedRulesType string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {}

resource "aws_networkfirewall_rule_group" "test" {
  capacity = 100
  name     = %[1]q
  type     = "STATEFUL"

  rule_group {
    rules_source {
      rules_source_list {
        generated_rules_type = %[2]q
        target_types         = ["HTTP_HOST"]
        targets              = ["test.example.com"]
      }
    }
  }

  encryption_configuration {
    key_id = aws_kms_key.test.arn
    type   = "CUSTOMER_KMS"
  }
}
`, rName, generatedRulesType)
}

// The KMS key resource must stay in state while removing encryption configuration. If not
// (ie. using the _basic config), the KMS key is deleted before the rule group is updated,
// leaving the group in a "misconfigured" state. This causes update to fail with:
//
// InvalidRequestException: rule group has KMS key misconfigured
func testAccRuleGroupConfig_encryptionConfigurationDisabled(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {}

resource "aws_networkfirewall_rule_group" "test" {
  capacity = 100
  name     = %[1]q
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
}
`, rName)
}
