// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package webfunctions

// Exports for use in tests only.
var (
	ResourceFunction = newFunctionResource
	ResourceEndpoint = newEndpointResource

	FindFunctionByName = findFunctionByName
	FindEndpointByName = findEndpointByName
)
