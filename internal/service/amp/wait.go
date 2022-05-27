package amp

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/prometheusservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	// Maximum amount of time to wait for a Workspace to be created, updated, or deleted
	workspaceTimeout = 5 * time.Minute
)

func waitAlertManagerDefinitionCreated(ctx context.Context, conn *prometheusservice.PrometheusService, id string) (*prometheusservice.AlertManagerDefinitionDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{prometheusservice.AlertManagerDefinitionStatusCodeCreating},
		Target:  []string{prometheusservice.AlertManagerDefinitionStatusCodeActive},
		Refresh: statusAlertManagerDefinition(ctx, conn, id),
		Timeout: workspaceTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*prometheusservice.AlertManagerDefinitionDescription); ok {
		if statusCode := aws.StringValue(output.Status.StatusCode); statusCode == prometheusservice.AlertManagerDefinitionStatusCodeCreationFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.Status.StatusReason)))
		}

		return output, err
	}

	return nil, err
}

func waitAlertManagerDefinitionUpdated(ctx context.Context, conn *prometheusservice.PrometheusService, id string) (*prometheusservice.AlertManagerDefinitionDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{prometheusservice.AlertManagerDefinitionStatusCodeUpdating},
		Target:  []string{prometheusservice.AlertManagerDefinitionStatusCodeActive},
		Refresh: statusAlertManagerDefinition(ctx, conn, id),
		Timeout: workspaceTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*prometheusservice.AlertManagerDefinitionDescription); ok {
		if statusCode := aws.StringValue(output.Status.StatusCode); statusCode == prometheusservice.AlertManagerDefinitionStatusCodeUpdateFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.Status.StatusReason)))
		}

		return output, err
	}

	return nil, err
}

func waitAlertManagerDefinitionDeleted(ctx context.Context, conn *prometheusservice.PrometheusService, id string) (*prometheusservice.AlertManagerDefinitionDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{prometheusservice.AlertManagerDefinitionStatusCodeDeleting},
		Target:  []string{},
		Refresh: statusAlertManagerDefinition(ctx, conn, id),
		Timeout: workspaceTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*prometheusservice.AlertManagerDefinitionDescription); ok {
		return output, err
	}

	return nil, err
}

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

func waitRuleGroupNamespaceDeleted(ctx context.Context, conn *prometheusservice.PrometheusService, id string) (*prometheusservice.RuleGroupsNamespaceDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{prometheusservice.RuleGroupsNamespaceStatusCodeDeleting},
		Target:  []string{},
		Refresh: statusRuleGroupNamespace(ctx, conn, id),
		Timeout: workspaceTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*prometheusservice.RuleGroupsNamespaceDescription); ok {
		return output, err
	}

	return nil, err
}

func waitRuleGroupNamespaceCreated(ctx context.Context, conn *prometheusservice.PrometheusService, id string) (*prometheusservice.RuleGroupsNamespaceDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{prometheusservice.RuleGroupsNamespaceStatusCodeCreating},
		Target:  []string{prometheusservice.RuleGroupsNamespaceStatusCodeActive},
		Refresh: statusRuleGroupNamespace(ctx, conn, id),
		Timeout: workspaceTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*prometheusservice.RuleGroupsNamespaceDescription); ok {
		return output, err
	}

	return nil, err
}

func waitRuleGroupNamespaceUpdated(ctx context.Context, conn *prometheusservice.PrometheusService, id string) (*prometheusservice.RuleGroupsNamespaceDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{prometheusservice.RuleGroupsNamespaceStatusCodeUpdating},
		Target:  []string{prometheusservice.RuleGroupsNamespaceStatusCodeActive},
		Refresh: statusRuleGroupNamespace(ctx, conn, id),
		Timeout: workspaceTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*prometheusservice.RuleGroupsNamespaceDescription); ok {
		return output, err
	}

	return nil, err
}
