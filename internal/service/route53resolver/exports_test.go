// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53resolver

// Exports for use in tests only.
var (
	ResourceDNSSECConfig                 = resourceDNSSECConfig
	ResourceEndpoint                     = resourceEndpoint
	ResourceFirewallConfig               = resourceFirewallConfig
	ResourceFirewallDomainList           = resourceFirewallDomainList
	ResourceFirewallRuleGroupAssociation = resourceFirewallRuleGroupAssociation
	ResourceFirewallRuleGroup            = resourceFirewallRuleGroup
	ResourceFirewallRule                 = resourceFirewallRule
	ResourceQueryLogConfigAssociation    = resourceQueryLogConfigAssociation
	ResourceQueryLogConfig               = resourceQueryLogConfig
	ResourceRuleAssociation              = resourceRuleAssociation
	ResourceRule                         = resourceRule

	FirewallRuleParseResourceID = firewallRuleParseResourceID
	ValidResolverName           = validResolverName

	FindResolverConfigByID                    = findResolverConfigByID
	FindResolverDNSSECConfigByID              = findResolverDNSSECConfigByID
	FindResolverEndpointByID                  = findResolverEndpointByID
	FindFirewallConfigByID                    = findFirewallConfigByID
	FindFirewallDomainListByID                = findFirewallDomainListByID
	FindFirewallRuleGroupAssociationByID      = findFirewallRuleGroupAssociationByID
	FindFirewallRuleGroupByID                 = findFirewallRuleGroupByID
	FindFirewallRuleByTwoPartKey              = findFirewallRuleByTwoPartKey
	FindResolverQueryLogConfigAssociationByID = findResolverQueryLogConfigAssociationByID
	FindResolverQueryLogConfigByID            = findResolverQueryLogConfigByID
	FindResolverRuleAssociationByID           = findResolverRuleAssociationByID
	FindResolverRuleByID                      = findResolverRuleByID
)
