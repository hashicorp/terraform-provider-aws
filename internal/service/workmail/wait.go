package workmail

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/workmail"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	OrganizationDeleteTimeout = 2 * time.Minute
	OrganizationReadyTimeout  = 1 * time.Minute

	organizationStateActive  = "Active"
	organizationStateDeleted = "Deleted"
)

func WaitOrganizationDeleted(ctx context.Context, conn *workmail.WorkMail, name string) (*workmail.GetOrganizationOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			StateActive,
			StateDeleted,
		},
		Target:  []string{},
		Refresh: StatusOrganizationState(ctx, conn, name),
		Timeout: OrganizationDeleteTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*workmail.GetOrganizationOutput); ok {
		return v, err
	}

	return nil, err
}

func WaitOrganizationActive(ctx context.Context, conn *workmail.WorkMail, name string) (*workmail.GetOrganizationOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			StateDeleted,
		},
		Target: []string{
			StateActive,
		},
		Refresh: StatusOrganizationState(ctx, conn, name),
		Timeout: OrganizationReadyTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*workmail.GetOrganizationOutput); ok {
		return v, err
	}

	return nil, err
}
