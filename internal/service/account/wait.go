package account

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/account"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	alternateContactCreateTimeout = 5 * time.Minute
	alternateContactUpdateTimeout = 5 * time.Minute
	alternateContactDeleteTimeout = 5 * time.Minute
)

func waitAlternateContactCreated(ctx context.Context, conn *account.Account, accountID, contactType string, timeout time.Duration) (*account.AlternateContact, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusFound},
		Refresh:                   statusAlternateContact(ctx, conn, accountID, contactType),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*account.AlternateContact); ok {
		return output, err
	}

	return nil, err
}

func waitAlternateContactUpdated(ctx context.Context, conn *account.Account, accountID, contactType, email, name, phone, title string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{statusNotUpdated},
		Target:                    []string{statusUpdated},
		Refresh:                   statusAlternateContactUpdate(ctx, conn, accountID, contactType, email, name, phone, title),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitAlternateContactDeleted(ctx context.Context, conn *account.Account, accountID, contactType string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{statusFound},
		Target:  []string{},
		Refresh: statusAlternateContact(ctx, conn, accountID, contactType),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}
