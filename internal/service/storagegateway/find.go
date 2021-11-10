package storagegateway

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindLocalDiskByDiskID(conn *storagegateway.StorageGateway, gatewayARN string, diskID string) (*storagegateway.Disk, error) {
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

func FindLocalDiskByDiskPath(conn *storagegateway.StorageGateway, gatewayARN string, diskPath string) (*storagegateway.Disk, error) {
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

func FindUploadBufferDisk(conn *storagegateway.StorageGateway, gatewayARN string, diskID string) (*string, error) {
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

func FindSMBFileShareByARN(conn *storagegateway.StorageGateway, arn string) (*storagegateway.SMBFileShareInfo, error) {
	input := &storagegateway.DescribeSMBFileSharesInput{
		FileShareARNList: aws.StringSlice([]string{arn}),
	}

	output, err := conn.DescribeSMBFileShares(input)

	if operationErrorCode(err) == operationErrCodeFileShareNotFound {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.SMBFileShareInfoList) == 0 || output.SMBFileShareInfoList[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.SMBFileShareInfoList); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.SMBFileShareInfoList[0], nil
}

func FindFileSystemAssociationByARN(conn *storagegateway.StorageGateway, arn string) (*storagegateway.FileSystemAssociationInfo, error) {
	input := &storagegateway.DescribeFileSystemAssociationsInput{
		FileSystemAssociationARNList: []*string{aws.String(arn)},
	}

	output, err := conn.DescribeFileSystemAssociations(input)

	if operationErrorCode(err) == operationErrCodeFileSystemAssociationNotFound {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.FileSystemAssociationInfoList) == 0 || output.FileSystemAssociationInfoList[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.FileSystemAssociationInfoList); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.FileSystemAssociationInfoList[0], nil
}
