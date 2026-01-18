// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package notifications

// Exports for use in tests only.
var (
	ResourceChannelAssociation                           = newChannelAssociationResource
	ResourceEventRule                                    = newEventRuleResource
	ResourceManagedNotificationAccountContactAssociation = newManagedNotificationAccountContactAssociationResource
	ResourceNotificationConfiguration                    = newNotificationConfigurationResource
	ResourceNotificationHub                              = newNotificationHubResource

	FindChannelAssociationByTwoPartKey                           = findChannelAssociationByTwoPartKey
	FindEventRuleByARN                                           = findEventRuleByARN
	FindManagedNotificationAccountContactAssociationByTwoPartKey = findManagedNotificationAccountContactAssociationByTwoPartKey
	FindNotificationConfigurationByARN                           = findNotificationConfigurationByARN
	FindNotificationHubByRegion                                  = findNotificationHubByRegion
)
