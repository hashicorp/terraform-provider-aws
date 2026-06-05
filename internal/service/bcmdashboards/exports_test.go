// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bcmdashboards

// Exports for use in tests only.
var (
	ResourceDashboard       = newDashboardResource
	ResourceScheduledReport = newScheduledReportResource

	FindDashboardByARN       = findDashboardByARN
	FindScheduledReportByARN = findScheduledReportByARN
)
