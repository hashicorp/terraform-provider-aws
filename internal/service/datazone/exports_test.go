// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datazone

// Exports for use in tests only.
var (
	ResourceDomain                            = newResourceDomain
	ResourceEnvironmentBlueprintConfiguration = newResourceEnvironmentBlueprintConfiguration
	IsResourceMissing                         = isResourceMissing
	ResourceProject                           = newResourceProject
	ResourceEnvironmentProfile                = newResourceEnvironmentProfile
	FindEnvironmentProfileByID                = findEnvironmentProfileByID
	ResourceGlossary                          = newResourceGlossary
	FindGlossaryByID                          = findGlossaryByID
)
