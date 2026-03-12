// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package docdb

// Exports for use in tests only.
var (
	ResourceCluster               = resourceCluster
	ResourceClusterInstance       = resourceClusterInstance
	ResourceClusterParameterGroup = resourceClusterParameterGroup
	ResourceClusterSnapshot       = resourceClusterSnapshot
	ResourceEventSubscription     = resourceEventSubscription
	ResourceGlobalCluster         = resourceGlobalCluster
	ResourceSubnetGroup           = resourceSubnetGroup

	FindClusterSnapshotByID           = findClusterSnapshotByID
	FindDBClusterByID                 = findDBClusterByID
	FindDBClusterParameterGroupByName = findDBClusterParameterGroupByName
	FindDBInstanceByID                = findDBInstanceByID
	FindDBSubnetGroupByName           = findDBSubnetGroupByName
	FindEventSubscriptionByName       = findEventSubscriptionByName
	FindGlobalClusterByID             = findGlobalClusterByID
)
