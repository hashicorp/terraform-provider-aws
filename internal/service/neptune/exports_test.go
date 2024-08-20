// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package neptune

// Exports for use in tests only.
var (
	ResourceClusterEndpoint       = resourceClusterEndpoint
	ResourceClusterInstance       = resourceClusterInstance
	ResourceClusterParameterGroup = resourceClusterParameterGroup
	ResourceClusterSnapshot       = resourceClusterSnapshot

	FindClusterEndpointByTwoPartKey   = findClusterEndpointByTwoPartKey
	FindDBClusterParameterGroupByName = findDBClusterParameterGroupByName
	FindClusterSnapshotByID           = findClusterSnapshotByID
	FindDBInstanceByID                = findDBInstanceByID
)
