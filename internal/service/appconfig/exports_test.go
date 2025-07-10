// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appconfig

// Exports for use in tests only.
var (
	ResourceApplication                = resourceApplication
	ResourceConfigurationProfile       = resourceConfigurationProfile
	ResourceDeployment                 = resourceDeployment
	ResourceDeploymentStrategy         = resourceDeploymentStrategy
	ResourceEnvironment                = newEnvironmentResource
	ResourceExtension                  = resourceExtension
	ResourceExtensionAssociation       = resourceExtensionAssociation
	ResourceHostedConfigurationVersion = resourceHostedConfigurationVersion

	FindApplicationByID                          = findApplicationByID
	FindConfigurationProfileByTwoPartKey         = findConfigurationProfileByTwoPartKey
	FindDeploymentByThreePartKey                 = findDeploymentByThreePartKey
	FindDeploymentStrategyByID                   = findDeploymentStrategyByID
	FindEnvironmentByTwoPartKey                  = findEnvironmentByTwoPartKey
	FindExtensionByID                            = findExtensionByID
	FindExtensionAssociationByID                 = findExtensionAssociationByID
	FindHostedConfigurationVersionByThreePartKey = findHostedConfigurationVersionByThreePartKey
)
