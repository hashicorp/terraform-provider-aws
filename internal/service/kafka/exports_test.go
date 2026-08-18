// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kafka

// Exports for use in tests only.
var (
	ResourceCluster                      = resourceCluster
	ResourceClusterPolicy                = resourceClusterPolicy
	ResourceConfiguration                = resourceConfiguration
	ResourceChannel                      = newChannelResource
	ResourceReplicator                   = resourceReplicator
	ResourceSCRAMSecretAssociation       = resourceSCRAMSecretAssociation
	ResourceSingleSCRAMSecretAssociation = newSingleSCRAMSecretAssociationResource
	ResourceServerlessCluster            = resourceServerlessCluster
	ResourceTopic                        = newTopicResource
	ResourceVPCConnection                = resourceVPCConnection

	FindChannelByTwoPartKey                      = findChannelByTwoPartKey
	FindClusterByARN                             = findClusterByARN
	FindClusterPolicyByARN                       = findClusterPolicyByARN
	FindConfigurationByARN                       = findConfigurationByARN
	FindReplicatorByARN                          = findReplicatorByARN
	FindSCRAMSecretAssociation                   = findSCRAMSecretAssociation
	FindSingleSCRAMSecretAssociationByTwoPartKey = findSingleSCRAMSecretAssociationByTwoPartKey
	FindServerlessClusterByARN                   = findServerlessClusterByARN
	FindTopicByTwoPartKey                        = findTopicByTwoPartKey
	FindVPCConnectionByARN                       = findVPCConnectionByARN

	ClusterUUIDFromARN    = clusterUUIDFromARN
	NormalizeKafkaVersion = normalizeKafkaVersion // nosemgrep:ci.kafka-in-var-name
	SortEndpointsString   = sortEndpointsString
)
