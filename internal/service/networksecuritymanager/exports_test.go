// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networksecuritymanager

// Exports for use in tests only.
var (
	ResourcePolicy = newPolicyResource
	ResourceRule   = newRuleResource
	ResourceScope  = newScopeResource

	FindPolicyByARN = findPolicyByARN
	FindRuleByARN   = findRuleByARN
	FindScopeByARN  = findScopeByARN
)
