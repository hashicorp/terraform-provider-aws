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
func newRuntimeVersionsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &runtimeVersionsDataSource{}, nil
}

const (
	DSNameRuntimeVersions = "Runtime Versions Data Source"
)

type runtimeVersionsDataSource struct {
	framework.DataSourceWithModel[runtimeVersionsDataSourceModel]
}

func (d *runtimeVersionsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID:       framework.IDAttribute(),
			"runtime_versions": framework.DataSourceComputedListOfObjectAttribute[runtimeVersionModel](ctx),
		},
	}
}

func (d *runtimeVersionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().SyntheticsClient(ctx)

	var data runtimeVersionsDataSourceModel
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

type runtimeVersionsDataSourceModel struct {
	framework.WithRegionModel
	ID              types.String                                         `tfsdk:"id"`
	RuntimeVersions fwtypes.ListNestedObjectValueOf[runtimeVersionModel] `tfsdk:"runtime_versions"`
}

type runtimeVersionModel struct {
	DeprecationDate timetypes.RFC3339 `tfsdk:"deprecation_date"`
	Description     types.String      `tfsdk:"description"`
	ReleaseDate     timetypes.RFC3339 `tfsdk:"release_date"`
	VersionName     types.String      `tfsdk:"version_name"`
}
