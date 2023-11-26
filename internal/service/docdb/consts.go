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
