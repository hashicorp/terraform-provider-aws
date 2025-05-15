// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backoff

import (
	"time"
)

type deadline time.Time

func NewDeadline(duration time.Duration) deadline {
	return deadline(time.Now().Add(duration))
}

func (d deadline) Remaining() time.Duration {
	if v := time.Until(time.Time(d)); v < 0 {
		return 0
	} else {
		return v
	}
}
