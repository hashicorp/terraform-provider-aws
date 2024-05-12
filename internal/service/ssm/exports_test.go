// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm

// Exports for use in tests only.
var (
	ResourceActivation           = resourceActivation
	ResourceAssociation          = resourceAssociation
	ResourceDefaultPatchBaseline = resourceDefaultPatchBaseline
	ResourceDocument             = resourceDocument
	ResourcePatchBaseline        = resourcePatchBaseline

	FindActivationByID    = findActivationByID
	FindAssociationByID   = findAssociationByID
	FindDocumentByName    = findDocumentByName
	FindPatchBaselineByID = findPatchBaselineByID
)
