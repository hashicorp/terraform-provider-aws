// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package notifications

// Exports for use in tests only.
var (
	ResourceChannelAssociation                              = newChannelAssociationResource
	ResourceEventRule                                       = newEventRuleResource
	ResourceManagedNotificationAccountContactAssociation    = newManagedNotificationAccountContactAssociationResource
	ResourceManagedNotificationAdditionalChannelAssociation = newManagedNotificationAdditionalChannelAssociationResource
	ResourceNotificationConfiguration                       = newNotificationConfigurationResource
	ResourceNotificationHub                                 = newNotificationHubResource

	FindAccessForOrganization                                       = findAccessForOrganization
	FindChannelAssociationByTwoPartKey                              = findChannelAssociationByTwoPartKey
	FindEventRuleByARN                                              = findEventRuleByARN
	FindManagedNotificationAccountContactAssociationByTwoPartKey    = findManagedNotificationAccountContactAssociationByTwoPartKey
	FindManagedNotificationAdditionalChannelAssociationByTwoPartKey = findManagedNotificationAdditionalChannelAssociationByTwoPartKey
	FindNotificationConfigurationByARN                              = findNotificationConfigurationByARN
	FindNotificationHubByRegion                                     = findNotificationHubByRegion
)
