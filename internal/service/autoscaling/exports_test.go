// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package autoscaling

// Exports for use in tests only.
var (
	ResourceAttachment          = resourceAttachment
	ResourceGroup               = resourceGroup
	ResourceGroupTag            = resourceGroupTag
	ResourceLaunchConfiguration = resourceLaunchConfiguration
	ResourceLifecycleHook       = resourceLifecycleHook

	FindAttachmentByLoadBalancerName = findAttachmentByLoadBalancerName
	FindAttachmentByTargetGroupARN   = findAttachmentByTargetGroupARN
	FindGroupByName                  = findGroupByName
	FindInstanceRefreshes            = findInstanceRefreshes
	FindLaunchConfigurationByName    = findLaunchConfigurationByName
	FindLifecycleHookByTwoPartKey    = findLifecycleHookByTwoPartKey
	FindTag                          = findTag
)
