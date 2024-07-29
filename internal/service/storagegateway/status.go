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
)

const (
	storediSCSIVolumeStatusNotFound = "NotFound"
)

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
