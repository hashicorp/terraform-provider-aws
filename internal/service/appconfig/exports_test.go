// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appconfig

// Exports for use in tests only.
var (
	ResourceApplication          = resourceApplication
	ResourceConfigurationProfile = resourceConfigurationProfile
	ResourceDeployment           = resourceDeployment
	ResourceDeploymentStrategy   = resourceDeploymentStrategy
	ResourceEnvironmentFW        = newResourceEnvironment
	ResourceExtension            = resourceExtension
	ResourceExtensionAssociation = resourceExtensionAssociation

	FindApplicationByID                  = findApplicationByID
	FindConfigurationProfileByTwoPartKey = findConfigurationProfileByTwoPartKey
	FindDeploymentByThreePartKey         = findDeploymentByThreePartKey
	FindDeploymentStrategyByID           = findDeploymentStrategyByID
	FindExtensionByID                    = findExtensionByID
	FindExtensionAssociationByID         = findExtensionAssociationByID
)
