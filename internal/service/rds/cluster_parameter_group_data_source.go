// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

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

// @FrameworkDataSource(name=Cluster Parameter Group)
func newDataSourceClusterParameterGroup(_ context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceClusterParameterGroup{}, nil
}

const (
	DSNameClusterParameterGroup = "Cluster Parameter Group Data Source"

	dbClusterParameterGroupPrefix = "DBClusterParameterGroup"
)

type dataSourceClusterParameterGroup struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceClusterParameterGroup) Metadata(_ context.Context, _ datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "aws_rds_cluster_parameter_group"
}

func (d *dataSourceClusterParameterGroup) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDescription: schema.StringAttribute{
				Computed: true,
			},
			names.AttrFamily: schema.StringAttribute{
				Computed: true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (d *dataSourceClusterParameterGroup) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	conn := d.Meta().RDSClient(ctx)
	var data dataSourceClusterParameterGroupData

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	output, err := findDBClusterParameterGroupByName(ctx, conn, data.Name.ValueString())

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.RDS, create.ErrActionReading, DSNameClusterParameterGroup, data.Name.String(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data, fwflex.WithFieldNamePrefix(dbClusterParameterGroupPrefix))...)

	if response.Diagnostics.HasError() {
		return
	}

	data.Family = fwflex.StringToFramework(ctx, output.DBParameterGroupFamily)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type dataSourceClusterParameterGroupData struct {
	ARN         types.String `tfsdk:"arn"`
	Description types.String `tfsdk:"description"`
	Family      types.String `tfsdk:"family"`
	Name        types.String `tfsdk:"name"`
}
