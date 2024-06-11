// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appfabric

// Exports for use in tests only.
var (
	ResourceAppBundle = newAppBundleResource
	ResourceIngestion = newIngestionResource

	FindAppBundleByID = findAppBundleByID
)
