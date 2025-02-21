// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package shield

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/shield"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_shield_protection", name="Protection")
func newDataSourceProtection(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceProtection{}, nil
}

const (
	DSNameProtection = "Protection Data Source"
)

type dataSourceProtection struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceProtection) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Computed: true,
			},
			"protection_arn": framework.ARNAttributeComputedOnly(),
			"protection_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			names.AttrResourceARN: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
		},
	}
}

func (d *dataSourceProtection) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.ExactlyOneOf(
			path.MatchRoot("protection_id"),
			path.MatchRoot(names.AttrResourceARN),
		),
	}
}

func (d *dataSourceProtection) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().ShieldClient(ctx)

	var data dataSourceProtectionData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &shield.DescribeProtectionInput{}
	if !data.ProtectionID.IsNull() {
		data.ID = types.StringValue(data.ProtectionID.ValueString())
		input.ProtectionId = data.ProtectionID.ValueStringPointer()
	} else {
		data.ID = types.StringValue(data.ResourceARN.ValueString())
		input.ResourceArn = data.ResourceARN.ValueStringPointer()
	}

	out, err := findProtection(ctx, conn, input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Shield, create.ErrActionReading, DSNameProtection, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ProtectionID = flex.StringToFramework(ctx, out.Id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceProtectionData struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	ProtectionARN types.String `tfsdk:"protection_arn"`
	ProtectionID  types.String `tfsdk:"protection_id"`
	ResourceARN   types.String `tfsdk:"resource_arn"`
}
