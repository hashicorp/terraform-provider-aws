// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package notifications

// Exports for use in tests only.
var (
	ResourceNotificationConfiguration = newNotificationConfigurationResource
	ResourceNotificationHub           = newNotificationHubResource
	ResourceChannelAssociation        = newResourceChannelAssociation
	ResourceEventRule                 = newResourceEventRule

	FindNotificationConfigurationByARN = findNotificationConfigurationByARN
	FindNotificationHubByRegion        = findNotificationHubByRegion
	FindChannelAssociationByARNs       = findChannelAssociationByARNs
	FindEventRuleByARN                 = findEventRuleByARN
)
