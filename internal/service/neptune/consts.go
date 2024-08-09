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
	engineNeptune = "neptune" // nosemgrep:ci.neptune-in-const-name,ci.neptune-in-var-name
)

func engine_Values() []string {
	return []string{
		engineNeptune,
	}
}

const (
	storageTypeStandard = "standard"
	storageTypeIopt1    = "iopt1"
)

func storageType_Values() []string {
	return []string{
		storageTypeStandard,
		storageTypeIopt1,
	}
}

const (
	clusterEndpointStatusAvailable = "available"
	clusterEndpointStatusCreating  = "creating"
	clusterEndpointStatusDeleting  = "deleting"
	clusterEndpointStatusModifying = "modifying"
)

const (
	clusterSnapshotStatusAvailable = "available"
	clusterSnapshotStatusCreating  = "creating"
)

const (
	clusterStatusAvailable                  = "available"
	clusterStatusBackingUp                  = "backing-up"
	clusterStatusConfiguringIAMDatabaseAuth = "configuring-iam-database-auth"
	clusterStatusCreating                   = "creating"
	clusterStatusDeleting                   = "deleting"
	clusterStatusMigrating                  = "migrating"
	clusterStatusModifying                  = "modifying"
	clusterStatusPreparingDataMigration     = "preparing-data-migration"
	clusterStatusUpgrading                  = "upgrading"
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
