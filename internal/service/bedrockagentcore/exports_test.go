// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

// Exports for use in tests only.
var (
	ResourceAgentRuntime             = newAgentRuntimeResource
	ResourceAgentRuntimeEndpoint     = newAgentRuntimeEndpointResource
	ResourceAPIKeyCredentialProvider = newAPIKeyCredentialProviderResource
	ResourceOAuth2CredentialProvider = newOAuth2CredentialProviderResource

	FindAgentRuntimeByID                 = findAgentRuntimeByID
	FindAgentRuntimeEndpointByTwoPartKey = findAgentRuntimeEndpointByTwoPartKey
	FindAPIKeyCredentialProviderByName   = findAPIKeyCredentialProviderByName
	FindOAuth2CredentialProviderByName   = findOAuth2CredentialProviderByName
)
