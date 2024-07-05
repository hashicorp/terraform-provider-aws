// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package efs

// Exports for use in tests only.
var (
	ResourceAccessPoint      = resourceAccessPoint
	ResourceFileSystem       = resourceFileSystem
	ResourceFileSystemPolicy = resourceFileSystemPolicy

	FindAccessPointByID      = findAccessPointByID
	FindFileSystemByID       = findFileSystemByID
	FindFileSystemPolicyByID = findFileSystemPolicyByID
)
