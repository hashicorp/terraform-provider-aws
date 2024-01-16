// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appmesh

import (
	"time"
)

const (
	// Maximum amount of time to wait for Appmesh changes to propagate
	propagationTimeout = 2 * time.Minute
)
