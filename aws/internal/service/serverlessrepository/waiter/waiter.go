package waiter

import (
	"time"
)

const (
	// Default maximum amount of time to wait for a Stack to be Created
	StackCreatedDefaultTimeout = 30 * time.Minute

	// Default maximum amount of time to wait for a Stack to be Updated
	StackUpdatedDefaultTimeout = 30 * time.Minute

	// Default maximum amount of time to wait for a Stack to be Deleted
	StackDeletedDefaultTimeout = 30 * time.Minute
)
