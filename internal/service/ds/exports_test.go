// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ds

// Exports for use in tests only.
var (
	ResourceConditionalForwarder    = resourceConditionalForwarder
	ResourceDirectory               = resourceDirectory
	ResourceIPRoute                 = newIPRouteResource
	ResourceLogSubscription         = resourceLogSubscription
	ResourceRadiusSettings          = resourceRadiusSettings
	ResourceRegion                  = resourceRegion
	ResourceSharedDirectory         = resourceSharedDirectory
	ResourceSharedDirectoryAccepter = resourceSharedDirectoryAccepter
	ResourceTrust                   = newTrustResource

	FindConditionalForwarderByTwoPartKey = findConditionalForwarderByTwoPartKey
	FindDirectoryByID                    = findDirectoryByID
	FindIPRouteByTwoPartKey              = findIPRouteByTwoPartKey
	FindIPRoutesByDirectoryID            = findIPRoutesByDirectoryID
	WaitIPRoutesAdded                    = waitIPRoutesAdded
	WaitIPRoutesRemoved                  = waitIPRoutesRemoved
	FindLogSubscriptionByID              = findLogSubscriptionByID
	FindRadiusSettingsByID               = findRadiusSettingsByID
	FindRegionByTwoPartKey               = findRegionByTwoPartKey
	FindSharedDirectoryByTwoPartKey      = findSharedDirectoryByTwoPartKey // nosemgrep:ci.ds-in-var-name
	FindTrustByTwoPartKey                = findTrustByTwoPartKey
)

// Type aliases for use in tests only.
type (
	IPRouteImportID = ipRouteImportID
)
