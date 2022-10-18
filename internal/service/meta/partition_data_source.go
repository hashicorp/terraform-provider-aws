package meta

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func init() {
	registerFrameworkDataSourceFactory(newDataSourcePartition)
}

// newDataSourcePartition instantiates a new DataSource for the aws_partition data source.
func newDataSourcePartition(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourcePartition{}, nil
}

type dataSourcePartition struct {
	meta *conns.AWSClient
}

// Metadata should return the full name of the data source, such as
// examplecloud_thing.
func (d *dataSourcePartition) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	response.TypeName = "aws_partition"
}

// GetSchema returns the schema for this data source.
func (d *dataSourcePartition) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	schema := tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"dns_suffix": {
				Type:     types.StringType,
				Computed: true,
			},
			"id": {
				Type:     types.StringType,
				Optional: true,
				Computed: true,
			},
			"partition": {
				Type:     types.StringType,
				Computed: true,
			},
			"reverse_dns_prefix": {
				Type:     types.StringType,
				Computed: true,
			},
		},
	}

	return schema, nil
}

// Configure enables provider-level data or clients to be set in the
// provider-defined DataSource type. It is separately executed for each
// ReadDataSource RPC.
func (d *dataSourcePartition) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	if v, ok := request.ProviderData.(*conns.AWSClient); ok {
		d.meta = v
	}
}

// Read is called when the provider must read data source values in order to update state.
// Config values should be read from the ReadRequest and new state values set on the ReadResponse.
func (d *dataSourcePartition) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data dataSourcePartitionData

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	data.DNSSuffix = types.String{Value: d.meta.DNSSuffix}
	data.ID = types.String{Value: d.meta.Partition}
	data.Partition = types.String{Value: d.meta.Partition}
	data.ReverseDNSPrefix = types.String{Value: d.meta.ReverseDNSPrefix}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type dataSourcePartitionData struct {
	DNSSuffix        types.String `tfsdk:"dns_suffix"`
	ID               types.String `tfsdk:"id"`
	Partition        types.String `tfsdk:"partition"`
	ReverseDNSPrefix types.String `tfsdk:"reverse_dns_prefix"`
}
