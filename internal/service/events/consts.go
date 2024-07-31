// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package events

import (
	"time"
)

const (
	DefaultEventBusName = "default"
)

const (
	targetInputTransformerMaxInputPaths = 100
)

const (
	propagationTimeout = 2 * time.Minute
)
