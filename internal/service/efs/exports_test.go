// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package efs

// Exports for use in tests only.
var (
	ResourceAccessPoint              = resourceAccessPoint
	ResourceBackupPolicy             = resourceBackupPolicy
	ResourceFileSystem               = resourceFileSystem
	ResourceFileSystemPolicy         = resourceFileSystemPolicy
	ResourceMountTarget              = resourceMountTarget
	ResourceReplicationConfiguration = resourceReplicationConfiguration

	FindAccessPointByID              = findAccessPointByID
	FindBackupPolicyByID             = findBackupPolicyByID
	FindFileSystemByID               = findFileSystemByID
	FindFileSystemPolicyByID         = findFileSystemPolicyByID
	FindMountTargetByID              = findMountTargetByID
	FindReplicationConfigurationByID = findReplicationConfigurationByID
)
