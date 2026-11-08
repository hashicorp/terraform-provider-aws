// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networksecuritymanager

// Exports for use in tests only.
var (
	ResourceAdminAccount = newAdminAccountResource
	ResourceDeployment   = newDeploymentResource
	ResourcePolicy       = newPolicyResource
	ResourceRule         = newRuleResource
	ResourceScope        = newScopeResource
	ResourceTemplate     = newTemplateResource

	FindAdminAccountByID = findAdminAccountByID
	FindDeploymentByARN  = findDeploymentByARN
	FindPolicyByARN      = findPolicyByARN
	FindRuleByARN        = findRuleByARN
	FindScopeByARN       = findScopeByARN
	FindTemplateByARN    = findTemplateByARN

	IsAccountNotOnboardedError = isAccountNotOnboardedError
)
