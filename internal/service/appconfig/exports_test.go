// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appconfig

// Exports for use in tests only.
var (
	ResourceApplication          = resourceApplication
	ResourceConfigurationProfile = resourceConfigurationProfile
	ResourceEnvironmentFW        = newResourceEnvironment

	FindApplicationByID                  = findApplicationByID
	FindConfigurationProfileByTwoPartKey = findConfigurationProfileByTwoPartKey
)
