// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package continuum

// Exports for use in tests only.
var (
	ResourceCollector      = newCollectorResource
	ResourceResourcePolicy = newResourcePolicyResource

	FindCollectorByID       = findCollectorByID
	FindResourcePolicyByARN = findResourcePolicyByARN
)
