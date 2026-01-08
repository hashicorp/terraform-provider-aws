// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

// Exports for use in tests only.
var (
	ResourceAgentRuntime             = newAgentRuntimeResource
	ResourceAgentRuntimeEndpoint     = newAgentRuntimeEndpointResource
	ResourceAPIKeyCredentialProvider = newAPIKeyCredentialProviderResource
	ResourceBrowser                  = newBrowserResource
	ResourceCodeInterpreter          = newCodeInterpreterResource
	ResourceGateway                  = newGatewayResource
	ResourceGatewayTarget            = newGatewayTargetResource
	ResourceMemory                   = newMemoryResource
	ResourceMemoryStrategy           = newResourceMemoryStrategy
	ResourceOAuth2CredentialProvider = newOAuth2CredentialProviderResource
	ResourceTokenVaultCMK            = newTokenVaultCMKResource
	ResourceWorkloadIdentity         = newWorkloadIdentityResource

	FindAgentRuntimeByID                 = findAgentRuntimeByID
	FindAgentRuntimeEndpointByTwoPartKey = findAgentRuntimeEndpointByTwoPartKey
	FindAPIKeyCredentialProviderByName   = findAPIKeyCredentialProviderByName
	FindBrowserByID                      = findBrowserByID
	FindCodeInterpreterByID              = findCodeInterpreterByID
	FindGatewayByID                      = findGatewayByID
	FindGatewayTargetByTwoPartKey        = findGatewayTargetByTwoPartKey
	FindMemoryByID                       = findMemoryByID
	FindMemoryStrategyByID               = findMemoryStrategyByTwoPartKey
	FindOAuth2CredentialProviderByName   = findOAuth2CredentialProviderByName
	FindTokenVaultByID                   = findTokenVaultByID
	FindWorkloadIdentityByName           = findWorkloadIdentityByName
)
