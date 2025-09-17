// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"slices"
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
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

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

	// Wait for instance to stop with periodic progress updates
	err = a.waitForInstanceStopped(ctx, conn, instanceID, timeout, resp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Timeout Waiting for Instance to Stop",
			fmt.Sprintf("EC2 instance %s did not stop within %s: %s", instanceID, timeout, err),
		)
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

// waitForInstanceStopped waits for an instance to reach the stopped state with progress updates
func (a *stopInstanceAction) waitForInstanceStopped(ctx context.Context, conn *ec2.Client, instanceID string, timeout time.Duration, resp *action.InvokeResponse) error {
	const (
		pollInterval     = 10 * time.Second
		progressInterval = 30 * time.Second
	)

	deadline := time.Now().Add(timeout)
	lastProgressUpdate := time.Now()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Check if we've exceeded the timeout
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout after %s", timeout)
		}

		// Get current instance state
		instance, err := findInstanceByID(ctx, conn, instanceID)
		if err != nil {
			return fmt.Errorf("describing instance: %w", err)
		}

		currentState := string(instance.State.Name)

		// Send progress update every 30 seconds
		if time.Since(lastProgressUpdate) >= progressInterval {
			resp.SendProgress(action.InvokeProgressEvent{
				Message: fmt.Sprintf("EC2 instance %s is currently in state '%s', continuing to wait for 'stopped'...", instanceID, currentState),
			})
			lastProgressUpdate = time.Now()
		}

		// Check if we've reached the target state
		if instance.State.Name == awstypes.InstanceStateNameStopped {
			return nil
		}

		// Check if we're in an unexpected state
		validStates := []awstypes.InstanceStateName{
			awstypes.InstanceStateNameRunning,
			awstypes.InstanceStateNameStopping,
			awstypes.InstanceStateNameShuttingDown,
		}
		if !slices.Contains(validStates, instance.State.Name) {
			return fmt.Errorf("instance entered unexpected state: %s", currentState)
		}

		// Wait before next poll
		time.Sleep(pollInterval)
	}
}
