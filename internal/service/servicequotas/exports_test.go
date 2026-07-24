// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package servicequotas

// Exports for use in tests only.
var (
	ResourceAutoManagement      = newAutoManagementResource
	ResourceServiceQuota        = resourceServiceQuota
	ResourceTemplate            = newTemplateResource
	ResourceTemplateAssociation = newTemplateAssociationResource

	FindAutoManagement         = findAutoManagement
	FindTemplateAssociation    = findTemplateAssociation
	FindTemplateByThreePartKey = findTemplateByThreePartKey
)
