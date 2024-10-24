// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

// Exports for use in tests only.
var (
	ResourceCustomModel                         = newCustomModelResource
	ResourceGuardrail                           = newResourceGuardrail
	ResourceGuardrailVersion                    = newGuardrailVersionResource
	ResourceModelInvocationLoggingConfiguration = newModelInvocationLoggingConfigurationResource

	FindCustomModelByID                     = findCustomModelByID
	FindGuardrailByTwoPartKey               = findGuardrailByTwoPartKey
	FindModelCustomizationJobByID           = findModelCustomizationJobByID
	FindModelInvocationLoggingConfiguration = findModelInvocationLoggingConfiguration
	FindProvisionedModelThroughputByID      = findProvisionedModelThroughputByID

	WaitModelCustomizationJobCompleted = waitModelCustomizationJobCompleted
)
