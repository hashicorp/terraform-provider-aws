// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

// Exports for use in tests only.
var (
	ResourceClusterCapacityProviders = resourceClusterCapacityProviders
	ResourceCapacityProvider         = resourceCapacityProvider
	ResourceCluster                  = resourceCluster
	ResourceService                  = resourceService
	ResourceTag                      = resourceTag
	ResourceTaskDefinition           = resourceTaskDefinition
	ResourceTaskSet                  = resourceTaskSet

	FindCapacityProviderByARN = findCapacityProviderByARN
	FindClusterByNameOrARN    = findClusterByNameOrARN
	FindServiceNoTagsByID     = findServiceNoTagsByID
)
