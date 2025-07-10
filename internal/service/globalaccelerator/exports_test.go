// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package globalaccelerator

// Exports for use in tests only.
var (
	ResourceAccelerator                = resourceAccelerator
	ResourceCrossAccountAttachment     = newCrossAccountAttachmentResource
	ResourceCustomRoutingAccelerator   = resourceCustomRoutingAccelerator
	ResourceCustomRoutingEndpointGroup = resourceCustomRoutingEndpointGroup
	ResourceCustomRoutingListener      = resourceCustomRoutingListener
	ResourceEndpointGroup              = resourceEndpointGroup
	ResourceListener                   = resourceListener

	FindAcceleratorByARN                = findAcceleratorByARN
	FindCrossAccountAttachmentByARN     = findCrossAccountAttachmentByARN
	FindCustomRoutingAcceleratorByARN   = findCustomRoutingAcceleratorByARN
	FindCustomRoutingEndpointGroupByARN = findCustomRoutingEndpointGroupByARN
	FindCustomRoutingListenerByARN      = findCustomRoutingListenerByARN
	FindEndpointGroupByARN              = findEndpointGroupByARN
	FindListenerByARN                   = findListenerByARN

	ListenerOrEndpointGroupARNToAcceleratorARN = listenerOrEndpointGroupARNToAcceleratorARN
	EndpointGroupARNToListenerARN              = endpointGroupARNToListenerARN
)
