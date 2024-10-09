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
	ResourceRecord                      = resourceRecord
	ResourceTrafficPolicy               = resourceTrafficPolicy
	ResourceTrafficPolicyInstance       = resourceTrafficPolicyInstance
	ResourceVPCAssociationAuthorization = resourceVPCAssociationAuthorization
	ResourceZone                        = resourceZone
	ResourceZoneAssociation             = resourceZoneAssociation

	CleanDelegationSetID                        = cleanDelegationSetID
	CleanRecordName                             = cleanRecordName
	CleanZoneID                                 = cleanZoneID
	ExpandRecordName                            = expandRecordName
	FindCIDRCollectionByID                      = findCIDRCollectionByID
	FindCIDRLocationByTwoPartKey                = findCIDRLocationByTwoPartKey
	FindDelegationSetByID                       = findDelegationSetByID
	FindHealthCheckByID                         = findHealthCheckByID
	FindHostedZoneByID                          = findHostedZoneByID
	FindHostedZoneDNSSECByZoneID                = findHostedZoneDNSSECByZoneID
	FindKeySigningKeyByTwoPartKey               = findKeySigningKeyByTwoPartKey
	FindQueryLoggingConfigByID                  = findQueryLoggingConfigByID
	FindResourceRecordSetByFourPartKey          = findResourceRecordSetByFourPartKey
	FindTrafficPolicyByID                       = findTrafficPolicyByID
	FindTrafficPolicyInstanceByID               = findTrafficPolicyInstanceByID
	FindVPCAssociationAuthorizationByTwoPartKey = findVPCAssociationAuthorizationByTwoPartKey
	FindZoneAssociationByThreePartKey           = findZoneAssociationByThreePartKey
	FQDN                                        = fqdn
	KeySigningKeyStatusActive                   = keySigningKeyStatusActive
	KeySigningKeyStatusInactive                 = keySigningKeyStatusInactive
	RecordParseResourceID                       = recordParseResourceID
	ServeSignatureNotSigning                    = serveSignatureNotSigning
	ServeSignatureSigning                       = serveSignatureSigning
	WaitChangeInsync                            = waitChangeInsync
)

type Route53TrafficPolicyDoc = route53TrafficPolicyDoc
