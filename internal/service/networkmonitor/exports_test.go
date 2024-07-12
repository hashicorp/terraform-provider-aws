// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmonitor

// Exports for use in tests only.
var (
	ResourceMonitor = newMonitorResource
	ResourceProbe   = newProbeResource

	FindMonitorByName     = findMonitorByName
	FindProbeByTwoPartKey = findProbeByTwoPartKey
)
