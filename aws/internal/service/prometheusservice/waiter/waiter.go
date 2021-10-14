package waiter

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/prometheusservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	// Maximum amount of time to wait for a Workspace to be created, updated, or deleted
	workspaceTimeout = 5 * time.Minute
)

// waitWorkspaceCreated waits for a Workspace to return "Active"
func waitWorkspaceCreated(ctx context.Context, conn *prometheusservice.PrometheusService, id string) (*prometheusservice.WorkspaceSummary, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{prometheusservice.WorkspaceStatusCodeCreating},
		Target:  []string{prometheusservice.WorkspaceStatusCodeActive},
		Refresh: statusWorkspaceCreated(ctx, conn, id),
		Timeout: workspaceTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*prometheusservice.WorkspaceSummary); ok {
		return v, err
	}

	return nil, err
}

// waitWorkspaceDeleted waits for a Workspace to return "Deleted"
func waitWorkspaceDeleted(ctx context.Context, conn *prometheusservice.PrometheusService, arn string) (*prometheusservice.WorkspaceSummary, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{prometheusservice.WorkspaceStatusCodeDeleting},
		Target:  []string{resourceStatusDeleted},
		Refresh: statusWorkspaceDeleted(ctx, conn, arn),
		Timeout: workspaceTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*prometheusservice.WorkspaceSummary); ok {
		return v, err
	}

	return nil, err
}
