package waiter

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/storagegateway/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	storageGatewayGatewayStatusConnected = "GatewayConnected"
	storediSCSIVolumeStatusNotFound      = "NotFound"
	storediSCSIVolumeStatusUnknown       = "Unknown"
	nfsFileShareStatusNotFound           = "NotFound"
	nfsFileShareStatusUnknown            = "Unknown"
	smbFileShareStatusNotFound           = "NotFound"
	smbFileShareStatusUnknown            = "Unknown"
	fileSystemAssociationStatusNotFound  = "NotFound"
	fileSystemAssociationStatusUnknown   = "Unknown"
)

func statusStorageGatewayGateway(conn *storagegateway.StorageGateway, gatewayARN string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &storagegateway.DescribeGatewayInformationInput{
			GatewayARN: aws.String(gatewayARN),
		}

		output, err := conn.DescribeGatewayInformation(input)

		if tfawserr.ErrMessageContains(err, storagegateway.ErrCodeInvalidGatewayRequestException, "The specified gateway is not connected") {
			return output, storagegateway.ErrorCodeGatewayNotConnected, nil
		}

		if err != nil {
			return output, "", err
		}

		return output, storageGatewayGatewayStatusConnected, nil
	}
}

func statusStorageGatewayGatewayJoinDomain(conn *storagegateway.StorageGateway, gatewayARN string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &storagegateway.DescribeSMBSettingsInput{
			GatewayARN: aws.String(gatewayARN),
		}

		output, err := conn.DescribeSMBSettings(input)

		if tfawserr.ErrMessageContains(err, storagegateway.ErrCodeInvalidGatewayRequestException, "The specified gateway is not connected") {
			return output, storagegateway.ActiveDirectoryStatusUnknownError, nil
		}

		if err != nil {
			return output, storagegateway.ActiveDirectoryStatusUnknownError, err
		}

		return output, aws.StringValue(output.ActiveDirectoryStatus), nil
	}
}

// statusStorediSCSIVolume fetches the Volume and its Status
func statusStorediSCSIVolume(conn *storagegateway.StorageGateway, volumeARN string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &storagegateway.DescribeStorediSCSIVolumesInput{
			VolumeARNs: []*string{aws.String(volumeARN)},
		}

		output, err := conn.DescribeStorediSCSIVolumes(input)

		if tfawserr.ErrCodeEquals(err, storagegateway.ErrorCodeVolumeNotFound) ||
			tfawserr.ErrMessageContains(err, storagegateway.ErrCodeInvalidGatewayRequestException, "The specified volume was not found") {
			return nil, storediSCSIVolumeStatusNotFound, nil
		}

		if err != nil {
			return nil, storediSCSIVolumeStatusUnknown, err
		}

		if output == nil || len(output.StorediSCSIVolumes) == 0 {
			return nil, storediSCSIVolumeStatusNotFound, nil
		}

		return output, aws.StringValue(output.StorediSCSIVolumes[0].VolumeStatus), nil
	}
}

func statusNFSFileShare(conn *storagegateway.StorageGateway, fileShareArn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &storagegateway.DescribeNFSFileSharesInput{
			FileShareARNList: []*string{aws.String(fileShareArn)},
		}

		log.Printf("[DEBUG] Reading Storage Gateway NFS File Share: %s", input)
		output, err := conn.DescribeNFSFileShares(input)
		if err != nil {
			if tfawserr.ErrMessageContains(err, storagegateway.ErrCodeInvalidGatewayRequestException, "The specified file share was not found.") {
				return nil, nfsFileShareStatusNotFound, nil
			}
			return nil, nfsFileShareStatusUnknown, fmt.Errorf("error reading Storage Gateway NFS File Share: %w", err)
		}

		if output == nil || len(output.NFSFileShareInfoList) == 0 || output.NFSFileShareInfoList[0] == nil {
			return nil, nfsFileShareStatusNotFound, nil
		}

		fileshare := output.NFSFileShareInfoList[0]

		return fileshare, aws.StringValue(fileshare.FileShareStatus), nil
	}
}

func statusSMBFileShare(conn *storagegateway.StorageGateway, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := finder.FindSMBFileShareByARN(conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.FileShareStatus), nil
	}
}

func statusFileSystemAssociation(conn *storagegateway.StorageGateway, fileSystemArn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {

		output, err := finder.FindFileSystemAssociationByARN(conn, fileSystemArn)

		// there was an unhandled error in the Finder
		if err != nil {
			return nil, "", err
		}

		// no error, and no File System Association found
		if output == nil {
			return nil, fileSystemAssociationStatusNotFound, nil
		}

		return output, aws.StringValue(output.FileSystemAssociationStatus), nil
	}
}
