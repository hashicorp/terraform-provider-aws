// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3files

// Exports for use in tests only.
var (
	ResourceFileSystem = newFileSystemResource
	ResourceFileSystemPolicy = newFileSystemPolicyResource

	FindFileSystemByID = findFileSystemByID
	FindFileSystemPolicyByID = findFileSystemPolicyByID
)
