// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"time"
)

type Deadline time.Time

func NewDeadline(duration time.Duration) Deadline {
	return Deadline(time.Now().Add(duration))
}

func (d Deadline) Remaining() time.Duration {
	if v := time.Until(time.Time(d)); v < 0 {
		return 0
	} else {
		return v
	}
}
