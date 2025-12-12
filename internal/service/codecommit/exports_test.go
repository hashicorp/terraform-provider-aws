// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package codecommit

// Exports for use in tests only.
var (
	ResourceApprovalRuleTemplate            = resourceApprovalRuleTemplate
	ResourceApprovalRuleTemplateAssociation = resourceApprovalRuleTemplateAssociation
	ResourceRepository                      = resourceRepository
	ResourceTrigger                         = resourceTrigger

	FindApprovalRuleTemplateAssociationByTwoPartKey = findApprovalRuleTemplateAssociationByTwoPartKey
	FindApprovalRuleTemplateByName                  = findApprovalRuleTemplateByName
	FindRepositoryByName                            = findRepositoryByName
	FindRepositoryTriggersByName                    = findRepositoryTriggersByName
)
