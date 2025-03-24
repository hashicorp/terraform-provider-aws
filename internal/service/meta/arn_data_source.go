// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package meta

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_arn", name="ARN")
func newARNDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	d := &arnDataSource{}

	return d, nil
}

type arnDataSource struct {
	framework.DataSourceWithConfigure
}

func (d *arnDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"account": schema.StringAttribute{
				Computed: true,
			},
			names.AttrARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			names.AttrID: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"partition": schema.StringAttribute{
				Computed: true,
			},
			names.AttrRegion: schema.StringAttribute{
				Computed: true,
			},
			"resource": schema.StringAttribute{
				Computed: true,
			},
			"service": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *arnDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data arnDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	arn := data.ARN.ValueARN()

	data.Account = fwflex.StringValueToFrameworkLegacy(ctx, arn.AccountID)
	data.ID = fwflex.StringValueToFrameworkLegacy(ctx, arn.String())
	data.Partition = fwflex.StringValueToFrameworkLegacy(ctx, arn.Partition)
	data.Region = fwflex.StringValueToFrameworkLegacy(ctx, arn.Region)
	data.Resource = fwflex.StringValueToFrameworkLegacy(ctx, arn.Resource)
	data.Service = fwflex.StringValueToFrameworkLegacy(ctx, arn.Service)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type arnDataSourceModel struct {
	Account   types.String `tfsdk:"account"`
	ARN       fwtypes.ARN  `tfsdk:"arn"`
	ID        types.String `tfsdk:"id"`
	Partition types.String `tfsdk:"partition"`
	Region    types.String `tfsdk:"region"`
	Resource  types.String `tfsdk:"resource"`
	Service   types.String `tfsdk:"service"`
}
