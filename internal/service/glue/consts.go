package glue

import "time"

const (
	devEndpointStatusFailed       = "FAILED"
	devEndpointStatusProvisioning = "PROVISIONING"
	devEndpointStatusReady        = "READY"
	devEndpointStatusTerminating  = "TERMINATING"
)

const (
	propagationTimeout = 2 * time.Minute
)
