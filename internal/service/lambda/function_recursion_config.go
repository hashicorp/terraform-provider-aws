// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package lambda

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_lambda_function_recursion_config", name="Function Recursion Config")
func newFunctionRecursionConfigResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &functionRecursionConfigResource{}, nil
}

const (
	ResNameFunctionRecursionConfig = "Function Recursion Config"
)

type functionRecursionConfigResource struct {
	framework.ResourceWithModel[functionRecursionConfigResourceModel]
}

func (r *functionRecursionConfigResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"function_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					functionNameValidator,
				},
			},
			"recursive_loop": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.RecursiveLoop](),
				Required:   true,
			},
		},
	}
}

func (r *functionRecursionConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan functionRecursionConfigResourceModel
	conn := r.Meta().LambdaClient(ctx)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &lambda.PutFunctionRecursionConfigInput{}
	resp.Diagnostics.Append(flex.Expand(ctx, &plan, in)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := conn.PutFunctionRecursionConfig(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Lambda, create.ErrActionCreating, ResNameFunctionRecursionConfig, plan.FunctionName.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Lambda, create.ErrActionCreating, ResNameFunctionRecursionConfig, plan.FunctionName.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *functionRecursionConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state functionRecursionConfigResourceModel
	conn := r.Meta().LambdaClient(ctx)

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findFunctionRecursionConfigByName(ctx, conn, state.FunctionName.ValueString())
	if retry.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Lambda, create.ErrActionSetting, ResNameFunctionRecursionConfig, state.FunctionName.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *functionRecursionConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state functionRecursionConfigResourceModel
	conn := r.Meta().LambdaClient(ctx)

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.RecursiveLoop.Equal(state.RecursiveLoop) {
		in := &lambda.PutFunctionRecursionConfigInput{
			FunctionName:  plan.FunctionName.ValueStringPointer(),
			RecursiveLoop: plan.RecursiveLoop.ValueEnum(),
		}

		out, err := conn.PutFunctionRecursionConfig(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Lambda, create.ErrActionUpdating, ResNameFunctionRecursionConfig, plan.FunctionName.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Lambda, create.ErrActionUpdating, ResNameFunctionRecursionConfig, plan.FunctionName.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete sets the Lambda function's recursion configuration to the default ("Terminate")
//
// Ref: https://docs.aws.amazon.com/lambda/latest/api/API_PutFunctionRecursionConfig.html
func (r *functionRecursionConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state functionRecursionConfigResourceModel
	conn := r.Meta().LambdaClient(ctx)

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &lambda.PutFunctionRecursionConfigInput{
		FunctionName:  state.FunctionName.ValueStringPointer(),
		RecursiveLoop: awstypes.RecursiveLoopTerminate,
	}

	_, err := conn.PutFunctionRecursionConfig(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Lambda, create.ErrActionDeleting, ResNameFunctionRecursionConfig, state.FunctionName.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *functionRecursionConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("function_name"), req, resp)
}

func findFunctionRecursionConfigByName(ctx context.Context, conn *lambda.Client, functionName string) (*lambda.GetFunctionRecursionConfigOutput, error) {
	in := &lambda.GetFunctionRecursionConfigInput{
		FunctionName: aws.String(functionName),
	}

	out, err := conn.GetFunctionRecursionConfig(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return out, nil
}

type functionRecursionConfigResourceModel struct {
	framework.WithRegionModel
	FunctionName  types.String                               `tfsdk:"function_name"`
	RecursiveLoop fwtypes.StringEnum[awstypes.RecursiveLoop] `tfsdk:"recursive_loop"`
}
