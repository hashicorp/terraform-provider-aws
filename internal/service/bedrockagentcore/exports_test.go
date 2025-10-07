// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

// Exports for use in tests only.
var (
	ResourceAgentRuntime         = newAgentRuntimeResource
	ResourceAgentRuntimeEndpoint = newAgentRuntimeEndpointResource
	ResourceCodeInterpreter      = newResourceCodeInterpreter

	FindAgentRuntimeByID                 = findAgentRuntimeByID
	FindAgentRuntimeEndpointByTwoPartKey = findAgentRuntimeEndpointByTwoPartKey
	FindCodeInterpreterByID              = findCodeInterpreterByID
)
