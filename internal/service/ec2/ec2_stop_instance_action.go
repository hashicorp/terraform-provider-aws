// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
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

// stopInstancePollInterval defines polling cadence for stop instance action.
const stopInstancePollInterval = 10 * time.Second

// @Action(aws_ec2_stop_instance, name="Stop Instance")
func newStopInstanceAction(_ context.Context) (action.ActionWithConfigure, error) {
	return &stopInstanceAction{}, nil
}

var (
	_ action.Action = (*stopInstanceAction)(nil)
)

type stopInstanceAction struct {
	framework.ActionWithModel[stopInstanceModel]
}

type stopInstanceModel struct {
	framework.WithRegionModel
	InstanceID types.String `tfsdk:"instance_id"`
	Force      types.Bool   `tfsdk:"force"`
	Timeout    types.Int64  `tfsdk:"timeout"`
}

func (a *stopInstanceAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Stops an EC2 instance. This action will gracefully stop the instance and wait for it to reach the stopped state.",
		Attributes: map[string]schema.Attribute{
			names.AttrInstanceID: schema.StringAttribute{
				Description: "The ID of the EC2 instance to stop",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexache.MustCompile(`^i-[0-9a-f]{8,17}$`),
						"must be a valid EC2 instance ID (e.g., i-1234567890abcdef0)",
					),
				},
			},
			"force": schema.BoolAttribute{
				Description: "Forces the instance to stop. The instance does not have an opportunity to flush file system caches or file system metadata. If you use this option, you must perform file system check and repair procedures. This option is not recommended for Windows instances.",
				Optional:    true,
			},
			names.AttrTimeout: schema.Int64Attribute{
				Description: "Timeout in seconds to wait for the instance to stop (default: 600)",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(30),
					int64validator.AtMost(3600),
				},
			},
		},
	}
}

func (a *stopInstanceAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var config stopInstanceModel

	// Parse configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get AWS client
	conn := a.Meta().EC2Client(ctx)

	instanceID := config.InstanceID.ValueString()
	force := config.Force.ValueBool()

	// Set default timeout if not provided
	timeout := 600 * time.Second
	if !config.Timeout.IsNull() {
		timeout = time.Duration(config.Timeout.ValueInt64()) * time.Second
	}

	tflog.Info(ctx, "Starting EC2 stop instance action", map[string]any{
		names.AttrInstanceID: instanceID,
		"force":              force,
		names.AttrTimeout:    timeout.String(),
	})

	// Send initial progress update
	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Starting stop operation for EC2 instance %s...", instanceID),
	})

	// Check current instance state first
	instance, err := findInstanceByID(ctx, conn, instanceID)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, errCodeInvalidInstanceIDNotFound) {
			resp.Diagnostics.AddError(
				"Instance Not Found",
				fmt.Sprintf("EC2 instance %s was not found", instanceID),
			)
			return
		}
		resp.Diagnostics.AddError(
			"Failed to Describe Instance",
			fmt.Sprintf("Could not describe EC2 instance %s: %s", instanceID, err),
		)
		return
	}

	currentState := string(instance.State.Name)
	tflog.Debug(ctx, "Current instance state", map[string]any{
		names.AttrInstanceID: instanceID,
		names.AttrState:      currentState,
	})

	// Check if instance is already stopped
	if instance.State.Name == awstypes.InstanceStateNameStopped {
		resp.SendProgress(action.InvokeProgressEvent{
			Message: fmt.Sprintf("EC2 instance %s is already stopped", instanceID),
		})
		tflog.Info(ctx, "Instance already stopped", map[string]any{
			names.AttrInstanceID: instanceID,
		})
		return
	}

	// Check if instance is in a state that can be stopped
	if !canStopInstance(instance.State.Name) {
		resp.Diagnostics.AddError(
			"Cannot Stop Instance",
			fmt.Sprintf("EC2 instance %s is in state '%s' and cannot be stopped. Instance must be in 'running' or 'stopping' state.", instanceID, currentState),
		)
		return
	}

	// If instance is already stopping, just wait for it
	if instance.State.Name == awstypes.InstanceStateNameStopping {
		resp.SendProgress(action.InvokeProgressEvent{
			Message: fmt.Sprintf("EC2 instance %s is already stopping, waiting for completion...", instanceID),
		})
	} else {
		// Stop the instance
		resp.SendProgress(action.InvokeProgressEvent{
			Message: fmt.Sprintf("Sending stop command to EC2 instance %s...", instanceID),
		})

		input := ec2.StopInstancesInput{
			Force:       aws.Bool(force),
			InstanceIds: []string{instanceID},
		}

		_, err = conn.StopInstances(ctx, &input)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to Stop Instance",
				fmt.Sprintf("Could not stop EC2 instance %s: %s", instanceID, err),
			)
			return
		}

		resp.SendProgress(action.InvokeProgressEvent{
			Message: fmt.Sprintf("Stop command sent to EC2 instance %s, waiting for instance to stop...", instanceID),
		})
	}

	// Wait for instance to stop with periodic progress updates using actionwait
	// Use fixed interval since EC2 instance state transitions are predictable and
	// relatively quick - consistent polling every 10s is optimal for this operation
	_, err = actionwait.WaitForStatus(ctx, func(ctx context.Context) (actionwait.FetchResult[struct{}], error) {
		instance, derr := findInstanceByID(ctx, conn, instanceID)
		if derr != nil {
			return actionwait.FetchResult[struct{}]{}, fmt.Errorf("describing instance: %w", derr)
		}
		state := string(instance.State.Name)
		return actionwait.FetchResult[struct{}]{Status: actionwait.Status(state)}, nil
	}, actionwait.Options[struct{}]{
		Timeout:          timeout,
		Interval:         actionwait.FixedInterval(stopInstancePollInterval),
		ProgressInterval: 30 * time.Second,
		SuccessStates:    []actionwait.Status{actionwait.Status(awstypes.InstanceStateNameStopped)},
		TransitionalStates: []actionwait.Status{
			actionwait.Status(awstypes.InstanceStateNameRunning),
			actionwait.Status(awstypes.InstanceStateNameStopping),
			actionwait.Status(awstypes.InstanceStateNameShuttingDown),
		},
		ProgressSink: func(fr actionwait.FetchResult[any], meta actionwait.ProgressMeta) {
			resp.SendProgress(action.InvokeProgressEvent{Message: fmt.Sprintf("EC2 instance %s is currently in state '%s', continuing to wait for 'stopped'...", instanceID, fr.Status)})
		},
	})
	if err != nil {
		var timeoutErr *actionwait.TimeoutError
		var unexpectedErr *actionwait.UnexpectedStateError
		if errors.As(err, &timeoutErr) {
			resp.Diagnostics.AddError(
				"Timeout Waiting for Instance to Stop",
				fmt.Sprintf("EC2 instance %s did not stop within %s: %s", instanceID, timeout, err),
			)
		} else if errors.As(err, &unexpectedErr) {
			resp.Diagnostics.AddError(
				"Unexpected Instance State",
				fmt.Sprintf("EC2 instance %s entered unexpected state while stopping: %s", instanceID, err),
			)
		} else {
			resp.Diagnostics.AddError(
				"Error Waiting for Instance to Stop",
				fmt.Sprintf("Error while waiting for EC2 instance %s to stop: %s", instanceID, err),
			)
		}
		return
	}

	// Final success message
	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("EC2 instance %s has been successfully stopped", instanceID),
	})

	tflog.Info(ctx, "EC2 stop instance action completed successfully", map[string]any{
		names.AttrInstanceID: instanceID,
	})
}

// canStopInstance checks if an instance can be stopped based on its current state
func canStopInstance(state awstypes.InstanceStateName) bool {
	switch state {
	case awstypes.InstanceStateNameRunning, awstypes.InstanceStateNameStopping:
		return true
	default:
		return false
	}
}
