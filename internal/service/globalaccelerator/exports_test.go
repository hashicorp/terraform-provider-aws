// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package globalaccelerator

// Exports for use in tests only.
var (
	ExpandResources                = expandResources
	FlattenResources               = flattenResources
	DiffResources                  = diffResources
	DiffPrincipals                 = diffPrincipals
	ResourceCrossAccountAttachment = newResourceCrossAccountAttachment
)
