// Copyright (c) HashiCorp, Inc.
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
