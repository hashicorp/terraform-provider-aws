// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package appfabric

// Exports for use in tests only.
var (
	ResourceAppAuthorization           = newAppAuthorizationResource
	ResourceAppAuthorizationConnection = newAppAuthorizationConnectionResource
	ResourceAppBundle                  = newAppBundleResource
	ResourceIngestion                  = newIngestionResource
	ResourceIngestionDestination       = newIngestionDestinationResource

	FindAppAuthorizationByTwoPartKey           = findAppAuthorizationByTwoPartKey
	FindAppAuthorizationConnectionByTwoPartKey = findAppAuthorizationConnectionByTwoPartKey
	FindAppBundleByID                          = findAppBundleByID
	FindIngestionByTwoPartKey                  = findIngestionByTwoPartKey
	FindIngestionDestinationByThreePartKey     = findIngestionDestinationByThreePartKey
)
