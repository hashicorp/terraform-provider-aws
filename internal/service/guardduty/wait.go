package guardduty

import (
	"time"

	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// Maximum amount of time to wait for an AdminAccount to return Enabled
	adminAccountEnabledTimeout = 5 * time.Minute

	// Maximum amount of time to wait for an AdminAccount to return NotFound
	adminAccountNotFoundTimeout = 5 * time.Minute

	// Maximum amount of time to wait for a PublishingDestination to return Publishing
	publishingDestinationCreatedTimeout    = 5 * time.Minute
	publishingDestinationCreatedMinTimeout = 3 * time.Second

	// Maximum amount of time to wait for membership to propagate
	// When removing Organization Admin Accounts, there is eventual
	// consistency even after the account is no longer listed.
	// Reference error message:
	// BadRequestException: The request is rejected because the current account cannot delete detector while it has invited or associated members.
	membershipPropagationTimeout = 2 * time.Minute
)

// waitAdminAccountEnabled waits for an AdminAccount to return Enabled
func waitAdminAccountEnabled(conn *guardduty.GuardDuty, adminAccountID string) (*guardduty.AdminAccount, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{adminStatusNotFound},
		Target:  []string{guardduty.AdminStatusEnabled},
		Refresh: statusAdminAccountAdmin(conn, adminAccountID),
		Timeout: adminAccountNotFoundTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*guardduty.AdminAccount); ok {
		return output, err
	}

	return nil, err
}

// waitAdminAccountNotFound waits for an AdminAccount to return NotFound
func waitAdminAccountNotFound(conn *guardduty.GuardDuty, adminAccountID string) (*guardduty.AdminAccount, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{guardduty.AdminStatusDisableInProgress},
		Target:  []string{adminStatusNotFound},
		Refresh: statusAdminAccountAdmin(conn, adminAccountID),
		Timeout: adminAccountNotFoundTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*guardduty.AdminAccount); ok {
		return output, err
	}

	return nil, err
}

// waitPublishingDestinationCreated waits for GuardDuty to return Publishing
func waitPublishingDestinationCreated(conn *guardduty.GuardDuty, destinationID, detectorID string) (*guardduty.CreatePublishingDestinationOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{guardduty.PublishingStatusPendingVerification},
		Target:  []string{guardduty.PublishingStatusPublishing},
		Refresh: statusPublishingDestination(conn, destinationID, detectorID),
		Timeout: publishingDestinationCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*guardduty.CreatePublishingDestinationOutput); ok {
		return v, err
	}

	return nil, err
}
