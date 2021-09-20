package appmesh

import (
	"time"
)

const (
	// Maximum amount of time to wait for Appmesh changes to propagate
	propagationTimeout = 2 * time.Minute
)
