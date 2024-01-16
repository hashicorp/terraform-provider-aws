// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ivs

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ivs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	statusNormal        = "Normal"
	statusChangePending = "Pending"
	statusUpdated       = "Updated"
)

func statusPlaybackKeyPair(ctx context.Context, conn *ivs.IVS, id string) retry.StateRefreshFunc {
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

func statusRecordingConfiguration(ctx context.Context, conn *ivs.IVS, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindRecordingConfigurationByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, aws.StringValue(out.State), nil
	}
}

func statusChannel(ctx context.Context, conn *ivs.IVS, arn string, updateDetails *ivs.UpdateChannelInput) retry.StateRefreshFunc {
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
			if (updateDetails.Authorized != nil && aws.BoolValue(updateDetails.Authorized) == aws.BoolValue(out.Authorized)) ||
				(updateDetails.LatencyMode != nil && aws.StringValue(updateDetails.LatencyMode) == aws.StringValue(out.LatencyMode)) ||
				(updateDetails.Name != nil && aws.StringValue(updateDetails.Name) == aws.StringValue(out.Name)) ||
				(updateDetails.RecordingConfigurationArn != nil && aws.StringValue(updateDetails.RecordingConfigurationArn) == aws.StringValue(out.RecordingConfigurationArn)) ||
				(updateDetails.Type != nil && aws.StringValue(updateDetails.Type) == aws.StringValue(out.Type)) {
				return out, statusUpdated, nil
			}
			return out, statusChangePending, nil
		}
	}
}
