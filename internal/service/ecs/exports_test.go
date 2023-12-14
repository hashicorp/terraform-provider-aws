// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

// Exports for use in tests only.
var (
	ResourceCapacityProvider = resourceCapacityProvider
	ResourceCluster          = resourceCluster

	FindCapacityProviderByARN = findCapacityProviderByARN
	FindClusterByNameOrARN    = findClusterByNameOrARN
)
