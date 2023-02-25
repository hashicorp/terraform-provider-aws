package workmail

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/workmail"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func waitOrganizationCreated(ctx context.Context, conn *workmail.Client, id string, timeout time.Duration) (*workmail.DescribeOrganizationOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{statusRequested, statusCreating},
		Target:                    []string{statusActive},
		Refresh:                   statusOrganization(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*workmail.DescribeOrganizationOutput); ok {
		return out, err
	}

	return nil, err
}

func waitOrganizationDeleted(ctx context.Context, conn *workmail.Client, id string, timeout time.Duration) (*workmail.DescribeOrganizationOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{statusDeleting},
		Target:  []string{statusDeleted},
		Refresh: statusOrganization(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*workmail.DescribeOrganizationOutput); ok {
		return out, err
	}

	return nil, err
}
