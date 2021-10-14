package ecr

import (
	"time"
)

const (
	// Maximum amount of time to wait for ECR changes to propagate
	propagationTimeout = 2 * time.Minute
)
