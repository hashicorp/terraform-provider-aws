// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package storagegateway

// Exports for use in tests only.
var (
	ResourceCache                 = resourceCache
	ResourceCachediSCSIVolume     = resourceCachediSCSIVolume
	ResourceFileSystemAssociation = resourceFileSystemAssociation
	ResourceGateway               = resourceGateway
	ResourceNFSFileShare          = resourceNFSFileShare
	ResourceSMBFileShare          = resourceSMBFileShare
	ResourceStorediSCSIVolume     = resourceStorediSCSIVolume
	ResourceTapePool              = resourceTapePool
	ResourceUploadBuffer          = resourceUploadBuffer

	FindCacheByTwoPartKey                = findCacheByTwoPartKey
	FindCachediSCSIVolumeByARN           = findCachediSCSIVolumeByARN
	FindFileSystemAssociationByARN       = findFileSystemAssociationByARN
	FindGatewayByARN                     = findGatewayByARN
	FindNFSFileShareByARN                = findNFSFileShareByARN
	FindSMBFileShareByARN                = findSMBFileShareByARN
	FindStorediSCSIVolumeByARN           = findStorediSCSIVolumeByARN
	FindTapePoolByARN                    = findTapePoolByARN
	FindUploadBufferDiskIDByTwoPartKey   = findUploadBufferDiskIDByTwoPartKey
	FindWorkingStorageDiskIDByTwoPartKey = findWorkingStorageDiskIDByTwoPartKey

	CacheParseResourceID                      = cacheParseResourceID
	ParseVolumeGatewayARNAndTargetNameFromARN = parseVolumeGatewayARNAndTargetNameFromARN
	UploadBufferParseResourceID               = uploadBufferParseResourceID
	WorkingStorageParseResourceID             = workingStorageParseResourceID
)
