// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package storagegateway

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func findLocalDisk(ctx context.Context, conn *storagegateway.StorageGateway, input *storagegateway.ListLocalDisksInput, filter tfslices.Predicate[*storagegateway.Disk]) (*storagegateway.Disk, error) {
	output, err := findLocalDisks(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findLocalDisks(ctx context.Context, conn *storagegateway.StorageGateway, input *storagegateway.ListLocalDisksInput, filter tfslices.Predicate[*storagegateway.Disk]) ([]*storagegateway.Disk, error) {
	output, err := conn.ListLocalDisksWithContext(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return tfslices.Filter(output.Disks, filter), nil
}

func findLocalDiskByGatewayARNAndDiskID(ctx context.Context, conn *storagegateway.StorageGateway, gatewayARN, diskID string) (*storagegateway.Disk, error) {
	input := &storagegateway.ListLocalDisksInput{
		GatewayARN: aws.String(gatewayARN),
	}

	return findLocalDisk(ctx, conn, input, func(v *storagegateway.Disk) bool {
		return aws.StringValue(v.DiskId) == diskID
	})
}

func findLocalDiskByGatewayARNAndDiskPath(ctx context.Context, conn *storagegateway.StorageGateway, gatewayARN, diskPath string) (*storagegateway.Disk, error) {
	input := &storagegateway.ListLocalDisksInput{
		GatewayARN: aws.String(gatewayARN),
	}

	return findLocalDisk(ctx, conn, input, func(v *storagegateway.Disk) bool {
		return aws.StringValue(v.DiskPath) == diskPath
	})
}

func FindUploadBufferDisk(ctx context.Context, conn *storagegateway.StorageGateway, gatewayARN string, diskID string) (*string, error) {
	input := &storagegateway.DescribeUploadBufferInput{
		GatewayARN: aws.String(gatewayARN),
	}

	var result *string

	output, err := conn.DescribeUploadBufferWithContext(ctx, input)

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

func FindGatewayByARN(ctx context.Context, conn *storagegateway.StorageGateway, arn string) (*storagegateway.DescribeGatewayInformationOutput, error) {
	input := &storagegateway.DescribeGatewayInformationInput{
		GatewayARN: aws.String(arn),
	}

	output, err := conn.DescribeGatewayInformationWithContext(ctx, input)

	if operationErrorCode(err) == operationErrCodeGatewayNotFound || tfawserr.ErrCodeEquals(err, storagegateway.ErrorCodeGatewayNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func FindNFSFileShareByARN(ctx context.Context, conn *storagegateway.StorageGateway, arn string) (*storagegateway.NFSFileShareInfo, error) {
	input := &storagegateway.DescribeNFSFileSharesInput{
		FileShareARNList: aws.StringSlice([]string{arn}),
	}

	output, err := conn.DescribeNFSFileSharesWithContext(ctx, input)

	if operationErrorCode(err) == operationErrCodeFileShareNotFound {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.NFSFileShareInfoList) == 0 || output.NFSFileShareInfoList[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.NFSFileShareInfoList); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output.NFSFileShareInfoList[0], nil
}

func FindSMBFileShareByARN(ctx context.Context, conn *storagegateway.StorageGateway, arn string) (*storagegateway.SMBFileShareInfo, error) {
	input := &storagegateway.DescribeSMBFileSharesInput{
		FileShareARNList: aws.StringSlice([]string{arn}),
	}

	output, err := conn.DescribeSMBFileSharesWithContext(ctx, input)

	if operationErrorCode(err) == operationErrCodeFileShareNotFound {
		return nil, &retry.NotFoundError{
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

func FindFileSystemAssociationByARN(ctx context.Context, conn *storagegateway.StorageGateway, arn string) (*storagegateway.FileSystemAssociationInfo, error) {
	input := &storagegateway.DescribeFileSystemAssociationsInput{
		FileSystemAssociationARNList: []*string{aws.String(arn)},
	}

	output, err := conn.DescribeFileSystemAssociationsWithContext(ctx, input)

	if operationErrorCode(err) == operationErrCodeFileSystemAssociationNotFound {
		return nil, &retry.NotFoundError{
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
