// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package signer

// Exports for use in tests only.
var (
	ResourceSigningJob               = resourceSigningJob
	ResourceSigningProfile           = resourceSigningProfile
	ResourceSigningProfilePermission = resourceSigningProfilePermission

	FindPermissionByTwoPartKey = findPermissionByTwoPartKey
	FindSigningJobByID         = findSigningJobByID
	FindSigningProfileByName   = findSigningProfileByName
)
