// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"time"
)

const (
	propagationTimeout = 2 * time.Minute
)

const (
	clusterStatusActive         = "ACTIVE"
	clusterStatusDeprovisioning = "DEPROVISIONING"
	clusterStatusInactive       = "INACTIVE"
	clusterStatusProvisioning   = "PROVISIONING"
)
