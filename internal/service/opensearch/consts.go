// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearch

import (
	"time"
)

const (
	// OpenSearch sometimes needs a longer IAM propagation time than most services,
	// especially with acceptance tests
	propagationTimeout = 10 * time.Minute
)
