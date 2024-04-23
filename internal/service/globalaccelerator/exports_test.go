// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package globalaccelerator

// Exports for use in tests only.
var (
	ResourceAccelerator            = resourceAccelerator
	ResourceCrossAccountAttachment = newCrossAccountAttachmentResource
	ResourceListener               = resourceListener

	FindAcceleratorByARN            = findAcceleratorByARN
	FindCrossAccountAttachmentByARN = findCrossAccountAttachmentByARN
	FindListenerByARN               = findListenerByARN
)
