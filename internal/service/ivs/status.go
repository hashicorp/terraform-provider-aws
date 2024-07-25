// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ivs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ivs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	statusNormal        = "Normal"
	statusChangePending = "Pending"
	statusUpdated       = "Updated"
)

func statusPlaybackKeyPair(ctx context.Context, conn *ivs.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindPlaybackKeyPairByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, statusNormal, nil
	}
}

func statusRecordingConfiguration(ctx context.Context, conn *ivs.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindRecordingConfigurationByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.State), nil
	}
}

func statusChannel(ctx context.Context, conn *ivs.Client, arn string, updateDetails *ivs.UpdateChannelInput) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindChannelByID(ctx, conn, arn)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if updateDetails == nil {
			return out, statusNormal, nil
		} else {
			if (updateDetails.Authorized == out.Authorized) ||
				(updateDetails.LatencyMode == out.LatencyMode) ||
				(updateDetails.Name != nil && aws.ToString(updateDetails.Name) == aws.ToString(out.Name)) ||
				(updateDetails.RecordingConfigurationArn != nil && aws.ToString(updateDetails.RecordingConfigurationArn) == aws.ToString(out.RecordingConfigurationArn)) ||
				(updateDetails.Type == out.Type) {
				return out, statusUpdated, nil
			}
			return out, statusChangePending, nil
		}
	}
}
