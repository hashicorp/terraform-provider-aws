// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm

// Exports for use in tests only.
var (
	ResourceActivation           = resourceActivation
	ResourceDefaultPatchBaseline = resourceDefaultPatchBaseline
	ResourcePatchBaseline        = resourcePatchBaseline

	FindActivationByID    = findActivationByID
	FindPatchBaselineByID = findPatchBaselineByID
)
