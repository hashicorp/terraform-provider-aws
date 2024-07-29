// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package storagegateway

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/storagegateway"
	awstypes "github.com/aws/aws-sdk-go-v2/service/storagegateway/types"
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
