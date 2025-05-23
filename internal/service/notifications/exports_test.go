// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package notifications

// Exports for use in tests only.
var (
	ResourceChannelAssociation        = newChannelAssociationResource
	ResourceNotificationConfiguration = newNotificationConfigurationResource
	ResourceNotificationHub           = newNotificationHubResource
	ResourceEventRule                 = newResourceEventRule

	FindChannelAssociationByTwoPartKey = findChannelAssociationByTwoPartKey
	FindNotificationConfigurationByARN = findNotificationConfigurationByARN
	FindNotificationHubByRegion        = findNotificationHubByRegion
	FindEventRuleByARN                 = findEventRuleByARN
)
