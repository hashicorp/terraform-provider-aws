// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package meta

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_service", name="Service")
func newServiceDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	d := &serviceDataSource{}

	return d, nil
}

type serviceDataSource struct {
	framework.DataSourceWithConfigure
}

func (d *serviceDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
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

func (d *serviceDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data serviceDataSourceModel
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

		data.Region = fwflex.StringValueToFrameworkLegacy(ctx, serviceParts[n-2])
		data.ReverseDNSPrefix = fwflex.StringValueToFrameworkLegacy(ctx, strings.Join(serviceParts[0:n-2], "."))
		data.ServiceID = fwflex.StringValueToFrameworkLegacy(ctx, serviceParts[n-1])
	}

	if !data.DNSName.IsNull() {
		v := data.DNSName.ValueString()
		serviceParts := tfslices.Reverse(strings.Split(v, "."))
		n := len(serviceParts)

		if n < 4 {
			response.Diagnostics.AddError("service DNS names must have at least 4 parts", fmt.Sprintf("%s has %d", v, n))

			return
		}

		data.Region = fwflex.StringValueToFrameworkLegacy(ctx, serviceParts[n-2])
		data.ReverseDNSPrefix = fwflex.StringValueToFrameworkLegacy(ctx, strings.Join(serviceParts[0:n-2], "."))
		data.ServiceID = fwflex.StringValueToFrameworkLegacy(ctx, serviceParts[n-1])
	}

	if data.Region.IsNull() {
		data.Region = fwflex.StringValueToFrameworkLegacy(ctx, d.Meta().Region(ctx))
	}

	if data.ServiceID.IsNull() {
		response.Diagnostics.AddError("service ID not provided directly or through a DNS name", "")

		return
	}

	if data.ReverseDNSPrefix.IsNull() {
		dnsParts := strings.Split(d.Meta().DNSSuffix(ctx), ".")
		data.ReverseDNSPrefix = fwflex.StringValueToFrameworkLegacy(ctx, strings.Join(tfslices.Reverse(dnsParts), "."))
	}

	reverseDNSName := fmt.Sprintf("%s.%s.%s", data.ReverseDNSPrefix.ValueString(), data.Region.ValueString(), data.ServiceID.ValueString())
	data.ReverseDNSName = fwflex.StringValueToFrameworkLegacy(ctx, reverseDNSName)
	data.DNSName = fwflex.StringValueToFrameworkLegacy(ctx, strings.ToLower(strings.Join(tfslices.Reverse(strings.Split(reverseDNSName, ".")), ".")))

	data.Supported = types.BoolValue(true)
	if partition, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), data.Region.ValueString()); ok {
		data.Partition = fwflex.StringValueToFrameworkLegacy(ctx, partition.ID())

		if _, ok := partition.Services()[data.ServiceID.ValueString()]; !ok {
			data.Supported = types.BoolValue(false)
		}
	} else {
		data.Partition = types.StringNull()
	}

	data.ID = fwflex.StringValueToFrameworkLegacy(ctx, reverseDNSName)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type serviceDataSourceModel struct {
	DNSName          types.String `tfsdk:"dns_name"`
	ID               types.String `tfsdk:"id"`
	Partition        types.String `tfsdk:"partition"`
	Region           types.String `tfsdk:"region"`
	ReverseDNSName   types.String `tfsdk:"reverse_dns_name"`
	ReverseDNSPrefix types.String `tfsdk:"reverse_dns_prefix"`
	ServiceID        types.String `tfsdk:"service_id"`
	Supported        types.Bool   `tfsdk:"supported"`
}
