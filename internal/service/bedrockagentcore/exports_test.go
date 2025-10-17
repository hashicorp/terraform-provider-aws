// Copyright (c) HashiCorp, Inc.
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
	ResourceOAuth2CredentialProvider = newOAuth2CredentialProviderResource
	ResourceWorkloadIdentity         = newWorkloadIdentityResource

	FindAgentRuntimeByID                 = findAgentRuntimeByID
	FindAgentRuntimeEndpointByTwoPartKey = findAgentRuntimeEndpointByTwoPartKey
	FindAPIKeyCredentialProviderByName   = findAPIKeyCredentialProviderByName
	FindBrowserByID                      = findBrowserByID
	FindCodeInterpreterByID              = findCodeInterpreterByID
	FindGatewayByID                      = findGatewayByID
	FindGatewayTargetByTwoPartKey        = findGatewayTargetByTwoPartKey
	FindOAuth2CredentialProviderByName   = findOAuth2CredentialProviderByName
	FindWorkloadIdentityByName           = findWorkloadIdentityByName
)
