// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafka

// Exports for use in tests only.
var (
	ResourceCluster                      = resourceCluster
	ResourceClusterPolicy                = resourceClusterPolicy
	ResourceConfiguration                = resourceConfiguration
	ResourceReplicator                   = resourceReplicator
	ResourceSCRAMSecretAssociation       = resourceSCRAMSecretAssociation
	ResourceSingleSCRAMSecretAssociation = newSingleSCRAMSecretAssociationResource
	ResourceServerlessCluster            = resourceServerlessCluster
	ResourceVPCConnection                = resourceVPCConnection

	FindClusterByARN                             = findClusterByARN
	FindClusterPolicyByARN                       = findClusterPolicyByARN
	FindConfigurationByARN                       = findConfigurationByARN
	FindReplicatorByARN                          = findReplicatorByARN
	FindSCRAMSecretAssociation                   = findSCRAMSecretAssociation
	FindSingleSCRAMSecretAssociationByTwoPartKey = findSingleSCRAMSecretAssociationByTwoPartKey
	FindServerlessClusterByARN                   = findServerlessClusterByARN
	FindVPCConnectionByARN                       = findVPCConnectionByARN
)
