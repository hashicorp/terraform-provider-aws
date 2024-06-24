// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appfabric

// Exports for use in tests only.
var (
	ResourceAppAuthorization           = newAppAuthorizationResource
	ResourceAppAuthorizationConnection = newAppAuthorizationConnectionResource
	ResourceAppBundle                  = newAppBundleResource
	ResourceIngestion                  = newIngestionResource
	ResourceIngestionDestination       = newResourceIngestionDestination

	FindAppAuthorizationByTwoPartKey           = findAppAuthorizationByTwoPartKey
	FindAppAuthorizationConnectionByTwoPartKey = findAppAuthorizationConnectionByTwoPartKey
	FindAppBundleByID                          = findAppBundleByID
	FindIngestionByTwoPartKey                  = findIngestionByTwoPartKey
	FindIngestionDestinationByID               = findIngestionDestinationByID
)
