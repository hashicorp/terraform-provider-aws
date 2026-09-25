// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networksecuritymanager

// Exports for use in tests only.
var (
	ResourceAdminAccount = newAdminAccountResource
	ResourcePolicy       = newPolicyResource
	ResourceRule         = newRuleResource
	ResourceScope        = newScopeResource
	ResourceTemplate     = newTemplateResource

	FindAdminAccountByID = findAdminAccountByID
	FindPolicyByARN      = findPolicyByARN
	FindRuleByARN        = findRuleByARN
	FindScopeByARN       = findScopeByARN
	FindTemplateByARN    = findTemplateByARN
)
