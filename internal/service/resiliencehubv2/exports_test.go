// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resiliencehubv2

// Exports for use in tests only.
var (
	ResourceInputSource = newInputSourceResource
	ResourcePolicy      = newPolicyResource
	ResourceService     = newServiceResource
	ResourceSystem      = newSystemResource

	FindInputSourceByID = findInputSourceByTwoPartKey
	FindPolicyByARN     = findPolicyByARN
	FindServiceByARN    = findServiceByARN
	FindSystemByARN     = findSystemByARN
)
