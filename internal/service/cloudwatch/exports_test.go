// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudwatch

// Exports for use in tests only.
var (
	ResourceDashboard    = resourceDashboard
	ResourceMetricAlarm  = resourceMetricAlarm
	ResourceMetricStream = resourceMetricStream

	FindDashboardByName    = findDashboardByName
	FindMetricAlarmByName  = findMetricAlarmByName
	FindMetricStreamByName = findMetricStreamByName
)
