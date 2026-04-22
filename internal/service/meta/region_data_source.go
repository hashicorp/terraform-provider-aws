// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package meta

import (
	"context"
	"fmt"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_region", name="Region")
// @Region(validateOverrideInPartition=false)
func newRegionDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	d := &regionDataSource{}

	return d, nil
}

type regionDataSource struct {
	framework.DataSourceWithModel[regionDataSourceModel]
}

func (d *regionDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrDescription: schema.StringAttribute{
				Computed: true,
			},
			names.AttrEndpoint: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			names.AttrID: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			names.AttrName: schema.StringAttribute{
				Optional:           true,
				Computed:           true,
				DeprecationMessage: "name is deprecated. Use region instead.",
			},
		},
	}
}

func (d *regionDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data regionDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	var region *endpoints.Region

	if !data.Endpoint.IsNull() {
		endpoint := data.Endpoint.ValueString()
		matchingRegion, err := findRegionByEC2Endpoint(ctx, endpoint)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("finding Region by endpoint (%s)", endpoint), err.Error())

			return
		}

		region = matchingRegion
	}

	var name string
	if !data.Region.IsNull() {
		name = data.Region.ValueString()
	} else if !data.Name.IsNull() {
		name = data.Name.ValueString()
	}
	if name != "" {
		matchingRegion, err := findRegionByName(ctx, name)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("finding Region by name (%s)", name), err.Error())

			return
		}

		if region != nil && region.ID() != matchingRegion.ID() {
			response.Diagnostics.AddError("multiple Regions matched", "use additional constraints to reduce matches to a single Region")

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

	regionEndpointEC2, err := ec2Endpoint(ctx, region)

	if err != nil {
		response.Diagnostics.AddError("resolving EC2 endpoint", err.Error())

		return
	}

	data.Description = fwflex.StringValueToFrameworkLegacy(ctx, region.Description())
	data.Endpoint = fwflex.StringValueToFrameworkLegacy(ctx, regionEndpointEC2.Host)
	data.ID = fwflex.StringValueToFrameworkLegacy(ctx, region.ID())
	data.Name = fwflex.StringValueToFrameworkLegacy(ctx, region.ID())
	data.Region = fwflex.StringValueToFrameworkLegacy(ctx, region.ID())

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (d *regionDataSource) ConfigValidators(context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.Conflicting(
			path.MatchRoot(names.AttrName),
			path.MatchRoot(names.AttrRegion),
		),
	}
}

type regionDataSourceModel struct {
	framework.WithRegionModel
	Description types.String `tfsdk:"description"`
	Endpoint    types.String `tfsdk:"endpoint"`
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
}

func findRegionByEC2Endpoint(ctx context.Context, endpoint string) (*endpoints.Region, error) {
	for _, partition := range endpoints.DefaultPartitions() {
		for _, region := range partition.Regions() {
			regionEndpointEC2, err := ec2Endpoint(ctx, &region)

			if err != nil {
				return nil, err
			}

			if regionEndpointEC2.Host == endpoint {
				return &region, nil
			}
		}
	}

	return nil, &retry.NotFoundError{}
}

func findRegionByName(_ context.Context, name string) (*endpoints.Region, error) {
	for _, partition := range endpoints.DefaultPartitions() {
		for _, region := range partition.Regions() {
			if region.ID() == name {
				return &region, nil
			}
		}
	}

	return nil, &retry.NotFoundError{}
}

func ec2Endpoint(ctx context.Context, region *endpoints.Region) (*url.URL, error) {
	endpoint, err := ec2.NewDefaultEndpointResolverV2().ResolveEndpoint(ctx, ec2.EndpointParameters{
		Region: aws.String(region.ID()),
	})
	if err != nil {
		return nil, err
	}

	return &endpoint.URI, nil
}
