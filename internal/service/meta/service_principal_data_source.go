// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package meta

import (
	"context"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource
func newDataSourceServicePrincipal(context.Context) (datasource.DataSourceWithConfigure, error) {
	d := &dataSourceServicePrincipal{}

	return d, nil
}

type dataSourceServicePrincipal struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceServicePrincipal) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	response.TypeName = "aws_service_principal"
}

func (d *dataSourceServicePrincipal) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			names.AttrName: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			names.AttrServiceName: schema.StringAttribute{
				Required: true,
			},
			"suffix": schema.StringAttribute{
				Computed: true,
			},
			names.AttrRegion: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
		},
	}
}

// Read is called when the provider must read data source values in order to update state.
// Config values should be read from the ReadRequest and new state values set on the ReadResponse.
func (d *dataSourceServicePrincipal) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data dataSourceServicePrincipalData

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	var region *endpoints.Region

	// find the region given by the user
	if !data.Region.IsNull() {
		matchingRegion, err := FindRegionByName(data.Region.ValueString())

		if err != nil {
			response.Diagnostics.AddError("finding Region by name", err.Error())

			return
		}

		region = matchingRegion
	}

	// Default to provider current region if no other filters matched
	if region == nil {
		matchingRegion, err := FindRegionByName(d.Meta().Region)

		if err != nil {
			response.Diagnostics.AddError("finding Region using the provider", err.Error())

			return
		}

		region = matchingRegion
	}

	partition := names.PartitionForRegion(region.ID())

	serviceName := ""

	if !data.ServiceName.IsNull() {
		serviceName = data.ServiceName.ValueString()
	}

	SourceServicePrincipal := names.ServicePrincipalNameForPartition(serviceName, partition)

	data.ID = types.StringValue(serviceName + "." + region.ID() + "." + SourceServicePrincipal)
	data.Name = types.StringValue(serviceName + "." + SourceServicePrincipal)
	data.Suffix = types.StringValue(SourceServicePrincipal)
	data.Region = types.StringValue(region.ID())
	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type dataSourceServicePrincipalData struct {
	Suffix      types.String `tfsdk:"suffix"`
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	ServiceName types.String `tfsdk:"service_name"`
	Region      types.String `tfsdk:"region"`
}
