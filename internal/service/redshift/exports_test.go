// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package redshift

// Exports for use in tests only.
var (
	ResourceAuthenticationProfile        = resourceAuthenticationProfile
	ResourceCluster                      = resourceCluster
	ResourceClusterIAMRoles              = resourceClusterIAMRoles
	ResourceClusterSnapshot              = resourceClusterSnapshot
	ResourceDataShareAuthorization       = newDataShareAuthorizationResource
	ResourceDataShareConsumerAssociation = newDataShareConsumerAssociationResource
	ResourceEndpointAccess               = resourceEndpointAccess
	ResourceEndpointAuthorization        = resourceEndpointAuthorization
	ResourceEventSubscription            = resourceEventSubscription
	ResourceHSMClientCertificate         = resourceHSMClientCertificate
	ResourceHSMConfiguration             = resourceHSMConfiguration
	ResourceIdcApplication               = resourceIdcApplication
	ResourceIntegration                  = newIntegrationResource
	ResourceLogging                      = newLoggingResource
	ResourceParameterGroup               = resourceParameterGroup
	ResourcePartner                      = resourcePartner
	ResourceResourcePolicy               = resourceResourcePolicy
	ResourceScheduledAction              = resourceScheduledAction
	ResourceSnapshotCopy                 = newSnapshotCopyResource
	ResourceSnapshotCopyGrant            = resourceSnapshotCopyGrant
	ResourceSnapshotSchedule             = resourceSnapshotSchedule
	ResourceSnapshotScheduleAssociation  = resourceSnapshotScheduleAssociation
	ResourceSubnetGroup                  = resourceSubnetGroup
	ResourceUsageLimit                   = resourceUsageLimit

	FindAuthenticationProfileByID                 = findAuthenticationProfileByID
	FindClusterByID                               = findClusterByID
	FindClusterSnapshotByID                       = findClusterSnapshotByID
	FindDataShareAuthorizationByTwoPartKey        = findDataShareAuthorizationByTwoPartKey
	FindDataShareConsumerAssociationByFourPartKey = findDataShareConsumerAssociationByFourPartKey
	FindEndpointAccessByName                      = findEndpointAccessByName
	FindEndpointAuthorizationByTwoPartKey         = findEndpointAuthorizationByTwoPartKey
	FindEventSubscriptionByName                   = findEventSubscriptionByName
	FindHSMClientCertificateByID                  = findHSMClientCertificateByID
	FindHSMConfigurationByID                      = findHSMConfigurationByID
	FindIDCApplicationByArn                       = findIDCApplicationByArn
	FindIntegrationByARN                          = findIntegrationByARN
	FindLoggingStatusByID                         = findLoggingStatusByID
	FindParameterGroupByName                      = findParameterGroupByName
	FindPartnerByFourPartKey                      = findPartnerByFourPartKey
	FindResourcePolicyByARN                       = findResourcePolicyByARN
	FindScheduledActionByName                     = findScheduledActionByName
	FindSnapshotCopyByID                          = findSnapshotCopyByID
	FindSnapshotCopyGrantByName                   = findSnapshotCopyGrantByName
	FindSnapshotScheduleAssociationByTwoPartKey   = findSnapshotScheduleAssociationByTwoPartKey
	FindSnapshotScheduleByID                      = findSnapshotScheduleByID
	FindSubnetGroupByName                         = findSubnetGroupByName
	FindUsageLimitByID                            = findUsageLimitByID

	WaitSnapshotScheduleAssociationCreated = waitSnapshotScheduleAssociationCreated
)
