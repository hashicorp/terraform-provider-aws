// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package docdb

import (
	"time"
)

const (
	propagationTimeout = 2 * time.Minute
)

const (
	engineDocDB = "docdb" // nosemgrep:ci.docdb-in-const-name,ci.docdb-in-var-name
)

func engine_Values() []string {
	return []string{
		engineDocDB,
	}
}

const (
	errCodeInvalidParameterValue = "InvalidParameterValue"
)

const (
	clusterStatusAvailable                  = "available"
	clusterStatusBackingUp                  = "backing-up"
	clusterStatusCreating                   = "creating"
	clusterStatusDeleting                   = "deleting"
	clusterStatusMigrating                  = "migrating"
	clusterStatusModifying                  = "modifying"
	clusterStatusPreparingDataMigration     = "preparing-data-migration"
	clusterStatusResettingMasterCredentials = "resetting-master-credentials"
	clusterStatusUpgrading                  = "upgrading"
)

const (
	clusterSnapshotStatusAvailable = "available"
	clusterSnapshotStatusCreating  = "creating"
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
	storageTypeIOpt1    = "iopt1"
	storageTypeStandard = "standard"
)

func storageType_Values() []string {
	return []string{
		storageTypeIOpt1,
		storageTypeStandard,
	}
}

const (
	RestoreTypeCopyOnWrite = "copy-on-write"
	RestoreTypeFullCopy    = "full-copy"
)

func RestoreType_Values() []string {
	return []string{
		RestoreTypeCopyOnWrite,
		RestoreTypeFullCopy,
	}
}
