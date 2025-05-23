// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package notifications

// Exports for use in tests only.
var (
	ResourceNotificationHub           = newNotificationHubResource
	ResourceNotificationConfiguration = newResourceNotificationConfiguration
	ResourceChannelAssociation        = newResourceChannelAssociation
	ResourceEventRule                 = newResourceEventRule

	FindNotificationHubByRegion        = findNotificationHubByRegion
	FindNotificationConfigurationByARN = findNotificationConfigurationByARN
	FindChannelAssociationByARNs       = findChannelAssociationByARNs
	FindEventRuleByARN                 = findEventRuleByARN
)
