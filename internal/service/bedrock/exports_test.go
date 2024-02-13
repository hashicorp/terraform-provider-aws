// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

// Exports for use in tests only.
var (
	ResourceCustomModel                         = newCustomModelResource
	ResourceModelInvocationLoggingConfiguration = newResourceModelInvocationLoggingConfiguration

	FindCustomModelByID                = findCustomModelByID
	FindModelCustomizationJobByID      = findModelCustomizationJobByID
	WaitModelCustomizationJobCompleted = waitModelCustomizationJobCompleted
)
