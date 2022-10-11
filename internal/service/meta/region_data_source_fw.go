package meta

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func init() {
	registerFWDataSourceFactory(newDataSourceRegion)
}

// newDataSourceRegion instantiates a new DataSource for the aws_region data source.
func newDataSourceRegion(context.Context) (datasource.DataSource, error) {
	return &dataSourceRegion{}, nil
}

type dataSourceRegion struct {
	meta any
}

// Metadata should return the full name of the data source, such as
// examplecloud_thing.
func (d *dataSourceRegion) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "aws_region"
}

// GetSchema returns the schema for this data source.
func (d *dataSourceRegion) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	schema := tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"description": {
				Type:     types.StringType,
				Computed: true,
			},
			"endpoint": {
				Type:     types.StringType,
				Optional: true,
				Computed: true,
			},
			"id": {
				Type:     types.StringType,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     types.StringType,
				Optional: true,
				Computed: true,
			},
		},
	}

	return schema, nil
}

// Configure enables provider-level data or clients to be set in the
// provider-defined DataSource type. It is separately executed for each
// ReadDataSource RPC.
func (d *dataSourceRegion) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	d.meta = request.ProviderData
}

// Read is called when the provider must read data source values in order to update state.
// Config values should be read from the ReadRequest and new state values set on the ReadResponse.
func (d *dataSourceRegion) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data dataSourceRegionData

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	var region *endpoints.Region

	if !data.Endpoint.IsNull() && !data.Endpoint.IsUnknown() {
		matchingRegion, err := FindRegionByEndpoint(data.Endpoint.Value)

		if err != nil {
			// TODO
			response.Diagnostics.AddError("", "")

			return
		}

		region = matchingRegion
	}

	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		matchingRegion, err := FindRegionByName(data.Name.Value)

		if err != nil {
			// TODO
			response.Diagnostics.AddError("", "")

			return
		}

		if region != nil && region.ID() != matchingRegion.ID() {
			// TODO
			// "multiple regions matched; use additional constraints to reduce matches to a single region"
			response.Diagnostics.AddError("", "")

			return
		}

		region = matchingRegion
	}

	// Default to provider current region if no other filters matched
	if region == nil {
		matchingRegion, err := FindRegionByName(d.meta.(*conns.AWSClient).Region)

		if err != nil {
			// TODO
			response.Diagnostics.AddError("", "")

			return
		}

		region = matchingRegion
	}

	regionEndpointEC2, err := region.ResolveEndpoint(endpoints.Ec2ServiceID)

	if err != nil {
		// TODO
		response.Diagnostics.AddError("", "")

		return
	}

	data.Description = types.String{Value: region.Description()}
	data.Endpoint = types.String{Value: strings.TrimPrefix(regionEndpointEC2.URL, "https://")}
	data.ID = types.String{Value: region.ID()}
	data.Name = types.String{Value: region.ID()}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type dataSourceRegionData struct {
	Description types.String `tfsdk:"description"`
	Endpoint    types.String `tfsdk:"endpoint"`
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
}
