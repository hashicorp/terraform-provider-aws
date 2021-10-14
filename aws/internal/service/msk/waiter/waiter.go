package waiter

import (
	"time"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	ClusterCreateTimeout = 120 * time.Minute
	ClusterUpdateTimeout = 120 * time.Minute
	ClusterDeleteTimeout = 120 * time.Minute
)
