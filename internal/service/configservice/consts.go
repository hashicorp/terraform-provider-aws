// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package configservice

import (
	"time"
)

const (
	organizationsPropagationTimeout = 1 * time.Minute // Organizations eventual consistency.
	propagationTimeout              = 2 * time.Minute // IAM eventual consistency.
)

const (
	defaultConfigurationRecorderName = "default"
	defaultDeliveryChannelName       = "default"
)
