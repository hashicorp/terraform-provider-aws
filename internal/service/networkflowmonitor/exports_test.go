// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package networkflowmonitor

// Exports for use in tests only.
var (
	ResourceMonitor = newMonitorResource
	ResourceScope   = newScopeResource

	FindMonitorByName = findMonitorByName
	FindScopeByID     = findScopeByID
)
