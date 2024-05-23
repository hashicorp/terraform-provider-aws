// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package shield

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/shield"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource(name="Protection")
func newDataSourceProtection(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceProtection{}, nil
}

const (
	DSNameProtection = "Protection Data Source"
)

type dataSourceProtection struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceProtection) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	resp.TypeName = "aws_shield_protection"
}

func (d *dataSourceProtection) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrResourceARN: schema.StringAttribute{
				Optional: true,
			},
			"protection_id": schema.StringAttribute{
				Optional: true,
			},
			names.AttrID: framework.IDAttribute(),
		},
	}
}

func (d *dataSourceProtection) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().ShieldClient(ctx)

	var data dataSourceProtectionData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := &shield.DescribeProtectionInput{
		ProtectionId: data.ProtectionId.ValueStringPointer(),
		ResourceArn:  data.ResourceArn.ValueStringPointer(),
	}

	out, err := findProtection(ctx, conn, input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Shield, create.ErrActionReading, DSNameProtection, fmt.Sprintf("%s%s", data.ID.String(), data.ResourceArn.String()), err),
			err.Error(),
		)
		return
	}

	data.ARN = flex.StringToFramework(ctx, out.ProtectionArn)
	data.ResourceArn = flex.StringToFramework(ctx, out.ResourceArn)
	data.ID = flex.StringToFramework(ctx, out.Id)
	data.ProtectionId = flex.StringToFramework(ctx, out.Id)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceProtectionData struct {
	ARN          types.String `tfsdk:"arn"`
	ResourceArn  types.String `tfsdk:"resource_arn"`
	ProtectionId types.String `tfsdk:"protection_id"`
	ID           types.String `tfsdk:"id"`
}
