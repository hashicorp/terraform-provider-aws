// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package autoscaling

import (
	"time"
)

const (
	TagResourceTypeGroup = `auto-scaling-group`
)

const (
	propagationTimeout = 2 * time.Minute
)

const (
	DefaultEnabledMetricsGranularity = "1Minute"
)

const (
	DefaultTerminationPolicy = "Default"
)

const (
	DefaultWarmPoolMaxGroupPreparedCapacity = -1
)

const (
	InstanceHealthStatusHealthy   = "Healthy"
	InstanceHealthStatusUnhealthy = "Unhealthy"
)

const (
	LoadBalancerStateAdding    = "Adding"
	LoadBalancerStateAdded     = "Added"
	LoadBalancerStateInService = "InService"
	LoadBalancerStateRemoving  = "Removing"
	LoadBalancerStateRemoved   = "Removed"
)

const (
	LoadBalancerTargetGroupStateAdding    = "Adding"
	LoadBalancerTargetGroupStateAdded     = "Added"
	LoadBalancerTargetGroupStateInService = "InService"
	LoadBalancerTargetGroupStateRemoving  = "Removing"
	LoadBalancerTargetGroupStateRemoved   = "Removed"
)

const (
	DesiredCapacityTypeMemoryMiB = "memory-mib"
	DesiredCapacityTypeUnits     = "units"
	DesiredCapacityTypeVCPU      = "vcpu"
)

func DesiredCapacityType_Values() []string {
	return []string{
		DesiredCapacityTypeMemoryMiB,
		DesiredCapacityTypeUnits,
		DesiredCapacityTypeVCPU,
	}
}

const (
	PolicyTypePredictiveScaling     = "PredictiveScaling"
	PolicyTypeSimpleScaling         = "SimpleScaling"
	PolicyTypeStepScaling           = "StepScaling"
	PolicyTypeTargetTrackingScaling = "TargetTrackingScaling"
)

func PolicyType_Values() []string {
	return []string{
		PolicyTypePredictiveScaling,
		PolicyTypeSimpleScaling,
		PolicyTypeStepScaling,
		PolicyTypeTargetTrackingScaling,
	}
}

const (
	TrafficSourceStateAdding    = "Adding"
	TrafficSourceStateAdded     = "Added"
	TrafficSourceStateInService = "InService"
	TrafficSourceStateRemoving  = "Removing"
	TrafficSourceStateRemoved   = "Removed"
)

const (
	launchTemplateIDUnknown = "unknown"
)

const (
	lifecycleHookDefaultResultAbandon  = "ABANDON"
	lifecycleHookDefaultResultContinue = "CONTINUE"
)

func lifecycleHookDefaultResult_Values() []string {
	return []string{
		lifecycleHookDefaultResultAbandon,
		lifecycleHookDefaultResultContinue,
	}
}

const (
	lifecycleHookLifecycleTransitionInstanceLaunching   = "autoscaling:EC2_INSTANCE_LAUNCHING"
	lifecycleHookLifecycleTransitionInstanceTerminating = "autoscaling:EC2_INSTANCE_TERMINATING"
)

func lifecycleHookLifecycleTransition_Values() []string {
	return []string{
		lifecycleHookLifecycleTransitionInstanceLaunching,
		lifecycleHookLifecycleTransitionInstanceTerminating,
	}
}
