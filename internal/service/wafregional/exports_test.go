// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafregional

// Exports for use in tests only.
var (
	ResourceByteMatchSet         = resourceByteMatchSet
	ResourceGeoMatchSet          = resourceGeoMatchSet
	ResourceIPSet                = resourceIPSet
	ResourceRateBasedRule        = resourceRateBasedRule
	ResourceRegexMatchSet        = resourceRegexMatchSet
	ResourceRegexPatternSet      = resourceRegexPatternSet
	ResourceRule                 = resourceRule
	ResourceRuleGroup            = resourceRuleGroup
	ResourceSizeConstraintSet    = resourceSizeConstraintSet
	ResourceSQLInjectionMatchSet = resourceSQLInjectionMatchSet
	ResourceWebACL               = resourceWebACL
	ResourceWebACLAssociation    = resourceWebACLAssociation
	ResourceXSSMatchSet          = resourceXSSMatchSet

	FindByteMatchSetByID    = findByteMatchSetByID
	FindWebACLByResourceARN = findWebACLByResourceARN
)
