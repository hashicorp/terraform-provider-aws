// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package notifications

// Exports for use in tests only.
var (
	ResourceChannelAssociation        = newChannelAssociationResource
	ResourceEventRule                 = newEventRuleResource
	ResourceNotificationConfiguration = newNotificationConfigurationResource
	ResourceNotificationHub           = newNotificationHubResource
	ResourceOrganizationsAccess       = newOrganizationsAccessResource

	FindChannelAssociationByTwoPartKey = findChannelAssociationByTwoPartKey
	FindEventRuleByARN                 = findEventRuleByARN
	FindNotificationConfigurationByARN = findNotificationConfigurationByARN
	FindNotificationHubByRegion        = findNotificationHubByRegion
	WaitOrganizationsAccessDisabled    = waitOrganizationsAccessDisabled
	WaitOrganizationsAccessEnabled     = waitOrganizationsAccessEnabled
	WaitOrganizationsAccessStable      = waitOrganizationsAccessStable

	OrganizationsAccessStableTimeout = organizationsAccessStableTimeout
)
