// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

// Exports for use in tests only.
var (
	ResourceAccountSettingDefault    = resourceAccountSettingDefault
	ResourceCapacityProvider         = resourceCapacityProvider
	ResourceCluster                  = resourceCluster
	ResourceClusterCapacityProviders = resourceClusterCapacityProviders
	ResourceExpressGatewayService    = newExpressGatewayServiceResource
	ResourceService                  = resourceService
	ResourceTag                      = resourceTag
	ResourceTaskDefinition           = resourceTaskDefinition
	ResourceTaskSet                  = resourceTaskSet

	ClusterNameFromARN                      = clusterNameFromARN
	FindCapacityProviderByARN               = findCapacityProviderByARN
	FindClusterByNameOrARN                  = findClusterByNameOrARN
	FindEffectiveAccountSettingByName       = findEffectiveAccountSettingByName
	FindExpressGatewayServiceByARN          = findExpressGatewayServiceByARN
	FindServiceNoTagsByTwoPartKey           = findServiceNoTagsByTwoPartKey
	FindTag                                 = findTag
	FindTaskDefinitionByFamilyOrARN         = findTaskDefinitionByFamilyOrARN
	FindTaskSetNoTagsByThreePartKey         = findTaskSetNoTagsByThreePartKey
	RoleNameFromARN                         = roleNameFromARN
	ServiceNameFromARN                      = serviceNameFromARN
	TaskDefinitionARNStripRevision          = taskDefinitionARNStripRevision
	ValidTaskDefinitionContainerDefinitions = validTaskDefinitionContainerDefinitions
)
