// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore

// Exports for use in tests only.
var (
	ResourceAgentRuntime             = newAgentRuntimeResource
	ResourceAgentRuntimeEndpoint     = newAgentRuntimeEndpointResource
	ResourceAPIKeyCredentialProvider = newAPIKeyCredentialProviderResource
	ResourceBrowser                  = newBrowserResource
	ResourceBrowserProfile           = newBrowserProfileResource
	ResourceCodeInterpreter          = newCodeInterpreterResource
	ResourceEvaluator                = newEvaluatorResource
	ResourceGateway                  = newGatewayResource
	ResourceGatewayTarget            = newGatewayTargetResource
	ResourceMemory                   = newMemoryResource
	ResourceResourcePolicy           = newResourcePolicyResource
	ResourceMemoryStrategy           = newResourceMemoryStrategy
	ResourceOAuth2CredentialProvider = newOAuth2CredentialProviderResource
	ResourcePolicy                   = newPolicyResource
	ResourceTokenVaultCMK            = newTokenVaultCMKResource
	ResourceHarness                  = newHarnessResource
	ResourceOnlineEvaluationConfig   = newOnlineEvaluationConfigResource
	ResourcePolicyEngine             = newPolicyEngineResource
	ResourceRegistry                 = newRegistryResource
	ResourcePaymentManager           = newPaymentManagerResource
	ResourceWorkloadIdentity         = newWorkloadIdentityResource

	FindAgentRuntimeByID                 = findAgentRuntimeByID
	FindHarnessByID                      = findHarnessByID
	FindAgentRuntimeEndpointByTwoPartKey = findAgentRuntimeEndpointByTwoPartKey
	FindAPIKeyCredentialProviderByName   = findAPIKeyCredentialProviderByName
	FindBrowserByID                      = findBrowserByID
	FindBrowserProfileByID               = findBrowserProfileByID
	FindCodeInterpreterByID              = findCodeInterpreterByID
	FindEvaluatorByID                    = findEvaluatorByID
	FindGatewayByID                      = findGatewayByID
	FindGatewayTargetByTwoPartKey        = findGatewayTargetByTwoPartKey
	FindMemoryByID                       = findMemoryByID
	FindMemoryStrategyByTwoPartKey       = findMemoryStrategyByTwoPartKey
	FindOAuth2CredentialProviderByName   = findOAuth2CredentialProviderByName
	FindOnlineEvaluationConfigByID       = findOnlineEvaluationConfigByID
	FindPolicyByTwoPartKey               = findPolicyByTwoPartKey
	FindResourcePolicyByARN              = findResourcePolicyByARN
	FindTokenVaultByID                   = findTokenVaultByID
	FindPolicyEngineByID                 = findPolicyEngineByID
	FindRegistryByID                     = findRegistryByID
	FindPaymentManagerByID               = findPaymentManagerByID
	FindWorkloadIdentityByName           = findWorkloadIdentityByName
)

type (
	CustomJWTAuthorizerConfigurationModel = customJWTAuthorizerConfigurationModel
	ManagedVPCResourceModel               = managedVPCResourceModel
	PrivateEndpointModel                  = privateEndpointModel
	SelfManagedLatticeResourceModel       = selfManagedLatticeResourceModel
)
