// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package devopsguru

// Exports for use in tests only.
var (
	ResourceEventSourcesConfig = newResourceEventSourcesConfig
	ResourceResourceCollection = newResourceResourceCollection

	FindEventSourcesConfig     = findEventSourcesConfig
	FindResourceCollectionByID = findResourceCollectionByID
)
