// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package devicefarm

// Exports for use in tests only.
var (
	ResourceDevicePool      = resourceDevicePool
	ResourceInstanceProfile = resourceInstanceProfile
	ResourceNetworkProfile  = resourceNetworkProfile
	ResourceProject         = resourceProject
	ResourceTestGridProject = resourceTestGridProject
	ResourceUpload          = resourceUpload

	FindDevicePoolByARN      = findDevicePoolByARN
	FindInstanceProfileByARN = findInstanceProfileByARN
	FindNetworkProfileByARN  = findNetworkProfileByARN
	FindProjectByARN         = findProjectByARN
	FindTestGridProjectByARN = findTestGridProjectByARN
	FindUploadByARN          = findUploadByARN
)
