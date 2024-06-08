// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package docdb

// Exports for use in tests only.
var (
	FindDBClusterByID                 = findDBClusterByID
	FindDBClusterParameterGroupByName = findDBClusterParameterGroupByName
	FindDBSubnetGroupByName           = findDBSubnetGroupByName
	FindClusterSnapshotByID           = findClusterSnapshotByID
	FindDBInstanceByID                = findDBInstanceByID
	FindEventSubscriptionByName       = findEventSubscriptionByName
	FindGlobalClusterByID             = findGlobalClusterByID
)
