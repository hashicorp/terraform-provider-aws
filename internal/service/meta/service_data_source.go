package meta

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/slices"
)

func init() {
	registerFrameworkDataSourceFactory(newDataSourceService)
}

// newDataSourceService instantiates a new DataSource for the aws_service data source.
func newDataSourceService(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceService{}, nil
}

type dataSourceService struct {
	meta *conns.AWSClient
}

// Metadata should return the full name of the data source, such as
// examplecloud_thing.
func (d *dataSourceService) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	response.TypeName = "aws_service"
}

// GetSchema returns the schema for this data source.
func (d *dataSourceService) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	schema := tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"dns_name": {
				Type:     types.StringType,
				Optional: true,
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
			"region": {
				Type:     types.StringType,
				Optional: true,
				Computed: true,
			},
			"reverse_dns_name": {
				Type:     types.StringType,
				Optional: true,
				Computed: true,
			},
			"reverse_dns_prefix": {
				Type:     types.StringType,
				Optional: true,
				Computed: true,
			},
			"service_id": {
				Type:     types.StringType,
				Optional: true,
				Computed: true,
			},
			"supported": {
				Type:     types.BoolType,
				Computed: true,
			},
		},
	}

	return schema, nil
}

// Configure enables provider-level data or clients to be set in the
// provider-defined DataSource type. It is separately executed for each
// ReadDataSource RPC.
func (d *dataSourceService) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	if v, ok := request.ProviderData.(*conns.AWSClient); ok {
		d.meta = v
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
		v := data.ReverseDNSName.Value
		serviceParts := strings.Split(v, ".")
		n := len(serviceParts)

		if n < 4 {
			response.Diagnostics.AddError("reverse service DNS names must have at least 4 parts", fmt.Sprintf("%s has %d", v, n))

			return
		}

		data.Region = types.String{Value: serviceParts[n-2]}
		data.ReverseDNSPrefix = types.String{Value: strings.Join(serviceParts[0:n-2], ".")}
		data.ServiceID = types.String{Value: serviceParts[n-1]}
	}

	if !data.DNSName.IsNull() {
		v := data.DNSName.Value
		serviceParts := slices.Reversed(strings.Split(v, "."))
		n := len(serviceParts)

		if n < 4 {
			response.Diagnostics.AddError("service DNS names must have at least 4 parts", fmt.Sprintf("%s has %d", v, n))

			return
		}

		data.Region = types.String{Value: serviceParts[n-2]}
		data.ReverseDNSPrefix = types.String{Value: strings.Join(serviceParts[0:n-2], ".")}
		data.ServiceID = types.String{Value: serviceParts[n-1]}
	}

	if data.Region.IsNull() {
		data.Region = types.String{Value: d.meta.Region}
	}

	if data.ServiceID.IsNull() {
		response.Diagnostics.AddError("service ID not provided directly or through a DNS name", "")

		return
	}

	if data.ReverseDNSPrefix.IsNull() || data.ReverseDNSPrefix.IsUnknown() {
		dnsParts := strings.Split(d.meta.DNSSuffix, ".")
		data.ReverseDNSPrefix = types.String{Value: strings.Join(slices.Reversed(dnsParts), ".")}
	}

	reverseDNSName := fmt.Sprintf("%s.%s.%s", data.ReverseDNSPrefix.Value, data.Region.Value, data.ServiceID.Value)
	data.ReverseDNSName = types.String{Value: reverseDNSName}
	data.DNSName = types.String{Value: strings.ToLower(strings.Join(slices.Reversed(strings.Split(reverseDNSName, ".")), "."))}

	data.Supported = types.Bool{Value: true}
	if partition, ok := endpoints.PartitionForRegion(endpoints.DefaultPartitions(), data.Region.Value); ok {
		data.Partition = types.String{Value: partition.ID()}

		if _, ok := partition.Services()[data.ServiceID.Value]; !ok {
			data.Supported.Value = false
		}
	} else {
		data.Partition = types.String{Null: true}
	}

	data.ID = types.String{Value: reverseDNSName}

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
