// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

// Exports for use in tests only.
var (
	ResourceCIDRCollection              = newCIDRCollectionResource
	ResourceCIDRLocation                = newCIDRLocationResource
	ResourceDelegationSet               = resourceDelegationSet
	ResourceHealthCheck                 = resourceHealthCheck
	ResourceHostedZoneDNSSEC            = resourceHostedZoneDNSSEC
	ResourceKeySigningKey               = resourceKeySigningKey
	ResourceQueryLog                    = resourceQueryLog
	ResourceTrafficPolicy               = resourceTrafficPolicy
	ResourceTrafficPolicyInstance       = resourceTrafficPolicyInstance
	ResourceVPCAssociationAuthorization = resourceVPCAssociationAuthorization
	ResourceZone                        = resourceZone
	ResourceZoneAssociation             = resourceZoneAssociation

	CleanDelegationSetID                        = cleanDelegationSetID
	FindCIDRCollectionByID                      = findCIDRCollectionByID
	FindCIDRLocationByTwoPartKey                = findCIDRLocationByTwoPartKey
	FindDelegationSetByID                       = findDelegationSetByID
	FindHealthCheckByID                         = findHealthCheckByID
	FindHostedZoneByID                          = findHostedZoneByID
	FindHostedZoneDNSSECByZoneID                = findHostedZoneDNSSECByZoneID
	FindKeySigningKeyByTwoPartKey               = findKeySigningKeyByTwoPartKey
	FindQueryLoggingConfigByID                  = findQueryLoggingConfigByID
	FindTrafficPolicyByID                       = findTrafficPolicyByID
	FindTrafficPolicyInstanceByID               = findTrafficPolicyInstanceByID
	FindVPCAssociationAuthorizationByTwoPartKey = findVPCAssociationAuthorizationByTwoPartKey
	FindZoneAssociationByThreePartKey           = findZoneAssociationByThreePartKey
	KeySigningKeyStatusActive                   = keySigningKeyStatusActive
	KeySigningKeyStatusInactive                 = keySigningKeyStatusInactive
	ServeSignatureNotSigning                    = serveSignatureNotSigning
	ServeSignatureSigning                       = serveSignatureSigning
	WaitChangeInsync                            = waitChangeInsync
)

type Route53TrafficPolicyDoc = route53TrafficPolicyDoc
