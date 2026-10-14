// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambdaweb

// Exports for use in tests only.
var (
	ResourceEndpoint       = newEndpointResource
	ResourceFunction       = newFunctionResource
	ResourceResourcePolicy = newResourcePolicyResource

	FindEndpointByName      = findEndpointByName
	FindResourcePolicyByARN = findResourcePolicyByARN
	FindWebFunctionByName   = findWebFunctionByName
)
