// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb

// Exports for use in tests only.
var (
	ResourceContributorInsights         = resourceContributorInsights
	ResourceGlobalTable                 = resourceGlobalTable
	ResourceKinesisStreamingDestination = resourceKinesisStreamingDestination
	ResourceTable                       = resourceTable
	ResourceTableExport                 = resourceTableExport
	ResourceTableItem                   = resourceTableItem
	ResourceTableReplica                = resourceTableReplica
	ResourceTag                         = resourceTag
	ResourceResourcePolicy              = newResourcePolicyResource

	ARNForNewRegion                              = arnForNewRegion
	ContributorInsightsParseResourceID           = contributorInsightsParseResourceID
	ExpandTableItemAttributes                    = expandTableItemAttributes
	ExpandTableItemQueryKey                      = expandTableItemQueryKey
	FindContributorInsightsByTwoPartKey          = findContributorInsightsByTwoPartKey
	FindGlobalTableByName                        = findGlobalTableByName
	FindKinesisDataStreamDestinationByTwoPartKey = findKinesisDataStreamDestinationByTwoPartKey
	FindResourcePolicyByARN                      = findResourcePolicyByARN
	FindTableByName                              = findTableByName
	FindTableExportByARN                         = findTableExportByARN
	FindTableItemByTwoPartKey                    = findTableItemByTwoPartKey
	FindTag                                      = findTag
	FlattenTableItemAttributes                   = flattenTableItemAttributes
	ListTags                                     = listTags
	RegionFromARN                                = regionFromARN
	ReplicaForRegion                             = replicaForRegion
	TableNameFromARN                             = tableNameFromARN
	TableReplicaParseResourceID                  = tableReplicaParseResourceID
	UpdateDiffGSI                                = updateDiffGSI
)
