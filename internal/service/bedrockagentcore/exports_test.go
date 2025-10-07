// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

// Exports for use in tests only.
var (
	ResourceAgentRuntime         = newAgentRuntimeResource
	ResourceAgentRuntimeEndpoint = newAgentRuntimeEndpointResource
	ResourceGateway              = newResourceGateway
	ResourceGatewayTarget        = newResourceGatewayTarget

	FindAgentRuntimeByID                 = findAgentRuntimeByID
	FindAgentRuntimeEndpointByTwoPartKey = findAgentRuntimeEndpointByTwoPartKey
	FindGatewayByID                      = findGatewayByID
	FindGatewayTargetByID                = findGatewayTargetByID
)
