// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkfirewall_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/networkfirewall"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfnetworkfirewall "github.com/hashicorp/terraform-provider-aws/internal/service/networkfirewall"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccNetworkFirewallProxyConfigurationRuleGroupAttachmentsExclusive_basic(t *testing.T) {
	t.Helper()

	ctx := acctest.Context(t)
	var v networkfirewall.DescribeProxyConfigurationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_proxy_configuration_rule_group_attachments_exclusive.test"
	proxyConfigResourceName := "aws_networkfirewall_proxy_configuration.test"
	ruleGroup1ResourceName := "aws_networkfirewall_proxy_rule_group.test1"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewall),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyConfigurationRuleGroupAttachmentsExclusiveDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyConfigurationRuleGroupAttachmentsExclusiveConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProxyConfigurationRuleGroupAttachmentsExclusiveExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, proxyConfigResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "proxy_configuration_arn", proxyConfigResourceName, names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "rule_group.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "rule_group.0.proxy_rule_group_name", ruleGroup1ResourceName, names.AttrName),
					resource.TestCheckResourceAttrSet(resourceName, "update_token"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"update_token"},
			},
		},
	})
}

func testAccNetworkFirewallProxyConfigurationRuleGroupAttachmentsExclusive_disappears(t *testing.T) {
	t.Helper()

	ctx := acctest.Context(t)
	var v networkfirewall.DescribeProxyConfigurationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_proxy_configuration_rule_group_attachments_exclusive.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewall),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyConfigurationRuleGroupAttachmentsExclusiveDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyConfigurationRuleGroupAttachmentsExclusiveConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProxyConfigurationRuleGroupAttachmentsExclusiveExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappearsWithStateFunc(ctx, t, tfnetworkfirewall.ResourceProxyConfigurationRuleGroupAttachmentsExclusive, resourceName, proxyConfigurationRuleGroupAttachmentsExclusiveDisappearsStateFunc),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccNetworkFirewallProxyConfigurationRuleGroupAttachmentsExclusive_updateAdd(t *testing.T) {
	t.Helper()

	ctx := acctest.Context(t)
	var v1, v2 networkfirewall.DescribeProxyConfigurationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_proxy_configuration_rule_group_attachments_exclusive.test"
	ruleGroup3ResourceName := "aws_networkfirewall_proxy_rule_group.test3"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewall),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyConfigurationRuleGroupAttachmentsExclusiveDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyConfigurationRuleGroupAttachmentsExclusiveConfig_twoRuleGroups(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProxyConfigurationRuleGroupAttachmentsExclusiveExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.#", "2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"update_token"},
			},
			{
				Config: testAccProxyConfigurationRuleGroupAttachmentsExclusiveConfig_threeRuleGroups(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProxyConfigurationRuleGroupAttachmentsExclusiveExists(ctx, t, resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "rule_group.#", "3"),
					resource.TestCheckResourceAttrPair(resourceName, "rule_group.2.proxy_rule_group_name", ruleGroup3ResourceName, names.AttrName),
				),
			},
		},
	})
}

func testAccNetworkFirewallProxyConfigurationRuleGroupAttachmentsExclusive_updateRemove(t *testing.T) {
	t.Helper()

	ctx := acctest.Context(t)
	var v1, v2 networkfirewall.DescribeProxyConfigurationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_proxy_configuration_rule_group_attachments_exclusive.test"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewall),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyConfigurationRuleGroupAttachmentsExclusiveDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyConfigurationRuleGroupAttachmentsExclusiveConfig_twoRuleGroups(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProxyConfigurationRuleGroupAttachmentsExclusiveExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.#", "2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"update_token"},
			},
			{
				Config: testAccProxyConfigurationRuleGroupAttachmentsExclusiveConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProxyConfigurationRuleGroupAttachmentsExclusiveExists(ctx, t, resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "rule_group.#", "1"),
				),
			},
		},
	})
}

func testAccNetworkFirewallProxyConfigurationRuleGroupAttachmentsExclusive_updateReorder(t *testing.T) {
	t.Helper()

	ctx := acctest.Context(t)
	var v1, v2 networkfirewall.DescribeProxyConfigurationOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_networkfirewall_proxy_configuration_rule_group_attachments_exclusive.test"
	ruleGroup1ResourceName := "aws_networkfirewall_proxy_rule_group.test1"
	ruleGroup2ResourceName := "aws_networkfirewall_proxy_rule_group.test2"

	acctest.Test(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.NetworkFirewall),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProxyConfigurationRuleGroupAttachmentsExclusiveDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccProxyConfigurationRuleGroupAttachmentsExclusiveConfig_twoRuleGroups(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProxyConfigurationRuleGroupAttachmentsExclusiveExists(ctx, t, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, "rule_group.#", "2"),
					resource.TestCheckResourceAttrPair(resourceName, "rule_group.0.proxy_rule_group_name", ruleGroup1ResourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, "rule_group.1.proxy_rule_group_name", ruleGroup2ResourceName, names.AttrName),
				),
			},
			{
				Config: testAccProxyConfigurationRuleGroupAttachmentsExclusiveConfig_twoRuleGroupsReversed(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckProxyConfigurationRuleGroupAttachmentsExclusiveExists(ctx, t, resourceName, &v2),
					resource.TestCheckResourceAttr(resourceName, "rule_group.#", "2"),
					resource.TestCheckResourceAttrPair(resourceName, "rule_group.0.proxy_rule_group_name", ruleGroup2ResourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, "rule_group.1.proxy_rule_group_name", ruleGroup1ResourceName, names.AttrName),
				),
			},
		},
	})
}

func testAccCheckProxyConfigurationRuleGroupAttachmentsExclusiveDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).NetworkFirewallClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_networkfirewall_proxy_configuration_rule_group_attachments_exclusive" {
				continue
			}

			out, err := tfnetworkfirewall.FindProxyConfigurationByARN(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			// Check if there are any rule groups attached
			if out != nil && out.ProxyConfiguration != nil && out.ProxyConfiguration.RuleGroups != nil {
				if len(out.ProxyConfiguration.RuleGroups) > 0 {
					return fmt.Errorf("NetworkFirewall Proxy Configuration Rule Group Attachment still exists: %s has %d rule groups", rs.Primary.ID, len(out.ProxyConfiguration.RuleGroups))
				}
			}
		}

		return nil
	}
}

func testAccCheckProxyConfigurationRuleGroupAttachmentsExclusiveExists(ctx context.Context, t *testing.T, n string, v *networkfirewall.DescribeProxyConfigurationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).NetworkFirewallClient(ctx)

		output, err := tfnetworkfirewall.FindProxyConfigurationByARN(ctx, conn, rs.Primary.Attributes["proxy_configuration_arn"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func proxyConfigurationRuleGroupAttachmentsExclusiveDisappearsStateFunc(ctx context.Context, state *tfsdk.State, is *terraform.InstanceState) error {
	// Set the id attribute (needed for Delete to find the resource)
	if v, ok := is.Attributes[names.AttrID]; ok {
		if diags := state.SetAttribute(ctx, path.Root(names.AttrID), types.StringValue(v)); diags.HasError() {
			return fmt.Errorf("setting id: %s", diags.Errors()[0].Detail())
		}
	}

	// Set the rule_group nested block from the instance state
	ruleGroupCount := 0
	if v, ok := is.Attributes["rule_group.#"]; ok {
		fmt.Sscanf(v, "%d", &ruleGroupCount)
	}

	if ruleGroupCount > 0 {
		var ruleGroups []tfnetworkfirewall.RuleGroupAttachmentModel
		for i := 0; i < ruleGroupCount; i++ {
			key := fmt.Sprintf("rule_group.%d.proxy_rule_group_name", i)
			if v, ok := is.Attributes[key]; ok {
				ruleGroups = append(ruleGroups, tfnetworkfirewall.RuleGroupAttachmentModel{
					ProxyRuleGroupName: types.StringValue(v),
				})
			}
		}

		ruleGroupsList := fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, ruleGroups)
		if diags := state.SetAttribute(ctx, path.Root("rule_group"), ruleGroupsList); diags.HasError() {
			return fmt.Errorf("setting rule_group: %s", diags.Errors()[0].Detail())
		}
	}

	return nil
}

func testAccProxyConfigurationRuleGroupAttachmentsExclusiveConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_networkfirewall_proxy_configuration" "test" {
  name = %[1]q

  default_rule_phase_actions {
    post_response = "ALLOW"
    pre_dns       = "ALLOW"
    pre_request   = "ALLOW"
  }
}

resource "aws_networkfirewall_proxy_rule_group" "test1" {
  name = "%[1]s-1"
}

resource "aws_networkfirewall_proxy_rule_group" "test2" {
  name = "%[1]s-2"
}

resource "aws_networkfirewall_proxy_rule_group" "test3" {
  name = "%[1]s-3"
}
`, rName)
}

func testAccProxyConfigurationRuleGroupAttachmentsExclusiveConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccProxyConfigurationRuleGroupAttachmentsExclusiveConfig_base(rName),
		`
resource "aws_networkfirewall_proxy_configuration_rule_group_attachments_exclusive" "test" {
  proxy_configuration_arn = aws_networkfirewall_proxy_configuration.test.arn

  rule_group {
    proxy_rule_group_name = aws_networkfirewall_proxy_rule_group.test1.name
  }
}
`)
}

func testAccProxyConfigurationRuleGroupAttachmentsExclusiveConfig_twoRuleGroups(rName string) string {
	return acctest.ConfigCompose(
		testAccProxyConfigurationRuleGroupAttachmentsExclusiveConfig_base(rName),
		`
resource "aws_networkfirewall_proxy_configuration_rule_group_attachments_exclusive" "test" {
  proxy_configuration_arn = aws_networkfirewall_proxy_configuration.test.arn

  rule_group {
    proxy_rule_group_name = aws_networkfirewall_proxy_rule_group.test1.name
  }

  rule_group {
    proxy_rule_group_name = aws_networkfirewall_proxy_rule_group.test2.name
  }
}
`)
}

func testAccProxyConfigurationRuleGroupAttachmentsExclusiveConfig_twoRuleGroupsReversed(rName string) string {
	return acctest.ConfigCompose(
		testAccProxyConfigurationRuleGroupAttachmentsExclusiveConfig_base(rName),
		`
resource "aws_networkfirewall_proxy_configuration_rule_group_attachments_exclusive" "test" {
  proxy_configuration_arn = aws_networkfirewall_proxy_configuration.test.arn

  rule_group {
    proxy_rule_group_name = aws_networkfirewall_proxy_rule_group.test2.name
  }

  rule_group {
    proxy_rule_group_name = aws_networkfirewall_proxy_rule_group.test1.name
  }
}
`)
}

func testAccProxyConfigurationRuleGroupAttachmentsExclusiveConfig_threeRuleGroups(rName string) string {
	return acctest.ConfigCompose(
		testAccProxyConfigurationRuleGroupAttachmentsExclusiveConfig_base(rName),
		`
resource "aws_networkfirewall_proxy_configuration_rule_group_attachments_exclusive" "test" {
  proxy_configuration_arn = aws_networkfirewall_proxy_configuration.test.arn

  rule_group {
    proxy_rule_group_name = aws_networkfirewall_proxy_rule_group.test1.name
  }

  rule_group {
    proxy_rule_group_name = aws_networkfirewall_proxy_rule_group.test2.name
  }

  rule_group {
    proxy_rule_group_name = aws_networkfirewall_proxy_rule_group.test3.name
  }
}
`)
}
