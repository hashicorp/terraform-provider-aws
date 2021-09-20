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
	fileSystemAssociationCreateTimeout = 3 * time.Minute
	fileSystemAssociationUpdateTimeout = 3 * time.Minute
	fileSystemAssociationDeleteTimeout = 3 * time.Minute
)

const (
	fileSystemAssociationStatusAvailable     = "AVAILABLE"
	fileSystemAssociationStatusCreating      = "CREATING"
	fileSystemAssociationStatusDeleting      = "DELETING"
	fileSystemAssociationStatusForceDeleting = "FORCE_DELETING"
	fileSystemAssociationStatusUpdating      = "UPDATING"
	fileSystemAssociationStatusError         = "ERROR"
)

func fileSystemAssociationStatusAvailableStatusPending() []string {
	return []string{fileSystemAssociationStatusCreating, fileSystemAssociationStatusUpdating}
}

func fileSystemAssociationStatusAvailableStatusTarget() []string {
	return []string{fileSystemAssociationStatusAvailable}
}

func fileSystemAssociationStatusDeletedStatusPending() []string {
	return []string{fileSystemAssociationStatusAvailable, fileSystemAssociationStatusDeleting, fileSystemAssociationStatusForceDeleting}
}

func fileSystemAssociationStatusDeletedStatusTarget() []string {
	return []string{}
}
