// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kinesis

// Exports for use in tests only.
var (
	ResourceAccountSettings = newAccountSettingsResource
	ResourceResourcePolicy  = newResourcePolicyResource
	ResourceStream          = resourceStream
	ResourceStreamConsumer  = resourceStreamConsumer

	FindAccountSettings     = findAccountSettings
	FindLimits              = findLimits
	FindResourcePolicyByARN = findResourcePolicyByARN
	FindStreamByName        = findStreamByName
	FindStreamConsumerByARN = findStreamConsumerByARN
)
