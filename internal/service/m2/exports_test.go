// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package m2

// Exports for use in tests only.
var (
	ResourceApplication = newApplicationResource
	ResourceDeployment  = newDeploymentResource
	ResourceEnvironment = newEnvironmentResource

	FindApplicationByID        = findApplicationByID
	FindDeploymentByTwoPartKey = findDeploymentByTwoPartKey
	FindEnvironmentByID        = findEnvironmentByID
)
