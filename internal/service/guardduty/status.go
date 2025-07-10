// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package guardduty

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/guardduty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

const (
	// AdminStatus NotFound
	adminStatusNotFound = "NotFound"

	// AdminStatus Unknown
	adminStatusUnknown = "Unknown"

	// Constants not currently provided by the AWS Go SDK
	publishingStatusFailed  = "Failed"
	publishingStatusUnknown = "Unknown"
)

// statusAdminAccountAdmin fetches the AdminAccount and its AdminStatus
func statusAdminAccountAdmin(ctx context.Context, conn *guardduty.Client, adminAccountID string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		adminAccount, err := getOrganizationAdminAccount(ctx, conn, adminAccountID)

		if err != nil {
			return nil, adminStatusUnknown, err
		}

		if adminAccount == nil {
			return adminAccount, adminStatusNotFound, nil
		}

		return adminAccount, string(adminAccount.AdminStatus), nil
	}
}

// statusPublishingDestination fetches the PublishingDestination and its Status
func statusPublishingDestination(ctx context.Context, conn *guardduty.Client, destinationID, detectorID string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		input := &guardduty.DescribePublishingDestinationInput{
			DetectorId:    aws.String(detectorID),
			DestinationId: aws.String(destinationID),
		}

		output, err := conn.DescribePublishingDestination(ctx, input)

		if err != nil {
			return output, publishingStatusFailed, err
		}

		if output == nil {
			return output, publishingStatusUnknown, nil
		}

		return output, string(output.Status), nil
	}
}
