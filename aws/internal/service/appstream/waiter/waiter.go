package waiter

import (
	"time"
)

const (
	// StackOperationTimeout Maximum amount of time to wait for Stack operation eventual consistency
	StackOperationTimeout = 4 * time.Minute
	// StackSleep Maximum amount of time to sleep for Stack operation after delete
	StackSleep = 15 * time.Second
)
