// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appfabric

// Exports for use in tests only.
var (
	ResourceAppAuthorization        = newAppAuthorizationResource
	ResourceConnectAppAuthorization = newAppAuthorizationConnectionResource
	ResourceAppBundle               = newAppBundleResource

	FindAppAuthorizationByTwoPartKey        = findAppAuthorizationByTwoPartKey
	FindConnectAppAuthorizationByTwoPartKey = findConnectAppAuthorizationByTwoPartKey
	FindAppBundleByID                       = findAppBundleByID
)
