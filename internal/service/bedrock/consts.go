// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"time"
)

const (
	propagationTimeout = 2 * time.Minute

	// evaluationTaskTypeGeneral is not yet in the AWS SDK but is required by the
	// API for model-as-judge evaluations.
	evaluationTaskTypeGeneral = "General"
)
