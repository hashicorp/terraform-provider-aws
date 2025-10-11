// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

// Exports for use in tests only.
var (
	ResourceAgentRuntime         = newAgentRuntimeResource
	ResourceAgentRuntimeEndpoint = newAgentRuntimeEndpointResource
	ResourceBrowser              = newBrowserResource

	FindAgentRuntimeByID                 = findAgentRuntimeByID
	FindAgentRuntimeEndpointByTwoPartKey = findAgentRuntimeEndpointByTwoPartKey
	FindBrowserByID                      = findBrowserByID
)
