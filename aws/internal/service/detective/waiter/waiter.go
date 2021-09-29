package waiter

import (
	"time"
)

const (
	// DetectiveOperationTimeout Maximum amount of time to wait for a detective graph to be created, deleted
	DetectiveOperationTimeout = 4 * time.Minute
)
