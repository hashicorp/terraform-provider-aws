// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambdaweb

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_lambdaweb_endpoint", name="Endpoint")
func newEndpointDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &endpointDataSource{}, nil
}

type endpointDataSource struct {
	framework.DataSourceWithModel[endpointDataSourceModel]
}

func (d *endpointDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				Computed: true,
			},
			"auth_type": schema.StringAttribute{
				Computed: true,
			},
			"auto_deployment_mode": schema.StringAttribute{
				Computed: true,
			},
			names.AttrDomainName: schema.StringAttribute{
				Computed: true,
			},
			"endpoint_name": schema.StringAttribute{
				Required: true,
			},
			names.AttrEndpointType: schema.StringAttribute{
				Computed: true,
			},
			"function_name": schema.StringAttribute{
				Required: true,
			},
			"regional_domain_names": schema.MapAttribute{
				CustomType:  fwtypes.MapOfStringType,
				Computed:    true,
				ElementType: types.StringType,
			},
			"regions": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				Computed:    true,
				ElementType: types.StringType,
			},
			names.AttrState: schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *endpointDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data endpointDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().LambdaWebClient(ctx)
	functionName := data.FunctionName.ValueString()
	endpointName := data.EndpointName.ValueString()

	out, err := findEndpointByName(ctx, conn, functionName, endpointName)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, endpointName)
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, out, &data))
	if response.Diagnostics.HasError() {
		return
	}
	data.ARN = fwflex.StringToFramework(ctx, out.EndpointArn)

	regionalDomainNames := map[string]attr.Value{}
	for region, ep := range out.RegionalEndpoints {
		// Only PerRegion endpoints carry per-region domains; MultiRegion
		// regional entries have no domain (traffic uses the global domain).
		if ep.DomainName != nil {
			regionalDomainNames[region] = fwflex.StringToFramework(ctx, ep.DomainName)
		}
	}
	data.RegionalDomainNames = fwtypes.NewMapValueOfMust[types.String](ctx, regionalDomainNames)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type endpointDataSourceModel struct {
	framework.WithRegionModel
	ARN                 types.String        `tfsdk:"arn"`
	AuthType            types.String        `tfsdk:"auth_type"`
	AutoDeploymentMode  types.String        `tfsdk:"auto_deployment_mode"`
	DomainName          types.String        `tfsdk:"domain_name"`
	EndpointName        types.String        `tfsdk:"endpoint_name"`
	EndpointType        types.String        `tfsdk:"endpoint_type"`
	FunctionName        types.String        `tfsdk:"function_name"`
	RegionalDomainNames fwtypes.MapOfString `tfsdk:"regional_domain_names"`
	Regions             fwtypes.SetOfString `tfsdk:"regions"`
	State               types.String        `tfsdk:"state"`
}
