// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ce

// Exports for use in tests only.
var (
	ResourceAnomalyMonitor      = resourceAnomalyMonitor
	ResourceAnomalySubscription = resourceAnomalySubscription

	FindAnomalyMonitorByARN      = findAnomalyMonitorByARN
	FindAnomalySubscriptionByARN = findAnomalySubscriptionByARN
)
