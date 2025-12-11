// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package devopsguru

// Exports for use in tests only.
var (
	ResourceEventSourcesConfig  = newEventSourcesConfigResource
	ResourceNotificationChannel = newNotificationChannelResource
	ResourceResourceCollection  = newResourceCollectionResource

	FindEventSourcesConfig      = findEventSourcesConfig
	FindNotificationChannelByID = findNotificationChannelByID
	FindResourceCollectionByID  = findResourceCollectionByID
	FindServiceIntegration      = findServiceIntegration
)
