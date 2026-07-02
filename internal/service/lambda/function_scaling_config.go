// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_lambda_function_scaling_config", name="Function Scaling Config")
// @IdentityAttribute("function_name")
// @IdentityAttribute("qualifier")
// @ImportIDHandler("functionScalingConfigImportID")
// @Testing(hasNoPreExistingResource=true)
// @Testing(importStateIdAttributes="function_name;qualifier", importStateIdAttributesSep="flex.ResourceIdSeparator")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/lambda;lambda.GetFunctionScalingConfigOutput")
// @Testing(preCheck="testAccCapacityProviderPreCheck")
func newFunctionScalingConfigResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &functionScalingConfigResource{}, nil
}

const (
	ResNameFunctionScalingConfig      = "Function Scaling Config"
	functionScalingConfigRetryTimeout = 10 * time.Minute
)

type functionScalingConfigResource struct {
	framework.ResourceWithModel[functionScalingConfigResourceModel]
	framework.WithImportByIdentity
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
		Blocks: map[string]schema.Block{
			"function_scaling_config": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[functionScalingConfigModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"max_execution_environments": schema.Int32Attribute{
							Optional:    true,
							Description: "Maximum number of execution environments that can be provisioned for the function.",
						},
						"min_execution_environments": schema.Int32Attribute{
							Optional:    true,
							Description: "Minimum number of execution environments to maintain for the function.",
						},
					},
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

	var input lambda.PutFunctionScalingConfigInput
	resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Capacity provider functions can take time to stabilize after publishing.
	// Retry on ResourceConflictException until the version is ready.
	_, err := tfresource.RetryWhenIsA[any, *awstypes.ResourceConflictException](ctx, functionScalingConfigRetryTimeout, func(ctx context.Context) (any, error) {
		return conn.PutFunctionScalingConfig(ctx, &input)
	})
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Lambda, create.ErrActionCreating, ResNameFunctionScalingConfig, plan.FunctionName.String(), err),
			err.Error(),
		)
		return
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
			create.ProblemStandardMessage(names.Lambda, create.ErrActionReading, ResNameFunctionScalingConfig, state.FunctionName.String(), err),
			err.Error(),
		)
		return
	}

	// Map the API response to the model. AutoFlex can't handle the field name mismatch
	// (RequestedFunctionScalingConfig/AppliedFunctionScalingConfig vs FunctionScalingConfig).
	scalingConfig := out.RequestedFunctionScalingConfig
	if scalingConfig == nil {
		scalingConfig = out.AppliedFunctionScalingConfig
	}

	model := functionScalingConfigModel{
		MinExecutionEnvironments: types.Int32PointerValue(scalingConfig.MinExecutionEnvironments),
		MaxExecutionEnvironments: types.Int32PointerValue(scalingConfig.MaxExecutionEnvironments),
	}
	state.FunctionScalingConfig = fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &model)

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

	if !plan.FunctionScalingConfig.Equal(state.FunctionScalingConfig) {
		var input lambda.PutFunctionScalingConfigInput
		resp.Diagnostics.Append(flex.Expand(ctx, plan, &input)...)
		if resp.Diagnostics.HasError() {
			return
		}

		_, err := tfresource.RetryWhenIsA[any, *awstypes.ResourceConflictException](ctx, lambdaPropagationTimeout, func(ctx context.Context) (any, error) {
			return conn.PutFunctionScalingConfig(ctx, &input)
		})
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Lambda, create.ErrActionUpdating, ResNameFunctionScalingConfig, plan.FunctionName.String(), err),
				err.Error(),
			)
			return
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

	_, err := tfresource.RetryWhenIsA[any, *awstypes.ResourceConflictException](ctx, functionScalingConfigRetryTimeout, func(ctx context.Context) (any, error) {
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

	// There is no dedicated delete API; the scaling configuration is "removed" by
	// resetting it (see Delete). After a reset, GetFunctionScalingConfig still returns
	// a result but with no execution environment values, so treat that as not found.
	scalingConfig := out.RequestedFunctionScalingConfig
	if scalingConfig == nil {
		scalingConfig = out.AppliedFunctionScalingConfig
	}
	if scalingConfig == nil || (scalingConfig.MinExecutionEnvironments == nil && scalingConfig.MaxExecutionEnvironments == nil) {
		return nil, &retry.NotFoundError{}
	}

	return out, nil
}

type functionScalingConfigResourceModel struct {
	framework.WithRegionModel
	FunctionName          types.String                                                `tfsdk:"function_name"`
	FunctionScalingConfig fwtypes.ListNestedObjectValueOf[functionScalingConfigModel] `tfsdk:"function_scaling_config"`
	Qualifier             types.String                                                `tfsdk:"qualifier"`
}

type functionScalingConfigModel struct {
	MaxExecutionEnvironments types.Int32 `tfsdk:"max_execution_environments"`
	MinExecutionEnvironments types.Int32 `tfsdk:"min_execution_environments"`
}

var _ inttypes.ImportIDParser = functionScalingConfigImportID{}

type functionScalingConfigImportID struct{}

func (functionScalingConfigImportID) Parse(id string) (string, map[string]any, error) {
	functionName, qualifier, found := strings.Cut(id, intflex.ResourceIdSeparator)
	if !found || functionName == "" || qualifier == "" {
		return "", nil, fmt.Errorf("id %q should be in the format <function-name>%s<qualifier>", id, intflex.ResourceIdSeparator)
	}

	result := map[string]any{
		"function_name": functionName,
		"qualifier":     qualifier,
	}

	return id, result, nil
}
