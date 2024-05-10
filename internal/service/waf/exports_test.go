// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waf

// Exports for use in tests only.
var (
	ResourceByteMatchSet = resourceByteMatchSet
	ResourceGeoMatchSet  = resourceGeoMatchSet
	ResourceIPSet        = resourceIPSet
	ResourceWebACL       = resourceWebACL

	DiffIPSetDescriptors = diffIPSetDescriptors
	FindByteMatchSetByID = findByteMatchSetByID
	FindGeoMatchSetByID  = findGeoMatchSetByID
	FindIPSetByID        = findIPSetByID
	FindWebACLByID       = findWebACLByID
	FlattenFieldToMatch  = flattenFieldToMatch
)
