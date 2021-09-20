package storagegateway

import (
	"time"

	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	storageGatewayGatewayConnectedMinTimeout                = 10 * time.Second
	storageGatewayGatewayConnectedContinuousTargetOccurence = 6
	storageGatewayGatewayJoinDomainJoinedTimeout            = 5 * time.Minute
	storediSCSIVolumeAvailableTimeout                       = 5 * time.Minute
	nfsFileShareAvailableDelay                              = 5 * time.Second
	nfsFileShareDeletedDelay                                = 5 * time.Second
	smbFileShareAvailableDelay                              = 5 * time.Second
	smbFileShareDeletedDelay                                = 5 * time.Second
	fileSystemAssociationAvailableDelay                     = 5 * time.Second
	fileSystemAssociationDeletedDelay                       = 5 * time.Second
)

func waitStorageGatewayGatewayConnected(conn *storagegateway.StorageGateway, gatewayARN string, timeout time.Duration) (*storagegateway.DescribeGatewayInformationOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{storagegateway.ErrorCodeGatewayNotConnected},
		Target:                    []string{storageGatewayGatewayStatusConnected},
		Refresh:                   statusStorageGatewayGateway(conn, gatewayARN),
		Timeout:                   timeout,
		MinTimeout:                storageGatewayGatewayConnectedMinTimeout,
		ContinuousTargetOccurence: storageGatewayGatewayConnectedContinuousTargetOccurence, // Gateway activations can take a few seconds and can trigger a reboot of the Gateway
	}

	outputRaw, err := stateConf.WaitForState()

	switch output := outputRaw.(type) {
	case *storagegateway.DescribeGatewayInformationOutput:
		return output, err
	default:
		return nil, err
	}
}

func waitStorageGatewayGatewayJoinDomainJoined(conn *storagegateway.StorageGateway, volumeARN string) (*storagegateway.DescribeSMBSettingsOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{storagegateway.ActiveDirectoryStatusJoining},
		Target:  []string{storagegateway.ActiveDirectoryStatusJoined},
		Refresh: statusStorageGatewayGatewayJoinDomain(conn, volumeARN),
		Timeout: storageGatewayGatewayJoinDomainJoinedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*storagegateway.DescribeSMBSettingsOutput); ok {
		return output, err
	}

	return nil, err
}

// waitStorediSCSIVolumeAvailable waits for a StoredIscsiVolume to return Available
func waitStorediSCSIVolumeAvailable(conn *storagegateway.StorageGateway, volumeARN string) (*storagegateway.DescribeStorediSCSIVolumesOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"BOOTSTRAPPING", "CREATING", "RESTORING"},
		Target:  []string{"AVAILABLE"},
		Refresh: statusStorediSCSIVolume(conn, volumeARN),
		Timeout: storediSCSIVolumeAvailableTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*storagegateway.DescribeStorediSCSIVolumesOutput); ok {
		return output, err
	}

	return nil, err
}

// waitNFSFileShareAvailable waits for a NFS File Share to return Available
func waitNFSFileShareAvailable(conn *storagegateway.StorageGateway, fileShareArn string, timeout time.Duration) (*storagegateway.NFSFileShareInfo, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"BOOTSTRAPPING", "CREATING", "RESTORING", "UPDATING"},
		Target:  []string{"AVAILABLE"},
		Refresh: statusNFSFileShare(conn, fileShareArn),
		Timeout: timeout,
		Delay:   nfsFileShareAvailableDelay,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*storagegateway.NFSFileShareInfo); ok {
		return output, err
	}

	return nil, err
}

func waitNFSFileShareDeleted(conn *storagegateway.StorageGateway, fileShareArn string, timeout time.Duration) (*storagegateway.NFSFileShareInfo, error) {
	stateConf := &resource.StateChangeConf{
		Pending:        []string{"AVAILABLE", "DELETING", "FORCE_DELETING"},
		Target:         []string{},
		Refresh:        statusNFSFileShare(conn, fileShareArn),
		Timeout:        timeout,
		Delay:          nfsFileShareDeletedDelay,
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*storagegateway.NFSFileShareInfo); ok {
		return output, err
	}

	return nil, err
}

func waitSMBFileShareCreated(conn *storagegateway.StorageGateway, arn string, timeout time.Duration) (*storagegateway.SMBFileShareInfo, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{fileShareStatusCreating},
		Target:  []string{fileShareStatusAvailable},
		Refresh: statusSMBFileShare(conn, arn),
		Timeout: timeout,
		Delay:   smbFileShareAvailableDelay,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*storagegateway.SMBFileShareInfo); ok {
		return output, err
	}

	return nil, err
}

func waitSMBFileShareDeleted(conn *storagegateway.StorageGateway, arn string, timeout time.Duration) (*storagegateway.SMBFileShareInfo, error) {
	stateConf := &resource.StateChangeConf{
		Pending:        []string{fileShareStatusAvailable, fileShareStatusDeleting, fileShareStatusForceDeleting},
		Target:         []string{},
		Refresh:        statusSMBFileShare(conn, arn),
		Timeout:        timeout,
		Delay:          smbFileShareDeletedDelay,
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*storagegateway.SMBFileShareInfo); ok {
		return output, err
	}

	return nil, err
}

func waitSMBFileShareUpdated(conn *storagegateway.StorageGateway, arn string, timeout time.Duration) (*storagegateway.SMBFileShareInfo, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{fileShareStatusUpdating},
		Target:  []string{fileShareStatusAvailable},
		Refresh: statusSMBFileShare(conn, arn),
		Timeout: timeout,
		Delay:   smbFileShareAvailableDelay,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*storagegateway.SMBFileShareInfo); ok {
		return output, err
	}

	return nil, err
}

// waitFileSystemAssociationAvailable waits for a File System Association to return Available
func waitFileSystemAssociationAvailable(conn *storagegateway.StorageGateway, fileSystemArn string, timeout time.Duration) (*storagegateway.FileSystemAssociationInfo, error) {
	stateConf := &resource.StateChangeConf{
		Pending: fileSystemAssociationStatusAvailableStatusPending(),
		Target:  fileSystemAssociationStatusAvailableStatusTarget(),
		Refresh: statusFileSystemAssociation(conn, fileSystemArn),
		Timeout: timeout,
		Delay:   fileSystemAssociationAvailableDelay,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*storagegateway.FileSystemAssociationInfo); ok {
		return output, err
	}

	return nil, err
}

func waitFileSystemAssociationDeleted(conn *storagegateway.StorageGateway, fileSystemArn string, timeout time.Duration) (*storagegateway.FileSystemAssociationInfo, error) {
	stateConf := &resource.StateChangeConf{
		Pending:        fileSystemAssociationStatusDeletedStatusPending(),
		Target:         fileSystemAssociationStatusDeletedStatusTarget(),
		Refresh:        statusFileSystemAssociation(conn, fileSystemArn),
		Timeout:        timeout,
		Delay:          fileSystemAssociationDeletedDelay,
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*storagegateway.FileSystemAssociationInfo); ok {
		return output, err
	}

	return nil, err
}
