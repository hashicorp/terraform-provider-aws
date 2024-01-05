// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kinesis

// Exports for use in tests only.
var (
	ResourceStream         = resourceStream
	ResourceStreamConsumer = resourceStreamConsumer

	FindStreamByName        = findStreamByName
	FindStreamConsumerByARN = findStreamConsumerByARN
)
