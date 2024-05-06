// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package meta

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource
func newDataSourcePartition(context.Context) (datasource.DataSourceWithConfigure, error) {
	d := &dataSourcePartition{}

	return d, nil
}

type dataSourcePartition struct {
	framework.DataSourceWithConfigure
}

// Metadata should return the full name of the data source, such as
// examplecloud_thing.
func (d *dataSourcePartition) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	response.TypeName = "aws_partition"
}

// Schema returns the schema for this data source.
func (d *dataSourcePartition) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"dns_suffix": schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"partition": schema.StringAttribute{
				Computed: true,
			},
			"reverse_dns_prefix": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

// Read is called when the provider must read data source values in order to update state.
// Config values should be read from the ReadRequest and new state values set on the ReadResponse.
func (d *dataSourcePartition) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data dataSourcePartitionData

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	data.DNSSuffix = types.StringValue(d.Meta().DNSSuffix(ctx))
	data.ID = types.StringValue(d.Meta().Partition)
	data.Partition = types.StringValue(d.Meta().Partition)
	data.ReverseDNSPrefix = types.StringValue(d.Meta().ReverseDNSPrefix(ctx))

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type dataSourcePartitionData struct {
	DNSSuffix        types.String `tfsdk:"dns_suffix"`
	ID               types.String `tfsdk:"id"`
	Partition        types.String `tfsdk:"partition"`
	ReverseDNSPrefix types.String `tfsdk:"reverse_dns_prefix"`
}
