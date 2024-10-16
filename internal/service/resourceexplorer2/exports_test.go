// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resourceexplorer2

// Exports for use in tests only.
var (
	ResourceIndex = newIndexResource
	ResourceView  = newResourceView

	FindIndex     = findIndex
	FindViewByARN = findViewByARN
)
