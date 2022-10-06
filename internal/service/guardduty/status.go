package guardduty

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/guardduty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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
func statusAdminAccountAdmin(conn *guardduty.GuardDuty, adminAccountID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		adminAccount, err := getOrganizationAdminAccount(conn, adminAccountID)

		if err != nil {
			return nil, adminStatusUnknown, err
		}

		if adminAccount == nil {
			return adminAccount, adminStatusNotFound, nil
		}

		return adminAccount, aws.StringValue(adminAccount.AdminStatus), nil
	}
}

// statusPublishingDestination fetches the PublishingDestination and its Status
func statusPublishingDestination(conn *guardduty.GuardDuty, destinationID, detectorID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &guardduty.DescribePublishingDestinationInput{
			DetectorId:    aws.String(detectorID),
			DestinationId: aws.String(destinationID),
		}

		output, err := conn.DescribePublishingDestination(input)

		if err != nil {
			return output, publishingStatusFailed, err
		}

		if output == nil {
			return output, publishingStatusUnknown, nil
		}

		return output, aws.StringValue(output.Status), nil
	}
}

// TODO: Migrate to shared internal package guardduty
func getOrganizationAdminAccount(conn *guardduty.GuardDuty, adminAccountID string) (*guardduty.AdminAccount, error) {
	input := &guardduty.ListOrganizationAdminAccountsInput{}
	var result *guardduty.AdminAccount

	err := conn.ListOrganizationAdminAccountsPages(input, func(page *guardduty.ListOrganizationAdminAccountsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, adminAccount := range page.AdminAccounts {
			if adminAccount == nil {
				continue
			}

			if aws.StringValue(adminAccount.AdminAccountId) == adminAccountID {
				result = adminAccount
				return false
			}
		}

		return !lastPage
	})

	return result, err
}
