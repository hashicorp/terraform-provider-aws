// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package notifications

// Exports for use in tests only.
var (
	ResourceNotificationConfiguration = newResourceNotificationConfiguration
	ResourceChannelAssociation        = newResourceChannelAssociation
	ResourceEventRule                 = newResourceEventRule

	FindNotificationConfigurationByARN = findNotificationConfigurationByARN
	FindChannelAssociationByARNs       = findChannelAssociationByARNs
	FindEventRuleByARN                 = findEventRuleByARN
)
