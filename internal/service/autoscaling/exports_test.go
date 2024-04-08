// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package autoscaling

// Exports for use in tests only.
var (
	ResourceAttachment = resourceAttachment
	ResourceGroupTag   = resourceGroupTag

	FindAttachmentByLoadBalancerName = findAttachmentByLoadBalancerName
	FindAttachmentByTargetGroupARN   = findAttachmentByTargetGroupARN
	FindGroupByName                  = findGroupByName
	FindTag                          = findTag
)
