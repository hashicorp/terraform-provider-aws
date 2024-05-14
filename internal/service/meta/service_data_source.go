// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package meta

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource
func newDataSourceService(context.Context) (datasource.DataSourceWithConfigure, error) {
	d := &dataSourceService{}

	return d, nil
}

type dataSourceService struct {
	framework.DataSourceWithConfigure
}

// Metadata should return the full name of the data source, such as
// examplecloud_thing.
func (d *dataSourceService) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	response.TypeName = "aws_service"
}

// Schema returns the schema for this data source.
func (d *dataSourceService) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrDNSName: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			names.AttrID: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"partition": schema.StringAttribute{
				Computed: true,
			},
			names.AttrRegion: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"reverse_dns_name": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"reverse_dns_prefix": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"service_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"supported": schema.BoolAttribute{
				Computed: true,
			},
		},
	}
}

// Read is called when the provider must read data source values in order to update state.
// Config values should be read from the ReadRequest and new state values set on the ReadResponse.
func (d *dataSourceService) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data dataSourceServiceData

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	if !data.ReverseDNSName.IsNull() {
		v := data.ReverseDNSName.ValueString()
		serviceParts := strings.Split(v, ".")
		n := len(serviceParts)

		if n < 4 {
			response.Diagnostics.AddError("reverse service DNS names must have at least 4 parts", fmt.Sprintf("%s has %d", v, n))

			return
		}

		data.Region = types.StringValue(serviceParts[n-2])
		data.ReverseDNSPrefix = types.StringValue(strings.Join(serviceParts[0:n-2], "."))
		data.ServiceID = types.StringValue(serviceParts[n-1])
	}

	if !data.DNSName.IsNull() {
		v := data.DNSName.ValueString()
		serviceParts := slices.Reverse(strings.Split(v, "."))
		n := len(serviceParts)

		if n < 4 {
			response.Diagnostics.AddError("service DNS names must have at least 4 parts", fmt.Sprintf("%s has %d", v, n))

			return
		}

		data.Region = types.StringValue(serviceParts[n-2])
		data.ReverseDNSPrefix = types.StringValue(strings.Join(serviceParts[0:n-2], "."))
		data.ServiceID = types.StringValue(serviceParts[n-1])
	}

	if data.Region.IsNull() {
		data.Region = types.StringValue(d.Meta().Region)
	}

	if data.ServiceID.IsNull() {
		response.Diagnostics.AddError("service ID not provided directly or through a DNS name", "")

		return
	}

	if data.ReverseDNSPrefix.IsNull() {
		dnsParts := strings.Split(d.Meta().DNSSuffix(ctx), ".")
		data.ReverseDNSPrefix = types.StringValue(strings.Join(slices.Reverse(dnsParts), "."))
	}

	reverseDNSName := fmt.Sprintf("%s.%s.%s", data.ReverseDNSPrefix.ValueString(), data.Region.ValueString(), data.ServiceID.ValueString())
	data.ReverseDNSName = types.StringValue(reverseDNSName)
	data.DNSName = types.StringValue(strings.ToLower(strings.Join(slices.Reverse(strings.Split(reverseDNSName, ".")), ".")))

	data.Supported = types.BoolValue(true)
	if partition, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), data.Region.ValueString()); ok {
		data.Partition = types.StringValue(partition.ID())

		if _, ok := partition.Services()[data.ServiceID.ValueString()]; !ok {
			data.Supported = types.BoolValue(false)
		}
	} else {
		data.Partition = types.StringNull()
	}

	data.ID = types.StringValue(reverseDNSName)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type dataSourceServiceData struct {
	DNSName          types.String `tfsdk:"dns_name"`
	ID               types.String `tfsdk:"id"`
	Partition        types.String `tfsdk:"partition"`
	Region           types.String `tfsdk:"region"`
	ReverseDNSName   types.String `tfsdk:"reverse_dns_name"`
	ReverseDNSPrefix types.String `tfsdk:"reverse_dns_prefix"`
	ServiceID        types.String `tfsdk:"service_id"`
	Supported        types.Bool   `tfsdk:"supported"`
}
