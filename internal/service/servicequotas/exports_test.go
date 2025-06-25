// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicequotas

// Exports for use in tests only.
var (
	ResourceServiceQuota        = resourceServiceQuota
	ResourceTemplate            = newTemplateResource
	ResourceTemplateAssociation = newTemplateAssociationResource

	FindTemplateAssociation    = findTemplateAssociation
	FindTemplateByThreePartKey = findTemplateByThreePartKey
)
