// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import "time"

const (
	propagationTimeout = 2 * time.Minute
	
	// Global Table consistency modes
	consistencyModeEventuallyConsistent = "EVENTUALLY_CONSISTENT"
	consistencyModeStronglyConsistent   = "STRONGLY_CONSISTENT"
)
