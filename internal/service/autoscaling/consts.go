package autoscaling

import "time"

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
