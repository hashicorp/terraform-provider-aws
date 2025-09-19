// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

// Exports for use in tests only.
var (
	ResourceAgentRuntime         = newResourceAgentRuntime
	ResourceAgentRuntimeEndpoint = newResourceAgentRuntimeEndpoint

	FindAgentRuntimeByID                       = findAgentRuntimeByID
	FindAgentRuntimeEndpointByRuntimeIDAndName = findAgentRuntimeEndpointByRuntimeIDAndName
)
