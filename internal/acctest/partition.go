// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acctest

import (
	"github.com/YakDriver/regexache"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func IsIsolatedRegion(region string) bool {
	partition := names.PartitionForRegion(region)

	return IsIsolatedPartition(partition.ID())
}

func IsIsolatedPartition(partition string) bool {
	pattern := `^aws-iso-?.*$`

	re := regexache.MustCompile(pattern)

	return re.MatchString(partition)
}

func IsStandardRegion(region string) bool {
	partition := names.PartitionForRegion(region)

	return IsStandardPartition(partition.ID())
}

func IsStandardPartition(partitionID string) bool {
	return partitionID == endpoints.AwsPartitionID
}

func RegionsInPartition(partitionName string) []string {
	var regions []string
	for _, partition := range endpoints.DefaultPartitions() {
		if partition.ID() == partitionName {
			for _, region := range partition.Regions() {
				regions = append(regions, region.ID())
			}
		}
	}

	return regions
}
