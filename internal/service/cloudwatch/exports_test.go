// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudwatch

// Exports for use in tests only.
var (
	ResourceCompositeAlarm                = resourceCompositeAlarm
	ResourceDashboard                     = resourceDashboard
	ResourceMetricAlarm                   = resourceMetricAlarm
	ResourceMetricStream                  = resourceMetricStream
	ResourceContributorInsightRule        = newResourceContributorInsightRule
	ResourceContributorManagedInsightRule = newResourceContributorManagedInsightRule

	FindCompositeAlarmByName                                   = findCompositeAlarmByName
	FindDashboardByName                                        = findDashboardByName
	FindMetricAlarmByName                                      = findMetricAlarmByName
	FindMetricStreamByName                                     = findMetricStreamByName
	FindContributorInsightRuleByName                           = findContributorInsightRuleByName
	FindContributorManagedInsightRuleDescriptionByTemplateName = findContributorManagedInsightRuleDescriptionByTemplateName
)
