// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apprunner

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apprunner"
	apprunner_types "github.com/aws/aws-sdk-go-v2/service/apprunner/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Start Deployment")
func newResourceStartDeployment(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceStartDeployment{}
	r.SetDefaultCreateTimeout(5 * time.Minute)
	r.SetDefaultReadTimeout(5 * time.Minute)

	return r, nil
}

type resourceStartDeployment struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceStartDeployment) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_apprunner_start_deployment"
}

const (
	ResNameStartDeployment = "Start Deployment"
)

func (r *resourceStartDeployment) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"aws_account_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"operation_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"started_at": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ended_at": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceStartDeployment) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().AppRunnerClient(ctx)

	var plan resourceStartDeploymentData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &apprunner.StartDeploymentInput{
		ServiceArn: aws.String(plan.ServiceARN.ValueString()),
	}

	out, err := conn.StartDeployment(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AppRunner, create.ErrActionCreating, ResNameStartDeployment, plan.ServiceARN.String(), err),
			err.Error(),
		)
		return
	}

	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AppRunner, create.ErrActionCreating, ResNameStartDeployment, plan.ServiceARN.String(), nil),
			"no output",
		)
		return
	}

	plan.OperationID = flex.StringToFramework(ctx, out.OperationId)

	_, err = waitStartDeploymentSucceeded(ctx, conn, plan.ServiceARN.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.QuickSight, create.ErrActionWaitingForCreation, ResNameStartDeployment, plan.ServiceARN.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceStartDeployment) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().AppRunnerClient(ctx)

	var state resourceStartDeploymentData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findStartDeploymentOperationByServiceARN(ctx, conn, state.ServiceARN.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AppRunner, create.ErrActionReading, ResNameStartDeployment, state.ServiceARN.String(), err),
			err.Error(),
		)
		return
	}

	state.OperationID = flex.StringToFramework(ctx, out.Id)
	state.StartedAt = flex.StringToFramework(ctx, out.StartedAt)
	state.EndedAt = flex.StringToFramework(ctx, out.EndedAt)
	state.Status = flex.StringToFramework(ctx, (*string)(&out.Status))

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceStartDeployment) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func waitStartDeploymentSucceeded(ctx context.Context, conn *apprunner.Client, arn string) (*apprunner_types.OperationSummary, error) {
	const (
		timeout = 15 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: []string{},
		Target:  []string{string(apprunner_types.OperationStatusSucceeded)},
		Refresh: statusStartDeployment(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*apprunner_types.OperationSummary); ok {
		return output, err
	}

	return nil, err
}

func statusStartDeployment(ctx context.Context, conn *apprunner.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findStartDeploymentOperationByServiceARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func findStartDeploymentOperationByServiceARN(ctx context.Context, conn *apprunner.Client, arn string) (*apprunner_types.OperationSummary, error) {
	input := &apprunner.ListOperationsInput{
		ServiceArn: aws.String(arn),
	}

	output, err := conn.ListOperations(ctx, input)

	if err != nil {
		return nil, err
	}

	if len(output.OperationSummaryList) == 0 {
		return nil, &retry.NotFoundError{
			Message:     "start deployment operation not found",
			LastRequest: input,
		}
	}

	var operation apprunner_types.OperationSummary
	var found bool
	for _, op := range output.OperationSummaryList {
		if aws.String(*op.TargetArn) == aws.String(arn) {
			operation = op
			found = true
			break
		}
	}

	if !found {
		return nil, &retry.NotFoundError{
			Message:     "start deployment operation not found",
			LastRequest: input,
		}
	}

	return &operation, nil
}

type resourceStartDeploymentData struct {
	ServiceARN  types.String `tf:"service_arn"`
	OperationID types.String `tf:"operation_id"`
	StartedAt   types.String `tf:"started_at"`
	EndedAt     types.String `tf:"ended_at"`
	Status      types.String `tf:"status"`
}
