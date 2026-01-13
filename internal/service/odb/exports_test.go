// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb

// Exports for use in tests only.
var (
	ResourceCloudAutonomousVMCluster   = newResourceCloudAutonomousVmCluster
	ResourceCloudExadataInfrastructure = newResourceCloudExadataInfrastructure

	FindCloudAutonomousVmClusterByID  = findCloudAutonomousVmClusterByID
	FindExadataInfraResourceByID      = findExadataInfraResourceByID
	FindCloudVmClusterForResourceByID = findCloudVmClusterForResourceByID
)
