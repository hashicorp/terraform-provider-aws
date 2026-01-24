// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatch

// Exports for use in tests only.
var (
	ResourceCompositeAlarm                = resourceCompositeAlarm
	ResourceDashboard                     = resourceDashboard
	ResourceMetricAlarm                   = resourceMetricAlarm
	ResourceMetricStream                  = resourceMetricStream
	ResourceContributorInsightRule        = newContributorInsightRuleResource
	ResourceContributorManagedInsightRule = newContributorManagedInsightRuleResource

	FindCompositeAlarmByName                                   = findCompositeAlarmByName
	FindDashboardByName                                        = findDashboardByName
	FindMetricAlarmByName                                      = findMetricAlarmByName
	FindMetricStreamByName                                     = findMetricStreamByName
	FindContributorInsightRuleByName                           = findContributorInsightRuleByName
	FindContributorManagedInsightRuleDescriptionByTemplateName = findContributorManagedInsightRuleDescriptionByTemplateName
)
