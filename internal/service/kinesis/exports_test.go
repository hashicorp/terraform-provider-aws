// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kinesis

// Exports for use in tests only.
var (
	ResourceResourcePolicy = newResourcePolicyResource
	ResourceStream         = resourceStream
	ResourceStreamConsumer = resourceStreamConsumer

	FindResourcePolicyByARN = findResourcePolicyByARN
	FindStreamByName        = findStreamByName
	FindStreamConsumerByARN = findStreamConsumerByARN
)
