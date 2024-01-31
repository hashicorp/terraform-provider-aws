// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

// Exports for use in tests only.
var (
	ResourceCustomModelName                     = newCustomModelResource
	ResourceModelInvocationLoggingConfiguration = newResourceModelInvocationLoggingConfiguration

	FindModelCustomizationJobByID = findModelCustomizationJobByID
)
