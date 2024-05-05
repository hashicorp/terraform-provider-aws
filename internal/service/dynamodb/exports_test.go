// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb

// Exports for use in tests only.
var (
	ResourceContributorInsights         = resourceContributorInsights
	ResourceKinesisStreamingDestination = resourceKinesisStreamingDestination
	ResourceTag                         = resourceTag
	ResourceResourcePolicy              = newResourcePolicyResource

	ARNForNewRegion                              = arnForNewRegion
	ContributorInsightsParseResourceID           = contributorInsightsParseResourceID
	FindContributorInsightsByTwoPartKey          = findContributorInsightsByTwoPartKey
	FindKinesisDataStreamDestinationByTwoPartKey = findKinesisDataStreamDestinationByTwoPartKey
	FindResourcePolicyByARN                      = findResourcePolicyByARN
	ListTags                                     = listTags
	RegionFromARN                                = regionFromARN
	TableNameFromARN                             = tableNameFromARN
)
