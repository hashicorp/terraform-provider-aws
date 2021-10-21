package prometheus

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/prometheusservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// Maximum amount of time to wait for a Workspace to be created, updated, or deleted
	workspaceTimeout = 5 * time.Minute

	// Maximum amount of time to wait for AlertManager to be created, updated, or deleted
	alertManagerTimeout = 5 * time.Minute
)

// waitAlertManagerActive waits for a AlertManager to return "Active"
func waitAlertManagerActive(ctx context.Context, conn *prometheusservice.PrometheusService, id string) (*prometheusservice.AlertManagerDefinitionDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			prometheusservice.AlertManagerDefinitionStatusCodeCreating,
			prometheusservice.AlertManagerDefinitionStatusCodeUpdating,
		},
		Target:  []string{prometheusservice.AlertManagerDefinitionStatusCodeActive},
		Refresh: statusAlertManager(ctx, conn, id),
		Timeout: alertManagerTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*prometheusservice.AlertManagerDefinitionDescription); ok {
		return v, err
	}

	return nil, err
}

// waitAlertManagerDeleted waits for a AlertManager to return "Deleted"
func waitAlertManagerDeleted(ctx context.Context, conn *prometheusservice.PrometheusService, arn string) (*prometheusservice.AlertManagerDefinitionDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{prometheusservice.AlertManagerDefinitionStatusCodeDeleting},
		Target:  []string{resourceStatusDeleted},
		Refresh: statusAlertManagerDeleted(ctx, conn, arn),
		Timeout: alertManagerTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*prometheusservice.AlertManagerDefinitionDescription); ok {
		return v, err
	}

	return nil, err
}

// waitWorkspaceActive waits for a Workspace to return "Active"
func waitWorkspaceActive(ctx context.Context, conn *prometheusservice.PrometheusService, id string) (*prometheusservice.WorkspaceSummary, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			prometheusservice.WorkspaceStatusCodeCreating,
			prometheusservice.WorkspaceStatusCodeUpdating,
		},
		Target:  []string{prometheusservice.WorkspaceStatusCodeActive},
		Refresh: statusWorkspace(ctx, conn, id),
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
