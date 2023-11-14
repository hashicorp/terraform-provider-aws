// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package neptune

import (
	"time"
)

const (
	propagationTimeout = 2 * time.Minute
)

const (
	dbInstanceStatusAvailable                     = "available"
	dbInstanceStatusBackingUp                     = "backing-up"
	dbInstanceStatusConfiguringEnhancedMonitoring = "configuring-enhanced-monitoring"
	dbInstanceStatusConfiguringIAMDatabaseAuth    = "configuring-iam-database-auth"
	dbInstanceStatusConfiguringLogExports         = "configuring-log-exports"
	dbInstanceStatusCreating                      = "creating"
	dbInstanceStatusDeleting                      = "deleting"
	dbInstanceStatusMaintenance                   = "maintenance"
	dbInstanceStatusModifying                     = "modifying"
	dbInstanceStatusRebooting                     = "rebooting"
	dbInstanceStatusRenaming                      = "renaming"
	dbInstanceStatusResettingMasterCredentials    = "resetting-master-credentials"
	dbInstanceStatusStarting                      = "starting"
	dbInstanceStatusStorageOptimization           = "storage-optimization"
	dbInstanceStatusUpgrading                     = "upgrading"
)

const (
	eventSubscriptionStatusActive    = "active"
	eventSubscriptionStatusCreating  = "creating"
	eventSubscriptionStatusDeleting  = "deleting"
	eventSubscriptionStatusModifying = "modifying"
)

const (
	globalClusterStatusAvailable = "available"
	globalClusterStatusCreating  = "creating"
	globalClusterStatusDeleted   = "deleted"
	globalClusterStatusDeleting  = "deleting"
	globalClusterStatusModifying = "modifying"
	globalClusterStatusUpgrading = "upgrading"
)

const (
	clusterEndpointTypeAny    = "ANY"
	clusterEndpointTypeReader = "READER"
	clusterEndpointTypeWriter = "WRITER"
)

func clusterEndpointType_Values() []string {
	return []string{
		clusterEndpointTypeAny,
		clusterEndpointTypeReader,
		clusterEndpointTypeWriter,
	}
}

const (
	errCodeInvalidParameterValue = "InvalidParameterValue"
)
