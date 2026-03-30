// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package opensearch

import (
	"time"
)

const (
	// OpenSearch sometimes needs a longer IAM propagation time than most services,
	// especially with acceptance tests
	propagationTimeout = 30 * time.Minute
)
