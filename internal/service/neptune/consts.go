package neptune

import (
	"time"
)

const (
	propagationTimeout                     = 2 * time.Minute
	ErrClusterStillAttachedToGlobalCluster = "neptune Cluster still exists in Neptune Global Cluster"
)
