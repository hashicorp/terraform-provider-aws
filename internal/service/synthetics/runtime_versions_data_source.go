// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package synthetics

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_synthetics_runtime_versions", name="Runtime Versions")
func newDataSourceRuntimeVersions(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceRuntimeVersions{}, nil
}

const (
	DSNameRuntimeVersions = "Runtime Versions Data Source"
)

type dataSourceRuntimeVersions struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceRuntimeVersions) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"runtime_versions": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[runtimeVersionModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"deprecation_date": schema.StringAttribute{
							CustomType: timetypes.RFC3339Type{},
							Computed:   true,
						},
						names.AttrDescription: schema.StringAttribute{
							Computed: true,
						},
						"release_date": schema.StringAttribute{
							CustomType: timetypes.RFC3339Type{},
							Computed:   true,
						},
						"version_name": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *dataSourceRuntimeVersions) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().SyntheticsClient(ctx)

	var data dataSourceRuntimeVersionsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findRuntimeVersions(ctx, conn)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Synthetics, create.ErrActionReading, DSNameRuntimeVersions, "", err),
			err.Error(),
		)
		return
	}

	data.ID = flex.StringValueToFramework(ctx, d.Meta().Region(ctx))
	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data.RuntimeVersions)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceRuntimeVersionsModel struct {
	ID              types.String                                         `tfsdk:"id"`
	RuntimeVersions fwtypes.ListNestedObjectValueOf[runtimeVersionModel] `tfsdk:"runtime_versions"`
}

type runtimeVersionModel struct {
	DeprecationDate timetypes.RFC3339 `tfsdk:"deprecation_date"`
	Description     types.String      `tfsdk:"description"`
	ReleaseDate     timetypes.RFC3339 `tfsdk:"release_date"`
	VersionName     types.String      `tfsdk:"version_name"`
}
