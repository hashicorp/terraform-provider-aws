// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waf

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
	ResourceXSSMatchSet          = resourceXSSMatchSet

	FindByteMatchSetByID         = findByteMatchSetByID
	FindGeoMatchSetByID          = findGeoMatchSetByID
	FindIPSetByID                = findIPSetByID
	FindRateBasedRuleByID        = findRateBasedRuleByID
	FindRegexMatchSetByID        = findRegexMatchSetByID
	FindRegexPatternSetByID      = findRegexPatternSetByID
	FindRuleByID                 = findRuleByID
	FindRuleGroupByID            = findRuleGroupByID
	FindSizeConstraintSetByID    = findSizeConstraintSetByID
	FindSQLInjectionMatchSetByID = findSQLInjectionMatchSetByID
	FindWebACLByID               = findWebACLByID
	FindXSSMatchSetByID          = findXSSMatchSetByID
	FlattenFieldToMatch          = flattenFieldToMatch
	RegexMatchSetTupleHash       = regexMatchSetTupleHash
)
