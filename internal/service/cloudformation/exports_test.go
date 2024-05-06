// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudformation

// Exports for use in tests only.
var (
	ResourceStack    = resourceStack
	ResourceStackSet = resourceStackSet
	ResourceType     = resourceType

	FindStackSetByName = findStackSetByName
	FindTypeByARN      = findTypeByARN
)
