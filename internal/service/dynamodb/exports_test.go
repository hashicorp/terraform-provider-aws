// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb

// Exports for use in tests only.
var (
	ResourceKinesisStreamingDestination = resourceKinesisStreamingDestination
	ResourceTag                         = resourceTag
	ResourceResourcePolicy              = newResourcePolicyResource

	FindKinesisDataStreamDestinationByTwoPartKey = findKinesisDataStreamDestinationByTwoPartKey
	FindResourcePolicyByARN                      = findResourcePolicyByARN
	ListTags                                     = listTags
)
