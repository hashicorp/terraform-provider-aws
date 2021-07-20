package storagegateway

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
