// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ERNameParameter = "Ephemeral Resource Parameter"
)

// @EphemeralResource(aws_ssm_parameter, name="Parameter")
func newEphemeralParameter(_ context.Context) (ephemeral.EphemeralResourceWithConfigure, error) {
	return &ephemeralParameter{}, nil
}

type ephemeralParameter struct {
	framework.EphemeralResourceWithConfigure
}

func (e *ephemeralParameter) Schema(ctx context.Context, _ ephemeral.SchemaRequest, response *ephemeral.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			names.AttrName: schema.StringAttribute{
				Computed: true,
			},
			names.AttrType: schema.StringAttribute{
				Computed: true,
			},
			names.AttrValue: schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
			},
			names.AttrVersion: schema.Int64Attribute{
				Computed: true,
			},
			"with_decryption": schema.BoolAttribute{
				Optional: true,
				Computed: true,
			},
		},
	}
}

func (e *ephemeralParameter) Open(ctx context.Context, request ephemeral.OpenRequest, response *ephemeral.OpenResponse) {
	var data epParameterData
	conn := e.Meta().SSMClient(ctx)

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Terraform does not have the notion of planning for ephemeral resources,
	// data sources, or providers. As a result, default handlers are not
	// implemented for these objects in the Terraform Plugin Framework.
	//
	// To align with the data source data.aws_ssm_parameter,
	// we default `with_decryption`.
	if data.WithDecryption.IsNull() {
		data.WithDecryption = types.BoolValue(true)
	}

	output, err := findParameterByName(ctx, conn, data.ARN.ValueString(), data.WithDecryption.ValueBool())
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.SSM, create.ErrActionReading, ERNameParameter, data.ARN.String(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.Result.Set(ctx, &data)...)
}

type epParameterData struct {
	ARN            fwtypes.ARN  `tfsdk:"arn"`
	Name           types.String `tfsdk:"name"`
	Type           types.String `tfsdk:"type"`
	Value          types.String `tfsdk:"value"`
	Version        types.Int64  `tfsdk:"version"`
	WithDecryption types.Bool   `tfsdk:"with_decryption"`
}
