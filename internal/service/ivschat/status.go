// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ivschat

import (
	"context"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ivschat"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
)

const (
	statusChangePending = "Pending"
	statusNormal        = "Normal"
	statusUpdated       = "Updated"
)

func statusLoggingConfiguration(conn *ivschat.Client, id string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findLoggingConfigurationByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.State), nil
	}
}

func statusRoom(conn *ivschat.Client, id string, updateDetails *ivschat.UpdateRoomInput) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findRoomByID(ctx, conn, id)
		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if updateDetails == nil {
			return out, statusNormal, nil
		} else {
			if (aws.ToInt32(updateDetails.MaximumMessageLength) != 0 && updateDetails.MaximumMessageLength == out.MaximumMessageLength) ||
				(aws.ToInt32(updateDetails.MaximumMessageRatePerSecond) != 0 && updateDetails.MaximumMessageRatePerSecond == out.MaximumMessageRatePerSecond) ||
				(updateDetails.MessageReviewHandler != nil && out.MessageReviewHandler != nil &&
					(updateDetails.MessageReviewHandler.FallbackResult == out.MessageReviewHandler.FallbackResult || aws.ToString(updateDetails.MessageReviewHandler.Uri) == aws.ToString(out.MessageReviewHandler.Uri))) ||
				(updateDetails.Name != nil && aws.ToString(updateDetails.Name) == aws.ToString(out.Name)) ||
				(updateDetails.LoggingConfigurationIdentifiers != nil &&
					(reflect.DeepEqual(updateDetails.LoggingConfigurationIdentifiers, out.LoggingConfigurationIdentifiers) || (len(updateDetails.LoggingConfigurationIdentifiers) == 0 && out.LoggingConfigurationIdentifiers == nil))) {
				return out, statusUpdated, nil
			}
			return out, statusChangePending, nil
		}
	}
}
