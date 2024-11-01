// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
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

func (e *ephemeralInvocation) Metadata(_ context.Context, _ ephemeral.MetadataRequest, response *ephemeral.MetadataResponse) {
	response.TypeName = "aws_lambda_invocation"
}

func (e *ephemeralInvocation) Schema(ctx context.Context, _ ephemeral.SchemaRequest, response *ephemeral.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"function_name": schema.StringAttribute{
				Required: true,
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
		},
	}
}

func (e *ephemeralInvocation) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	conn := e.Meta().LambdaClient(ctx)
	d := &epInvocationData{}

	resp.Diagnostics.Append(req.Config.Get(ctx, &d)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &lambda.InvokeInput{}
	resp.Diagnostics.Append(fwflex.Expand(ctx, d, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	output, err := conn.Invoke(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Lambda, create.ErrActionCreating, ResNameInvocation, d.FunctionName.String(), err),
			err.Error(),
		)
		return
	}

	if output.FunctionError != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Lambda, create.ErrActionCreating, ResNameInvocation, d.FunctionName.String(), errors.New(aws.ToString(output.FunctionError))),
			err.Error(),
		)
		return
	}

	d.Result = types.StringValue(string(output.Payload))
	resp.Diagnostics.Append(resp.Result.Set(ctx, &d)...)
}

type epInvocationData struct {
	FunctionName types.String `tfsdk:"function_name"`
	Payload      types.String `tfsdk:"input"`
	Qualifier    types.String `tfsdk:"qualifier"`
	Result       types.String `tfsdk:"result"`
}
