// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package auditmanager

// Exports for use in tests only.
var (
	ResourceAccountRegistration                  = newResourceAccountRegistration
	ResourceOrganizationAdminAccountRegistration = newResourceOrganizationAdminAccountRegistration
	ResourceAssessment                           = newResourceAssessment
	ResourceAssessmentDelegation                 = newResourceAssessmentDelegation
	ResourceAssessmentReport                     = newResourceAssessmentReport
	ResourceControl                              = newResourceControl
	ResourceFramework                            = newResourceFramework
	ResourceFrameworkShare                       = newResourceFrameworkShare
)
