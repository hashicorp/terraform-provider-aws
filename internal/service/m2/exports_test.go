// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package m2

// Exports for use in tests only.
var (
	ResourceApplication = newApplicationResource
	ResourceDeployment  = newResourceDeployment
	ResourceEnvironment = newEnvironmentResource

	FindApplicationByID = findApplicationByID
	FindEnvironmentByID = findEnvironmentByID
)
