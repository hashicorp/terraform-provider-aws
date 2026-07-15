// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package webfunctions

// Exports for use in tests only.
var (
	ResourceFunction = newFunctionResource
	ResourceEndpoint = newEndpointResource

	FindFunctionByName = findFunctionByName
	FindEndpointByName = findEndpointByName
)
