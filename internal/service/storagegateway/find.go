// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package storagegateway

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/storagegateway"
	awstypes "github.com/aws/aws-sdk-go-v2/service/storagegateway/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func findLocalDisk(ctx context.Context, conn *storagegateway.Client, input *storagegateway.ListLocalDisksInput, filter tfslices.Predicate[awstypes.Disk]) (*awstypes.Disk, error) {
	output, err := findLocalDisks(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findLocalDisks(ctx context.Context, conn *storagegateway.Client, input *storagegateway.ListLocalDisksInput, filter tfslices.Predicate[awstypes.Disk]) ([]awstypes.Disk, error) {
	output, err := conn.ListLocalDisks(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return tfslices.Filter(output.Disks, filter), nil
}

func findLocalDiskByGatewayARNAndDiskID(ctx context.Context, conn *storagegateway.Client, gatewayARN, diskID string) (*awstypes.Disk, error) {
	input := &storagegateway.ListLocalDisksInput{
		GatewayARN: aws.String(gatewayARN),
	}

	return findLocalDisk(ctx, conn, input, func(v awstypes.Disk) bool {
		return aws.ToString(v.DiskId) == diskID
	})
}

func findLocalDiskByGatewayARNAndDiskPath(ctx context.Context, conn *storagegateway.Client, gatewayARN, diskPath string) (*awstypes.Disk, error) {
	input := &storagegateway.ListLocalDisksInput{
		GatewayARN: aws.String(gatewayARN),
	}

	return findLocalDisk(ctx, conn, input, func(v awstypes.Disk) bool {
		return aws.ToString(v.DiskPath) == diskPath
	})
}

func findUploadBufferDisk(ctx context.Context, conn *storagegateway.Client, gatewayARN string, diskID string) (*string, error) {
	input := &storagegateway.DescribeUploadBufferInput{
		GatewayARN: aws.String(gatewayARN),
	}

	var result string

	output, err := conn.DescribeUploadBuffer(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	for _, diskId := range output.DiskIds {
		if diskId == diskID {
			result = diskId
			break
		}
	}

	return &result, err
}

func findGatewayByARN(ctx context.Context, conn *storagegateway.Client, arn string) (*storagegateway.DescribeGatewayInformationOutput, error) {
	input := &storagegateway.DescribeGatewayInformationInput{
		GatewayARN: aws.String(arn),
	}

	output, err := conn.DescribeGatewayInformation(ctx, input)

	if operationErrorCode(err) == operationErrCodeGatewayNotFound || tfawserr.ErrCodeEquals(err, string(awstypes.ErrorCodeGatewayNotFound)) {
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

func findNFSFileShareByARN(ctx context.Context, conn *storagegateway.Client, arn string) (*awstypes.NFSFileShareInfo, error) {
	input := &storagegateway.DescribeNFSFileSharesInput{
		FileShareARNList: []string{arn},
	}

	output, err := conn.DescribeNFSFileShares(ctx, input)

	if operationErrorCode(err) == operationErrCodeFileShareNotFound {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.NFSFileShareInfoList) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.NFSFileShareInfoList); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return &output.NFSFileShareInfoList[0], nil
}

func findSMBFileShareByARN(ctx context.Context, conn *storagegateway.Client, arn string) (*awstypes.SMBFileShareInfo, error) {
	input := &storagegateway.DescribeSMBFileSharesInput{
		FileShareARNList: []string{arn},
	}

	output, err := conn.DescribeSMBFileShares(ctx, input)

	if operationErrorCode(err) == operationErrCodeFileShareNotFound {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.SMBFileShareInfoList) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.SMBFileShareInfoList); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return &output.SMBFileShareInfoList[0], nil
}

func findFileSystemAssociationByARN(ctx context.Context, conn *storagegateway.Client, arn string) (*awstypes.FileSystemAssociationInfo, error) {
	input := &storagegateway.DescribeFileSystemAssociationsInput{
		FileSystemAssociationARNList: []string{arn},
	}

	output, err := conn.DescribeFileSystemAssociations(ctx, input)

	if operationErrorCode(err) == operationErrCodeFileSystemAssociationNotFound {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.FileSystemAssociationInfoList) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.FileSystemAssociationInfoList); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return &output.FileSystemAssociationInfoList[0], nil
}
