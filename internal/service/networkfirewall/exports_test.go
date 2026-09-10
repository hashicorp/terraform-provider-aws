// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkfirewall

// Exports for use in tests only.
var (
	ResourceContainerAssociation                     = newContainerAssociationResource
	ResourceFirewall                                 = resourceFirewall
	ResourceFirewallPolicy                           = resourceFirewallPolicy
	ResourceFirewallTransitGatewayAttachmentAccepter = newFirewallTransitGatewayAttachmentAccepterResource
	ResourceLoggingConfiguration                     = resourceLoggingConfiguration
	ResourceResourcePolicy                           = resourceResourcePolicy
	ResourceRuleGroup                                = resourceRuleGroup
	ResourceTLSInspectionConfiguration               = newTLSInspectionConfigurationResource
	ResourceVPCEndpointAssociation                   = newVPCEndpointAssociationResource

	FindContainerAssociationByARN       = findContainerAssociationByARN
	FindFirewallByARN                   = findFirewallByARN
	FindFirewallPolicyByARN             = findFirewallPolicyByARN
	FindLoggingConfigurationByARN       = findLoggingConfigurationByARN
	FindResourcePolicyByARN             = findResourcePolicyByARN
	FindRuleGroupByARN                  = findRuleGroupByARN
	FindTLSInspectionConfigurationByARN = findTLSInspectionConfigurationByARN
	FindVPCEndpointAssociationByARN     = findVPCEndpointAssociationByARN
)
