package finder

import (
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
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

func FileSystemAssociationByARN(conn *storagegateway.StorageGateway, fileSystemAssociationARN string) (*storagegateway.FileSystemAssociationInfo, error) {

	input := &storagegateway.DescribeFileSystemAssociationsInput{
		FileSystemAssociationARNList: []*string{aws.String(fileSystemAssociationARN)},
	}
	log.Printf("[DEBUG] Reading Storage Gateway FSx File Associations: %s", input)

	output, err := conn.DescribeFileSystemAssociations(input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, storagegateway.ErrCodeInvalidGatewayRequestException) {
			var igrex *storagegateway.InvalidGatewayRequestException
			if ok := errors.As(err, &igrex); ok {
				if err := igrex.Error_; err != nil {
					if aws.StringValue(err.ErrorCode) == "FileSystemAssociationNotFound" {
						log.Printf("[WARN] FSX File System %q not found", fileSystemAssociationARN)
						return nil, nil
					}
				}
			}
		}

		return nil, fmt.Errorf("error reading Storage Gateway FSx File System: %w", err)
	}

	if output == nil || len(output.FileSystemAssociationInfoList) == 0 || output.FileSystemAssociationInfoList[0] == nil {
		log.Printf("[WARN] FSX File System %q not found", fileSystemAssociationARN)
		return nil, nil
	}

	filesystem := output.FileSystemAssociationInfoList[0]

	return filesystem, nil
}
