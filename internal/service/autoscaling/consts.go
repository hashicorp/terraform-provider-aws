package autoscaling

import "time"

const (
	TagResourceTypeGroup = `auto-scaling-group`
)

const (
	iamPropagationTimeout = 2 * time.Minute
)
