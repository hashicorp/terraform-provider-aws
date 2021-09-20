package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	// Maximum amount of time to wait for an AdminAccount to return Enabled
	AdminAccountEnabledTimeout = 5 * time.Minute

	// Maximum amount of time to wait for an AdminAccount to return NotFound
	AdminAccountNotFoundTimeout = 5 * time.Minute

	StandardsSubscriptionCreateTimeout = 3 * time.Minute
	StandardsSubscriptionDeleteTimeout = 3 * time.Minute
)

// AdminAccountEnabled waits for an AdminAccount to return Enabled
func AdminAccountEnabled(conn *securityhub.SecurityHub, adminAccountID string) (*securityhub.AdminAccount, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{AdminStatusNotFound},
		Target:  []string{securityhub.AdminStatusEnabled},
		Refresh: AdminAccountAdminStatus(conn, adminAccountID),
		Timeout: AdminAccountEnabledTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*securityhub.AdminAccount); ok {
		return output, err
	}

	return nil, err
}

// AdminAccountNotFound waits for an AdminAccount to return NotFound
func AdminAccountNotFound(conn *securityhub.SecurityHub, adminAccountID string) (*securityhub.AdminAccount, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{securityhub.AdminStatusDisableInProgress},
		Target:  []string{AdminStatusNotFound},
		Refresh: AdminAccountAdminStatus(conn, adminAccountID),
		Timeout: AdminAccountNotFoundTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*securityhub.AdminAccount); ok {
		return output, err
	}

	return nil, err
}

func StandardsSubscriptionCreated(conn *securityhub.SecurityHub, arn string) (*securityhub.StandardsSubscription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{securityhub.StandardsStatusPending},
		Target:  []string{securityhub.StandardsStatusReady, securityhub.StandardsStatusIncomplete},
		Refresh: StandardsSubscriptionStatus(conn, arn),
		Timeout: StandardsSubscriptionCreateTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*securityhub.StandardsSubscription); ok {
		return output, err
	}

	return nil, err
}

func StandardsSubscriptionDeleted(conn *securityhub.SecurityHub, arn string) (*securityhub.StandardsSubscription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{securityhub.StandardsStatusDeleting},
		Target:  []string{StandardsStatusNotFound, securityhub.StandardsStatusIncomplete},
		Refresh: StandardsSubscriptionStatus(conn, arn),
		Timeout: StandardsSubscriptionDeleteTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*securityhub.StandardsSubscription); ok {
		return output, err
	}

	return nil, err
}
