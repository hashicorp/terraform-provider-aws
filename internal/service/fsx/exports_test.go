// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fsx

// Exports for use in tests only.
var (
	ResourceBackup                    = resourceBackup
	ResourceDataRepositoryAssociation = resourceDataRepositoryAssociation
	ResourceFileCache                 = resourceFileCache
	ResourceOpenZFSSnapshot           = resourceOpenZFSSnapshot

	FindBackupByID                    = findBackupByID
	FindDataRepositoryAssociationByID = findDataRepositoryAssociationByID
	FindFileCacheByID                 = findFileCacheByID
)
