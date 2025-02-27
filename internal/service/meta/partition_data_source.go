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
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_partition", name="Partition")
func newPartitionDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	d := &partitionDataSource{}

	return d, nil
}

type partitionDataSource struct {
	framework.DataSourceWithConfigure
}

func (d *partitionDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
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

func (d *partitionDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data partitionDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	data.DNSSuffix = fwflex.StringValueToFrameworkLegacy(ctx, d.Meta().DNSSuffix(ctx))
	data.ID = fwflex.StringValueToFrameworkLegacy(ctx, d.Meta().Partition(ctx))
	data.Partition = fwflex.StringValueToFrameworkLegacy(ctx, d.Meta().Partition(ctx))
	data.ReverseDNSPrefix = fwflex.StringValueToFrameworkLegacy(ctx, d.Meta().ReverseDNSPrefix(ctx))

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type partitionDataSourceModel struct {
	DNSSuffix        types.String `tfsdk:"dns_suffix"`
	ID               types.String `tfsdk:"id"`
	Partition        types.String `tfsdk:"partition"`
	ReverseDNSPrefix types.String `tfsdk:"reverse_dns_prefix"`
}
