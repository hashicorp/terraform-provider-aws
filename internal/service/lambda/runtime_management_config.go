// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_lambda_runtime_management_config", name="Runtime Management Config")
func newResourceRuntimeManagementConfig(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceRuntimeManagementConfig{}, nil
}

const (
	ResNameRuntimeManagementConfig = "Runtime Management Config"
	runtimeManagementConfigIDParts = 2
)

type resourceRuntimeManagementConfig struct {
	framework.ResourceWithConfigure
	framework.WithNoOpDelete
}

func (r *resourceRuntimeManagementConfig) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_lambda_runtime_management_config"
}

func (r *resourceRuntimeManagementConfig) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrFunctionARN: framework.ARNAttributeComputedOnly(),
			"function_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"qualifier": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"runtime_version_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
			},
			"update_runtime_on": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.UpdateRuntimeOn](),
				Optional:   true,
			},
		},
	}
}

func (r *resourceRuntimeManagementConfig) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().LambdaClient(ctx)

	var plan resourceRuntimeManagementConfigData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &lambda.PutRuntimeManagementConfigInput{}
	resp.Diagnostics.Append(flex.Expand(ctx, plan, in)...)

	out, err := conn.PutRuntimeManagementConfig(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Lambda, create.ErrActionCreating, ResNameRuntimeManagementConfig, plan.FunctionName.String(), err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Lambda, create.ErrActionCreating, ResNameRuntimeManagementConfig, plan.FunctionName.String(), nil),
			errors.New("empty output").Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceRuntimeManagementConfig) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().LambdaClient(ctx)

	var state resourceRuntimeManagementConfigData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findRuntimeManagementConfigByTwoPartKey(ctx, conn, state.FunctionName.ValueString(), state.Qualifier.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Lambda, create.ErrActionSetting, ResNameRuntimeManagementConfig, state.FunctionName.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceRuntimeManagementConfig) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().LambdaClient(ctx)

	var plan, state resourceRuntimeManagementConfigData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.RuntimeVersionARN.Equal(state.RuntimeVersionARN) ||
		!plan.UpdateRuntimeOn.Equal(state.UpdateRuntimeOn) {
		in := &lambda.PutRuntimeManagementConfigInput{}
		resp.Diagnostics.Append(flex.Expand(ctx, plan, in)...)

		out, err := conn.PutRuntimeManagementConfig(ctx, in)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Lambda, create.ErrActionUpdating, ResNameRuntimeManagementConfig, plan.FunctionName.String(), err),
				err.Error(),
			)
			return
		}
		if out == nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Lambda, create.ErrActionUpdating, ResNameRuntimeManagementConfig, plan.FunctionName.String(), nil),
				errors.New("empty output").Error(),
			)
			return
		}

		resp.Diagnostics.Append(flex.Flatten(ctx, out, &plan)...)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceRuntimeManagementConfig) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().LambdaClient(ctx)

	var state resourceRuntimeManagementConfigData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &lambda.PutRuntimeManagementConfigInput{
		FunctionName:    aws.String(state.FunctionName.ValueString()),
		UpdateRuntimeOn: awstypes.UpdateRuntimeOnAuto,
	}
	if !state.Qualifier.IsNull() && state.Qualifier.ValueString() != "" {
		in.Qualifier = aws.String(state.Qualifier.ValueString())
	}

	_, err := conn.PutRuntimeManagementConfig(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Lambda, create.ErrActionDeleting, ResNameRuntimeManagementConfig, state.FunctionName.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *resourceRuntimeManagementConfig) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, err := intflex.ExpandResourceId(req.ID, runtimeManagementConfigIDParts, true)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: function_name,qualifier. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("function_name"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("qualifier"), parts[1])...)
}

func findRuntimeManagementConfigByTwoPartKey(ctx context.Context, conn *lambda.Client, functionName, qualifier string) (*lambda.GetRuntimeManagementConfigOutput, error) {
	in := &lambda.GetRuntimeManagementConfigInput{
		FunctionName: aws.String(functionName),
	}
	if qualifier != "" {
		in.Qualifier = aws.String(qualifier)
	}

	out, err := conn.GetRuntimeManagementConfig(ctx, in)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}

type resourceRuntimeManagementConfigData struct {
	FunctionARN       types.String                                 `tfsdk:"function_arn"`
	FunctionName      types.String                                 `tfsdk:"function_name"`
	Qualifier         types.String                                 `tfsdk:"qualifier"`
	RuntimeVersionARN fwtypes.ARN                                  `tfsdk:"runtime_version_arn"`
	UpdateRuntimeOn   fwtypes.StringEnum[awstypes.UpdateRuntimeOn] `tfsdk:"update_runtime_on"`
}
