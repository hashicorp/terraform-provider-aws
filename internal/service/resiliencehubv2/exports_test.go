// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resiliencehubv2

// Exports for use in tests only.
var (
	ResourcePolicy  = newPolicyResource
	ResourceService = newServiceResource
	ResourceSystem  = newSystemResource

	FindPolicyByARN  = findPolicyByARN
	FindServiceByARN = findServiceByARN
	FindSystemByARN  = findSystemByARN
)
