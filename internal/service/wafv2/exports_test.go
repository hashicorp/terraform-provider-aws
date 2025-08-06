// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafv2

// Exports for use in tests only.
var (
	ResourceIPSet                      = resourceIPSet
	ResourceRegexPatternSet            = resourceRegexPatternSet
	ResourceRuleGroup                  = resourceRuleGroup
	ResourceWebACL                     = resourceWebACL
	ResourceWebACLAssociation          = resourceWebACLAssociation
	ResourceWebACLLoggingConfiguration = resourceWebACLLoggingConfiguration
	ResourceAPIKey                     = newAPIKeyResource
	ResourceWebACLRuleGroupAssociation = newResourceWebACLRuleGroupAssociation
	ResourceWebACLRule                 = newResourceWebACLRule

	FindAPIKeyByTwoPartKey            = findAPIKeyByTwoPartKey
	FindIPSetByThreePartKey           = findIPSetByThreePartKey
	FindLoggingConfigurationByARN     = findLoggingConfigurationByARN
	FindRegexPatternSetByThreePartKey = findRegexPatternSetByThreePartKey
	FindRuleGroupByThreePartKey       = findRuleGroupByThreePartKey
	FindWebACLByResourceARN           = findWebACLByResourceARN
	FindWebACLByThreePartKey          = findWebACLByThreePartKey
	ListRuleGroupsPages               = listRuleGroupsPages
	ListWebACLsPages                  = listWebACLsPages
	ParseWebACLARN                    = parseWebACLARN
	ParseWebACLRuleID                 = parseWebACLRuleID
)
