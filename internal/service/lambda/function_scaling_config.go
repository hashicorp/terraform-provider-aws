// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
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
// @Testing(importIgnore="function_state")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/lambda;lambda.GetFunctionScalingConfigOutput")
// @Testing(preCheck="testAccCapacityProviderPreCheck")
func newFunctionScalingConfigResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &functionScalingConfigResource{}

	r.SetDefaultCreateTimeout(10 * time.Minute)
	r.SetDefaultUpdateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(10 * time.Minute)

	return r, nil
}

const (
	ResNameFunctionScalingConfig = "Function Scaling Config"
)

type functionScalingConfigResource struct {
	framework.ResourceWithModel[functionScalingConfigResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *functionScalingConfigResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"applied_function_scaling_config": framework.ResourceComputedListOfObjectsAttribute[functionScalingConfigModel](ctx, listplanmodifier.UseStateForUnknown()),
			"function_arn": schema.StringAttribute{
				CustomType:  fwtypes.ARNType,
				Computed:    true,
				Description: "ARN of the Lambda function.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
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
			"function_state": schema.StringAttribute{
				CustomType:  fwtypes.StringEnumType[awstypes.State](),
				Computed:    true,
				Description: "State of the function after applying the scaling configuration.",
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
					listvalidator.IsRequired(),
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
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *functionScalingConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().LambdaClient(ctx)

	var plan functionScalingConfigResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	var input lambda.PutFunctionScalingConfigInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	// Capacity provider functions can take time to stabilize after publishing.
	// Retry on ResourceConflictException until the version is ready.
	createTimeout := r.CreateTimeout(ctx, plan.Timeouts)
	out, err := tfresource.RetryWhenIsA[*lambda.PutFunctionScalingConfigOutput, *awstypes.ResourceConflictException](ctx, createTimeout, func(ctx context.Context) (*lambda.PutFunctionScalingConfigOutput, error) {
		return conn.PutFunctionScalingConfig(ctx, &input)
	})
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.FunctionName.ValueString())
		return
	}

	// function_state is only returned by the Put operation.
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan), smerr.ID, plan.FunctionName.ValueString())
	if resp.Diagnostics.HasError() {
		return
	}

	// Read back to populate computed attributes (function_arn, applied config).
	// AppliedFunctionScalingConfig is populated by AWS asynchronously and may not
	// be present immediately; surface whatever is currently returned.
	scOut, err := findFunctionScalingConfigByTwoPartKey(ctx, conn, plan.FunctionName.ValueString(), plan.Qualifier.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.FunctionName.ValueString())
		return
	}
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, scOut, &plan), smerr.ID, plan.FunctionName.ValueString())
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan), smerr.ID, plan.FunctionName.ValueString())
}

func (r *functionScalingConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().LambdaClient(ctx)

	var state functionScalingConfigResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findFunctionScalingConfigByTwoPartKey(ctx, conn, state.FunctionName.ValueString(), state.Qualifier.ValueString())
	if retry.NotFound(err) {
		smerr.AddOne(ctx, &resp.Diagnostics, fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.FunctionName.ValueString())
		return
	}

	// AutoFlex maps function_arn and applied_function_scaling_config directly.
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &state), smerr.ID, state.FunctionName.ValueString())
	if resp.Diagnostics.HasError() {
		return
	}

	// The API returns the configured (requested) scaling config under a different
	// field name than the Put input, so map it explicitly into function_scaling_config.
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out.RequestedFunctionScalingConfig, &state.FunctionScalingConfig), smerr.ID, state.FunctionName.ValueString())
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state), smerr.ID, state.FunctionName.ValueString())
}

func (r *functionScalingConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().LambdaClient(ctx)

	var plan, state functionScalingConfigResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.FunctionScalingConfig.Equal(state.FunctionScalingConfig) {
		var input lambda.PutFunctionScalingConfigInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, plan, &input))
		if resp.Diagnostics.HasError() {
			return
		}

		updateTimeout := r.UpdateTimeout(ctx, plan.Timeouts)
		out, err := tfresource.RetryWhenIsA[*lambda.PutFunctionScalingConfigOutput, *awstypes.ResourceConflictException](ctx, updateTimeout, func(ctx context.Context) (*lambda.PutFunctionScalingConfigOutput, error) {
			return conn.PutFunctionScalingConfig(ctx, &input)
		})
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.FunctionName.ValueString())
			return
		}

		// function_state is only returned by the Put operation.
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &plan), smerr.ID, plan.FunctionName.ValueString())
		if resp.Diagnostics.HasError() {
			return
		}

		// Read back to populate computed attributes (function_arn, applied config).
		// AppliedFunctionScalingConfig is populated by AWS asynchronously and may not
		// be present immediately; surface whatever is currently returned.
		scOut, err := findFunctionScalingConfigByTwoPartKey(ctx, conn, plan.FunctionName.ValueString(), plan.Qualifier.ValueString())
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.FunctionName.ValueString())
			return
		}
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, scOut, &plan), smerr.ID, plan.FunctionName.ValueString())
		if resp.Diagnostics.HasError() {
			return
		}
	} else {
		plan.FunctionARN = state.FunctionARN
		plan.FunctionState = state.FunctionState
		plan.AppliedFunctionScalingConfig = state.AppliedFunctionScalingConfig
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan), smerr.ID, plan.FunctionName.ValueString())
}

// Delete resets the scaling configuration by calling PutFunctionScalingConfig with nil config.
// There is no dedicated DeleteFunctionScalingConfig API.
func (r *functionScalingConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().LambdaClient(ctx)

	var state functionScalingConfigResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	input := lambda.PutFunctionScalingConfigInput{
		FunctionName: state.FunctionName.ValueStringPointer(),
		Qualifier:    state.Qualifier.ValueStringPointer(),
	}

	deleteTimeout := r.DeleteTimeout(ctx, state.Timeouts)
	_, err := tfresource.RetryWhenIsA[any, *awstypes.ResourceConflictException](ctx, deleteTimeout, func(ctx context.Context) (any, error) {
		return conn.PutFunctionScalingConfig(ctx, &input)
	})
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return
		}
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.FunctionName.ValueString())
		return
	}
}

func findFunctionScalingConfigByTwoPartKey(ctx context.Context, conn *lambda.Client, functionName, qualifier string) (*lambda.GetFunctionScalingConfigOutput, error) {
	input := lambda.GetFunctionScalingConfigInput{
		FunctionName: aws.String(functionName),
		Qualifier:    aws.String(qualifier),
	}

	out, err := conn.GetFunctionScalingConfig(ctx, &input)
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, smarterr.NewError(&retry.NotFoundError{
				LastError: err,
			})
		}

		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

type functionScalingConfigResourceModel struct {
	framework.WithRegionModel
	AppliedFunctionScalingConfig fwtypes.ListNestedObjectValueOf[functionScalingConfigModel] `tfsdk:"applied_function_scaling_config"`
	FunctionARN                  fwtypes.ARN                                                 `tfsdk:"function_arn"`
	FunctionName                 types.String                                                `tfsdk:"function_name"`
	FunctionScalingConfig        fwtypes.ListNestedObjectValueOf[functionScalingConfigModel] `tfsdk:"function_scaling_config"`
	FunctionState                fwtypes.StringEnum[awstypes.State]                          `tfsdk:"function_state"`
	Qualifier                    types.String                                                `tfsdk:"qualifier"`
	Timeouts                     timeouts.Value                                              `tfsdk:"timeouts"`
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
