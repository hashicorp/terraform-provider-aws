// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package guardduty

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/guardduty"
	awstypes "github.com/aws/aws-sdk-go-v2/service/guardduty/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
)

const (
	// Maximum amount of time to wait for an AdminAccount to return Enabled
	adminAccountEnabledTimeout = 5 * time.Minute

	// Maximum amount of time to wait for an AdminAccount to return NotFound
	adminAccountNotFoundTimeout = 5 * time.Minute

	// Maximum amount of time to wait for a PublishingDestination to return Publishing
	publishingDestinationCreatedTimeout = 5 * time.Minute

	// Maximum amount of time to wait for membership to propagate
	// When removing Organization Admin Accounts, there is eventual
	// consistency even after the account is no longer listed.
	// Reference error message:
	// BadRequestException: The request is rejected because the current account cannot delete detector while it has invited or associated members.
	membershipPropagationTimeout = 2 * time.Minute
)

// waitAdminAccountEnabled waits for an AdminAccount to return Enabled
func waitAdminAccountEnabled(ctx context.Context, conn *guardduty.Client, adminAccountID string) (*awstypes.AdminAccount, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{adminStatusNotFound},
		Target:  enum.Slice(awstypes.AdminStatusEnabled),
		Refresh: statusAdminAccountAdmin(conn, adminAccountID),
		Timeout: adminAccountEnabledTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.AdminAccount); ok {
		return output, err
	}

	return nil, err
}

// waitAdminAccountNotFound waits for an AdminAccount to return NotFound
func waitAdminAccountNotFound(ctx context.Context, conn *guardduty.Client, adminAccountID string) (*awstypes.AdminAccount, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AdminStatusDisableInProgress),
		Target:  []string{adminStatusNotFound},
		Refresh: statusAdminAccountAdmin(conn, adminAccountID),
		Timeout: adminAccountNotFoundTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.AdminAccount); ok {
		return output, err
	}

	return nil, err
}

// waitPublishingDestinationCreated waits for GuardDuty to return Publishing
func waitPublishingDestinationCreated(ctx context.Context, conn *guardduty.Client, destinationID, detectorID string) (*guardduty.CreatePublishingDestinationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.PublishingStatusPendingVerification),
		Target:  enum.Slice(awstypes.PublishingStatusPublishing),
		Refresh: statusPublishingDestination(conn, destinationID, detectorID),
		Timeout: publishingDestinationCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*guardduty.CreatePublishingDestinationOutput); ok {
		return v, err
	}

	return nil, err
}
