package globalaccelerator

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/globalaccelerator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/fwtypes"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func init() {
	registerFrameworkDataSourceFactory(newDataSourceAccelerator)
}

// newDataSourceAccelerator instantiates a new DataSource for the aws_globalaccelerator_accelerator data source.
func newDataSourceAccelerator(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceAccelerator{}, nil
}

type dataSourceAccelerator struct {
	meta *conns.AWSClient
}

// Metadata should return the full name of the data source, such as
// examplecloud_thing.
func (d *dataSourceAccelerator) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "aws_globalaccelerator_accelerator"
}

// GetSchema returns the schema for this data source.
func (d *dataSourceAccelerator) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	schema := tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"arn": {
				Type:     fwtypes.ARNType,
				Optional: true,
				Computed: true,
			},
			"attributes": {
				Type: types.ListType{ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"flow_logs_enabled":   types.BoolType,
						"flow_logs_s3_bucket": types.StringType,
						"flow_logs_s3_prefix": types.StringType,
					},
				}},
				Computed: true,
			},
			"dns_name": {
				Type:     types.StringType,
				Computed: true,
			},
			"enabled": {
				Type:     types.BoolType,
				Computed: true,
			},
			"hosted_zone_id": {
				Type:     types.StringType,
				Computed: true,
			},
			"id": {
				Type:     types.StringType,
				Optional: true,
				Computed: true,
			},
			"ip_address_type": {
				Type:     types.StringType,
				Computed: true,
			},
			"ip_sets": {
				Type: types.ListType{ElemType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"ip_addresses": types.ListType{ElemType: types.StringType},
						"ip_family":    types.StringType,
					},
				}},
				Computed: true,
			},
			"name": {
				Type:     types.StringType,
				Optional: true,
				Computed: true,
			},
			"tags": tftags.TagsAttributeComputed(),
		},
	}

	return schema, nil
}

// Configure enables provider-level data or clients to be set in the
// provider-defined DataSource type. It is separately executed for each
// ReadDataSource RPC.
func (d *dataSourceAccelerator) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	if v, ok := request.ProviderData.(*conns.AWSClient); ok {
		d.meta = v
	}
}

// Read is called when the provider must read data source values in order to update state.
// Config values should be read from the ReadRequest and new state values set on the ReadResponse.
func (d *dataSourceAccelerator) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data dataSourceAcceleratorData

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := d.meta.GlobalAcceleratorConn
	ignoreTagsConfig := d.meta.IgnoreTagsConfig

	var results []*globalaccelerator.Accelerator
	err := conn.ListAcceleratorsPagesWithContext(ctx, &globalaccelerator.ListAcceleratorsInput{}, func(page *globalaccelerator.ListAcceleratorsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, accelerator := range page.Accelerators {
			if accelerator == nil {
				continue
			}

			if !data.ARN.IsNull() && data.ARN.Value.String() != aws.StringValue(accelerator.AcceleratorArn) {
				continue
			}

			if !data.Name.IsNull() && data.Name.Value != aws.StringValue(accelerator.Name) {
				continue
			}

			results = append(results, accelerator)
		}

		return !lastPage
	})

	if err != nil {
		response.Diagnostics.AddError("listing Global Accelerator Accelerators", err.Error())

		return
	}

	if n := len(results); n == 0 {
		response.Diagnostics.AddError("no matching Global Accelerator Accelerator found", "")

		return
	} else if n > 1 {
		response.Diagnostics.AddError("multiple Global Accelerator Accelerators matched; use additional constraints to reduce matches to a single Global Accelerator Accelerator", "")

		return
	}

	accelerator := results[0]
	acceleratorARN := aws.StringValue(accelerator.AcceleratorArn)
	if v, err := arn.Parse(acceleratorARN); err != nil {
		response.Diagnostics.AddError("parsing ARN", err.Error())
	} else {
		data.ARN = fwtypes.ARN{Value: v}
	}
	data.DnsName = types.String{Value: aws.StringValue(accelerator.DnsName)}
	data.Enabled = types.Bool{Value: aws.BoolValue(accelerator.Enabled)}
	data.HostedZoneID = types.String{Value: route53ZoneID}
	data.ID = types.String{Value: acceleratorARN}
	data.IpAddressType = types.String{Value: aws.StringValue(accelerator.IpAddressType)}
	data.IpSets = flattenIPSetsFramework(ctx, accelerator.IpSets)
	data.Name = types.String{Value: aws.StringValue(accelerator.Name)}

	attributes, err := FindAcceleratorAttributesByARN(ctx, conn, acceleratorARN)

	if err != nil {
		response.Diagnostics.AddError("reading Global Accelerator Accelerator attributes", err.Error())

		return
	}

	data.Attributes = flattenAcceleratorAttributesFramework(ctx, attributes)

	tags, err := ListTagsWithContext(ctx, conn, acceleratorARN)

	if err != nil {
		response.Diagnostics.AddError("listing tags for Global Accelerator Accelerator", err.Error())

		return
	}

	data.Tags = flex.FlattenFrameworkStringValueMap(ctx, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map())

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type dataSourceAcceleratorData struct {
	ARN           fwtypes.ARN  `tfsdk:"arn"`
	Attributes    types.List   `tfsdk:"attributes"`
	DnsName       types.String `tfsdk:"dns_name"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	HostedZoneID  types.String `tfsdk:"hosted_zone_id"`
	ID            types.String `tfsdk:"id"`
	IpAddressType types.String `tfsdk:"ip_address_type"`
	IpSets        types.List   `tfsdk:"ip_sets"`
	Name          types.String `tfsdk:"name"`
	Tags          types.Map    `tfsdk:"tags"`
}

func flattenIPSetFramework(ctx context.Context, apiObject *globalaccelerator.IpSet) types.Object {
	attrTypes := map[string]attr.Type{
		"ip_addresses": types.ListType{ElemType: types.StringType},
		"ip_family":    types.StringType,
	}

	if apiObject == nil {
		return types.Object{AttrTypes: attrTypes, Null: true}
	}

	attrs := map[string]attr.Value{}

	if v := apiObject.IpAddresses; v != nil {
		attrs["ip_addresses"] = flex.FlattenFrameworkStringList(ctx, v)
	} else {
		attrs["ip_addresses"] = types.List{Null: true}
	}

	if v := apiObject.IpFamily; v != nil {
		attrs["ip_family"] = types.String{Value: aws.StringValue(v)}
	} else {
		attrs["ip_family"] = types.String{Null: true}
	}

	return types.Object{AttrTypes: attrTypes, Attrs: attrs}
}

func flattenIPSetsFramework(ctx context.Context, apiObjects []*globalaccelerator.IpSet) types.List {
	elemType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"ip_addresses": types.ListType{ElemType: types.StringType},
		"ip_family":    types.StringType,
	}}

	if len(apiObjects) == 0 {
		return types.List{ElemType: elemType, Null: true}
	}

	var elems []attr.Value

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		elems = append(elems, flattenIPSetFramework(ctx, apiObject))
	}

	return types.List{ElemType: elemType, Elems: elems}
}

func flattenAcceleratorAttributesFramework(_ context.Context, apiObject *globalaccelerator.AcceleratorAttributes) types.List {
	attrTypes := map[string]attr.Type{
		"flow_logs_enabled":   types.BoolType,
		"flow_logs_s3_bucket": types.StringType,
		"flow_logs_s3_prefix": types.StringType,
	}
	elemType := types.ObjectType{
		AttrTypes: attrTypes,
	}

	if apiObject == nil {
		return types.List{ElemType: elemType, Null: true}
	}

	attrs := map[string]attr.Value{}

	if v := apiObject.FlowLogsEnabled; v != nil {
		attrs["flow_logs_enabled"] = types.Bool{Value: aws.BoolValue(v)}
	} else {
		attrs["flow_logs_enabled"] = types.Bool{Null: true}
	}

	if v := apiObject.FlowLogsS3Bucket; v != nil {
		attrs["flow_logs_s3_bucket"] = types.String{Value: aws.StringValue(v)}
	} else {
		attrs["flow_logs_s3_bucket"] = types.String{Value: ""}
	}

	if v := apiObject.FlowLogsS3Prefix; v != nil {
		attrs["flow_logs_s3_prefix"] = types.String{Value: aws.StringValue(v)}
	} else {
		attrs["flow_logs_s3_prefix"] = types.String{Value: ""}
	}

	return types.List{ElemType: elemType, Elems: []attr.Value{types.Object{AttrTypes: attrTypes, Attrs: attrs}}}
}
