// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package autoscaling

// Exports for use in tests only.
var (
	ResourceAttachment          = resourceAttachment
	ResourceGroup               = resourceGroup
	ResourceGroupTag            = resourceGroupTag
	ResourceLaunchConfiguration = resourceLaunchConfiguration

	FindAttachmentByLoadBalancerName = findAttachmentByLoadBalancerName
	FindAttachmentByTargetGroupARN   = findAttachmentByTargetGroupARN
	FindGroupByName                  = findGroupByName
	FindInstanceRefreshes            = findInstanceRefreshes
	FindLaunchConfigurationByName    = findLaunchConfigurationByName
	FindTag                          = findTag
)
