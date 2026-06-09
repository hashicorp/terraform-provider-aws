// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resiliencehubv2

// Exports for use in tests only.
var (
	ResourcePolicy = newResourcePolicy
	ResourceSystem = newResourceSystem

	FindPolicyByARN = findPolicyByARN
	FindSystemByARN = findSystemByARN
)
