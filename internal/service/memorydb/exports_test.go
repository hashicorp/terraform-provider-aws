// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package memorydb

// Exports for use in tests only.
var (
	ResourceACL                = resourceACL
	ResourceCluster            = resourceCluster
	ResourceMultiRegionCluster = newMultiRegionClusterResource
	ResourceParameterGroup     = resourceParameterGroup
	ResourceSnapshot           = resourceSnapshot
	ResourceSubnetGroup        = resourceSubnetGroup
	ResourceUser               = resourceUser

	FindACLByName                = findACLByName
	FindClusterByName            = findClusterByName
	FindMultiRegionClusterByName = findMultiRegionClusterByName
	FindParameterGroupByName     = findParameterGroupByName
	FindSnapshotByName           = findSnapshotByName
	FindSubnetGroupByName        = findSubnetGroupByName
	FindUserByName               = findUserByName
	ParameterChanges             = parameterChanges
	ParameterHash                = parameterHash
)
