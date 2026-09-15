// Copyright IBM Corp. 2014, 2026
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
	ResourceS3AccessPointAttachment    = newS3AccessPointAttachmentResource

	FindBackupByID                    = findBackupByID
	FindDataRepositoryAssociationByID = findDataRepositoryAssociationByID
	FindFileCacheByID                 = findFileCacheByID
	FindLustreFileSystemByID          = findLustreFileSystemByID
	FindONTAPFileSystemByID           = findONTAPFileSystemByID
	FindONTAPVolumeByID               = findONTAPVolumeByID
	FindOpenZFSFileSystemByID         = findOpenZFSFileSystemByID
	FindOpenZFSVolumeByID             = findOpenZFSVolumeByID
	FindS3AccessPointAttachmentByName = findS3AccessPointAttachmentByName
	FindStorageVirtualMachineByID     = findStorageVirtualMachineByID
	FindSnapshotByID                  = findSnapshotByID
	FindWindowsFileSystemByID         = findWindowsFileSystemByID
)
