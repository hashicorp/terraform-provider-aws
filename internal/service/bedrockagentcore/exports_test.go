// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

// Exports for use in tests only.
var (
	ResourceAgentRuntime             = newResourceAgentRuntime
	ResourceOAuth2CredentialProvider = newResourceOAuth2CredentialProvider

	FindAgentRuntimeByID               = findAgentRuntimeByID
	FindOAuth2CredentialProviderByName = findOAuth2CredentialProviderByName
)
