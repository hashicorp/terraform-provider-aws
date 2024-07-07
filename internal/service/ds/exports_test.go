// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ds

// Exports for use in tests only.
var (
	ResourceConditionalForwarder = resourceConditionalForwarder
	ResourceDirectory            = resourceDirectory
	ResourceLogSubscription      = resourceLogSubscription
	ResourceRadiusSettings       = resourceRadiusSettings
	ResourceSharedDirectory      = resourceSharedDirectory

	FindConditionalForwarderByTwoPartKey = findConditionalForwarderByTwoPartKey
	FindDirectoryByID                    = findDirectoryByID
	FindLogSubscriptionByID              = findLogSubscriptionByID
	FindRadiusSettingsByID               = findRadiusSettingsByID
	FindSharedDirectoryByTwoPartKey      = findSharedDirectoryByTwoPartKey
)
