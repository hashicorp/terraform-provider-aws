// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkfirewall_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccNetworkFirewallFirewallPolicyDataSource_arn(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_networkfirewall_firewall_policy.test"
	datasourceName := "data.aws_networkfirewall_firewall_policy.test"

	resource.ParallelTest(t, resource.TestCase{
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
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.0.tls_inspection_coniguration_arn", resourceName, "firewall_policy.0.tls_inspection_coniguration_arn"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func TestAccNetworkFirewallFirewallPolicyDataSource_name(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_networkfirewall_firewall_policy.test"
	datasourceName := "data.aws_networkfirewall_firewall_policy.test"

	resource.ParallelTest(t, resource.TestCase{
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
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.0.tls_inspection_coniguration_arn", resourceName, "firewall_policy.0.tls_inspection_coniguration_arn"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func TestAccNetworkFirewallFirewallPolicyDataSource_nameAndARN(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_networkfirewall_firewall_policy.test"
	datasourceName := "data.aws_networkfirewall_firewall_policy.test"

	resource.ParallelTest(t, resource.TestCase{
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
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.0.tls_inspection_coniguration_arn", resourceName, "firewall_policy.0.tls_inspection_coniguration_arn"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func TestAccNetworkFirewallFirewallPolicyDataSource_withOverriddenManagedRuleGroup(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_networkfirewall_firewall_policy.test"
	datasourceName := "data.aws_networkfirewall_firewall_policy.test"

	resource.ParallelTest(t, resource.TestCase{
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
					resource.TestCheckResourceAttrPair(datasourceName, "firewall_policy.0.stateful_rule_group_reference.override.action", resourceName, "firewall_policy.0.stateful_rule_group_reference.override.action"),
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
      resource_arn = "arn:${data.aws_partition.current.partition}:network-firewall:${data.aws_region.current.name}:aws-managed:stateful-rulegroup/MalwareDomainsActionOrder"

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
