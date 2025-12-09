// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package networkfirewall

// Exports for use in tests only.
var (
	ResourceFirewall                                 = resourceFirewall
	ResourceFirewallPolicy                           = resourceFirewallPolicy
	ResourceFirewallTransitGatewayAttachmentAccepter = newFirewallTransitGatewayAttachmentAccepterResource
	ResourceLoggingConfiguration                     = resourceLoggingConfiguration
	ResourceResourcePolicy                           = resourceResourcePolicy
	ResourceRuleGroup                                = resourceRuleGroup
	ResourceTLSInspectionConfiguration               = newTLSInspectionConfigurationResource
	ResourceVPCEndpointAssociation                   = newVPCEndpointAssociationResource

	FindFirewallByARN                   = findFirewallByARN
	FindFirewallPolicyByARN             = findFirewallPolicyByARN
	FindLoggingConfigurationByARN       = findLoggingConfigurationByARN
	FindResourcePolicyByARN             = findResourcePolicyByARN
	FindRuleGroupByARN                  = findRuleGroupByARN
	FindTLSInspectionConfigurationByARN = findTLSInspectionConfigurationByARN
	FindVPCEndpointAssociationByARN     = findVPCEndpointAssociationByARN
)
