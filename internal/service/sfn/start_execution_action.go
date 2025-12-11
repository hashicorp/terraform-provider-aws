// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package sfn

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @Action(aws_sfn_start_execution, name="Start Execution")
func newStartExecutionAction(_ context.Context) (action.ActionWithConfigure, error) {
	return &startExecutionAction{}, nil
}

var (
	_ action.Action = (*startExecutionAction)(nil)
)

type startExecutionAction struct {
	framework.ActionWithModel[startExecutionActionModel]
}

type startExecutionActionModel struct {
	framework.WithRegionModel
	StateMachineArn types.String `tfsdk:"state_machine_arn"`
	Input           types.String `tfsdk:"input"`
	Name            types.String `tfsdk:"name"`
	TraceHeader     types.String `tfsdk:"trace_header"`
}

func (a *startExecutionAction) Schema(ctx context.Context, req action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Starts a Step Functions state machine execution with the specified input data.",
		Attributes: map[string]schema.Attribute{
			"state_machine_arn": schema.StringAttribute{
				Description: "The ARN of the state machine to execute. Can be unqualified, version-qualified, or alias-qualified.",
				Required:    true,
			},
			"input": schema.StringAttribute{
				Description: "JSON input data for the execution. Defaults to '{}'.",
				Optional:    true,
				Validators: []validator.String{
					validators.JSON(),
				},
			},
			names.AttrName: schema.StringAttribute{
				Description: "Name of the execution. Must be unique within the account/region/state machine for 90 days. Auto-generated if not provided.",
				Optional:    true,
			},
			"trace_header": schema.StringAttribute{
				Description: "AWS X-Ray trace header for distributed tracing.",
				Optional:    true,
			},
		},
	}
}

func (a *startExecutionAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var config startExecutionActionModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := a.Meta().SFNClient(ctx)

	stateMachineArn := config.StateMachineArn.ValueString()
	input := "{}"
	if !config.Input.IsNull() {
		input = config.Input.ValueString()
	}

	tflog.Info(ctx, "Starting Step Functions execution", map[string]any{
		"state_machine_arn": stateMachineArn,
		"input_length":      len(input),
		"has_name":          !config.Name.IsNull(),
		"has_trace_header":  !config.TraceHeader.IsNull(),
	})

	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Starting execution for state machine %s...", stateMachineArn),
	})

	startInput := &sfn.StartExecutionInput{
		StateMachineArn: aws.String(stateMachineArn),
		Input:           aws.String(input),
	}

	if !config.Name.IsNull() {
		startInput.Name = config.Name.ValueStringPointer()
	}

	if !config.TraceHeader.IsNull() {
		startInput.TraceHeader = config.TraceHeader.ValueStringPointer()
	}

	output, err := conn.StartExecution(ctx, startInput)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to Start Step Functions Execution",
			fmt.Sprintf("Could not start execution for state machine %s: %s", stateMachineArn, err),
		)
		return
	}

	executionArn := aws.ToString(output.ExecutionArn)
	resp.SendProgress(action.InvokeProgressEvent{
		Message: fmt.Sprintf("Execution started successfully with ARN %s", executionArn),
	})

	tflog.Info(ctx, "Step Functions execution started successfully", map[string]any{
		"state_machine_arn": stateMachineArn,
		"execution_arn":     executionArn,
		"start_date":        output.StartDate,
	})
}
