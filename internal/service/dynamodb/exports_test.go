// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb

// Exports for use in tests only.
var (
	ResourceKinesisStreamingDestination = resourceKinesisStreamingDestination
	ResourceTag                         = resourceTag

	FindKinesisDataStreamDestinationByTwoPartKey = findKinesisDataStreamDestinationByTwoPartKey
	ListTags                                     = listTags
)
