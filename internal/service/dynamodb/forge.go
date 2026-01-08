// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import (
	"maps"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func stripGSIUpdatableAttributes(in map[string]any) map[string]any {
	attrs := stripCapacityAttributes(in)
	attrs = stripNonKeyAttributes(attrs)
	attrs = stripOnDemandThroughputAttributes(attrs)
	attrs = stripWarmThroughputAttributes(attrs)
	// Remove empty hash_keys/range_keys to avoid false positive diffs
	if hks, ok := attrs["hash_keys"]; ok {
		if l, ok := hks.([]any); ok && len(l) == 0 {
			delete(attrs, "hash_keys")
		}
	}
	if rks, ok := attrs["range_keys"]; ok {
		if set, ok := rks.(*schema.Set); ok && set.Len() == 0 {
			delete(attrs, "range_keys")
		}
	}

	return attrs
}

func stripCapacityAttributes(in map[string]any) map[string]any {
	mapCopy := maps.Clone(in)

	delete(mapCopy, "write_capacity")
	delete(mapCopy, "read_capacity")

	return mapCopy
}

func stripNonKeyAttributes(in map[string]any) map[string]any {
	mapCopy := maps.Clone(in)

	delete(mapCopy, "non_key_attributes")

	return mapCopy
}

func stripOnDemandThroughputAttributes(in map[string]any) map[string]any {
	mapCopy := maps.Clone(in)

	delete(mapCopy, "on_demand_throughput")

	return mapCopy
}

func stripWarmThroughputAttributes(in map[string]any) map[string]any {
	mapCopy := maps.Clone(in)

	delete(mapCopy, "warm_throughput")

	return mapCopy
}
