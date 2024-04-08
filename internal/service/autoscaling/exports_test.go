// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package autoscaling

// Exports for use in tests only.
var (
	ResourceAttachment = resourceAttachment

	FindAttachmentByLoadBalancerName = findAttachmentByLoadBalancerName
	FindAttachmentByTargetGroupARN   = findAttachmentByTargetGroupARN
	FindGroupByName                  = findGroupByName
)
