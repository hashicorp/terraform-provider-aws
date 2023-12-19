// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

// Exports for use in tests only.
var (
	ResourceCapacityProvider = resourceCapacityProvider
	ResourceCluster          = resourceCluster
	ResourceService          = resourceService

	FindCapacityProviderByARN = findCapacityProviderByARN
	FindClusterByNameOrARN    = findClusterByNameOrARN
	FindServiceNoTagsByID     = findServiceNoTagsByID
)
