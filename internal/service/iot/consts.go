// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot

import (
	"time"
)

const (
	propagationTimeout          = 2 * time.Minute
	deprecatePropagationTimeout = 6 * time.Minute
)
