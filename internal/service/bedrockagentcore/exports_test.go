// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

// Exports for use in tests only.
var (
	ResourceAgentRuntime             = newResourceAgentRuntime
	ResourceCodeInterpreter          = newResourceCodeInterpreter
	ResourceGateway                  = newResourceGateway
	ResourceGatewayTarget            = newResourceGatewayTarget
	ResourceMemory                   = newResourceMemory
	ResourceMemoryStrategy           = newResourceMemoryStrategy
	ResourceOAuth2CredentialProvider = newResourceOAuth2CredentialProvider

	FindAgentRuntimeByID               = findAgentRuntimeByID
	FindCodeInterpreterByID            = findCodeInterpreterByID
	FindGatewayByID                    = findGatewayByID
	FindGatewayTargetByID              = findGatewayTargetByID
	FindMemoryByID                     = findMemoryByID
	FindOAuth2CredentialProviderByName = findOAuth2CredentialProviderByName
	FindMemoryStrategyByID             = findMemoryStrategyByID
)
