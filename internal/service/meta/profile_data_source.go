// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package meta

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
)

// @FrameworkDataSource
func newDataSourceProfile(context.Context) (datasource.DataSourceWithConfigure, error) {
	d := &dataSourceProfile{}
	d.SetMigratedFromPluginSDK(true)

	return d, nil
}

type dataSourceProfile struct {
	framework.DataSourceWithConfigure
}

// Metadata should return the full name of the data source, such as
// examplecloud_thing.
func (d *dataSourceProfile) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	response.TypeName = "aws_profile"
}

// Schema returns the schema for this data source.
func (d *dataSourceProfile) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"name": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
		},
	}
}

// Read is called when the provider must read data source values in order to update state.
// Config values should be read from the ReadRequest and new state values set on the ReadResponse.
func (d *dataSourceProfile) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data dataSourceProfileData

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	var profile string

	if profile = d.Meta().Profile; profile == "" {
		response.Diagnostics.AddError("Getting AWS profile", fmt.Errorf("AWS profile not set").Error())
		return
	}

	data.ID = types.StringValue(profile)
	data.Name = types.StringValue(profile)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type dataSourceProfileData struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}
