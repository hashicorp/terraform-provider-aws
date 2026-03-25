// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resourceexplorer2

// Exports for use in tests only.
var (
	ResourceIndex = newIndexResource
	ResourceView  = newViewResource

	FindIndex     = findIndex
	FindViewByARN = findViewByARN
)
