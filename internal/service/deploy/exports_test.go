// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package deploy

// Exports for use in tests only.
var (
	ResourceApp              = resourceApp
	ResourceDeploymentConfig = resourceDeploymentConfig // nosemgrep:ci.deploy-in-var-name
	ResourceDeploymentGroup  = resourceDeploymentGroup  // nosemgrep:ci.deploy-in-var-name

	FindApplicationByName           = findApplicationByName
	FindDeploymentConfigByName      = findDeploymentConfigByName      // nosemgrep:ci.deploy-in-var-name
	FindDeploymentGroupByTwoPartKey = findDeploymentGroupByTwoPartKey // nosemgrep:ci.deploy-in-var-name
)
