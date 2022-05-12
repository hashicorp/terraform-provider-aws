package networkfirewall_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkfirewall"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfnetworkfirewall "github.com/hashicorp/terraform-provider-aws/internal/service/networkfirewall"
)

func TestAccNetworkFirewallFirewallPolicy_basic(t *testing.T) {
	var firewallPolicy networkfirewall.DescribeFirewallPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, networkfirewall.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFirewallPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicy_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(resourceName, &firewallPolicy),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "network-firewall", fmt.Sprintf("firewall-policy/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_fragment_default_actions.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "firewall_policy.0.stateless_fragment_default_actions.*", "aws:drop"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_default_actions.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "firewall_policy.0.stateless_default_actions.*", "aws:pass"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccNetworkFirewallFirewallPolicy_statefulDefaultActions(t *testing.T) {
	var firewallPolicy networkfirewall.DescribeFirewallPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, networkfirewall.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFirewallPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicy_statefulDefaultActions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_default_actions.#", "1"),
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
	var firewallPolicy networkfirewall.DescribeFirewallPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, networkfirewall.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFirewallPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicy_statefulEngineOptions(rName, "STRICT_ORDER"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_engine_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_engine_options.0.rule_order", networkfirewall.RuleOrderStrictOrder),
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
	var firewallPolicy1, firewallPolicy2, firewallPolicy3 networkfirewall.DescribeFirewallPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, networkfirewall.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFirewallPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicy_statefulEngineOptions(rName, "DEFAULT_ACTION_ORDER"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(resourceName, &firewallPolicy1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_engine_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_engine_options.0.rule_order", networkfirewall.RuleOrderDefaultActionOrder),
				),
			},
			{
				Config: testAccFirewallPolicy_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(resourceName, &firewallPolicy2),
					testAccCheckFirewallPolicyNotRecreated(&firewallPolicy1, &firewallPolicy2),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", "1"),
				),
			},
			{
				Config: testAccFirewallPolicy_statefulEngineOptions(rName, "STRICT_ORDER"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(resourceName, &firewallPolicy3),
					testAccCheckFirewallPolicyRecreated(&firewallPolicy2, &firewallPolicy3),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_engine_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_engine_options.0.rule_order", networkfirewall.RuleOrderStrictOrder),
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
	var firewallPolicy networkfirewall.DescribeFirewallPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy.test"
	ruleGroupResourceName := "aws_networkfirewall_rule_group.test.0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, networkfirewall.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFirewallPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicy_statefulRuleGroupReference(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_default_actions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_engine_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_rule_group_reference.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_policy.0.stateful_rule_group_reference.*.resource_arn", ruleGroupResourceName, "arn"),
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

func TestAccNetworkFirewallFirewallPolicy_statefulRuleGroupReferenceManaged(t *testing.T) {
	var firewallPolicy networkfirewall.DescribeFirewallPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, networkfirewall.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFirewallPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicy_statefulRuleGroupReferenceManaged(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(resourceName, &firewallPolicy),
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

func TestAccNetworkFirewallFirewallPolicy_updateStatefulRuleGroupReference(t *testing.T) {
	var firewallPolicy networkfirewall.DescribeFirewallPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy.test"
	ruleGroupResourceName := "aws_networkfirewall_rule_group.test.0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, networkfirewall.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFirewallPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicy_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(resourceName, &firewallPolicy),
				),
			},
			{
				Config: testAccFirewallPolicy_statefulRuleGroupReference(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_rule_group_reference.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_policy.0.stateful_rule_group_reference.*.resource_arn", ruleGroupResourceName, "arn"),
				),
			},
			{
				Config: testAccFirewallPolicy_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(resourceName, &firewallPolicy),
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
	var firewallPolicy networkfirewall.DescribeFirewallPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy.test"
	ruleGroupResourceName1 := "aws_networkfirewall_rule_group.test.0"
	ruleGroupResourceName2 := "aws_networkfirewall_rule_group.test.1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, networkfirewall.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFirewallPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicy_multipleStatefulRuleGroupReferences(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_rule_group_reference.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_policy.0.stateful_rule_group_reference.*.resource_arn", ruleGroupResourceName1, "arn"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_policy.0.stateful_rule_group_reference.*.resource_arn", ruleGroupResourceName2, "arn"),
				),
			},
			{
				Config: testAccFirewallPolicy_singleStatefulRuleGroupReference(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_rule_group_reference.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_policy.0.stateful_rule_group_reference.*.resource_arn", ruleGroupResourceName1, "arn"),
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

func TestAccNetworkFirewallFirewallPolicy_statefulRuleGroupPriorityReference(t *testing.T) {
	var firewallPolicy networkfirewall.DescribeFirewallPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy.test"
	ruleGroupResourceName := "aws_networkfirewall_rule_group.test.0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, networkfirewall.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFirewallPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicy_statefulRuleGroupPriorityReference(rName, "1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_rule_group_reference.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_rule_group_reference.0.priority", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_policy.0.stateful_rule_group_reference.*.resource_arn", ruleGroupResourceName, "arn"),
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

func TestAccNetworkFirewallFirewallPolicy_updateStatefulRuleGroupPriorityReference(t *testing.T) {
	var firewallPolicy1, firewallPolicy2 networkfirewall.DescribeFirewallPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy.test"
	ruleGroupResourceName := "aws_networkfirewall_rule_group.test.0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, networkfirewall.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFirewallPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicy_statefulRuleGroupPriorityReference(rName, "1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(resourceName, &firewallPolicy1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_rule_group_reference.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_rule_group_reference.0.priority", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_policy.0.stateful_rule_group_reference.*.resource_arn", ruleGroupResourceName, "arn"),
				),
			},
			{
				Config: testAccFirewallPolicy_statefulRuleGroupPriorityReference(rName, "2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(resourceName, &firewallPolicy2),
					testAccCheckFirewallPolicyNotRecreated(&firewallPolicy1, &firewallPolicy2),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_rule_group_reference.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_rule_group_reference.0.priority", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_policy.0.stateful_rule_group_reference.*.resource_arn", ruleGroupResourceName, "arn"),
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
	var firewallPolicy networkfirewall.DescribeFirewallPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy.test"
	ruleGroupResourceName := "aws_networkfirewall_rule_group.test.0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, networkfirewall.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFirewallPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicy_statelessRuleGroupReference(rName, 20),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_rule_group_reference.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_policy.0.stateless_rule_group_reference.*.resource_arn", ruleGroupResourceName, "arn"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firewall_policy.0.stateless_rule_group_reference.*", map[string]string{
						"priority": "20",
					}),
				),
			},
			{
				Config: testAccFirewallPolicy_statelessRuleGroupReference(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_rule_group_reference.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firewall_policy.0.stateless_rule_group_reference.*", map[string]string{
						"priority": "1",
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
	var firewallPolicy networkfirewall.DescribeFirewallPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy.test"
	ruleGroupResourceName := "aws_networkfirewall_rule_group.test.0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, networkfirewall.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFirewallPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicy_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(resourceName, &firewallPolicy),
				),
			},
			{
				Config: testAccFirewallPolicy_statelessRuleGroupReference(rName, 20),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_policy.0.stateless_rule_group_reference.*.resource_arn", ruleGroupResourceName, "arn"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firewall_policy.0.stateless_rule_group_reference.*", map[string]string{
						"priority": "20",
					}),
				),
			},
			{
				Config: testAccFirewallPolicy_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_rule_group_reference.#", "0"),
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
	var firewallPolicy networkfirewall.DescribeFirewallPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy.test"
	ruleGroupResourceName1 := "aws_networkfirewall_rule_group.test.0"
	ruleGroupResourceName2 := "aws_networkfirewall_rule_group.test.1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, networkfirewall.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFirewallPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicy_multipleStatelessRuleGroupReferences(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_rule_group_reference.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_policy.0.stateless_rule_group_reference.*.resource_arn", ruleGroupResourceName1, "arn"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firewall_policy.0.stateless_rule_group_reference.*", map[string]string{
						"priority": "1",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_policy.0.stateless_rule_group_reference.*.resource_arn", ruleGroupResourceName2, "arn"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firewall_policy.0.stateless_rule_group_reference.*", map[string]string{
						"priority": "2",
					}),
				),
			},
			{
				Config: testAccFirewallPolicy_singleStatelessRuleGroupReference(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_rule_group_reference.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firewall_policy.0.stateless_rule_group_reference.*", map[string]string{
						"priority": "1",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_policy.0.stateless_rule_group_reference.*.resource_arn", ruleGroupResourceName1, "arn"),
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
	var firewallPolicy networkfirewall.DescribeFirewallPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, networkfirewall.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFirewallPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicy_statelessCustomAction(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_custom_action.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firewall_policy.0.stateless_custom_action.*", map[string]string{
						"action_definition.#":                                     "1",
						"action_definition.0.publish_metric_action.#":             "1",
						"action_definition.0.publish_metric_action.0.dimension.#": "1",
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
	var firewallPolicy1, firewallPolicy2, firewallPolicy3, firewallPolicy4 networkfirewall.DescribeFirewallPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, networkfirewall.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFirewallPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicy_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(resourceName, &firewallPolicy1),
				),
			},
			{
				Config: testAccFirewallPolicy_statelessCustomAction(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(resourceName, &firewallPolicy2),
					testAccCheckFirewallPolicyRecreated(&firewallPolicy1, &firewallPolicy2),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_custom_action.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firewall_policy.0.stateless_custom_action.*", map[string]string{
						"action_name":         "CustomAction",
						"action_definition.#": "1",
						"action_definition.0.publish_metric_action.#":             "1",
						"action_definition.0.publish_metric_action.0.dimension.#": "1",
					}),
				),
			},
			{
				Config: testAccFirewallPolicy_updateStatelessCustomAction(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(resourceName, &firewallPolicy3),
					testAccCheckFirewallPolicyRecreated(&firewallPolicy2, &firewallPolicy3),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_custom_action.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firewall_policy.0.stateless_custom_action.*", map[string]string{
						"action_name":         "updated",
						"action_definition.#": "1",
						"action_definition.0.publish_metric_action.#":             "1",
						"action_definition.0.publish_metric_action.0.dimension.#": "1",
					}),
				),
			},
			{
				Config: testAccFirewallPolicy_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(resourceName, &firewallPolicy4),
					testAccCheckFirewallPolicyRecreated(&firewallPolicy3, &firewallPolicy4),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_custom_action.#", "0"),
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
	var firewallPolicy1, firewallPolicy2 networkfirewall.DescribeFirewallPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, networkfirewall.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFirewallPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicy_multipleStatelessCustomActions(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(resourceName, &firewallPolicy1),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_custom_action.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firewall_policy.0.stateless_custom_action.*", map[string]string{
						"action_name":         "CustomAction",
						"action_definition.#": "1",
						"action_definition.0.publish_metric_action.#":             "1",
						"action_definition.0.publish_metric_action.0.dimension.#": "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firewall_policy.0.stateless_custom_action.*", map[string]string{
						"action_name":         "CustomAction2",
						"action_definition.#": "1",
						"action_definition.0.publish_metric_action.#":             "1",
						"action_definition.0.publish_metric_action.0.dimension.#": "1",
					}),
				),
			},
			{
				Config: testAccFirewallPolicy_statelessCustomAction(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(resourceName, &firewallPolicy2),
					testAccCheckFirewallPolicyRecreated(&firewallPolicy1, &firewallPolicy2),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateless_custom_action.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firewall_policy.0.stateless_custom_action.*", map[string]string{
						"action_name":         "CustomAction",
						"action_definition.#": "1",
						"action_definition.0.publish_metric_action.#":             "1",
						"action_definition.0.publish_metric_action.0.dimension.#": "1",
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
	var firewallPolicy networkfirewall.DescribeFirewallPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy.test"
	ruleGroupResourceName := "aws_networkfirewall_rule_group.test.0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, networkfirewall.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFirewallPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicy_statefulRuleGroupReferenceAndStatelessCustomAction(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_rule_group_reference.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_policy.0.stateful_rule_group_reference.*.resource_arn", ruleGroupResourceName, "arn"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "firewall_policy.0.stateless_custom_action.*", map[string]string{
						"action_name":         "CustomAction",
						"action_definition.#": "1",
						"action_definition.0.publish_metric_action.#":             "1",
						"action_definition.0.publish_metric_action.0.dimension.#": "1",
					}),
				),
			},
			{
				Config: testAccFirewallPolicy_statefulRuleGroupReference(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "firewall_policy.0.stateful_rule_group_reference.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "firewall_policy.0.stateful_rule_group_reference.*.resource_arn", ruleGroupResourceName, "arn"),
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
	var firewallPolicy networkfirewall.DescribeFirewallPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, networkfirewall.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFirewallPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicy_oneTag(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				Config: testAccFirewallPolicy_twoTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Description", "updated"),
				),
			},
			{
				Config: testAccFirewallPolicy_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(resourceName, &firewallPolicy),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccNetworkFirewallFirewallPolicy_disappears(t *testing.T) {
	var firewallPolicy networkfirewall.DescribeFirewallPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_firewall_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, networkfirewall.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckFirewallPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicy_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFirewallPolicyExists(resourceName, &firewallPolicy),
					acctest.CheckResourceDisappears(acctest.Provider, tfnetworkfirewall.ResourceFirewallPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckFirewallPolicyDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_networkfirewall_firewall_policy" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkFirewallConn
		output, err := tfnetworkfirewall.FindFirewallPolicy(context.Background(), conn, rs.Primary.ID)
		if tfawserr.ErrCodeEquals(err, networkfirewall.ErrCodeResourceNotFoundException) {
			continue
		}
		if err != nil {
			return err
		}
		if output != nil {
			return fmt.Errorf("NetworkFirewall Firewall Policy still exists: %s", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckFirewallPolicyExists(n string, p *networkfirewall.DescribeFirewallPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No NetworkFirewall Firewall Policy ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).NetworkFirewallConn
		output, err := tfnetworkfirewall.FindFirewallPolicy(context.Background(), conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("NetworkFirewall Firewall Policy (%s) not found", rs.Primary.ID)
		}

		*p = *output

		return nil
	}
}

func testAccCheckFirewallPolicyNotRecreated(i, j *networkfirewall.DescribeFirewallPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.StringValue(i.FirewallPolicyResponse.FirewallPolicyId), aws.StringValue(j.FirewallPolicyResponse.FirewallPolicyId); before != after {
			return fmt.Errorf("NetworkFirewall Firewall Policy was recreated. got: %s, expected: %s", after, before)
		}
		return nil
	}
}

func testAccCheckFirewallPolicyRecreated(i, j *networkfirewall.DescribeFirewallPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.StringValue(i.FirewallPolicyResponse.FirewallPolicyId), aws.StringValue(j.FirewallPolicyResponse.FirewallPolicyId); before == after {
			return fmt.Errorf("NetworkFirewall Firewall Policy (%s) was not recreated", before)
		}
		return nil
	}
}

func testAccFirewallPolicyStatelessRuleGroupDependencies(rName string, count int) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_rule_group" "test" {
  count    = %d
  capacity = 100
  name     = "%s-${count.index}"
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

func testAccFirewallPolicyStatefulRuleGroupDependencies(rName string, count int) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_rule_group" "test" {
  count    = %d
  capacity = 100
  name     = "%s-${count.index}"
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

func testAccFirewallPolicyStatefulRuleGroupStrictDependencies(rName string, count int) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_rule_group" "test" {
  count    = %d
  capacity = 100
  name     = "%s-${count.index}"
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
          keyword = "sid:1"
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

func testAccFirewallPolicy_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %q
  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]
  }
}
`, rName)
}

func testAccFirewallPolicy_oneTag(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q
  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]
  }
  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccFirewallPolicy_twoTags(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q
  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]
  }
  tags = {
    Name        = %[1]q
    Description = "updated"
  }
}
`, rName)
}

func testAccFirewallPolicy_statefulEngineOptions(rName, rule_order string) string {
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
`, rName, rule_order)
}

func testAccFirewallPolicy_statefulDefaultActions(rName string) string {
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

func testAccFirewallPolicy_statefulRuleGroupReference(rName string) string {
	return acctest.ConfigCompose(
		testAccFirewallPolicyStatefulRuleGroupDependencies(rName, 1),
		fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %q
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

func testAccFirewallPolicy_statefulRuleGroupReferenceManaged(rName string) string {
	return acctest.ConfigCompose(
		testAccFirewallPolicyStatefulRuleGroupDependencies(rName, 1),
		fmt.Sprintf(`
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

func testAccFirewallPolicy_multipleStatefulRuleGroupReferences(rName string) string {
	return acctest.ConfigCompose(
		testAccFirewallPolicyStatefulRuleGroupDependencies(rName, 2),
		fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %q
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

func testAccFirewallPolicy_statefulRuleGroupPriorityReference(rName, priority string) string {
	return acctest.ConfigCompose(
		testAccFirewallPolicyStatefulRuleGroupStrictDependencies(rName, 1),
		fmt.Sprintf(`
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

func testAccFirewallPolicy_singleStatefulRuleGroupReference(rName string) string {
	return acctest.ConfigCompose(
		testAccFirewallPolicyStatefulRuleGroupDependencies(rName, 2),
		fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %q
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

func testAccFirewallPolicy_statelessRuleGroupReference(rName string, priority int) string {
	return acctest.ConfigCompose(
		testAccFirewallPolicyStatelessRuleGroupDependencies(rName, 1),
		fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %q
  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]
    stateless_rule_group_reference {
      priority     = %d
      resource_arn = aws_networkfirewall_rule_group.test[0].arn
    }
  }
}
`, rName, priority))
}

func testAccFirewallPolicy_multipleStatelessRuleGroupReferences(rName string) string {
	return acctest.ConfigCompose(
		testAccFirewallPolicyStatelessRuleGroupDependencies(rName, 2),
		fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %q
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

func testAccFirewallPolicy_singleStatelessRuleGroupReference(rName string) string {
	return acctest.ConfigCompose(
		testAccFirewallPolicyStatelessRuleGroupDependencies(rName, 2),
		fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %q
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

func testAccFirewallPolicy_statelessCustomAction(rName string) string {
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

func testAccFirewallPolicy_updateStatelessCustomAction(rName string) string {
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

func testAccFirewallPolicy_multipleStatelessCustomActions(rName string) string {
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

func testAccFirewallPolicy_statefulRuleGroupReferenceAndStatelessCustomAction(rName string) string {
	return acctest.ConfigCompose(
		testAccFirewallPolicyStatefulRuleGroupDependencies(rName, 1),
		fmt.Sprintf(`
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
