// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datazone

// Exports for use in tests only.
var (
	ResourceAssetType                         = newResourceAssetType
	ResourceDomain                            = newResourceDomain
	ResourceEnvironmentBlueprintConfiguration = newResourceEnvironmentBlueprintConfiguration
	ResourceEnvironment                       = newResourceEnvironment
	ResourceEnvironmentProfile                = newResourceEnvironmentProfile
	ResourceFormType                          = newResourceFormType
	ResourceGlossary                          = newResourceGlossary
	ResourceGlossaryTerm                      = newResourceGlossaryTerm
	ResourceProject                           = newResourceProject
	ResourceUserProfile                       = newResourceUserProfile

	FindAssetTypeByID          = findAssetTypeByID
	FindEnvironmentByID        = findEnvironmentByID
	FindEnvironmentProfileByID = findEnvironmentProfileByID
	FindFormTypeByID           = findFormTypeByID
	FindGlossaryByID           = findGlossaryByID
	FindGlossaryTermByID       = findGlossaryTermByID
	FindUserProfileByID        = findUserProfileByID

	IsResourceMissing = isResourceMissing
)
