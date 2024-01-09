// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafka

// Exports for use in tests only.
var (
	ResourceCluster                = resourceCluster
	ResourceConfiguration          = resourceConfiguration
	ResourceReplicator             = resourceReplicator
	ResourceSCRAMSecretAssociation = resourceSCRAMSecretAssociation
	ResourceVPCConnection          = resourceVPCConnection

	FindClusterByARN             = findClusterByARN
	FindConfigurationByARN       = findConfigurationByARN
	FindReplicatorByARN          = findReplicatorByARN
	FindSCRAMSecretsByClusterARN = findSCRAMSecretsByClusterARN
	FindVPCConnectionByARN       = findVPCConnectionByARN
)
