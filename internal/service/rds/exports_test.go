// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

// Exports for use in tests only.
var (
	ResourceActivityStream                      = resourceActivityStream
	ResourceCertificate                         = resourceCertificate
	ResourceCluster                             = resourceCluster
	ResourceClusterActivityStream               = resourceClusterActivityStream
	ResourceClusterEndpoint                     = resourceClusterEndpoint
	ResourceClusterInstance                     = resourceClusterInstance
	ResourceClusterParameterGroup               = resourceClusterParameterGroup
	ResourceClusterRoleAssociation              = resourceClusterRoleAssociation
	ResourceClusterSnapshot                     = resourceClusterSnapshot
	ResourceCustomDBEngineVersion               = resourceCustomDBEngineVersion
	ResourceEventSubscription                   = resourceEventSubscription
	ResourceGlobalCluster                       = resourceGlobalCluster
	ResourceInstance                            = resourceInstance
	ResourceInstanceAutomatedBackupsReplication = resourceInstanceAutomatedBackupsReplication
	ResourceInstanceRoleAssociation             = resourceInstanceRoleAssociation
	ResourceIntegration                         = newIntegrationResource
	ResourceOptionGroup                         = resourceOptionGroup
	ResourceParameterGroup                      = resourceParameterGroup
	ResourceProxy                               = resourceProxy
	ResourceProxyDefaultTargetGroup             = resourceProxyDefaultTargetGroup
	ResourceProxyEndpoint                       = resourceProxyEndpoint
	ResourceProxyTarget                         = resourceProxyTarget
	ResourceReservedInstance                    = resourceReservedInstance
	ResourceSnapshot                            = resourceSnapshot
	ResourceSnapshotCopy                        = resourceSnapshotCopy
	ResourceSubnetGroup                         = resourceSubnetGroup

	ClusterIDAndRegionFromARN                  = clusterIDAndRegionFromARN
	FindCustomDBEngineVersionByTwoPartKey      = findCustomDBEngineVersionByTwoPartKey
	FindDBClusterByID                          = findDBClusterByID
	FindDBClusterEndpointByID                  = findDBClusterEndpointByID
	FindDBClusterParameterGroupByName          = findDBClusterParameterGroupByName
	FindDBClusterRoleByTwoPartKey              = findDBClusterRoleByTwoPartKey
	FindDBClusterSnapshotByID                  = findDBClusterSnapshotByID
	FindDBClusterWithActivityStream            = findDBClusterWithActivityStream
	FindDBInstanceAutomatedBackupByARN         = findDBInstanceAutomatedBackupByARN
	FindDBInstanceByID                         = findDBInstanceByID
	FindDBInstanceRoleByTwoPartKey             = findDBInstanceRoleByTwoPartKey
	FindDBInstanceWithActivityStream           = findDBInstanceWithActivityStream
	FindDBParameterGroupByName                 = findDBParameterGroupByName
	FindDBProxyByName                          = findDBProxyByName
	FindDBProxyEndpointByTwoPartKey            = findDBProxyEndpointByTwoPartKey
	FindDBProxyTargetByFourPartKey             = findDBProxyTargetByFourPartKey
	FindDBSnapshotByID                         = findDBSnapshotByID
	FindDBSubnetGroupByName                    = findDBSubnetGroupByName
	FindDefaultCertificate                     = findDefaultCertificate
	FindDefaultDBProxyTargetGroupByDBProxyName = findDefaultDBProxyTargetGroupByDBProxyName
	FindEventSubscriptionByID                  = findEventSubscriptionByID
	FindGlobalClusterByID                      = findGlobalClusterByID
	FindIntegrationByARN                       = findIntegrationByARN
	FindOptionGroupByName                      = findOptionGroupByName
	FindReservedDBInstanceByID                 = findReservedDBInstanceByID
	ListTags                                   = listTags
	NewBlueGreenOrchestrator                   = newBlueGreenOrchestrator
	ParameterGroupModifyChunk                  = parameterGroupModifyChunk
	ParseDBInstanceARN                         = parseDBInstanceARN
	ProxyTargetParseResourceID                 = proxyTargetParseResourceID
	WaitBlueGreenDeploymentDeleted             = waitBlueGreenDeploymentDeleted
	WaitBlueGreenDeploymentAvailable           = waitBlueGreenDeploymentAvailable
	WaitDBInstanceAvailable                    = waitDBInstanceAvailable
	WaitDBInstanceDeleted                      = waitDBInstanceDeleted

	ErrCodeInvalidAction               = errCodeInvalidAction
	ErrCodeInvalidParameterCombination = errCodeInvalidParameterCombination
	ErrCodeInvalidParameterValue       = errCodeInvalidParameterValue
)
