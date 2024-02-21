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
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Deployment")
func newResourceDeployment(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceDeployment{}
	r.SetDefaultCreateTimeout(20 * time.Minute)
	r.SetDefaultReadTimeout(20 * time.Minute)

	return r, nil
}

type resourceDeployment struct {
	framework.ResourceWithConfigure
	framework.WithTimeouts
}

func (r *resourceDeployment) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_apprunner_deployment"
}

const (
	ResNameDeployment = "Deployment"
)

func (r *resourceDeployment) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"service_arn": schema.StringAttribute{
				Required: true,
			},
			"id": schema.StringAttribute{
				Computed: true,
			},
			"status": schema.StringAttribute{
				Computed: true,
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

func (r *resourceDeployment) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().AppRunnerClient(ctx)

	var plan resourceDeploymentData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &apprunner.StartDeploymentInput{
		ServiceArn: aws.String(plan.ServiceArn.ValueString()),
	}

	out, err := conn.StartDeployment(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AppRunner, create.ErrActionCreating, ResNameDeployment, plan.ServiceArn.String(), err),
			err.Error(),
		)
		return
	}

	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AppRunner, create.ErrActionCreating, ResNameDeployment, plan.ServiceArn.String(), nil),
			"no output",
		)
		return
	}

	plan.ID = flex.StringToFramework(ctx, out.OperationId)

	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	_, err = waitDeploymentSucceeded(ctx, conn, plan.ServiceArn.ValueString(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AppRunner, create.ErrActionWaitingForCreation, ResNameDeployment, plan.ServiceArn.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceDeployment) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().AppRunnerClient(ctx)

	var state resourceDeploymentData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findDeploymentOperationByServiceARN(ctx, conn, state.ServiceArn.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.AppRunner, create.ErrActionReading, ResNameDeployment, state.ServiceArn.String(), err),
			err.Error(),
		)
		return
	}

	state.ID = flex.StringToFramework(ctx, out.Id)
	state.Status = flex.StringValueToFramework(ctx, out.Status)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceDeployment) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

// Delete does not need to explicitly call resp.State.RemoveResource() as this is automatically handled by the framework.
func (r *resourceDeployment) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func waitDeploymentSucceeded(ctx context.Context, conn *apprunner.Client, arn string, timeout time.Duration) (*apprunner_types.OperationSummary, error) {
	stateConf := &retry.StateChangeConf{
		Pending:        []string{},
		Target:         enum.Slice(apprunner_types.OperationStatusSucceeded),
		Refresh:        statusDeployment(ctx, conn, arn),
		Timeout:        timeout,
		PollInterval:   30 * time.Second,
		NotFoundChecks: 30,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*apprunner_types.OperationSummary); ok {
		return output, err
	}

	return nil, err
}

func statusDeployment(ctx context.Context, conn *apprunner.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findDeploymentOperationByServiceARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func findDeploymentOperationByServiceARN(ctx context.Context, conn *apprunner.Client, arn string) (*apprunner_types.OperationSummary, error) {
	input := &apprunner.ListOperationsInput{
		ServiceArn: aws.String(arn),
	}

	output, err := conn.ListOperations(ctx, input)

	if err != nil {
		return nil, err
	}

	if len(output.OperationSummaryList) == 0 {
		return nil, &retry.NotFoundError{
			Message:     "deployment operation not found",
			LastRequest: input,
		}
	}

	var operation apprunner_types.OperationSummary
	var found bool
	for _, op := range output.OperationSummaryList {
		if aws.ToString(op.TargetArn) == arn {
			operation = op
			found = true
			break
		}
	}

	if !found {
		return nil, &retry.NotFoundError{
			Message:     "deployment operation not found",
			LastRequest: input,
		}
	}

	return &operation, nil
}

type resourceDeploymentData struct {
	ServiceArn types.String   `tfsdk:"service_arn"`
	ID         types.String   `tfsdk:"id"`
	Status     types.String   `tfsdk:"status"`
	Timeouts   timeouts.Value `tfsdk:"timeouts"`
}
