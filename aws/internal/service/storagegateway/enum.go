package storagegateway

import "time"

const (
	AuthenticationActiveDirectory = "ActiveDirectory"
	AuthenticationGuestAccess     = "GuestAccess"
)

func Authentication_Values() []string {
	return []string{
		AuthenticationActiveDirectory,
		AuthenticationGuestAccess,
	}
}

const (
	DefaultStorageClassS3IntelligentTiering = "S3_INTELLIGENT_TIERING"
	DefaultStorageClassS3OneZoneIA          = "S3_ONEZONE_IA"
	DefaultStorageClassS3Standard           = "S3_STANDARD"
	DefaultStorageClassS3StandardIA         = "S3_STANDARD_IA"
)

func DefaultStorageClass_Values() []string {
	return []string{
		DefaultStorageClassS3IntelligentTiering,
		DefaultStorageClassS3OneZoneIA,
		DefaultStorageClassS3Standard,
		DefaultStorageClassS3StandardIA,
	}
}

const (
	FileShareStatusAvailable     = "AVAILABLE"
	FileShareStatusCreating      = "CREATING"
	FileShareStatusDeleting      = "DELETING"
	FileShareStatusForceDeleting = "FORCE_DELETING"
	FileShareStatusUpdating      = "UPDATING"
)

const (
	AssociateFileSystemCreateTimeout = 3 * time.Minute
	AssociateFileSystemUpdateTimeout = 3 * time.Minute
	AssociateFileSystemDeleteTimeout = 3 * time.Minute
)

const (
	FsxFileSystemAssociationStatusAvailable     = "AVAILABLE"
	FsxFileSystemAssociationStatusCreating      = "CREATING"
	FsxFileSystemAssociationStatusDeleting      = "DELETING"
	FsxFileSystemAssociationStatusForceDeleting = "FORCE_DELETING"
	FsxFileSystemAssociationStatusUpdating      = "UPDATING"
	FsxFileSystemAssociationStatusError         = "ERROR"
)

func FsxFileSystemStatusAvailableStatusPending() []string {
	return []string{FsxFileSystemAssociationStatusCreating, FsxFileSystemAssociationStatusUpdating}
}

func FsxFileSystemStatusAvailableStatusTarget() []string {
	return []string{FsxFileSystemAssociationStatusAvailable}
}

func FsxFileSystemStatusDeletedStatusPending() []string {
	return []string{FsxFileSystemAssociationStatusAvailable, FsxFileSystemAssociationStatusDeleting, FsxFileSystemAssociationStatusForceDeleting}
}

func FsxFileSystemStatusDeletedStatusTarget() []string {
	return []string{}
}
