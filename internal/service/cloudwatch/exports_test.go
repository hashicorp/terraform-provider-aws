// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cloudwatch

// Exports for use in tests only.
var (
	ResourceAlarmMuteRule                 = newAlarmMuteRuleResource
	ResourceCompositeAlarm                = resourceCompositeAlarm
	ResourceContributorInsightRule        = newContributorInsightRuleResource
	ResourceContributorManagedInsightRule = newContributorManagedInsightRuleResource
	ResourceDashboard                     = resourceDashboard
	ResourceMetricAlarm                   = resourceMetricAlarm
	ResourceMetricStream                  = resourceMetricStream
	ResourceOtelEnrichment                = newOTelEnrichmentResource

	FindAlarmMuteRuleByName                                    = findAlarmMuteRuleByName
	FindCompositeAlarmByName                                   = findCompositeAlarmByName
	FindContributorInsightRuleByName                           = findContributorInsightRuleByName
	FindContributorManagedInsightRuleDescriptionByTemplateName = findContributorManagedInsightRuleDescriptionByTemplateName
	FindDashboardByName                                        = findDashboardByName
	FindMetricAlarmByName                                      = findMetricAlarmByName
	FindMetricStreamByName                                     = findMetricStreamByName
	FindOtelEnrichment                                         = findOTelEnrichment
)
