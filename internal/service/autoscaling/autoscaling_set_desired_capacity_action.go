// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package autoscaling

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	awstypes "github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
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

// @Action(aws_autoscaling_set_desired_capacity, name="Set Desired Capacity")
func newSetDesiredCapacityAction(_ context.Context) (action.ActionWithConfigure, error) {
	return &setDesiredCapacityAction{}, nil
}

type setDesiredCapacityAction struct {
	framework.ActionWithModel[setDesiredCapacityModel]
}

type setDesiredCapacityModel struct {
	framework.WithRegionModel
	AutoScalingGroupName types.String `tfsdk:"autoscaling_group_name"`
	DesiredCapacity      types.Int64  `tfsdk:"desired_capacity"`
	MinSize              types.Int64  `tfsdk:"min_size"`
	MaxSize              types.Int64  `tfsdk:"max_size"`
	Timeout              types.Int64  `tfsdk:"timeout"`
}

func (a *setDesiredCapacityAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Sets the desired capacity of an Auto Scaling group and optionally updates min/max sizes.",
		Attributes: map[string]schema.Attribute{
			"autoscaling_group_name": schema.StringAttribute{
				Description: "The name of the Auto Scaling group",
				Required:    true,
			},
			"desired_capacity": schema.Int64Attribute{
				Description: "The desired capacity for the Auto Scaling group",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"min_size": schema.Int64Attribute{
				Description: "The minimum size of the Auto Scaling group",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"max_size": schema.Int64Attribute{
				Description: "The maximum size of the Auto Scaling group",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			names.AttrTimeout: schema.Int64Attribute{
				Description: "Timeout in seconds (default: 900)",
				Optional:    true,
				Validators: []validator.Int64{
					int64validator.Between(60, 3600),
				},
			},
		},
	}
}

func (a *setDesiredCapacityAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var config setDesiredCapacityModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := a.Meta().AutoScalingClient(ctx)

	asgName := config.AutoScalingGroupName.ValueString()
	desiredCapacity := config.DesiredCapacity.ValueInt64()

	timeout := 900 * time.Second
	if !config.Timeout.IsNull() {
		timeout = time.Duration(config.Timeout.ValueInt64()) * time.Second
	}

	tflog.Info(ctx, "Starting Auto Scaling set desired capacity action", map[string]any{
		"autoscaling_group_name": asgName,
		"desired_capacity":       desiredCapacity,
		names.AttrTimeout:        timeout.String(),
	})

	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Setting desired capacity for Auto Scaling group %s to %d...", asgName, desiredCapacity),
	})

	// Get current ASG to validate bounds
	asg, err := findAutoScalingGroupByName(ctx, conn, asgName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Auto Scaling Group Not Found",
			fmt.Sprintf("Auto Scaling group %s was not found: %s", asgName, err),
		)
		return
	}

	// Validate capacity bounds
	minSize := aws.ToInt32(asg.MinSize)
	maxSize := aws.ToInt32(asg.MaxSize)

	if !config.MinSize.IsNull() {
		minSize = int32(config.MinSize.ValueInt64())
	}
	if !config.MaxSize.IsNull() {
		maxSize = int32(config.MaxSize.ValueInt64())
	}

	if int64(minSize) > desiredCapacity || desiredCapacity > int64(maxSize) {
		resp.Diagnostics.AddError(
			"Invalid Capacity",
			fmt.Sprintf("Desired capacity %d must be between min_size %d and max_size %d", desiredCapacity, minSize, maxSize),
		)
		return
	}

	// Update min/max if specified
	if !config.MinSize.IsNull() || !config.MaxSize.IsNull() {
		updateInput := &autoscaling.UpdateAutoScalingGroupInput{
			AutoScalingGroupName: aws.String(asgName),
		}
		if !config.MinSize.IsNull() {
			updateInput.MinSize = aws.Int32(int32(config.MinSize.ValueInt64()))
		}
		if !config.MaxSize.IsNull() {
			updateInput.MaxSize = aws.Int32(int32(config.MaxSize.ValueInt64()))
		}

		_, err = conn.UpdateAutoScalingGroup(ctx, updateInput)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to Update Auto Scaling Group",
				fmt.Sprintf("Could not update Auto Scaling group %s: %s", asgName, err),
			)
			return
		}
	}

	// Set desired capacity
	setCapacityInput := &autoscaling.SetDesiredCapacityInput{
		AutoScalingGroupName: aws.String(asgName),
		DesiredCapacity:      aws.Int32(int32(desiredCapacity)),
		HonorCooldown:        aws.Bool(false),
	}

	_, err = conn.SetDesiredCapacity(ctx, setCapacityInput)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Set Desired Capacity",
			fmt.Sprintf("Could not set desired capacity for Auto Scaling group %s: %s", asgName, err),
		)
		return
	}

	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Waiting for Auto Scaling group %s to reach desired capacity %d...", asgName, desiredCapacity),
	})

	// Wait for desired capacity to be reached
	_, err = actionwait.WaitForStatus(ctx, func(ctx context.Context) (actionwait.FetchResult[struct{}], error) {
		asg, gerr := findAutoScalingGroupByName(ctx, conn, asgName)
		if gerr != nil {
			return actionwait.FetchResult[struct{}]{}, fmt.Errorf("describing Auto Scaling group: %w", gerr)
		}

		// Count instances in service
		inServiceCount := 0
		for _, instance := range asg.Instances {
			if instance.LifecycleState == awstypes.LifecycleStateInService {
				inServiceCount++
			}
		}

		if int64(inServiceCount) == desiredCapacity {
			return actionwait.FetchResult[struct{}]{Status: "READY"}, nil
		}
		return actionwait.FetchResult[struct{}]{Status: "SCALING"}, nil
	}, actionwait.Options[struct{}]{
		Timeout:          timeout,
		Interval:         actionwait.FixedInterval(15 * time.Second),
		ProgressInterval: 30 * time.Second,
		SuccessStates:    []actionwait.Status{"READY"},
		TransitionalStates: []actionwait.Status{"SCALING"},
		ProgressSink: func(fr actionwait.FetchResult[any], meta actionwait.ProgressMeta) {
			resp.SendProgress(action.InvokeProgressEvent{
				Message: fmt.Sprintf("Auto Scaling group %s is scaling to desired capacity %d...", asgName, desiredCapacity),
			})
		},
	})

	if err != nil {
		var timeoutErr *actionwait.TimeoutError
		if errors.As(err, &timeoutErr) {
			resp.Diagnostics.AddError(
				"Timeout Waiting for Desired Capacity",
				fmt.Sprintf("Auto Scaling group %s did not reach desired capacity %d within %s", asgName, desiredCapacity, timeout),
			)
		} else {
			resp.Diagnostics.AddError(
				"Error Waiting for Desired Capacity",
				fmt.Sprintf("Error while waiting for Auto Scaling group %s: %s", asgName, err),
			)
		}
		return
	}

	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Auto Scaling group %s successfully reached desired capacity %d", asgName, desiredCapacity),
	})

	tflog.Info(ctx, "Auto Scaling set desired capacity action completed successfully", map[string]any{
		"autoscaling_group_name": asgName,
		"desired_capacity":       desiredCapacity,
	})
}

func findAutoScalingGroupByName(ctx context.Context, conn *autoscaling.Client, name string) (*awstypes.AutoScalingGroup, error) {
	input := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []string{name},
	}

	output, err := conn.DescribeAutoScalingGroups(ctx, input)
	if err != nil {
		return nil, err
	}

	if len(output.AutoScalingGroups) == 0 {
		return nil, fmt.Errorf("Auto Scaling group %s not found", name)
	}

	return &output.AutoScalingGroups[0], nil
}
