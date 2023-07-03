// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package comprehend

import (
	"time"
)

const iamPropagationTimeout = 2 * time.Minute

// Avoid service throttling
const entityRegcognizerCreatedDelay = 10 * time.Minute
const entityRegcognizerStoppedDelay = 0
const entityRegcognizerDeletedDelay = 5 * time.Minute
const entityRegcognizerPollInterval = 1 * time.Minute

const documentClassifierCreatedDelay = 15 * time.Minute
const documentClassifierStoppedDelay = 0
const documentClassifierDeletedDelay = 5 * time.Minute
const documentClassifierPollInterval = 1 * time.Minute
