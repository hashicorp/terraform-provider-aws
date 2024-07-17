// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

// Exports for use in tests only.
var (
	ResourceAccountSettingDefault    = resourceAccountSettingDefault
	ResourceCapacityProvider         = resourceCapacityProvider
	ResourceCluster                  = resourceCluster
	ResourceClusterCapacityProviders = resourceClusterCapacityProviders
	ResourceTag                      = resourceTag

	FindCapacityProviderByARN               = findCapacityProviderByARN
	FindClusterByNameOrARN                  = findClusterByNameOrARN
	FindEffectiveAccountSettingByName       = findEffectiveAccountSettingByName
	FindTag                                 = findTag
	ValidTaskDefinitionContainerDefinitions = validTaskDefinitionContainerDefinitions
)
