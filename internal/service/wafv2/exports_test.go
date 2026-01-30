// Copyright IBM Corp. 2014, 2026
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

	CloudFrontDistributionIDFromARN   = cloudFrontDistributionIDFromARN
	FindAPIKeyByTwoPartKey            = findAPIKeyByTwoPartKey
	FindIPSetByThreePartKey           = findIPSetByThreePartKey
	FindLoggingConfigurationByARN     = findLoggingConfigurationByARN
	FindRegexPatternSetByThreePartKey = findRegexPatternSetByThreePartKey
	FindRuleGroupByThreePartKey       = findRuleGroupByThreePartKey
	FindWebACLByResourceARN           = findWebACLByResourceARN
	FindWebACLByThreePartKey          = findWebACLByThreePartKey
	IsCloudFrontDistributionARN       = isCloudFrontDistributionARN
	ListRuleGroupsPages               = listRuleGroupsPages
	ListWebACLsPages                  = listWebACLsPages
	ParseWebACLARN                    = parseWebACLARN
)
