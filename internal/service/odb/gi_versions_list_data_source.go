//Copyright © 2025, Oracle and/or its affiliates. All rights reserved.

package odb

import (
	"context"
	"fmt"

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
// @FrameworkDataSource("aws_odb_gi_versions_list", name="Gi Versions List")
func newDataSourceGiVersionsList(context.Context) (datasource.DataSourceWithConfigure, error) {
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
		},
		Blocks: map[string]schema.Block{
			"gi_versions": schema.ListNestedBlock{
				Description: fmt.Sprint(" (structure)\n " +
					"Information about a specific version of Oracle Grid\n" +
					"Infrastructure (GI) software that can be installed on a VM\n " +
					"cluster.\n\n " +
					"version -> (string)\n " +
					"The GI software version."),
				CustomType: fwtypes.NewListNestedObjectTypeOf[giVersionSummaryModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"version": schema.StringAttribute{
							Computed: true,
						},
					},
				},
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
	out, err := conn.ListGiVersions(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ODB, create.ErrActionReading, DSNameGiVersionsList, "", err),
			err.Error(),
		)
		return
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
