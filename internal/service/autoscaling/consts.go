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
	defaultEnabledMetricsGranularity = "1Minute"
)

const (
	defaultTerminationPolicy = "Default"
)

const (
	defaultWarmPoolMaxGroupPreparedCapacity = -1
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

type desiredCapacityType string

const (
	desiredCapacityTypeMemoryMiB desiredCapacityType = "memory-mib"
	desiredCapacityTypeUnits     desiredCapacityType = "units"
	desiredCapacityTypeVCPU      desiredCapacityType = "vcpu"
)

func (desiredCapacityType) Values() []desiredCapacityType {
	return []desiredCapacityType{
		desiredCapacityTypeMemoryMiB,
		desiredCapacityTypeUnits,
		desiredCapacityTypeVCPU,
	}
}

type policyType string

const (
	policyTypePredictiveScaling     policyType = "PredictiveScaling"
	policyTypeSimpleScaling         policyType = "SimpleScaling"
	policyTypeStepScaling           policyType = "StepScaling"
	policyTypeTargetTrackingScaling policyType = "TargetTrackingScaling"
)

func (policyType) Values() []policyType {
	return []policyType{
		policyTypePredictiveScaling,
		policyTypeSimpleScaling,
		policyTypeStepScaling,
		policyTypeTargetTrackingScaling,
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

type lifecycleHookDefaultResult string

const (
	lifecycleHookDefaultResultAbandon  lifecycleHookDefaultResult = "ABANDON"
	lifecycleHookDefaultResultContinue lifecycleHookDefaultResult = "CONTINUE"
)

func (lifecycleHookDefaultResult) Values() []lifecycleHookDefaultResult {
	return []lifecycleHookDefaultResult{
		lifecycleHookDefaultResultAbandon,
		lifecycleHookDefaultResultContinue,
	}
}

type lifecycleHookLifecycleTransition string

const (
	lifecycleHookLifecycleTransitionInstanceLaunching   lifecycleHookLifecycleTransition = "autoscaling:EC2_INSTANCE_LAUNCHING"
	lifecycleHookLifecycleTransitionInstanceTerminating lifecycleHookLifecycleTransition = "autoscaling:EC2_INSTANCE_TERMINATING"
)

func (lifecycleHookLifecycleTransition) Values() []lifecycleHookLifecycleTransition {
	return []lifecycleHookLifecycleTransition{
		lifecycleHookLifecycleTransitionInstanceLaunching,
		lifecycleHookLifecycleTransitionInstanceTerminating,
	}
}

const (
	elbInstanceStateInService = "InService"
)
