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

	ListRuleGroupsPages = listRuleGroupsPages
	ListWebACLsPages    = listWebACLsPages
)
