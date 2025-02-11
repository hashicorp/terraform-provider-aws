// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rum

// Exports for use in tests only.
var (
	ResourceAppMonitor         = resourceAppMonitor
	ResourceMetricsDestination = resourceMetricsDestination

	FindAppMonitorByName         = findAppMonitorByName
	FindMetricsDestinationByName = findMetricsDestinationByName
)
