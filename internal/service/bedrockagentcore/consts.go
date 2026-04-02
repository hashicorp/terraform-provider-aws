// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

import (
	"time"
)

const (
	propagationTimeout = 2 * time.Minute

	samplingPercentageMin = 0.01
	samplingPercentageMax = 100.0
)
