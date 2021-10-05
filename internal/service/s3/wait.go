package s3

import (
	"time"
)

const (
	bucketCreatedTimeout = 2 * time.Minute
	propagationTimeout   = 1 * time.Minute
)
