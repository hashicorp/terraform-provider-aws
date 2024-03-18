// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datasync

// Exports for use in tests only.
var (
	ResourceLocationAzureBlob            = resourceLocationAzureBlob
	ResourceLocationEFS                  = resourceLocationEFS
	ResourceLocationFSxLustreFileSystem  = resourceLocationFSxLustreFileSystem
	ResourceLocationFSxONTAPFileSystem   = resourceLocationFSxONTAPFileSystem
	ResourceLocationFSxOpenZFSFileSystem = resourceLocationFSxOpenZFSFileSystem
	ResourceLocationFSxWindowsFileSystem = resourceLocationFSxWindowsFileSystem
	ResourceLocationHDFS                 = resourceLocationHDFS
	ResourceLocationNFS                  = resourceLocationNFS
	ResourceLocationObjectStorage        = resourceLocationObjectStorage
	ResourceLocationS3                   = resourceLocationS3
	ResourceLocationSMB                  = resourceLocationSMB
	ResourceTask                         = resourceTask

	FindLocationAzureBlobByARN     = findLocationAzureBlobByARN
	FindLocationEFSByARN           = findLocationEFSByARN
	FindLocationFSxLustreByARN     = findLocationFSxLustreByARN
	FindLocationFSxONTAPByARN      = findLocationFSxONTAPByARN
	FindLocationFSxOpenZFSByARN    = findLocationFSxOpenZFSByARN
	FindLocationFSxWindowsByARN    = findLocationFSxWindowsByARN
	FindLocationHDFSByARN          = findLocationHDFSByARN
	FindLocationNFSByARN           = findLocationNFSByARN
	FindLocationObjectStorageByARN = findLocationObjectStorageByARN
	FindLocationS3ByARN            = findLocationS3ByARN
	FindLocationSMBByARN           = findLocationSMBByARN
	FindTaskByARN                  = findTaskByARN
)
