// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkfirewall

// Exports for use in tests only.
var (
	ResourceFirewallPolicy             = resourceFirewallPolicy
	ResourceLoggingConfiguration       = resourceLoggingConfiguration
	ResourceResourcePolicy             = resourceResourcePolicy
	ResourceTLSInspectionConfiguration = newTLSInspectionConfigurationResource

	FindFirewallPolicyByARN             = findFirewallPolicyByARN
	FindLoggingConfigurationByARN       = findLoggingConfigurationByARN
	FindResourcePolicyByARN             = findResourcePolicyByARN
	FindTLSInspectionConfigurationByARN = findTLSInspectionConfigurationByARN
)
