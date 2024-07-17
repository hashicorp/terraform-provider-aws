// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

// Exports for use in tests only.
var (
	ResourceAccountSettingDefault    = resourceAccountSettingDefault
	ResourceCapacityProvider         = resourceCapacityProvider
	ResourceClusterCapacityProviders = resourceClusterCapacityProviders
	ResourceTag                      = resourceTag

	FindCapacityProviderByARN               = findCapacityProviderByARN
	FindEffectiveAccountSettingByName       = findEffectiveAccountSettingByName
	FindTag                                 = findTag
	ValidTaskDefinitionContainerDefinitions = validTaskDefinitionContainerDefinitions
)
