// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datazone

// Exports for use in tests only.
var (
	ResourceAssetType                         = newAssetTypeResource
	ResourceDomain                            = newDomainResource
	ResourceEnvironmentBlueprintConfiguration = newEnvironmentBlueprintConfigurationResource
	ResourceEnvironment                       = newEnvironmentResource
	ResourceEnvironmentProfile                = newEnvironmentProfileResource
	ResourceFormType                          = newFormTypeResource
	ResourceGlossary                          = newGlossaryResource
	ResourceGlossaryTerm                      = newGlossaryTermResource
	ResourceProject                           = newProjectResource
	ResourceUserProfile                       = newUserProfileResource

	FindAssetTypeByID          = findAssetTypeByID
	FindDomainByID             = findDomainByID
	FindEnvironmentByID        = findEnvironmentByID
	FindEnvironmentProfileByID = findEnvironmentProfileByID
	FindFormTypeByID           = findFormTypeByID
	FindGlossaryByID           = findGlossaryByID
	FindGlossaryTermByID       = findGlossaryTermByID
	FindUserProfileByID        = findUserProfileByID

	IsResourceMissing = isResourceMissing
)
