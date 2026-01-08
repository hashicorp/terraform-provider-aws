// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkmonitor

// Exports for use in tests only.
var (
	ResourceMonitor = newMonitorResource
	ResourceProbe   = newProbeResource

	FindMonitorByName     = findMonitorByName
	FindProbeByTwoPartKey = findProbeByTwoPartKey
)
