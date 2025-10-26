// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/actionwait"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @Action(aws_rds_reboot_instance, name="Reboot DB Instance")
func newRebootInstanceAction(_ context.Context) (action.ActionWithConfigure, error) {
	return &rebootInstanceAction{}, nil
}

type rebootInstanceAction struct {
	framework.ActionWithModel[rebootInstanceModel]
}

type rebootInstanceModel struct {
	framework.WithRegionModel
	DBInstanceIdentifier types.String `tfsdk:"db_instance_identifier"`
	ForceFailover        types.Bool   `tfsdk:"force_failover"`
	Timeout              types.Int64  `tfsdk:"timeout"`
}

func (a *rebootInstanceAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reboots an Amazon RDS database instance to apply parameter changes or perform maintenance.",
		Attributes: map[string]schema.Attribute{
			"db_instance_identifier": schema.StringAttribute{
				Description: "The DB instance identifier of the RDS instance to reboot.",
				Required:    true,
			},
			"force_failover": schema.BoolAttribute{
				Description: "When true, the reboot is conducted through a Multi-AZ failover. This is only applicable for Multi-AZ instances.",
				Optional:    true,
			},
			names.AttrTimeout: schema.Int64Attribute{
				Description: "Timeout in seconds to wait for the reboot to complete (300-3600, default: 1800).",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.Between(300, 3600),
				},
			},
		},
	}
}

func (a *rebootInstanceAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var config rebootInstanceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := a.Meta().RDSClient(ctx)

	dbInstanceId := config.DBInstanceIdentifier.ValueString()
	forceFailover := config.ForceFailover.ValueBool()

	timeout := 1800 * time.Second
	if !config.Timeout.IsNull() {
		timeout = time.Duration(config.Timeout.ValueInt64()) * time.Second
	}

	tflog.Info(ctx, "Starting RDS reboot instance action", map[string]any{
		"db_instance_identifier": dbInstanceId,
		"force_failover":         forceFailover,
		names.AttrTimeout:        timeout.String(),
	})

	// If force_failover is requested, validate the instance is Multi-AZ
	if forceFailover {
		resp.SendProgress(action.InvokeProgressEvent{
			Message: fmt.Sprintf("Validating Multi-AZ configuration for RDS instance %s...", dbInstanceId),
		})

		describeInput := &rds.DescribeDBInstancesInput{
			DBInstanceIdentifier: aws.String(dbInstanceId),
		}
		describeOutput, err := conn.DescribeDBInstances(ctx, describeInput)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to Describe DB Instance",
				fmt.Sprintf("Could not describe RDS instance %s: %s", dbInstanceId, err),
			)
			return
		}

		if len(describeOutput.DBInstances) == 0 {
			resp.Diagnostics.AddError(
				"DB Instance Not Found",
				fmt.Sprintf("RDS instance %s was not found", dbInstanceId),
			)
			return
		}

		instance := describeOutput.DBInstances[0]
		if !aws.ToBool(instance.MultiAZ) {
			resp.Diagnostics.AddError(
				"Invalid Force Failover Request",
				fmt.Sprintf("Cannot force failover for RDS instance %s because it is not configured for Multi-AZ", dbInstanceId),
			)
			return
		}
	}

	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Rebooting RDS instance %s...", dbInstanceId),
	})

	// Reboot the instance
	input := &rds.RebootDBInstanceInput{
		DBInstanceIdentifier: aws.String(dbInstanceId),
		ForceFailover:        aws.Bool(forceFailover),
	}

	_, err := conn.RebootDBInstance(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Reboot DB Instance",
			fmt.Sprintf("Could not reboot RDS instance %s: %s", dbInstanceId, err),
		)
		return
	}

	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Reboot initiated for RDS instance %s, waiting for completion...", dbInstanceId),
	})

	// Wait for instance to return to available state
	_, err = actionwait.WaitForStatus(ctx, func(ctx context.Context) (actionwait.FetchResult[struct{}], error) {
		input := rds.DescribeDBInstancesInput{
			DBInstanceIdentifier: aws.String(dbInstanceId),
		}
		output, err := conn.DescribeDBInstances(ctx, &input)
		if err != nil {
			return actionwait.FetchResult[struct{}]{}, fmt.Errorf("describing instance: %w", err)
		}
		if len(output.DBInstances) == 0 {
			return actionwait.FetchResult[struct{}]{}, fmt.Errorf("instance %s not found", dbInstanceId)
		}
		status := aws.ToString(output.DBInstances[0].DBInstanceStatus)
		return actionwait.FetchResult[struct{}]{Status: actionwait.Status(status)}, nil
	}, actionwait.Options[struct{}]{
		Timeout:          timeout,
		Interval:         actionwait.FixedInterval(30 * time.Second),
		ProgressInterval: 2 * time.Minute,
		SuccessStates:    []actionwait.Status{"available"},
		TransitionalStates: []actionwait.Status{
			"rebooting",
			"modifying",
		},
		FailureStates: []actionwait.Status{
			"failed",
			"incompatible-parameters",
			"incompatible-restore",
		},
		ProgressSink: func(fr actionwait.FetchResult[any], meta actionwait.ProgressMeta) {
			resp.SendProgress(action.InvokeProgressEvent{
				Message: fmt.Sprintf("RDS instance %s is currently '%s', continuing to wait...", dbInstanceId, fr.Status),
			})
		},
	})

	if err != nil {
		var timeoutErr *actionwait.TimeoutError
		var failureErr *actionwait.FailureStateError
		var unexpectedErr *actionwait.UnexpectedStateError

		if errors.As(err, &timeoutErr) {
			resp.Diagnostics.AddError(
				"Timeout Rebooting DB Instance",
				fmt.Sprintf("RDS instance %s did not return to available state within %s", dbInstanceId, timeout),
			)
		} else if errors.As(err, &failureErr) {
			resp.Diagnostics.AddError(
				"DB Instance Reboot Failed",
				fmt.Sprintf("RDS instance %s reboot failed with status: %s", dbInstanceId, failureErr.Status),
			)
		} else if errors.As(err, &unexpectedErr) {
			resp.Diagnostics.AddError(
				"Unexpected Instance Status",
				fmt.Sprintf("RDS instance %s entered unexpected status: %s", dbInstanceId, unexpectedErr.Status),
			)
		} else {
			resp.Diagnostics.AddError(
				"Error Rebooting DB Instance",
				fmt.Sprintf("Error while rebooting instance %s: %s", dbInstanceId, err),
			)
		}
		return
	}

	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("RDS instance %s rebooted successfully and is now available", dbInstanceId),
	})

	tflog.Info(ctx, "RDS reboot instance action completed successfully", map[string]any{
		"db_instance_identifier": dbInstanceId,
		"force_failover":         forceFailover,
	})
}
