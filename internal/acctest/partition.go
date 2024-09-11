// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acctest

import (
	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func IsIsolatedRegion(region string) bool {
	partition := names.PartitionForRegion(region)

	return IsIsolatedPartition(partition)
}

func IsIsolatedPartition(partition string) bool {
	pattern := `^aws-iso-?.*$`

	re := regexache.MustCompile(pattern)

	return re.MatchString(partition)
}

func IsStandardRegion(region string) bool {
	partition := names.PartitionForRegion(region)

	return IsStandardPartition(partition)
}

func IsStandardPartition(partitionId string) bool {
	return partitionId == names.StandardPartitionID
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
