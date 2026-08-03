// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"time"
)

const (
	devEndpointStatusFailed       = "FAILED"
	devEndpointStatusProvisioning = "PROVISIONING"
	devEndpointStatusReady        = "READY"
	devEndpointStatusTerminating  = "TERMINATING"
)

const (
	jobCommandNameApacheSparkETL          = "glueetl"
	jobCommandNameApacheSparkStreamingETL = "gluestreaming"
	jobCommandNameRay                     = "glueray"
)

const (
	propagationTimeout = 2 * time.Minute
)
