// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opsworks

import (
	"time"
)

const (
	defaultBerkshelfVersion = "3.2.0"
)

const (
	instanceStatusBooting      = "booting"
	instanceStatusOnline       = "online"
	instanceStatusPending      = "pending"
	instanceStatusRequested    = "requested"
	instanceStatusRunning      = "running"
	instanceStatusRunningSetup = "running_setup"
	instanceStatusShuttingDown = "shutting_down"
	instanceStatusStopped      = "stopped"
	instanceStatusStopping     = "stopping"
	instanceStatusTerminated   = "terminated"
	instanceStatusTerminating  = "terminating"
)

const (
	propagationTimeout = 2 * time.Minute
)
