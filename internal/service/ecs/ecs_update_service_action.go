// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/actionwait"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const updateServicePollInterval = 15 * time.Second

// @Action(aws_ecs_update_service, name="Update Service")
func newUpdateServiceAction(_ context.Context) (action.ActionWithConfigure, error) {
	return &updateServiceAction{}, nil
}

type updateServiceAction struct {
	framework.ActionWithModel[updateServiceModel]
}

type updateServiceModel struct {
	framework.WithRegionModel
	ClusterName                     types.String `tfsdk:"cluster_name"`
	ServiceName                     types.String `tfsdk:"service_name"`
	TaskDefinition                  types.String `tfsdk:"task_definition"`
	DesiredCount                    types.Int64  `tfsdk:"desired_count"`
	ForceNewDeployment              types.Bool   `tfsdk:"force_new_deployment"`
	HealthCheckGracePeriodSeconds   types.Int64  `tfsdk:"health_check_grace_period_seconds"`
	Timeout                         types.Int64  `tfsdk:"timeout"`
}

func (a *updateServiceAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Updates an Amazon ECS service to force new deployments, update task definitions, or modify service configuration.",
		Attributes: map[string]schema.Attribute{
			"cluster_name": schema.StringAttribute{
				Description: "The name or ARN of the ECS cluster containing the service.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexache.MustCompile(`^[a-zA-Z0-9\-_]+$|^arn:aws:ecs:[a-z0-9\-]+:[0-9]{12}:cluster/[a-zA-Z0-9\-_]+$`),
						"must be a valid cluster name or ARN",
					),
				},
			},
			"service_name": schema.StringAttribute{
				Description: "The name or ARN of the ECS service to update.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexache.MustCompile(`^[a-zA-Z0-9\-_]+$|^arn:aws:ecs:[a-z0-9\-]+:[0-9]{12}:service/[a-zA-Z0-9\-_/]+$`),
						"must be a valid service name or ARN",
					),
				},
			},
			"task_definition": schema.StringAttribute{
				Description: "The task definition ARN to update the service to use.",
				Optional:    true,
			},
			"desired_count": schema.Int64Attribute{
				Description: "The desired number of tasks to run in the service.",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"force_new_deployment": schema.BoolAttribute{
				Description: "Force a new deployment of the service without changing the task definition or desired count.",
				Optional:    true,
			},
			"health_check_grace_period_seconds": schema.Int64Attribute{
				Description: "The period of time that the ECS service scheduler ignores unhealthy load balancer target health checks after a task has first started (0-2147483647).",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.Between(0, 2147483647),
				},
			},
			names.AttrTimeout: schema.Int64Attribute{
				Description: "Timeout in seconds to wait for the service update to complete (60-3600, default: 1200).",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.Between(60, 3600),
				},
			},
		},
	}
}

func (a *updateServiceAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var config updateServiceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := a.Meta().ECSClient(ctx)

	clusterName := config.ClusterName.ValueString()
	serviceName := config.ServiceName.ValueString()

	timeout := 1200 * time.Second
	if !config.Timeout.IsNull() {
		timeout = time.Duration(config.Timeout.ValueInt64()) * time.Second
	}

	tflog.Info(ctx, "Starting ECS update service action", map[string]any{
		"cluster_name": clusterName,
		"service_name": serviceName,
		names.AttrTimeout: timeout.String(),
	})

	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Starting update for ECS service %s in cluster %s...", serviceName, clusterName),
	})

	// Check current service state
	service, err := findServiceByName(ctx, conn, clusterName, serviceName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Service Not Found",
			fmt.Sprintf("ECS service %s was not found in cluster %s: %s", serviceName, clusterName, err),
		)
		return
	}

	// Wait for existing deployment if active
	if hasActiveDeployment(service) {
		resp.SendProgress(action.InvokeProgressEvent{
			Message: fmt.Sprintf("ECS service %s has an active deployment, waiting for completion...", serviceName),
		})
		
		_, err = waitForServiceStable(ctx, conn, clusterName, serviceName, timeout)
		if err != nil {
			resp.Diagnostics.AddError(
				"Existing Deployment Failed",
				fmt.Sprintf("Existing deployment for service %s failed: %s", serviceName, err),
			)
			return
		}
	}

	// Build update input
	input := &ecs.UpdateServiceInput{
		Cluster: aws.String(clusterName),
		Service: aws.String(serviceName),
	}

	if !config.TaskDefinition.IsNull() {
		input.TaskDefinition = config.TaskDefinition.ValueStringPointer()
	}

	if !config.DesiredCount.IsNull() {
		input.DesiredCount = aws.Int32(int32(config.DesiredCount.ValueInt64()))
	}

	if !config.ForceNewDeployment.IsNull() {
		input.ForceNewDeployment = config.ForceNewDeployment.ValueBool()
	}

	if !config.HealthCheckGracePeriodSeconds.IsNull() {
		input.HealthCheckGracePeriodSeconds = aws.Int32(int32(config.HealthCheckGracePeriodSeconds.ValueInt64()))
	}

	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Sending update request to ECS service %s...", serviceName),
	})

	// Update service
	_, err = conn.UpdateService(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Update Service",
			fmt.Sprintf("Could not update ECS service %s: %s", serviceName, err),
		)
		return
	}

	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("ECS service %s update initiated, waiting for deployment to stabilize...", serviceName),
	})

	// Wait for deployment to complete
	_, err = actionwait.WaitForStatus(ctx, func(ctx context.Context) (actionwait.FetchResult[struct{}], error) {
		service, err := findServiceByName(ctx, conn, clusterName, serviceName)
		if err != nil {
			return actionwait.FetchResult[struct{}]{}, fmt.Errorf("describing service: %w", err)
		}
		
		status := getDeploymentStatus(service)
		return actionwait.FetchResult[struct{}]{Status: actionwait.Status(status)}, nil
	}, actionwait.Options[struct{}]{
		Timeout:          timeout,
		Interval:         actionwait.FixedInterval(updateServicePollInterval),
		ProgressInterval: 30 * time.Second,
		SuccessStates:    []actionwait.Status{"STABLE"},
		TransitionalStates: []actionwait.Status{"UPDATING", "DRAINING"},
		FailureStates:    []actionwait.Status{"FAILED"},
		ProgressSink: func(fr actionwait.FetchResult[any], meta actionwait.ProgressMeta) {
			resp.SendProgress(action.InvokeProgressEvent{
				Message: fmt.Sprintf("ECS service %s deployment in progress (status: %s)", serviceName, fr.Status),
			})
		},
	})

	if err != nil {
		var timeoutErr *actionwait.TimeoutError
		var failureErr *actionwait.FailureStateError
		var unexpectedErr *actionwait.UnexpectedStateError

		if errors.As(err, &timeoutErr) {
			resp.Diagnostics.AddError(
				"Timeout Waiting for Service Update",
				fmt.Sprintf("ECS service %s did not reach stable state within %s", serviceName, timeout),
			)
		} else if errors.As(err, &failureErr) {
			resp.Diagnostics.AddError(
				"Service Update Failed",
				fmt.Sprintf("ECS service %s deployment failed with status: %s", serviceName, failureErr.Status),
			)
		} else if errors.As(err, &unexpectedErr) {
			resp.Diagnostics.AddError(
				"Unexpected Service Status",
				fmt.Sprintf("ECS service %s entered unexpected status: %s", serviceName, unexpectedErr.Status),
			)
		} else {
			resp.Diagnostics.AddError(
				"Error Updating Service",
				fmt.Sprintf("Error while updating ECS service %s: %s", serviceName, err),
			)
		}
		return
	}

	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("ECS service %s has been successfully updated and is stable", serviceName),
	})

	tflog.Info(ctx, "ECS update service action completed successfully", map[string]any{
		"cluster_name": clusterName,
		"service_name": serviceName,
	})
}

// Helper functions
func findServiceByName(ctx context.Context, conn *ecs.Client, cluster, service string) (*awstypes.Service, error) {
	input := &ecs.DescribeServicesInput{
		Cluster:  aws.String(cluster),
		Services: []string{service},
	}

	output, err := conn.DescribeServices(ctx, input)
	if err != nil {
		return nil, err
	}

	if len(output.Services) == 0 {
		return nil, fmt.Errorf("service not found")
	}

	return &output.Services[0], nil
}

func hasActiveDeployment(service *awstypes.Service) bool {
	for _, deployment := range service.Deployments {
		if deployment.Status != nil && *deployment.Status == "PRIMARY" && deployment.RolloutState == awstypes.DeploymentRolloutStateInProgress {
			return true
		}
	}
	return false
}

func getDeploymentStatus(service *awstypes.Service) string {
	if isServiceStable(service) {
		return "STABLE"
	}
	
	for _, deployment := range service.Deployments {
		if deployment.Status != nil && *deployment.Status == "PRIMARY" {
			switch deployment.RolloutState {
			case awstypes.DeploymentRolloutStateInProgress:
				return "UPDATING"
			case awstypes.DeploymentRolloutStateFailed:
				return "FAILED"
			}
		}
	}
	
	return "UPDATING"
}

func isServiceStable(service *awstypes.Service) bool {
	if len(service.Deployments) == 0 {
		return false
	}

	primaryDeployment := service.Deployments[0]
	return primaryDeployment.RolloutState == awstypes.DeploymentRolloutStateCompleted &&
		primaryDeployment.RunningCount == primaryDeployment.DesiredCount
}

func waitForServiceStable(ctx context.Context, conn *ecs.Client, cluster, service string, timeout time.Duration) (*awstypes.Service, error) {
	fr, err := actionwait.WaitForStatus(ctx, func(ctx context.Context) (actionwait.FetchResult[*awstypes.Service], error) {
		svc, err := findServiceByName(ctx, conn, cluster, service)
		if err != nil {
			return actionwait.FetchResult[*awstypes.Service]{}, err
		}
		
		status := getDeploymentStatus(svc)
		return actionwait.FetchResult[*awstypes.Service]{Status: actionwait.Status(status), Value: svc}, nil
	}, actionwait.Options[*awstypes.Service]{
		Timeout:          timeout,
		Interval:         actionwait.FixedInterval(updateServicePollInterval),
		SuccessStates:    []actionwait.Status{"STABLE"},
		TransitionalStates: []actionwait.Status{"UPDATING", "DRAINING"},
		FailureStates:    []actionwait.Status{"FAILED"},
	})

	if err != nil {
		return nil, err
	}

	return fr.Value, nil
}
