// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

// Exports for use in tests only.
var (
	ResourceCustomModel                         = newCustomModelResource
	ResourceModelInvocationLoggingConfiguration = newModelInvocationLoggingConfigurationResource

	FindCustomModelByID                     = findCustomModelByID
	FindModelCustomizationJobByID           = findModelCustomizationJobByID
	FindModelInvocationLoggingConfiguration = findModelInvocationLoggingConfiguration
	FindProvisionedModelThroughputByID      = findProvisionedModelThroughputByID
	WaitModelCustomizationJobCompleted      = waitModelCustomizationJobCompleted
)
