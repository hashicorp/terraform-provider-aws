// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
)

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
	ResourceGlobalSecondaryIndex        = newResourceGlobalSecondaryIndex

	ARNForNewRegion                              = arnForNewRegion
	ContributorInsightsParseResourceID           = contributorInsightsParseResourceID
	ExpandTableItemAttributes                    = expandTableItemAttributes
	ExpandTableItemQueryKey                      = expandTableItemQueryKey
	FindContributorInsightsByTwoPartKey          = findContributorInsightsByTwoPartKey
	FindGlobalTableByName                        = findGlobalTableByName
	FindGSIByTwoPartKey                          = findGSIByTwoPartKey
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

const (
	GlobalSecondaryIndexExperimentalFlagEnvVar = globalSecondaryIndexExperimentalFlagEnvVar
)

func TestGSILegacyKeySchemaEquivalent(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		state     cty.Value
		plan      cty.Value
		wantEqual bool
	}{
		{
			name: "matching hash and range keys",
			state: cty.ObjectVal(map[string]cty.Value{
				"hash_key":  cty.StringVal("hk"),
				"range_key": cty.StringVal("rk"),
			}),
			plan: cty.ObjectVal(map[string]cty.Value{
				"key_schema": cty.ListVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"attribute_name": cty.StringVal("hk"),
						"key_type":       cty.StringVal("HASH"),
					}),
					cty.ObjectVal(map[string]cty.Value{
						"attribute_name": cty.StringVal("rk"),
						"key_type":       cty.StringVal("RANGE"),
					}),
				}),
			}),
			wantEqual: true,
		},
		{
			name: "mismatching keys",
			state: cty.ObjectVal(map[string]cty.Value{
				"hash_key":  cty.StringVal("hk"),
				"range_key": cty.StringVal("rk"),
			}),
			plan: cty.ObjectVal(map[string]cty.Value{
				"key_schema": cty.ListVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"attribute_name": cty.StringVal("different"),
						"key_type":       cty.StringVal("HASH"),
					}),
					cty.ObjectVal(map[string]cty.Value{
						"attribute_name": cty.StringVal("rk"),
						"key_type":       cty.StringVal("RANGE"),
					}),
				}),
			}),
			wantEqual: false,
		},
		{
			name: "unknown key_schema value does not imply equivalence",
			state: cty.ObjectVal(map[string]cty.Value{
				"hash_key":  cty.StringVal("hk"),
				"range_key": cty.StringVal("rk"),
			}),
			plan: cty.ObjectVal(map[string]cty.Value{
				"key_schema": cty.ListVal([]cty.Value{
					cty.ObjectVal(map[string]cty.Value{
						"attribute_name": cty.UnknownVal(cty.String),
						"key_type":       cty.StringVal("HASH"),
					}),
					cty.ObjectVal(map[string]cty.Value{
						"attribute_name": cty.StringVal("rk"),
						"key_type":       cty.StringVal("RANGE"),
					}),
				}),
			}),
			wantEqual: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := gsiLegacyKeySchemaEquivalent(testCase.state, testCase.plan)

			if got != testCase.wantEqual {
				t.Fatalf("gsiLegacyKeySchemaEquivalent() = %t, want %t", got, testCase.wantEqual)
			}
		})
	}
}
