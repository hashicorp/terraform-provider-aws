package opsworks

import "time"

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
