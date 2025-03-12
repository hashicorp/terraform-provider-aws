// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package meta

import (
	"context"
	"fmt"

	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_service_principal", name="Service Principal")
func newServicePrincipalDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	d := &servicePrincipalDataSource{}

	return d, nil
}

type servicePrincipalDataSource struct {
	framework.DataSourceWithConfigure
}

func (d *servicePrincipalDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: schema.StringAttribute{
				Computed: true,
			},
			names.AttrName: schema.StringAttribute{
				Computed: true,
			},
			names.AttrRegion: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			names.AttrServiceName: schema.StringAttribute{
				Required: true,
			},
			"suffix": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *servicePrincipalDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data servicePrincipalDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	var region *endpoints.Region

	// find the region given by the user
	if !data.Region.IsNull() {
		name := data.Region.ValueString()
		matchingRegion, err := findRegionByName(ctx, name)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("finding Region by name (%s)", name), err.Error())

			return
		}

		region = matchingRegion
	}

	// Default to provider current Region if no other filters matched.
	if region == nil {
		name := d.Meta().Region(ctx)
		matchingRegion, err := findRegionByName(ctx, name)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("finding Region by name (%s)", name), err.Error())

			return
		}

		region = matchingRegion
	}

	regionID := region.ID()
	serviceName := fwflex.StringValueFromFramework(ctx, data.ServiceName)
	sourceServicePrincipal := servicePrincipalNameForPartition(serviceName, names.PartitionForRegion(regionID))

	data.ID = fwflex.StringValueToFrameworkLegacy(ctx, serviceName+"."+regionID+"."+sourceServicePrincipal)
	data.Name = fwflex.StringValueToFrameworkLegacy(ctx, serviceName+"."+sourceServicePrincipal)
	data.Suffix = fwflex.StringValueToFrameworkLegacy(ctx, sourceServicePrincipal)
	data.Region = fwflex.StringValueToFrameworkLegacy(ctx, regionID)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type servicePrincipalDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Region      types.String `tfsdk:"region"`
	ServiceName types.String `tfsdk:"service_name"`
	Suffix      types.String `tfsdk:"suffix"`
}

// SPN region unique taken from
// https://github.com/aws/aws-cdk/blob/main/packages/aws-cdk-lib/region-info/lib/default.ts
func servicePrincipalNameForPartition(service string, partition endpoints.Partition) string {
	if partitionID := partition.ID(); service != "" && partitionID != endpoints.AwsPartitionID {
		switch partitionID {
		case endpoints.AwsIsoPartitionID:
			switch service {
			case "cloudhsm",
				"config",
				"logs",
				"workspaces":
				return partition.DNSSuffix()
			}
		case endpoints.AwsIsoBPartitionID:
			switch service {
			case "dms",
				"logs":
				return partition.DNSSuffix()
			}
		case endpoints.AwsCnPartitionID:
			switch service {
			case "codedeploy",
				"elasticmapreduce",
				"logs":
				return partition.DNSSuffix()
			}
		}
	}

	return "amazonaws.com"
}
