package finder

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	tfstoragegateway "github.com/hashicorp/terraform-provider-aws/aws/internal/service/storagegateway"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func LocalDiskByDiskId(conn *storagegateway.StorageGateway, gatewayARN string, diskID string) (*storagegateway.Disk, error) {
	input := &storagegateway.ListLocalDisksInput{
		GatewayARN: aws.String(gatewayARN),
	}

	output, err := conn.ListLocalDisks(input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	for _, disk := range output.Disks {
		if aws.StringValue(disk.DiskId) == diskID {
			return disk, nil
		}
	}

	return nil, nil
}

func LocalDiskByDiskPath(conn *storagegateway.StorageGateway, gatewayARN string, diskPath string) (*storagegateway.Disk, error) {
	input := &storagegateway.ListLocalDisksInput{
		GatewayARN: aws.String(gatewayARN),
	}

	output, err := conn.ListLocalDisks(input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	for _, disk := range output.Disks {
		if aws.StringValue(disk.DiskPath) == diskPath {
			return disk, nil
		}
	}

	return nil, nil
}

func UploadBufferDisk(conn *storagegateway.StorageGateway, gatewayARN string, diskID string) (*string, error) {
	input := &storagegateway.DescribeUploadBufferInput{
		GatewayARN: aws.String(gatewayARN),
	}

	var result *string

	output, err := conn.DescribeUploadBuffer(input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	for _, diskId := range output.DiskIds {
		if aws.StringValue(diskId) == diskID {
			result = diskId
			break
		}
	}

	return result, err
}

func SMBFileShareByARN(conn *storagegateway.StorageGateway, arn string) (*storagegateway.SMBFileShareInfo, error) {
	input := &storagegateway.DescribeSMBFileSharesInput{
		FileShareARNList: aws.StringSlice([]string{arn}),
	}

	output, err := conn.DescribeSMBFileShares(input)

	if tfstoragegateway.OperationErrorCode(err) == tfstoragegateway.OperationErrCodeFileShareNotFound {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.SMBFileShareInfoList) == 0 || output.SMBFileShareInfoList[0] == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	// TODO Check for multiple results.
	// TODO https://github.com/hashicorp/terraform-provider-aws/pull/17613.
	return output.SMBFileShareInfoList[0], nil
}

func FileSystemAssociationByARN(conn *storagegateway.StorageGateway, fileSystemAssociationARN string) (*storagegateway.FileSystemAssociationInfo, error) {

	input := &storagegateway.DescribeFileSystemAssociationsInput{
		FileSystemAssociationARNList: []*string{aws.String(fileSystemAssociationARN)},
	}
	log.Printf("[DEBUG] Reading Storage Gateway File System Associations: %s", input)

	output, err := conn.DescribeFileSystemAssociations(input)
	if err != nil {
		if tfstoragegateway.InvalidGatewayRequestErrCodeEquals(err, tfstoragegateway.FileSystemAssociationNotFound) {
			log.Printf("[WARN] Storage Gateway File System Association (%s) not found", fileSystemAssociationARN)
			return nil, nil
		}

		return nil, fmt.Errorf("error reading Storage Gateway File System Association (%s): %w", fileSystemAssociationARN, err)
	}

	if output == nil || len(output.FileSystemAssociationInfoList) == 0 || output.FileSystemAssociationInfoList[0] == nil {
		log.Printf("[WARN] Storage Gateway File System Association (%s) not found", fileSystemAssociationARN)
		return nil, nil
	}

	filesystem := output.FileSystemAssociationInfoList[0]

	return filesystem, nil
}
