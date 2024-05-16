// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

// Exports for use in tests only.
var (
	ResourceCIDRCollection   = newResourceCIDRCollection
	ResourceCIDRLocation     = newResourceCIDRLocation
	ResourceHealthCheck      = resourceHealthCheck
	ResourceHostedZoneDNSSEC = resourceHostedZoneDNSSEC
	ResourceKeySigningKey    = resourceKeySigningKey

	CIDRLocationParseResourceID   = cidrLocationParseResourceID
	FindCIDRCollectionByID        = findCIDRCollectionByID
	FindCIDRLocationByTwoPartKey  = findCIDRLocationByTwoPartKey
	FindHealthCheckByID           = findHealthCheckByID
	FindHostedZoneDNSSECByZoneID  = findHostedZoneDNSSECByZoneID
	FindKeySigningKeyByTwoPartKey = findKeySigningKeyByTwoPartKey
	KeySigningKeyStatusActive     = keySigningKeyStatusActive
	KeySigningKeyStatusInactive   = keySigningKeyStatusInactive
	ServeSignatureNotSigning      = serveSignatureNotSigning
	ServeSignatureSigning         = serveSignatureSigning
)
