// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datazone

// Exports for use in tests only.
var (
	ResourceDomain                            = newResourceDomain
	ResourceEnvironmentBlueprintConfiguration = newResourceEnvironmentBlueprintConfiguration
	IsResourceMissing                         = isResourceMissing
	ResourceFormType                          = newResourceFormType
	FindFormTypeByID                          = findFormTypeByID
	ResourceProject                           = newResourceProject
	ResourceGlossary                          = newResourceGlossary
	FindGlossaryByID                          = findGlossaryByID
	ResourceAssetType                         = newResourceAssetType
	FindAssetTypeByID                         = findAssetTypeByID
)
