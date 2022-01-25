package storagegateway

import "time"

const (
	authenticationActiveDirectory = "ActiveDirectory"
	authenticationGuestAccess     = "GuestAccess"
)

func authentication_Values() []string {
	return []string{
		authenticationActiveDirectory,
		authenticationGuestAccess,
	}
}

const (
	defaultStorageClassS3IntelligentTiering = "S3_INTELLIGENT_TIERING"
	defaultStorageClassS3OneZoneIA          = "S3_ONEZONE_IA"
	defaultStorageClassS3Standard           = "S3_STANDARD"
	defaultStorageClassS3StandardIA         = "S3_STANDARD_IA"
)

func defaultStorageClass_Values() []string {
	return []string{
		defaultStorageClassS3IntelligentTiering,
		defaultStorageClassS3OneZoneIA,
		defaultStorageClassS3Standard,
		defaultStorageClassS3StandardIA,
	}
}

const (
	fileShareStatusAvailable     = "AVAILABLE"
	fileShareStatusCreating      = "CREATING"
	fileShareStatusDeleting      = "DELETING"
	fileShareStatusForceDeleting = "FORCE_DELETING"
	fileShareStatusUpdating      = "UPDATING"
)

const (
	fileSystemAssociationCreateTimeout = 10 * time.Minute
	fileSystemAssociationUpdateTimeout = 10 * time.Minute
	fileSystemAssociationDeleteTimeout = 10 * time.Minute
)

//nolint:deadcode,varcheck // These constants are missing from the AWS SDK
const (
	fileSystemAssociationStatusAvailable     = "AVAILABLE"
	fileSystemAssociationStatusCreating      = "CREATING"
	fileSystemAssociationStatusDeleting      = "DELETING"
	fileSystemAssociationStatusForceDeleting = "FORCE_DELETING"
	fileSystemAssociationStatusUpdating      = "UPDATING"
	fileSystemAssociationStatusError         = "ERROR"
)
