// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrock

// Exports for use in tests only.
var (
	ResourceCustomModel                         = newCustomModelResource
	ResourceEnforcedGuardrailConfiguration      = newEnforcedGuardrailConfigurationResource
	ResourceGuardrail                           = newGuardrailResource
	ResourceGuardrailVersion                    = newGuardrailVersionResource
	ResourceModelInvocationLoggingConfiguration = newModelInvocationLoggingConfigurationResource
	ResourceInferenceProfile                    = newInferenceProfileResource

	FindCustomModelByID                     = findCustomModelByID
	FindEnforcedGuardrailConfiguration      = findEnforcedGuardrailConfiguration
	FindGuardrailByTwoPartKey               = findGuardrailByTwoPartKey
	FindModelCustomizationJobByID           = findModelCustomizationJobByID
	FindModelInvocationLoggingConfiguration = findModelInvocationLoggingConfiguration
	FindProvisionedModelThroughputByID      = findProvisionedModelThroughputByID

	WaitModelCustomizationJobCompleted = waitModelCustomizationJobCompleted
)
