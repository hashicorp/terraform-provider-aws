// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package storagegateway

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/storagegateway"
	awstypes "github.com/aws/aws-sdk-go-v2/service/storagegateway/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	gatewayStatusConnected          = "GatewayConnected"
	storediSCSIVolumeStatusNotFound = "NotFound"
)

func statusGateway(ctx context.Context, conn *storagegateway.Client, gatewayARN string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &storagegateway.DescribeGatewayInformationInput{
			GatewayARN: aws.String(gatewayARN),
		}

		output, err := conn.DescribeGatewayInformation(ctx, input)

		if errs.IsAErrorMessageContains[*awstypes.InvalidGatewayRequestException](err, "The specified gateway is not connected") {
			return output, string(awstypes.ErrorCodeGatewayNotConnected), nil
		}

		if err != nil {
			return output, "", err
		}

		return output, gatewayStatusConnected, nil
	}
}

func statusGatewayJoinDomain(ctx context.Context, conn *storagegateway.Client, gatewayARN string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &storagegateway.DescribeSMBSettingsInput{
			GatewayARN: aws.String(gatewayARN),
		}

		output, err := conn.DescribeSMBSettings(ctx, input)

		if errs.IsAErrorMessageContains[*awstypes.InvalidGatewayRequestException](err, "The specified gateway is not connected") {
			return output, string(awstypes.ActiveDirectoryStatusUnknownError), nil
		}

		if err != nil {
			return output, string(awstypes.ActiveDirectoryStatusUnknownError), err
		}

		return output, string(output.ActiveDirectoryStatus), nil
	}
}

// statusStorediSCSIVolume fetches the Volume and its Status
func statusStorediSCSIVolume(ctx context.Context, conn *storagegateway.Client, volumeARN string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &storagegateway.DescribeStorediSCSIVolumesInput{
			VolumeARNs: []string{volumeARN},
		}

		output, err := conn.DescribeStorediSCSIVolumes(ctx, input)

		if tfawserr.ErrCodeEquals(err, string(awstypes.ErrorCodeVolumeNotFound)) ||
			errs.IsAErrorMessageContains[*awstypes.InvalidGatewayRequestException](err, "The specified volume was not found") {
			return nil, storediSCSIVolumeStatusNotFound, nil
		}

		if err != nil {
			return nil, "", err
		}

		if output == nil || len(output.StorediSCSIVolumes) == 0 {
			return nil, storediSCSIVolumeStatusNotFound, nil
		}

		return output, aws.ToString(output.StorediSCSIVolumes[0].VolumeStatus), nil
	}
}

func statusNFSFileShare(ctx context.Context, conn *storagegateway.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findNFSFileShareByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.FileShareStatus), nil
	}
}

func statusSMBFileShare(ctx context.Context, conn *storagegateway.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findSMBFileShareByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.FileShareStatus), nil
	}
}

func statusFileSystemAssociation(ctx context.Context, conn *storagegateway.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findFileSystemAssociationByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.FileSystemAssociationStatus), nil
	}
}
