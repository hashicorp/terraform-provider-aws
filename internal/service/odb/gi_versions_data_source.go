// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package odb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/odb"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for datasource registration to the Provider. DO NOT EDIT.
// @FrameworkDataSource("aws_odb_gi_versions", name="Gi Versions")
func newDataSourceGiVersions(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceGiVersionsList{}, nil
}

const (
	DSNameGiVersionsList = "Gi Versions List Data Source"
)

type dataSourceGiVersionsList struct {
	framework.DataSourceWithModel[giVersionDataSourceModel]
}

func (d *dataSourceGiVersionsList) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"shape": schema.StringAttribute{
				Optional:    true,
				Description: "The system shape.",
			},
			"gi_versions": schema.ListAttribute{
				Computed:    true,
				CustomType:  fwtypes.NewListNestedObjectTypeOf[giVersionSummaryModel](ctx),
				Description: "Information about a specific version of Oracle Grid Infrastructure (GI) software that can be installed on a VM cluster.",
			},
		},
	}
}

func (d *dataSourceGiVersionsList) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().ODBClient(ctx)
	var data giVersionDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var input odb.ListGiVersionsInput
	if !data.Shape.IsNull() {
		input.Shape = data.Shape.ValueStringPointer()
	}
	paginator := odb.NewListGiVersionsPaginator(conn, &input)
	var out odb.ListGiVersionsOutput
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.ODB, create.ErrActionReading, DSNameGiVersionsList, "", err),
				err.Error(),
			)
			return
		}
		if page != nil && len(page.GiVersions) > 0 {
			out.GiVersions = append(out.GiVersions, page.GiVersions...)
		}
	}
	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type giVersionDataSourceModel struct {
	framework.WithRegionModel
	GiVersions fwtypes.ListNestedObjectValueOf[giVersionSummaryModel] `tfsdk:"gi_versions"`
	Shape      types.String                                           `tfsdk:"shape"`
}

type giVersionSummaryModel struct {
	Version types.String `tfsdk:"version"`
}
