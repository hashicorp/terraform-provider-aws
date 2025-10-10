// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ds

// Exports for use in tests only.
var (
	ResourceConditionalForwarder    = resourceConditionalForwarder
	ResourceDirectory               = resourceDirectory
	ResourceLogSubscription         = resourceLogSubscription
	ResourceRadiusSettings          = resourceRadiusSettings
	ResourceRegion                  = resourceRegion
	ResourceSharedDirectory         = resourceSharedDirectory
	ResourceSharedDirectoryAccepter = resourceSharedDirectoryAccepter
	ResourceTrust                   = newTrustResource

	FindConditionalForwarderByTwoPartKey = findConditionalForwarderByTwoPartKey
	FindDirectoryByID                    = findDirectoryByID
	FindLogSubscriptionByID              = findLogSubscriptionByID
	FindRadiusSettingsByID               = findRadiusSettingsByID
	FindRegionByTwoPartKey               = findRegionByTwoPartKey
	FindSharedDirectoryByTwoPartKey      = findSharedDirectoryByTwoPartKey // nosemgrep:ci.ds-in-var-name
	FindTrustByTwoPartKey                = findTrustByTwoPartKey
)
