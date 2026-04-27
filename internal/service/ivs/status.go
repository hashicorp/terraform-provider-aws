// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ivs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ivs"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
)

const (
	statusNormal        = "Normal"
	statusChangePending = "Pending"
	statusUpdated       = "Updated"
)

func statusPlaybackKeyPair(conn *ivs.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := FindPlaybackKeyPairByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, statusNormal, nil
	}
}

func statusRecordingConfiguration(conn *ivs.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := FindRecordingConfigurationByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.State), nil
	}
}

func statusChannel(conn *ivs.Client, arn string, updateDetails *ivs.UpdateChannelInput) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := FindChannelByID(ctx, conn, arn)
		if retry.NotFound(err) {
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
