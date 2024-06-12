// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appfabric

// Exports for use in tests only.
var (
	ResourceAppAuthorization = newAppAuthorizationResource
	ResourceAppBundle        = newAppBundleResource
	ResourceIngestion        = newIngestionResource

	FindAppAuthorizationByTwoPartKey = findAppAuthorizationByTwoPartKey
	FindAppBundleByID                = findAppBundleByID
	FindIngestionByTwoPartKey        = findIngestionByTwoPartKey
)
