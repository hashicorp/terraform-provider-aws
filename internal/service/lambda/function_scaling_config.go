// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"context"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/int32validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_lambda_function_scaling_config", name="Function Scaling Config")
func newFunctionScalingConfigResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &functionScalingConfigResource{}, nil
}

const (
	ResNameFunctionScalingConfig = "Function Scaling Config"
)

type functionScalingConfigResource struct {
	framework.ResourceWithModel[functionScalingConfigResourceModel]
}

func (r *functionScalingConfigResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"function_name": schema.StringAttribute{
				Required:    true,
				Description: "Name or ARN of the Lambda function.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					functionNameValidator,
				},
			},
			"function_arn": schema.StringAttribute{
				Computed:    true,
				Description: "ARN of the Lambda function.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"max_execution_environments": schema.Int32Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Maximum number of execution environments that can be provisioned for the function.",
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int32{
					int32validator.AtLeast(1),
				},
			},
			"min_execution_environments": schema.Int32Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Minimum number of execution environments to maintain for the function.",
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Int32{
					int32validator.AtLeast(0),
				},
			},
			"qualifier": schema.StringAttribute{
				Required:    true,
				Description: "Qualifier for the scaling configuration. Valid values: $LATEST.PUBLISHED or a numeric version number.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^(\$LATEST\.PUBLISHED|[0-9]+)$`), "must be $LATEST.PUBLISHED or a numeric version"),
				},
			},
		},
	}
}

func (r *functionScalingConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().LambdaClient(ctx)

	var plan functionScalingConfigResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &lambda.PutFunctionScalingConfigInput{
		FunctionName:          plan.FunctionName.ValueStringPointer(),
		Qualifier:             plan.Qualifier.ValueStringPointer(),
		FunctionScalingConfig: &awstypes.FunctionScalingConfig{},
	}

	if !plan.MinExecutionEnvironments.IsNull() && !plan.MinExecutionEnvironments.IsUnknown() {
		in.FunctionScalingConfig.MinExecutionEnvironments = plan.MinExecutionEnvironments.ValueInt32Pointer()
	}

	if !plan.MaxExecutionEnvironments.IsNull() && !plan.MaxExecutionEnvironments.IsUnknown() {
		in.FunctionScalingConfig.MaxExecutionEnvironments = plan.MaxExecutionEnvironments.ValueInt32Pointer()
	}

	// Capacity provider functions can take time to stabilize after publishing.
	// Retry on ResourceConflictException until the version is ready.
	_, err := tfresource.RetryWhenIsA[any, *awstypes.ResourceConflictException](ctx, 10*time.Minute, func(ctx context.Context) (any, error) {
		return conn.PutFunctionScalingConfig(ctx, in)
	})
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Lambda, create.ErrActionCreating, ResNameFunctionScalingConfig, plan.FunctionName.String(), err),
			err.Error(),
		)
		return
	}

	// Read back the scaling config to populate computed attributes.
	out, err := findFunctionScalingConfigByTwoPartKey(ctx, conn, plan.FunctionName.ValueString(), plan.Qualifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Lambda, create.ErrActionCreating, ResNameFunctionScalingConfig, plan.FunctionName.String(), err),
			err.Error(),
		)
		return
	}

	plan.FunctionARN = types.StringPointerValue(out.FunctionArn)
	if out.RequestedFunctionScalingConfig != nil {
		plan.MinExecutionEnvironments = types.Int32PointerValue(out.RequestedFunctionScalingConfig.MinExecutionEnvironments)
		plan.MaxExecutionEnvironments = types.Int32PointerValue(out.RequestedFunctionScalingConfig.MaxExecutionEnvironments)
	} else if out.AppliedFunctionScalingConfig != nil {
		plan.MinExecutionEnvironments = types.Int32PointerValue(out.AppliedFunctionScalingConfig.MinExecutionEnvironments)
		plan.MaxExecutionEnvironments = types.Int32PointerValue(out.AppliedFunctionScalingConfig.MaxExecutionEnvironments)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *functionScalingConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().LambdaClient(ctx)

	var state functionScalingConfigResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findFunctionScalingConfigByTwoPartKey(ctx, conn, state.FunctionName.ValueString(), state.Qualifier.ValueString())
	if retry.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Lambda, create.ErrActionSetting, ResNameFunctionScalingConfig, state.FunctionName.String(), err),
			err.Error(),
		)
		return
	}

	state.FunctionARN = types.StringPointerValue(out.FunctionArn)
	if out.RequestedFunctionScalingConfig != nil {
		state.MinExecutionEnvironments = types.Int32PointerValue(out.RequestedFunctionScalingConfig.MinExecutionEnvironments)
		state.MaxExecutionEnvironments = types.Int32PointerValue(out.RequestedFunctionScalingConfig.MaxExecutionEnvironments)
	} else if out.AppliedFunctionScalingConfig != nil {
		state.MinExecutionEnvironments = types.Int32PointerValue(out.AppliedFunctionScalingConfig.MinExecutionEnvironments)
		state.MaxExecutionEnvironments = types.Int32PointerValue(out.AppliedFunctionScalingConfig.MaxExecutionEnvironments)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *functionScalingConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().LambdaClient(ctx)

	var plan, state functionScalingConfigResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.MinExecutionEnvironments.Equal(state.MinExecutionEnvironments) ||
		!plan.MaxExecutionEnvironments.Equal(state.MaxExecutionEnvironments) {

		in := &lambda.PutFunctionScalingConfigInput{
			FunctionName:          plan.FunctionName.ValueStringPointer(),
			Qualifier:             plan.Qualifier.ValueStringPointer(),
			FunctionScalingConfig: &awstypes.FunctionScalingConfig{},
		}

		if !plan.MinExecutionEnvironments.IsNull() && !plan.MinExecutionEnvironments.IsUnknown() {
			in.FunctionScalingConfig.MinExecutionEnvironments = plan.MinExecutionEnvironments.ValueInt32Pointer()
		}

		if !plan.MaxExecutionEnvironments.IsNull() && !plan.MaxExecutionEnvironments.IsUnknown() {
			in.FunctionScalingConfig.MaxExecutionEnvironments = plan.MaxExecutionEnvironments.ValueInt32Pointer()
		}

		_, err := tfresource.RetryWhenIsA[any, *awstypes.ResourceConflictException](ctx, lambdaPropagationTimeout, func(ctx context.Context) (any, error) {
			return conn.PutFunctionScalingConfig(ctx, in)
		})
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Lambda, create.ErrActionUpdating, ResNameFunctionScalingConfig, plan.FunctionName.String(), err),
				err.Error(),
			)
			return
		}

		// Read back to get computed values.
		out, err := findFunctionScalingConfigByTwoPartKey(ctx, conn, plan.FunctionName.ValueString(), plan.Qualifier.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Lambda, create.ErrActionUpdating, ResNameFunctionScalingConfig, plan.FunctionName.String(), err),
				err.Error(),
			)
			return
		}

		plan.FunctionARN = types.StringPointerValue(out.FunctionArn)
		if out.RequestedFunctionScalingConfig != nil {
			plan.MinExecutionEnvironments = types.Int32PointerValue(out.RequestedFunctionScalingConfig.MinExecutionEnvironments)
			plan.MaxExecutionEnvironments = types.Int32PointerValue(out.RequestedFunctionScalingConfig.MaxExecutionEnvironments)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete resets the scaling configuration by calling PutFunctionScalingConfig with nil config.
// There is no dedicated DeleteFunctionScalingConfig API.
func (r *functionScalingConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().LambdaClient(ctx)

	var state functionScalingConfigResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &lambda.PutFunctionScalingConfigInput{
		FunctionName: state.FunctionName.ValueStringPointer(),
		Qualifier:    state.Qualifier.ValueStringPointer(),
	}

	_, err := tfresource.RetryWhenIsA[any, *awstypes.ResourceConflictException](ctx, 10*time.Minute, func(ctx context.Context) (any, error) {
		return conn.PutFunctionScalingConfig(ctx, in)
	})
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Lambda, create.ErrActionDeleting, ResNameFunctionScalingConfig, state.FunctionName.String(), err),
			err.Error(),
		)
		return
	}
}

func (r *functionScalingConfigResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ":")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			"Expected import identifier with format: function_name:qualifier",
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("function_name"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("qualifier"), parts[1])...)
}

func findFunctionScalingConfigByTwoPartKey(ctx context.Context, conn *lambda.Client, functionName, qualifier string) (*lambda.GetFunctionScalingConfigOutput, error) {
	in := &lambda.GetFunctionScalingConfigInput{
		FunctionName: aws.String(functionName),
		Qualifier:    aws.String(qualifier),
	}

	out, err := conn.GetFunctionScalingConfig(ctx, in)
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

type functionScalingConfigResourceModel struct {
	framework.WithRegionModel
	FunctionARN              types.String `tfsdk:"function_arn"`
	FunctionName             types.String `tfsdk:"function_name"`
	MaxExecutionEnvironments types.Int32  `tfsdk:"max_execution_environments"`
	MinExecutionEnvironments types.Int32  `tfsdk:"min_execution_environments"`
	Qualifier                types.String `tfsdk:"qualifier"`
}
