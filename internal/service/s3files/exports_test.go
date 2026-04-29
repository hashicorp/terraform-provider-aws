// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3files

// Exports for use in tests only.
var (
	ResourceAccessPoint                  = newAccessPointResource
	ResourceFileSystem                   = newFileSystemResource
	ResourceFileSystemPolicy             = newFileSystemPolicyResource
	ResourceMountTarget                  = newMountTargetResource
	ResourceSynchronizationConfiguration = newSynchronizationConfigurationResource

	FindAccessPointByID                            = findAccessPointByID
	FindFileSystemByID                             = findFileSystemByID
	FindFileSystemPolicyByID                       = findFileSystemPolicyByID
	FindMountTargetByID                            = findMountTargetByID
	FindSynchronizationConfigurationByFileSystemID = findSynchronizationConfigurationByFileSystemID
)
