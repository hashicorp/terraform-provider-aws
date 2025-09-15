// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

// Exports for use in tests only.
var (
	ResourceAgentRuntime             = newResourceAgentRuntime
	ResourceAPIKeyCredentialProvider = newResourceAPIKeyCredentialProvider
	ResourceBrowser                  = newResourceBrowser
	ResourceCodeInterpreter          = newResourceCodeInterpreter
	ResourceGateway                  = newResourceGateway
	ResourceGatewayTarget            = newResourceGatewayTarget
	ResourceMemory                   = newResourceMemory
	ResourceMemoryStrategy           = newResourceMemoryStrategy
	ResourceOAuth2CredentialProvider = newResourceOAuth2CredentialProvider
	ResourceWorkloadIdentity         = newResourceWorkloadIdentity

	FindAgentRuntimeByID               = findAgentRuntimeByID
	FindAPIKeyCredentialProviderByName = findAPIKeyCredentialProviderByName
	FindBrowserByID                    = findBrowserByID
	FindCodeInterpreterByID            = findCodeInterpreterByID
	FindGatewayByID                    = findGatewayByID
	FindGatewayTargetByID              = findGatewayTargetByID
	FindMemoryByID                     = findMemoryByID
	FindOAuth2CredentialProviderByName = findOAuth2CredentialProviderByName
	FindMemoryStrategyByID             = findMemoryStrategyByID
	FindWorkloadIdentityByName         = findWorkloadIdentityByName
)
