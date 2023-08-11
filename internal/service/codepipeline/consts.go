// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codepipeline

import (
	"time"
)

const (
	ResNameWebhook  = "Webhook"
	ResNamePipeline = "Pipeline"
)

const (
	propagationTimeout = 2 * time.Minute
)
