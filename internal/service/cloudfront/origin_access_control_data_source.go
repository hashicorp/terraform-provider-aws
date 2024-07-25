// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource(name="Origin Access Control")
func newDataSourceOriginAccessControl(_ context.Context) (datasource.DataSourceWithConfigure, error) {
	d := &dataSourceOriginAccessControl{}

	return d, nil
}

type dataSourceOriginAccessControl struct {
	framework.DataSourceWithConfigure
}

const (
	DSNameOriginAccessControl = "Origin Access Control Data Source"
)

func (d *dataSourceOriginAccessControl) Metadata(_ context.Context, _ datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "aws_cloudfront_origin_access_control"
}

func (d *dataSourceOriginAccessControl) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrDescription: schema.StringAttribute{
				Computed: true,
			},
			"etag": schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: schema.StringAttribute{
				Required: true,
			},
			names.AttrName: schema.StringAttribute{
				Computed: true,
			},
			"origin_access_control_origin_type": schema.StringAttribute{
				Computed: true,
			},
			"signing_behavior": schema.StringAttribute{
				Computed: true,
			},
			"signing_protocol": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *dataSourceOriginAccessControl) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	conn := d.Meta().CloudFrontClient(ctx)
	var data dataSourceOriginAccessControlData

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	output, err := findOriginAccessControlByID(ctx, conn, data.ID.ValueString())

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CloudFront, create.ErrActionReading, DSNameOriginAccessControl, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output.OriginAccessControl.OriginAccessControlConfig, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	data.Etag = fwflex.StringToFramework(ctx, output.ETag)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type dataSourceOriginAccessControlData struct {
	Description                   types.String `tfsdk:"description"`
	Etag                          types.String `tfsdk:"etag"`
	ID                            types.String `tfsdk:"id"`
	Name                          types.String `tfsdk:"name"`
	OriginAccessControlOriginType types.String `tfsdk:"origin_access_control_origin_type"`
	SigningBehavior               types.String `tfsdk:"signing_behavior"`
	SigningProtocol               types.String `tfsdk:"signing_protocol"`
}
