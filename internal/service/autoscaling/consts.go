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
	DefaultWarmPoolMaxGroupPreparedCapacity = -1
)

const (
	InstanceHealthStatusHealthy   = "Healthy"
	InstanceHealthStatusUnhealthy = "Unhealthy"
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
