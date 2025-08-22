// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package auditmanager

// Exports for use in tests only.
var (
	ResourceAccountRegistration                  = newAccountRegistrationResource
	ResourceOrganizationAdminAccountRegistration = newOrganizationAdminAccountRegistrationResource
	ResourceAssessment                           = newAssessmentResource
	ResourceAssessmentDelegation                 = newAssessmentDelegationResource
	ResourceAssessmentReport                     = newAssessmentReportResource
	ResourceControl                              = newControlResource
	ResourceFramework                            = newFrameworkResource
	ResourceFrameworkShare                       = newFrameworkShareResource

	FindAccountRegistration                = findAccountRegistration
	FindAssessmentByID                     = findAssessmentByID
	FindAssessmentDelegationByThreePartKey = findAssessmentDelegationByThreePartKey
	FindAssessmentReportByID               = findAssessmentReportByID
	FindControlByID                        = findControlByID
	FindFrameworkByID                      = findFrameworkByID
	FindFrameworkShareByID                 = findFrameworkShareByID
	FindOrganizationAdminAccount           = findOrganizationAdminAccount
)
