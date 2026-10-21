// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambdaweb

// Exports for use in tests only.
var (
	ResourceFunction         = newFunctionResource
	ResourceFunctionEndpoint = newFunctionEndpointResource
	ResourceResourcePolicy   = newResourcePolicyResource

	FindResourcePolicyByARN             = findResourcePolicyByARN
	FindWebFunctionEndpointByTwoPartKey = findWebFunctionEndpointByTwoPartKey
	FindWebFunctionByName               = findWebFunctionByName
)
