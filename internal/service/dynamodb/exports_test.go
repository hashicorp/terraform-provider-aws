// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb

// Exports for use in tests only.
var (
	ResourceKinesisStreamingDestination = resourceKinesisStreamingDestination
	ResourceTag                         = resourceTag
	ResourceResourcePolicy              = newResourcePolicyResource

	ARNForNewRegion                              = arnForNewRegion
	FindKinesisDataStreamDestinationByTwoPartKey = findKinesisDataStreamDestinationByTwoPartKey
	FindResourcePolicyByARN                      = findResourcePolicyByARN
	ListTags                                     = listTags
	RegionFromARN                                = regionFromARN
	TableNameFromARN                             = tableNameFromARN
)
