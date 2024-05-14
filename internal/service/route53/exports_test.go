// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

// Exports for use in tests only.
var (
	ResourceCIDRCollection = newResourceCIDRCollection
	ResourceCIDRLocation   = newResourceCIDRLocation
	ResourceKeySigningKey  = resourceKeySigningKey

	CIDRLocationParseResourceID   = cidrLocationParseResourceID
	FindCIDRCollectionByID        = findCIDRCollectionByID
	FindCIDRLocationByTwoPartKey  = findCIDRLocationByTwoPartKey
	FindKeySigningKeyByTwoPartKey = findKeySigningKeyByTwoPartKey
	KeySigningKeyStatusActive     = keySigningKeyStatusActive
	KeySigningKeyStatusInactive   = keySigningKeyStatusInactive
)
