// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"time"
)

const (
	// Maximum amount of time to wait for an Operation to return Deleted
	iamPropagationTimeout = 2 * time.Minute
)
