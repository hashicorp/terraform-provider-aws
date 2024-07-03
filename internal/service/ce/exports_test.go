// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ce

// Exports for use in tests only.
var (
	ResourceAnomalyMonitor      = resourceAnomalyMonitor      // nosemgrep:ci.ce-in-var-name
	ResourceAnomalySubscription = resourceAnomalySubscription // nosemgrep:ci.ce-in-var-name
	ResourceCostAllocationTag   = resourceCostAllocationTag   // nosemgrep:ci.ce-in-var-name
	ResourceCostCategory        = resourceCostCategory        // nosemgrep:ci.ce-in-var-name

	FindAnomalyMonitorByARN       = findAnomalyMonitorByARN
	FindAnomalySubscriptionByARN  = findAnomalySubscriptionByARN
	FindCostAllocationTagByTagKey = findCostAllocationTagByTagKey
	FindCostCategoryByARN         = findCostCategoryByARN
)
