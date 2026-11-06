// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networksecuritymanager

// Exports for use in tests only.
var (
	ResourceDeployment = newDeploymentResource
	ResourcePolicy     = newPolicyResource
	ResourceRule       = newRuleResource
	ResourceScope      = newScopeResource
	ResourceTemplate   = newTemplateResource

	FindDeploymentByARN = findDeploymentByARN
	FindPolicyByARN     = findPolicyByARN
	FindRuleByARN       = findRuleByARN
	FindScopeByARN      = findScopeByARN
	FindTemplateByARN   = findTemplateByARN
)
