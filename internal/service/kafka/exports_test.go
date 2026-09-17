// Copyright IBM Corp. 2014, 2026
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
	ResourceTopic                        = newTopicResource
	ResourceVPCConnection                = resourceVPCConnection

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

	ExpandKafkaCluster                      = expandKafkaCluster                      // nosemgrep:ci.kafka-in-var-name
	ExpandApacheKafkaCluster                = expandApacheKafkaCluster                // nosemgrep:ci.kafka-in-var-name
	ExpandKafkaClusterClientAuthentication  = expandKafkaClusterClientAuthentication  // nosemgrep:ci.kafka-in-var-name
	ExpandKafkaClusterEncryptionInTransit   = expandKafkaClusterEncryptionInTransit   // nosemgrep:ci.kafka-in-var-name
	FlattenApacheKafkaCluster               = flattenApacheKafkaCluster               // nosemgrep:ci.kafka-in-var-name
	FlattenKafkaClusterClientAuthentication = flattenKafkaClusterClientAuthentication // nosemgrep:ci.kafka-in-var-name
	FlattenKafkaClusterEncryptionInTransit  = flattenKafkaClusterEncryptionInTransit  // nosemgrep:ci.kafka-in-var-name
	KafkaClusterIdentifier                  = kafkaClusterIdentifier                  // nosemgrep:ci.kafka-in-var-name
)
