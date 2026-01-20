// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package docdb

// Exports for use in tests only.
var (
	ResourceCluster               = resourceCluster
	ResourceClusterInstance       = resourceClusterInstance
	ResourceClusterParameterGroup = resourceClusterParameterGroup
	ResourceEventSubscription     = resourceEventSubscription
	ResourceGlobalCluster         = resourceGlobalCluster
	ResourceSubnetGroup           = resourceSubnetGroup

	FindDBClusterByID                 = findDBClusterByID
	FindDBClusterParameterGroupByName = findDBClusterParameterGroupByName
	FindDBSubnetGroupByName           = findDBSubnetGroupByName
	FindClusterSnapshotByID           = findClusterSnapshotByID
	FindDBInstanceByID                = findDBInstanceByID
	FindEventSubscriptionByName       = findEventSubscriptionByName
	FindGlobalClusterByID             = findGlobalClusterByID
)
