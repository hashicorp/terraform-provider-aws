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

	FindClusterEndpointByTwoPartKey   = findClusterEndpointByTwoPartKey
	FindClusterSnapshotByID           = findClusterSnapshotByID
	FindDBClusterByID                 = findDBClusterByID
	FindDBClusterParameterGroupByName = findDBClusterParameterGroupByName
	FindDBInstanceByID                = findDBInstanceByID
	FindEventSubscriptionByName       = findEventSubscriptionByName
)
