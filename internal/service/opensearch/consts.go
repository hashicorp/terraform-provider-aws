package opensearch

import (
	"time"
)

const (
	// OpenSearch sometimes needs a longer IAM propagation time than most servicces,
	// especially with acceptnace tests
	propagationTimeout = 10 * time.Minute
)
