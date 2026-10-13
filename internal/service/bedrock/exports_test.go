// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrock

// Exports for use in tests only.
var (
	ResourceCustomModel                         = newCustomModelResource
	ResourceEvaluationJob                       = newEvaluationJobResource
	ResourceGuardrail                           = newGuardrailResource
	ResourceGuardrailVersion                    = newGuardrailVersionResource
	ResourceModelInvocationLoggingConfiguration = newModelInvocationLoggingConfigurationResource
	ResourceModelInvocationJob                  = newModelInvocationJobResource
	ResourceInferenceProfile                    = newInferenceProfileResource
	ResourceFoundationModelAgreement            = newFoundationModelAgreementResource

	FindCustomModelByID                     = findCustomModelByID
	FindEvaluationJobByARN                  = findEvaluationJobByARN
	FindGuardrailByTwoPartKey               = findGuardrailByTwoPartKey
	FindModelCustomizationJobByID           = findModelCustomizationJobByID
	FindModelInvocationJobByARN             = findModelInvocationJobByARN
	FindModelInvocationLoggingConfiguration = findModelInvocationLoggingConfiguration
	FindProvisionedModelThroughputByID      = findProvisionedModelThroughputByID
	FindFoundationModelAgreementByID        = findFoundationModelAgreementByID

	WaitModelCustomizationJobCompleted = waitModelCustomizationJobCompleted
)
