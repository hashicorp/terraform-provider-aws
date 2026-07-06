// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package autoscaling

import (
	"time"
)

const (
	tagResourceTypeGroup = `auto-scaling-group`
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
	instanceHealthStatusHealthy   = "Healthy"
	instanceHealthStatusUnhealthy = "Unhealthy"
)

const (
	loadBalancerStateAdding    = "Adding"
	loadBalancerStateAdded     = "Added"
	loadBalancerStateInService = "InService"
	loadBalancerStateRemoving  = "Removing"
	loadBalancerStateRemoved   = "Removed"
)

const (
	loadBalancerTargetGroupStateAdding    = "Adding"
	loadBalancerTargetGroupStateAdded     = "Added"
	loadBalancerTargetGroupStateInService = "InService"
	loadBalancerTargetGroupStateRemoving  = "Removing"
	loadBalancerTargetGroupStateRemoved   = "Removed"
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
	trafficSourceStateAdding    = "Adding"
	trafficSourceStateAdded     = "Added"
	trafficSourceStateInService = "InService"
	trafficSourceStateRemoving  = "Removing"
	trafficSourceStateRemoved   = "Removed"
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
