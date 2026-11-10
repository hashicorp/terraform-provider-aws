// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambdamicrovms

var (
	ResourceImage   = newImageResource
	ResourceMicroVM = newMicroVMResource

	FindImageByARN  = findImageByARN
	FindMicroVMByID = findMicroVMByID
)
