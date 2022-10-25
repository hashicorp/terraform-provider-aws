package meta

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

func init() {
	registerFrameworkDataSourceFactory(newDataSourceRegions)
}

// newDataSourceRegions instantiates a new DataSource for the aws_regions data source.
func newDataSourceRegions(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceRegions{}, nil
}

type dataSourceRegions struct {
	meta *conns.AWSClient
}

// Metadata should return the full name of the data source, such as
// examplecloud_thing.
func (d *dataSourceRegions) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	response.TypeName = "aws_regions"
}

// GetSchema returns the schema for this data source.
func (d *dataSourceRegions) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	schema := tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"all_regions": {
				Type:     types.BoolType,
				Optional: true,
			},
			"id": {
				Type:     types.StringType,
				Optional: true,
				Computed: true,
			},
			"names": {
				Type:     types.SetType{ElemType: types.StringType},
				Computed: true,
			},
		},
		Blocks: map[string]tfsdk.Block{
			"filter": tfec2.CustomFiltersBlock(),
		},
	}

	return schema, nil
}

// Configure enables provider-level data or clients to be set in the
// provider-defined DataSource type. It is separately executed for each
// ReadDataSource RPC.
func (d *dataSourceRegions) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	if v, ok := request.ProviderData.(*conns.AWSClient); ok {
		d.meta = v
	}
}

// Read is called when the provider must read data source values in order to update state.
// Config values should be read from the ReadRequest and new state values set on the ReadResponse.
func (d *dataSourceRegions) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data dataSourceRegionsData

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	input := &ec2.DescribeRegionsInput{
		Filters: tfec2.BuildCustomFilters(ctx, data.Filters),
	}

	if !data.AllRegions.IsNull() {
		input.AllRegions = aws.Bool(data.AllRegions.Value)
	}

	output, err := d.meta.EC2Conn.DescribeRegionsWithContext(ctx, input)

	if err != nil {
		response.Diagnostics.AddError("reading Regions", err.Error())

		return
	}

	var names []string
	for _, v := range output.Regions {
		names = append(names, aws.StringValue(v.RegionName))
	}

	data.ID = types.String{Value: d.meta.Partition}
	data.Names = flex.FlattenFrameworkStringValueSet(ctx, names)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type dataSourceRegionsData struct {
	AllRegions types.Bool   `tfsdk:"all_regions"`
	Filters    types.Set    `tfsdk:"filter"`
	ID         types.String `tfsdk:"id"`
	Names      types.Set    `tfsdk:"names"`
}
