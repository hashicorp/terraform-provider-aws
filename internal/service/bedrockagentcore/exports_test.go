// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

// Exports for use in tests only.
var (
	ResourceAgentRuntime             = newAgentRuntimeResource
	ResourceAgentRuntimeEndpoint     = newAgentRuntimeEndpointResource
	ResourceAPIKeyCredentialProvider = newAPIKeyCredentialProviderResource
	ResourceBrowser                  = newBrowserResource
	ResourceGateway                  = newGatewayResource
	ResourceGatewayTarget            = newGatewayTargetResource
	ResourceCodeInterpreter          = newCodeInterpreterResource

	FindAgentRuntimeByID                 = findAgentRuntimeByID
	FindAgentRuntimeEndpointByTwoPartKey = findAgentRuntimeEndpointByTwoPartKey
	FindAPIKeyCredentialProviderByName   = findAPIKeyCredentialProviderByName
	FindBrowserByID                      = findBrowserByID
	FindGatewayByID                      = findGatewayByID
	FindGatewayTargetByTwoPartKey        = findGatewayTargetByTwoPartKey
	FindCodeInterpreterByID              = findCodeInterpreterByID
)
