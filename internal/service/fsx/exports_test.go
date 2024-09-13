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
	ResourceOpenZFSFileSystem          = resourceOpenZFSFileSystem
	ResourceOpenZFSSnapshot            = resourceOpenZFSSnapshot
	ResourceOpenZFSVolume              = resourceOpenZFSVolume

	FindBackupByID                    = findBackupByID
	FindDataRepositoryAssociationByID = findDataRepositoryAssociationByID
	FindFileCacheByID                 = findFileCacheByID
	FindLustreFileSystemByID          = findLustreFileSystemByID
	FindONTAPFileSystemByID           = findONTAPFileSystemByID
	FindONTAPVolumeByID               = findONTAPVolumeByID
	FindOpenZFSFileSystemByID         = findOpenZFSFileSystemByID
	FindOpenZFSVolumeByID             = findOpenZFSVolumeByID
	FindStorageVirtualMachineByID     = findStorageVirtualMachineByID
	FindSnapshotByID                  = findSnapshotByID
	FindWindowsFileSystemByID         = findWindowsFileSystemByID
)
