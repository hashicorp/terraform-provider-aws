// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafka

// Exports for use in tests only.
var (
	ResourceConfiguration          = resourceConfiguration
	ResourceReplicator             = resourceReplicator
	ResourceSCRAMSecretAssociation = resourceSCRAMSecretAssociation
	ResourceVPCConnection          = resourceVPCConnection

	FindConfigurationByARN       = findConfigurationByARN
	FindReplicatorByARN          = findReplicatorByARN
	FindSCRAMSecretsByClusterARN = findSCRAMSecretsByClusterARN
	FindVPCConnectionByARN       = findVPCConnectionByARN
)
