// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkfirewall_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

// Network Firewall Proxy resources are in preview and have strict concurrency limits,
// so all proxy acceptance tests must run serially.
func TestAccNetworkFirewallProxy_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Proxy": {
			acctest.CtBasic:       testAccNetworkFirewallProxy_basic,
			acctest.CtDisappears:  testAccNetworkFirewallProxy_disappears,
			"tlsInterceptEnabled": testAccNetworkFirewallProxy_tlsInterceptEnabled,
			"logging":             testAccNetworkFirewallProxy_logging,
		},
		"ProxyConfiguration": {
			acctest.CtBasic:      testAccNetworkFirewallProxyConfiguration_basic,
			acctest.CtDisappears: testAccNetworkFirewallProxyConfiguration_disappears,
			"tags":               testAccNetworkFirewallProxyConfiguration_tags,
		},
		"ProxyRuleGroup": {
			acctest.CtBasic:      testAccNetworkFirewallProxyRuleGroup_basic,
			acctest.CtDisappears: testAccNetworkFirewallProxyRuleGroup_disappears,
			"tags":               testAccNetworkFirewallProxyRuleGroup_tags,
		},
		"ProxyRulesExclusive": {
			acctest.CtBasic:         testAccNetworkFirewallProxyRulesExclusive_basic,
			acctest.CtDisappears:    testAccNetworkFirewallProxyRulesExclusive_disappears,
			"updateAdd":             testAccNetworkFirewallProxyRulesExclusive_updateAdd,
			"updateModify":          testAccNetworkFirewallProxyRulesExclusive_updateModify,
			"updateRemove":          testAccNetworkFirewallProxyRulesExclusive_updateRemove,
			"multipleRulesPerPhase": testAccNetworkFirewallProxyRulesExclusive_multipleRulesPerPhase,
		},
		"ProxyConfigurationRuleGroupAttachmentsExclusive": {
			acctest.CtBasic:      testAccNetworkFirewallProxyConfigurationRuleGroupAttachmentsExclusive_basic,
			acctest.CtDisappears: testAccNetworkFirewallProxyConfigurationRuleGroupAttachmentsExclusive_disappears,
			"updateAdd":          testAccNetworkFirewallProxyConfigurationRuleGroupAttachmentsExclusive_updateAdd,
			"updateRemove":       testAccNetworkFirewallProxyConfigurationRuleGroupAttachmentsExclusive_updateRemove,
			"updateReorder":      testAccNetworkFirewallProxyConfigurationRuleGroupAttachmentsExclusive_updateReorder,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
