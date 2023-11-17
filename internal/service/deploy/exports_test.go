// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package deploy

// Exports for use in tests only.
var (
	ResourceApp              = resourceApp
	ResourceDeploymentConfig = resourceDeploymentConfig

	FindApplicationByName      = findApplicationByName
	FindDeploymentConfigByName = findDeploymentConfigByName
)
