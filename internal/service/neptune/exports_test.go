// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package neptune

// Exports for use in tests only.
var (
	ResourceCluster               = resourceCluster
	ResourceClusterEndpoint       = resourceClusterEndpoint
	ResourceClusterInstance       = resourceClusterInstance
	ResourceClusterParameterGroup = resourceClusterParameterGroup
	ResourceClusterSnapshot       = resourceClusterSnapshot
	ResourceEventSubscription     = resourceEventSubscription
	ResourceGlobalCluster         = resourceGlobalCluster
	ResourceParameterGroup        = resourceParameterGroup
	ResourceSubnetGroup           = resourceSubnetGroup

	FindClusterEndpointByTwoPartKey   = findClusterEndpointByTwoPartKey
	FindClusterSnapshotByID           = findClusterSnapshotByID
	FindDBClusterByID                 = findDBClusterByID
	FindDBClusterParameterGroupByName = findDBClusterParameterGroupByName
	FindDBInstanceByID                = findDBInstanceByID
	FindDBParameterGroupByName        = findDBParameterGroupByName
	FindEventSubscriptionByName       = findEventSubscriptionByName
	FindGlobalClusterByID             = findGlobalClusterByID
	FindSubnetGroupByName             = findSubnetGroupByName
)
