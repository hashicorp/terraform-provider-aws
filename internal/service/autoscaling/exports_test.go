// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package autoscaling

// Exports for use in tests only.
var (
	ResourceAttachment              = resourceAttachment
	ResourceGroup                   = resourceGroup
	ResourceGroupTag                = resourceGroupTag
	ResourceLaunchConfiguration     = resourceLaunchConfiguration
	ResourceLifecycleHook           = resourceLifecycleHook
	ResourceNotification            = resourceNotification
	ResourcePolicy                  = resourcePolicy
	ResourceSchedule                = resourceSchedule
	ResourceTrafficSourceAttachment = resourceTrafficSourceAttachment

	FindAttachmentByLoadBalancerName          = findAttachmentByLoadBalancerName
	FindAttachmentByTargetGroupARN            = findAttachmentByTargetGroupARN
	FindInstanceRefreshes                     = findInstanceRefreshes
	FindLaunchConfigurationByName             = findLaunchConfigurationByName
	FindLifecycleHookByTwoPartKey             = findLifecycleHookByTwoPartKey
	FindNotificationsByTwoPartKey             = findNotificationsByTwoPartKey
	FindScheduleByTwoPartKey                  = findScheduleByTwoPartKey
	FindTag                                   = findTag
	FindTrafficSourceAttachmentByThreePartKey = findTrafficSourceAttachmentByThreePartKey
)
