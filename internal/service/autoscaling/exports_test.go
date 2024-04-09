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
	ResourceNotification        = resourceNotification
	ResourcePolicy              = resourcePolicy

	FindAttachmentByLoadBalancerName = findAttachmentByLoadBalancerName
	FindAttachmentByTargetGroupARN   = findAttachmentByTargetGroupARN
	FindGroupByName                  = findGroupByName
	FindInstanceRefreshes            = findInstanceRefreshes
	FindLaunchConfigurationByName    = findLaunchConfigurationByName
	FindLifecycleHookByTwoPartKey    = findLifecycleHookByTwoPartKey
	FindNotificationsByTwoPartKey    = findNotificationsByTwoPartKey
	FindScalingPolicyByTwoPartKey    = findScalingPolicyByTwoPartKey
	FindTag                          = findTag
)
