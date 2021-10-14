package serverlessapprepo

import (
	"time"
)

const (
	// Default maximum amount of time to wait for a Stack to be Created
	cloudFormationStackCreatedDefaultTimeout = 30 * time.Minute

	// Default maximum amount of time to wait for a Stack to be Updated
	cloudFormationStackUpdatedDefaultTimeout = 30 * time.Minute

	// Default maximum amount of time to wait for a Stack to be Deleted
	cloudFormationStackDeletedDefaultTimeout = 30 * time.Minute
)
