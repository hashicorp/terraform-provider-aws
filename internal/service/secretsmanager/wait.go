package secretsmanager

import (
	"time"
)

const (
	// Maximum amount of time to wait for Secrets Manager changes to propagate
	propagationTimeout = 2 * time.Minute
)
