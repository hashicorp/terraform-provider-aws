// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

// Exports for use in tests only.
var (
	ResourceCIDRCollection        = newCIDRCollectionResource
	ResourceCIDRLocation          = newCIDRLocationResource
	ResourceDelegationSet         = resourceDelegationSet
	ResourceHealthCheck           = resourceHealthCheck
	ResourceHostedZoneDNSSEC      = resourceHostedZoneDNSSEC
	ResourceKeySigningKey         = resourceKeySigningKey
	ResourceQueryLog              = resourceQueryLog
	ResourceTrafficPolicy         = resourceTrafficPolicy
	ResourceTrafficPolicyInstance = resourceTrafficPolicyInstance

	CleanDelegationSetID          = cleanDelegationSetID
	FindCIDRCollectionByID        = findCIDRCollectionByID
	FindCIDRLocationByTwoPartKey  = findCIDRLocationByTwoPartKey
	FindDelegationSetByID         = findDelegationSetByID
	FindHealthCheckByID           = findHealthCheckByID
	FindHostedZoneDNSSECByZoneID  = findHostedZoneDNSSECByZoneID
	FindKeySigningKeyByTwoPartKey = findKeySigningKeyByTwoPartKey
	FindQueryLoggingConfigByID    = findQueryLoggingConfigByID
	FindTrafficPolicyByID         = findTrafficPolicyByID
	FindTrafficPolicyInstanceByID = findTrafficPolicyInstanceByID
	KeySigningKeyStatusActive     = keySigningKeyStatusActive
	KeySigningKeyStatusInactive   = keySigningKeyStatusInactive
	ServeSignatureNotSigning      = serveSignatureNotSigning
	ServeSignatureSigning         = serveSignatureSigning
)

type Route53TrafficPolicyDoc = route53TrafficPolicyDoc
