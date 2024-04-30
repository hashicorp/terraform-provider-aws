// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fsx

// Exports for use in tests only.
var (
	ResourceBackup                     = resourceBackup
	ResourceDataRepositoryAssociation  = resourceDataRepositoryAssociation
	ResourceFileCache                  = resourceFileCache
	ResourceLustreFileSystem           = resourceLustreFileSystem
	ResourceONTAPFileSystem            = resourceONTAPFileSystem
	ResourceONTAPStorageVirtualMachine = resourceONTAPStorageVirtualMachine
	ResourceONTAPVolume                = resourceONTAPVolume
	ResourceOpenZFSSnapshot            = resourceOpenZFSSnapshot

	FindBackupByID                    = findBackupByID
	FindDataRepositoryAssociationByID = findDataRepositoryAssociationByID
	FindFileCacheByID                 = findFileCacheByID
	FindLustreFileSystemByID          = findLustreFileSystemByID
	FindONTAPFileSystemByID           = findONTAPFileSystemByID
	FindONTAPVolumeByID               = findONTAPVolumeByID
	FindStorageVirtualMachineByID     = findStorageVirtualMachineByID
	FindSnapshotByID                  = findSnapshotByID
)
