// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	awstypes "github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_api_gateway_account", name="Account")
func newResourceAccount(context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceAccount{}

	return r, nil
}

type resourceAccount struct {
	framework.ResourceWithConfigure
}

func (r *resourceAccount) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	s := schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key_version": schema.StringAttribute{
				Computed: true,
			},
			"cloudwatch_role_arn": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.Any(
						validators.ARN(),
						stringvalidator.OneOf(""),
					),
				},
				Default: stringdefault.StaticString(""), // Needed for backwards compatibility with SDK resource
			},
			"features": schema.SetAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			names.AttrID: framework.IDAttributeDeprecatedNoReplacement(),
			"reset_on_delete": schema.BoolAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				DeprecationMessage: `The "reset_on_delete" attribute will be removed in a future version of the provider`,
			},
			"throttle_settings": framework.DataSourceComputedListOfObjectAttribute[throttleSettingsModel](ctx),
		},
	}

	response.Schema = s
}

func (r *resourceAccount) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resourceAccountModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().APIGatewayClient(ctx)

	input := apigateway.UpdateAccountInput{}

	if data.CloudwatchRoleARN.IsNull() || data.CloudwatchRoleARN.ValueString() == "" {
		input.PatchOperations = []awstypes.PatchOperation{
			{
				Op:    awstypes.OpReplace,
				Path:  aws.String("/cloudwatchRoleArn"),
				Value: nil,
			},
		}
	} else {
		input.PatchOperations = []awstypes.PatchOperation{
			{
				Op:    awstypes.OpReplace,
				Path:  aws.String("/cloudwatchRoleArn"),
				Value: data.CloudwatchRoleARN.ValueStringPointer(),
			},
		}
	}

	output, err := tfresource.RetryGWhen(ctx, propagationTimeout,
		func() (*apigateway.UpdateAccountOutput, error) {
			return conn.UpdateAccount(ctx, &input)
		},
		func(err error) (bool, error) {
			if errs.IsAErrorMessageContains[*awstypes.BadRequestException](err, "The role ARN does not have required permissions") {
				return true, err
			}
			if errs.IsAErrorMessageContains[*awstypes.BadRequestException](err, "API Gateway could not successfully write to CloudWatch Logs using the ARN specified") {
				return true, err
			}
			return false, err
		},
	)
	if err != nil {
		response.Diagnostics.AddError("creating API Gateway Account", err.Error())
		return
	}

	response.Diagnostics.Append(flex.Flatten(ctx, output, &data)...)
	data.ID = types.StringValue("api-gateway-account")

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceAccount) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resourceAccountModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().APIGatewayClient(ctx)

	account, err := findAccount(ctx, conn)
	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		response.Diagnostics.AddError("reading API Gateway Account", err.Error())
		return
	}

	response.Diagnostics.Append(flex.Flatten(ctx, account, &data)...)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceAccount) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state, plan resourceAccountModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	diff, d := flex.Diff(ctx, plan, state)
	response.Diagnostics.Append(d...)
	if response.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		conn := r.Meta().APIGatewayClient(ctx)

		input := apigateway.UpdateAccountInput{}

		if plan.CloudwatchRoleARN.IsNull() || plan.CloudwatchRoleARN.ValueString() == "" {
			input.PatchOperations = []awstypes.PatchOperation{
				{
					Op:    awstypes.OpReplace,
					Path:  aws.String("/cloudwatchRoleArn"),
					Value: nil,
				},
			}
		} else {
			input.PatchOperations = []awstypes.PatchOperation{
				{
					Op:    awstypes.OpReplace,
					Path:  aws.String("/cloudwatchRoleArn"),
					Value: plan.CloudwatchRoleARN.ValueStringPointer(),
				},
			}
		}

		output, err := tfresource.RetryGWhen(ctx, propagationTimeout,
			func() (*apigateway.UpdateAccountOutput, error) {
				return conn.UpdateAccount(ctx, &input)
			},
			func(err error) (bool, error) {
				if errs.IsAErrorMessageContains[*awstypes.BadRequestException](err, "The role ARN does not have required permissions") {
					return true, err
				}
				if errs.IsAErrorMessageContains[*awstypes.BadRequestException](err, "API Gateway could not successfully write to CloudWatch Logs using the ARN specified") {
					return true, err
				}
				return false, err
			},
		)
		if err != nil {
			response.Diagnostics.AddError("updating API Gateway Account", err.Error())
			return
		}

		response.Diagnostics.Append(flex.Flatten(ctx, output, &plan)...)
	}

	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func (r *resourceAccount) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resourceAccountModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if data.ResetOnDelete.ValueBool() {
		conn := r.Meta().APIGatewayClient(ctx)

		input := apigateway.UpdateAccountInput{}

		input.PatchOperations = []awstypes.PatchOperation{{
			Op:    awstypes.OpReplace,
			Path:  aws.String("/cloudwatchRoleArn"),
			Value: nil,
		}}

		_, err := conn.UpdateAccount(ctx, &input)
		if err != nil {
			response.Diagnostics.AddError("resetting API Gateway Account", err.Error())
		}
	} else {
		response.Diagnostics.AddWarning(
			"Resource Destruction",
			"This resource has only been removed from Terraform state. "+
				"Manually use the AWS Console to fully destroy this resource. "+
				"Setting the attribute \"reset_on_delete\" will also fully destroy resources of this type.",
		)
	}
}

func (r *resourceAccount) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), request, response)
}

type resourceAccountModel struct {
	ApiKeyVersion     types.String                                           `tfsdk:"api_key_version"`
	CloudwatchRoleARN types.String                                           `tfsdk:"cloudwatch_role_arn" autoflex:",legacy"`
	Features          types.Set                                              `tfsdk:"features"`
	ID                types.String                                           `tfsdk:"id"`
	ResetOnDelete     types.Bool                                             `tfsdk:"reset_on_delete"`
	ThrottleSettings  fwtypes.ListNestedObjectValueOf[throttleSettingsModel] `tfsdk:"throttle_settings"`
}

type throttleSettingsModel struct {
	BurstLimit types.Int32   `tfsdk:"burst_limit"`
	RateLimit  types.Float64 `tfsdk:"rate_limit"`
}

func (r *resourceAccount) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	// If the entire plan is null, the resource is planned for destruction.
	if request.Plan.Raw.IsNull() {
		var resetOnDelete types.Bool
		response.Diagnostics.Append(request.State.GetAttribute(ctx, path.Root("reset_on_delete"), &resetOnDelete)...)
		if response.Diagnostics.HasError() {
			return
		}

		if !resetOnDelete.ValueBool() {
			response.Diagnostics.AddWarning(
				"Resource Destruction",
				"Applying this resource destruction will only remove the resource from Terraform state and will not reset account settings. "+
					"Either manually use the AWS Console to fully destroy this resource or "+
					"update the resource with \"reset_on_delete\" set to true.",
			)
		}
	}
}

func findAccount(ctx context.Context, conn *apigateway.Client) (*apigateway.GetAccountOutput, error) {
	input := apigateway.GetAccountInput{}

	output, err := conn.GetAccount(ctx, &input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
