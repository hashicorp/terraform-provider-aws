package storagegateway

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	gatewayStatusConnected          = "GatewayConnected"
	storediSCSIVolumeStatusNotFound = "NotFound"
)

func statusGateway(ctx context.Context, conn *storagegateway.StorageGateway, gatewayARN string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &storagegateway.DescribeGatewayInformationInput{
			GatewayARN: aws.String(gatewayARN),
		}

		output, err := conn.DescribeGatewayInformationWithContext(ctx, input)

		if tfawserr.ErrMessageContains(err, storagegateway.ErrCodeInvalidGatewayRequestException, "The specified gateway is not connected") {
			return output, storagegateway.ErrorCodeGatewayNotConnected, nil
		}

		if err != nil {
			return output, "", err
		}

		return output, gatewayStatusConnected, nil
	}
}

func statusGatewayJoinDomain(ctx context.Context, conn *storagegateway.StorageGateway, gatewayARN string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &storagegateway.DescribeSMBSettingsInput{
			GatewayARN: aws.String(gatewayARN),
		}

		output, err := conn.DescribeSMBSettingsWithContext(ctx, input)

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
func statusStorediSCSIVolume(ctx context.Context, conn *storagegateway.StorageGateway, volumeARN string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &storagegateway.DescribeStorediSCSIVolumesInput{
			VolumeARNs: []*string{aws.String(volumeARN)},
		}

		output, err := conn.DescribeStorediSCSIVolumesWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, storagegateway.ErrorCodeVolumeNotFound) ||
			tfawserr.ErrMessageContains(err, storagegateway.ErrCodeInvalidGatewayRequestException, "The specified volume was not found") {
			return nil, storediSCSIVolumeStatusNotFound, nil
		}

		if err != nil {
			return nil, "", err
		}

		if output == nil || len(output.StorediSCSIVolumes) == 0 {
			return nil, storediSCSIVolumeStatusNotFound, nil
		}

		return output, aws.StringValue(output.StorediSCSIVolumes[0].VolumeStatus), nil
	}
}

func statusNFSFileShare(ctx context.Context, conn *storagegateway.StorageGateway, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindNFSFileShareByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.FileShareStatus), nil
	}
}

func statusSMBFileShare(ctx context.Context, conn *storagegateway.StorageGateway, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindSMBFileShareByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.FileShareStatus), nil
	}
}

func statusFileSystemAssociation(ctx context.Context, conn *storagegateway.StorageGateway, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindFileSystemAssociationByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.FileSystemAssociationStatus), nil
	}
}
