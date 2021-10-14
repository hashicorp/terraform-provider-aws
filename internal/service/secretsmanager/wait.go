package secretsmanager

import (
	"time"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	// Maximum amount of time to wait for Secrets Manager changes to propagate
	PropagationTimeout = 2 * time.Minute
)
