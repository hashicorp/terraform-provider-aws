// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resiliencehubv2

// Exports for use in tests only.
var (
	ResourceInputSource = newResourceInputSource
	ResourcePolicy      = newResourcePolicy
	ResourceService     = newResourceService
	ResourceSystem      = newResourceSystem

	FindInputSourceByID = findInputSourceByID
	FindPolicyByARN     = findPolicyByARN
	FindServiceByARN    = findServiceByARN
	FindSystemByARN     = findSystemByARN
)
