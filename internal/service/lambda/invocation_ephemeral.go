// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @EphemeralResource("aws_lambda_invocation", name="Invocation")
func newEphemeralInvocation(_ context.Context) (ephemeral.EphemeralResourceWithConfigure, error) {
	return &ephemeralInvocation{}, nil
}

const (
	ResNameInvocation = "Invocation"
)

type ephemeralInvocation struct {
	framework.EphemeralResourceWithConfigure
}

func (e *ephemeralInvocation) Schema(ctx context.Context, _ ephemeral.SchemaRequest, response *ephemeral.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"client_context": schema.StringAttribute{
				Optional: true,
			},
			"executed_version": schema.StringAttribute{
				Computed: true,
			},
			"function_error": schema.StringAttribute{
				Computed: true,
			},
			"function_name": schema.StringAttribute{
				Required: true,
			},
			"log_result": schema.StringAttribute{
				Computed: true,
			},
			"log_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.LogType](),
				Optional:   true,
			},
			"payload": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					validators.JSON(),
				},
			},
			"qualifier": schema.StringAttribute{
				Optional: true,
			},
			"result": schema.StringAttribute{
				Computed: true,
			},
			names.AttrStatusCode: schema.Int32Attribute{
				Computed: true,
			},
		},
	}
}

func (e *ephemeralInvocation) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	conn := e.Meta().LambdaClient(ctx)
	data := epInvocationData{}

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &lambda.InvokeInput{
		InvocationType: awstypes.InvocationTypeRequestResponse,
	}
	resp.Diagnostics.Append(flex.Expand(ctx, data, input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if input.FunctionName == nil {
		data.Result = types.StringValue("")
		resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)
		return
	}

	output, err := conn.Invoke(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Lambda, create.ErrActionOpening, ResNameInvocation, data.FunctionName.String(), err),
			err.Error(),
		)
		return
	}

	if output.FunctionError != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Lambda, create.ErrActionOpening, ResNameInvocation, data.FunctionName.String(), errors.New(aws.ToString(output.FunctionError))),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, output, &data)...)
	data.Result = flex.StringValueToFramework(ctx, string(output.Payload))
	resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)
}

type epInvocationData struct {
	ClientContext   types.String                         `tfsdk:"client_context"`
	ExecutedVersion types.String                         `tfsdk:"executed_version"`
	FunctionError   types.String                         `tfsdk:"function_error"`
	FunctionName    types.String                         `tfsdk:"function_name"`
	LogResult       types.String                         `tfsdk:"log_result"`
	LogType         fwtypes.StringEnum[awstypes.LogType] `tfsdk:"log_type"`
	Payload         types.String                         `tfsdk:"payload"`
	Qualifier       types.String                         `tfsdk:"qualifier"`
	Result          types.String                         `tfsdk:"result"`
	StatusCode      types.Int32                          `tfsdk:"status_code"`
}
