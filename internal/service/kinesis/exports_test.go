// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package kinesis

// Exports for use in tests only.
var (
	ResourceResourcePolicy = newResourcePolicyResource
	ResourceStream         = resourceStream
	ResourceStreamConsumer = resourceStreamConsumer

	FindLimits              = findLimits
	FindResourcePolicyByARN = findResourcePolicyByARN
	FindStreamByName        = findStreamByName
	FindStreamConsumerByARN = findStreamConsumerByARN
)
