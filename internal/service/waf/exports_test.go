// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waf

// Exports for use in tests only.
var (
	ResourceByteMatchSet = resourceByteMatchSet
	ResourceGeoMatchSet  = resourceGeoMatchSet
	ResourceWebACL       = resourceWebACL

	FindByteMatchSetByID = findByteMatchSetByID
	FindGeoMatchSetByID  = findGeoMatchSetByID
	FindWebACLByID       = findWebACLByID
	FlattenFieldToMatch  = flattenFieldToMatch
)
