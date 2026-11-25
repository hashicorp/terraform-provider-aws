// Copyright IBM Corp. 2014, 2026
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
	ResourceRecordSetProtection         = newRecordSetProtectionResource
	ResourceTrafficPolicy               = resourceTrafficPolicy
	ResourceTrafficPolicyInstance       = resourceTrafficPolicyInstance
	ResourceVPCAssociationAuthorization = resourceVPCAssociationAuthorization
	ResourceZone                        = resourceZone
	ResourceZoneAssociation             = resourceZoneAssociation

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
	FindRecordSetProtectionByThreePartKey       = findRecordSetProtectionByThreePartKey
	IsProtectionNotFound                        = isProtectionNotFound
	ProtectionBatchError                        = protectionBatchError
	FindResourceRecordSetByFourPartKey          = findResourceRecordSetByFourPartKey
	FindResourceRecordSetsForHostedZone         = findResourceRecordSetsForHostedZone
	FindTrafficPolicyByID                       = findTrafficPolicyByID
	FindTrafficPolicyInstanceByID               = findTrafficPolicyInstanceByID
	FindVPCAssociationAuthorizationByTwoPartKey = findVPCAssociationAuthorizationByTwoPartKey
	FindZoneAssociationByThreePartKey           = findZoneAssociationByThreePartKey
	KeySigningKeyStatusActive                   = keySigningKeyStatusActive
	KeySigningKeyStatusInactive                 = keySigningKeyStatusInactive
	RecordBatchReadEnvVar                       = recordBatchReadEnvVar
	ServeSignatureNotSigning                    = serveSignatureNotSigning
	ServeSignatureSigning                       = serveSignatureSigning
	WaitChangeInsync                            = waitChangeInsync
)

type (
	RecordSetProtectionImportID = recordSetProtectionImportID
	Route53TrafficPolicyDoc     = route53TrafficPolicyDoc
)
