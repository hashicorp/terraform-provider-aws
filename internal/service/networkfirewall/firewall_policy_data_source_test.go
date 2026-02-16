// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkfirewall_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNetworkFirewallFirewallPolicyDataSource_arn(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, "resource-test-terraform")
	resourceName := "aws_networkfirewall_firewall_policy.test"
	datasourceName := "data.aws_networkfirewall_firewall_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyDataSourceConfig_arn(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.#", resourceName, "firewall_policy.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.0.stateless_fragment_default_actions.#", resourceName, "firewall_policy.0.stateless_fragment_default_actions.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.0.stateless_fragment_default_actions.0", resourceName, "firewall_policy.0.stateless_fragment_default_actions.0"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.0.stateless_default_actions.#", resourceName, "firewall_policy.0.stateless_default_actions.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.0.stateless_default_actions.0", resourceName, "firewall_policy.0.stateless_default_actions.0"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.0.tls_inspection_configuration_arn", resourceName, "firewall_policy.0.tls_inspection_configuration_arn"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func TestAccNetworkFirewallFirewallPolicyDataSource_name(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, "resource-test-terraform")
	resourceName := "aws_networkfirewall_firewall_policy.test"
	datasourceName := "data.aws_networkfirewall_firewall_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyDataSourceConfig_name(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.#", resourceName, "firewall_policy.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.0.stateless_fragment_default_actions.#", resourceName, "firewall_policy.0.stateless_fragment_default_actions.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.0.stateless_fragment_default_actions.0", resourceName, "firewall_policy.0.stateless_fragment_default_actions.0"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.0.stateless_default_actions.#", resourceName, "firewall_policy.0.stateless_default_actions.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.0.stateless_default_actions.0", resourceName, "firewall_policy.0.stateless_default_actions.0"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.0.tls_inspection_configuration_arn", resourceName, "firewall_policy.0.tls_inspection_configuration_arn"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func TestAccNetworkFirewallFirewallPolicyDataSource_nameAndARN(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, "resource-test-terraform")
	resourceName := "aws_networkfirewall_firewall_policy.test"
	datasourceName := "data.aws_networkfirewall_firewall_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyDataSourceConfig_nameAndARN(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.#", resourceName, "firewall_policy.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.0.stateless_fragment_default_actions.#", resourceName, "firewall_policy.0.stateless_fragment_default_actions.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.0.stateless_fragment_default_actions.0", resourceName, "firewall_policy.0.stateless_fragment_default_actions.0"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.0.stateless_default_actions.#", resourceName, "firewall_policy.0.stateless_default_actions.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.0.stateless_default_actions.0", resourceName, "firewall_policy.0.stateless_default_actions.0"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.0.tls_inspection_configuration_arn", resourceName, "firewall_policy.0.tls_inspection_configuration_arn"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func TestAccNetworkFirewallFirewallPolicyDataSource_withOverriddenManagedRuleGroup(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, "resource-test-terraform")
	resourceName := "aws_networkfirewall_firewall_policy.test"
	datasourceName := "data.aws_networkfirewall_firewall_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyDataSourceConfig_withOverriddenManagedRuleGroup(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN), resource.TestCheckResourceAttrPair(datasourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.#", resourceName, "firewall_policy.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.0.stateless_fragment_default_actions.#", resourceName, "firewall_policy.0.stateless_fragment_default_actions.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.0.stateless_fragment_default_actions.0", resourceName, "firewall_policy.0.stateless_fragment_default_actions.0"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.0.stateless_default_actions.#", resourceName, "firewall_policy.0.stateless_default_actions.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.0.stateless_default_actions.0", resourceName, "firewall_policy.0.stateless_default_actions.0"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.0.stateful_rule_group_reference.#", resourceName, "firewall_policy.0.stateful_rule_group_reference.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.0.stateful_rule_group_reference.0", resourceName, "firewall_policy.0.stateful_rule_group_reference.0"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.0.stateful_rule_group_reference.0.deep_threat_inspection", resourceName, "firewall_policy.0.stateful_rule_group_reference.0.deep_threat_inspection"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.0.stateful_rule_group_reference.override.action", resourceName, "firewall_policy.0.stateful_rule_group_reference.override.action"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func TestAccNetworkFirewallFirewallPolicyDataSource_withPolicyVariables(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, "resource-test-terraform")
	resourceName := "aws_networkfirewall_firewall_policy.test"
	datasourceName := "data.aws_networkfirewall_firewall_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyDataSourceConfig_withPolicyVariables(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.#", resourceName, "firewall_policy.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.rule_variables.#", resourceName, "firewall_policy.rule_variables.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.rule_variables.ip_set.#", resourceName, "firewall_policy.rule_variables.ip_set.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.rule_variables.ip_set.0.definition.#", resourceName, "firewall_policy.rule_variables.ip_set.0.definition.#"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func TestAccNetworkFirewallFirewallPolicyDataSource_activeThreatDefense(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, "resource-test-terraform")
	resourceName := "aws_networkfirewall_firewall_policy.test"
	datasourceName := "data.aws_networkfirewall_firewall_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyDataSourceConfig_activeThreatDefense(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN), resource.TestCheckResourceAttrPair(datasourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.#", resourceName, "firewall_policy.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.0.stateless_fragment_default_actions.#", resourceName, "firewall_policy.0.stateless_fragment_default_actions.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.0.stateless_fragment_default_actions.0", resourceName, "firewall_policy.0.stateless_fragment_default_actions.0"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.0.stateful_rule_group_reference.#", resourceName, "firewall_policy.0.stateful_rule_group_reference.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.0.stateful_rule_group_reference.0", resourceName, "firewall_policy.0.stateful_rule_group_reference.0"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.0.stateful_rule_group_reference.0.deep_threat_inspection", resourceName, "firewall_policy.0.stateful_rule_group_reference.0.deep_threat_inspection"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func TestAccNetworkFirewallFirewallPolicyDataSource_statefulEngineOptions(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, "resource-test-terraform")
	resourceName := "aws_networkfirewall_firewall_policy.test"
	datasourceName := "data.aws_networkfirewall_firewall_policy.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyDataSourceConfig_statefulEngineOptions(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN), resource.TestCheckResourceAttrPair(datasourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.#", resourceName, "firewall_policy.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.0.stateless_fragment_default_actions.#", resourceName, "firewall_policy.0.stateless_fragment_default_actions.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.0.stateless_fragment_default_actions.0", resourceName, "firewall_policy.0.stateless_fragment_default_actions.0"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.0.stateful_engine_options.#", resourceName, "firewall_policy.0.stateful_engine_options.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.0.stateful_engine_options.0.flow_timeouts.#", resourceName, "firewall_policy.0.stateful_engine_options.0.flow_timeouts.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.0.stateful_engine_options.0.flow_timeouts.0.tcp_idle_timeout_seconds", resourceName, "firewall_policy.0.stateful_engine_options.0.flow_timeouts.0.tcp_idle_timeout_seconds"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.0.stateful_engine_options.0.rule_order", resourceName, "firewall_policy.0.stateful_engine_options.0.rule_order"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.0.stateful_engine_options.0.stream_exception_policy", resourceName, "firewall_policy.0.stateful_engine_options.0.stream_exception_policy"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func TestAccNetworkFirewallFirewallPolicyDataSource_multipleStatefulRuleGroupReferences(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, "resource-test-terraform")
	resourceName := "aws_networkfirewall_firewall_policy.test"
	datasourceName := "data.aws_networkfirewall_firewall_policy.test"
	ruleGroupResourceName1 := "aws_networkfirewall_rule_group.test.0"
	ruleGroupResourceName2 := "aws_networkfirewall_rule_group.test.1"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewallServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFirewallPolicyDataSourceConfig_multipleStatefulRuleGroupReferences(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN), resource.TestCheckResourceAttrPair(datasourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.#", resourceName, "firewall_policy.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.0.stateful_rule_group_reference.#", resourceName, "firewall_policy.0.stateful_rule_group_reference.#"),
					resource.TestCheckTypeSetElemAttrPair(datasourceName, "firewall_policy.0.stateful_rule_group_reference.*.resource_arn", ruleGroupResourceName1, names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(datasourceName, "firewall_policy.0.stateful_rule_group_reference.*.resource_arn", ruleGroupResourceName2, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func testAccFirewallPolicyDataSourceConfig_basic(rName string) string {
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

func testAccFirewallPolicyDataSourceConfig_arn(rName string) string {
	return acctest.ConfigCompose(
		testAccFirewallPolicyDataSourceConfig_basic(rName),
		`
data "aws_networkfirewall_firewall_policy" "test" {
  arn = aws_networkfirewall_firewall_policy.test.arn
}`)
}

func testAccFirewallPolicyDataSourceConfig_name(rName string) string {
	return acctest.ConfigCompose(
		testAccFirewallPolicyDataSourceConfig_basic(rName),
		`
data "aws_networkfirewall_firewall_policy" "test" {
  name = aws_networkfirewall_firewall_policy.test.name
}`)
}

func testAccFirewallPolicyDataSourceConfig_nameAndARN(rName string) string {
	return acctest.ConfigCompose(
		testAccFirewallPolicyDataSourceConfig_basic(rName),
		`
data "aws_networkfirewall_firewall_policy" "test" {
  arn  = aws_networkfirewall_firewall_policy.test.arn
  name = aws_networkfirewall_firewall_policy.test.name
}`)
}

func testAccFirewallPolicyDataSourceConfig_withOverriddenManagedRuleGroup(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}
data "aws_partition" "current" {}

resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q

  firewall_policy {
    stateless_default_actions          = ["aws:forward_to_sfe"]
    stateless_fragment_default_actions = ["aws:forward_to_sfe"]

    # Managed rule group required for override block.
    stateful_rule_group_reference {
      resource_arn = "arn:${data.aws_partition.current.partition}:network-firewall:${data.aws_region.current.region}:aws-managed:stateful-rulegroup/MalwareDomainsActionOrder"

      override {
        action = "DROP_TO_ALERT"
      }
    }
  }
}

data "aws_networkfirewall_firewall_policy" "test" {
  arn = aws_networkfirewall_firewall_policy.test.arn
}`, rName)
}

func testAccFirewallPolicyDataSourceConfig_withPolicyVariables(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q
  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]
    policy_variables {
      rule_variables {
        key = "HOME_NET"
        ip_set {
          definition = ["10.0.0.0/16", "10.1.0.0/24"]
        }
      }
    }
  }
}

data "aws_networkfirewall_firewall_policy" "test" {
  arn = aws_networkfirewall_firewall_policy.test.arn
}`, rName)
}

func testAccFirewallPolicyDataSourceConfig_activeThreatDefense(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}
data "aws_partition" "current" {}

resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q

  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]

    stateful_rule_group_reference {
      deep_threat_inspection = true
      resource_arn           = "arn:${data.aws_partition.current.partition}:network-firewall:${data.aws_region.current.region}:aws-managed:stateful-rulegroup/AttackInfrastructureActionOrder"
    }
  }
}

data "aws_networkfirewall_firewall_policy" "test" {
  arn = aws_networkfirewall_firewall_policy.test.arn
}`, rName)
}

func testAccFirewallPolicyDataSourceConfig_statefulEngineOptions(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q

  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]

    stateful_engine_options {
      flow_timeouts {
        tcp_idle_timeout_seconds = 60
      }
      rule_order              = "STRICT_ORDER"
      stream_exception_policy = "DROP"
    }
  }
}
data "aws_networkfirewall_firewall_policy" "test" {
  arn = aws_networkfirewall_firewall_policy.test.arn
}
`, rName)
}

func testAccFirewallPolicyDataSourceConfig_multipleStatefulRuleGroupReferences(rName string) string {
	return acctest.ConfigCompose(testAccFirewallPolicyConfig_multipleStatefulRuleGroupReferences(rName), `
data "aws_networkfirewall_firewall_policy" "test" {
  arn = aws_networkfirewall_firewall_policy.test.arn
}
`)
}
