// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	awstypes "github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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
func newAccountResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &accountResource{}

	return r, nil
}

type accountResource struct {
	framework.ResourceWithModel[accountResourceModel]
	framework.WithImportByID
}

func (r *accountResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
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
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
			names.AttrID:        framework.IDAttributeDeprecatedNoReplacement(),
			"throttle_settings": framework.DataSourceComputedListOfObjectAttribute[throttleSettingsModel](ctx),
		},
	}

	response.Schema = s
}

func (r *accountResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data accountResourceModel
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
	data.ID = flex.StringValueToFramework(ctx, r.Meta().AccountID(ctx))

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *accountResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data accountResourceModel
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

func (r *accountResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state, plan accountResourceModel
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

func (r *accountResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data accountResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

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
}

type accountResourceModel struct {
	framework.WithRegionModel
	ApiKeyVersion     types.String                                           `tfsdk:"api_key_version"`
	CloudwatchRoleARN types.String                                           `tfsdk:"cloudwatch_role_arn" autoflex:",legacy"`
	Features          fwtypes.SetOfString                                    `tfsdk:"features"`
	ID                types.String                                           `tfsdk:"id"`
	ThrottleSettings  fwtypes.ListNestedObjectValueOf[throttleSettingsModel] `tfsdk:"throttle_settings"`
}

type throttleSettingsModel struct {
	BurstLimit types.Int32   `tfsdk:"burst_limit"`
	RateLimit  types.Float64 `tfsdk:"rate_limit"`
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
