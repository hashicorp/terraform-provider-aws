// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package bedrock

// Exports for use in tests only.
var (
	ResourceCustomModel                         = newCustomModelResource
	ResourceGuardrail                           = newGuardrailResource
	ResourceGuardrailVersion                    = newGuardrailVersionResource
	ResourceInferenceProfile                    = newInferenceProfileResource
	ResourceModelInvocationLoggingConfiguration = newModelInvocationLoggingConfigurationResource
	ResourcePromptRouter                        = newPromptRouterResource

	FindCustomModelByID                     = findCustomModelByID
	FindGuardrailByTwoPartKey               = findGuardrailByTwoPartKey
	FindModelCustomizationJobByID           = findModelCustomizationJobByID
	FindModelInvocationLoggingConfiguration = findModelInvocationLoggingConfiguration
	FindPromptRouterByARN                   = findPromptRouterByARN
	FindProvisionedModelThroughputByID      = findProvisionedModelThroughputByID

	WaitModelCustomizationJobCompleted = waitModelCustomizationJobCompleted
)
