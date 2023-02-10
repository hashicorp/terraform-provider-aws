package globalaccelerator

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/globalaccelerator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func init() {
	_sp.registerFrameworkDataSourceFactory(newDataSourceAccelerator)
}

// newDataSourceAccelerator instantiates a new DataSource for the aws_globalaccelerator_accelerator data source.
func newDataSourceAccelerator(context.Context) (datasource.DataSourceWithConfigure, error) {
	d := &dataSourceAccelerator{}
	d.SetMigratedFromPluginSDK(true)

	return d, nil
}

type dataSourceAccelerator struct {
	framework.DataSourceWithConfigure
}

// Metadata should return the full name of the data source, such as
// examplecloud_thing.
func (d *dataSourceAccelerator) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "aws_globalaccelerator_accelerator"
}

// Schema returns the schema for this data source.
func (d *dataSourceAccelerator) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				Computed:   true,
			},
			"attributes": schema.ListAttribute{
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"flow_logs_enabled":   types.BoolType,
						"flow_logs_s3_bucket": types.StringType,
						"flow_logs_s3_prefix": types.StringType,
					},
				},
				Computed: true,
			},
			"dns_name": schema.StringAttribute{
				Computed: true,
			},
			"enabled": schema.BoolAttribute{
				Computed: true,
			},
			"hosted_zone_id": schema.StringAttribute{
				Computed: true,
			},
			"id": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"ip_address_type": schema.StringAttribute{
				Computed: true,
			},
			"ip_sets": schema.ListAttribute{
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"ip_addresses": types.ListType{ElemType: types.StringType},
						"ip_family":    types.StringType,
					},
				},
				Computed: true,
			},
			"name": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"tags": tftags.TagsAttributeComputedOnly(),
		},
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

	conn := d.Meta().GlobalAcceleratorConn()
	ignoreTagsConfig := d.Meta().IgnoreTagsConfig

	var results []*globalaccelerator.Accelerator
	err := conn.ListAcceleratorsPagesWithContext(ctx, &globalaccelerator.ListAcceleratorsInput{}, func(page *globalaccelerator.ListAcceleratorsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, accelerator := range page.Accelerators {
			if accelerator == nil {
				continue
			}

			if !data.ARN.IsNull() && data.ARN.ValueARN().String() != aws.StringValue(accelerator.AcceleratorArn) {
				continue
			}

			if !data.Name.IsNull() && data.Name.ValueString() != aws.StringValue(accelerator.Name) {
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
		data.ARN = fwtypes.ARNValue(v)
	}
	data.DnsName = flex.StringToFrameworkLegacy(ctx, accelerator.DnsName)
	data.Enabled = flex.BoolToFrameworkLegacy(ctx, accelerator.Enabled)
	data.HostedZoneID = types.StringValue(route53ZoneID)
	data.ID = types.StringValue(acceleratorARN)
	data.IpAddressType = flex.StringToFrameworkLegacy(ctx, accelerator.IpAddressType)
	data.IpSets = d.flattenIPSetsFramework(ctx, accelerator.IpSets)
	data.Name = flex.StringToFrameworkLegacy(ctx, accelerator.Name)

	attributes, err := FindAcceleratorAttributesByARN(ctx, conn, acceleratorARN)

	if err != nil {
		response.Diagnostics.AddError("reading Global Accelerator Accelerator attributes", err.Error())

		return
	}

	data.Attributes = d.flattenAcceleratorAttributesFramework(ctx, attributes)

	tags, err := ListTags(ctx, conn, acceleratorARN)

	if err != nil {
		response.Diagnostics.AddError("listing tags for Global Accelerator Accelerator", err.Error())

		return
	}

	data.Tags = flex.FlattenFrameworkStringValueMapLegacy(ctx, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map())

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (d *dataSourceAccelerator) flattenIPSetFramework(ctx context.Context, apiObject *globalaccelerator.IpSet) types.Object {
	attributeTypes := map[string]attr.Type{
		"ip_addresses": types.ListType{ElemType: types.StringType},
		"ip_family":    types.StringType,
	}

	if apiObject == nil {
		return types.ObjectNull(attributeTypes)
	}

	attributes := map[string]attr.Value{
		"ip_addresses": flex.FlattenFrameworkStringListLegacy(ctx, apiObject.IpAddresses),
		"ip_family":    flex.StringToFrameworkLegacy(ctx, apiObject.IpFamily),
	}

	return types.ObjectValueMust(attributeTypes, attributes)
}

func (d *dataSourceAccelerator) flattenIPSetsFramework(ctx context.Context, apiObjects []*globalaccelerator.IpSet) types.List {
	elementType := types.ObjectType{AttrTypes: map[string]attr.Type{
		"ip_addresses": types.ListType{ElemType: types.StringType},
		"ip_family":    types.StringType,
	}}
	var elements []attr.Value

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		elements = append(elements, d.flattenIPSetFramework(ctx, apiObject))
	}

	return types.ListValueMust(elementType, elements)
}

func (d *dataSourceAccelerator) flattenAcceleratorAttributesFramework(ctx context.Context, apiObject *globalaccelerator.AcceleratorAttributes) types.List {
	attributeTypes := map[string]attr.Type{
		"flow_logs_enabled":   types.BoolType,
		"flow_logs_s3_bucket": types.StringType,
		"flow_logs_s3_prefix": types.StringType,
	}
	elementType := types.ObjectType{
		AttrTypes: attributeTypes,
	}

	if apiObject == nil {
		return types.ListNull(elementType)
	}

	attributes := map[string]attr.Value{
		"flow_logs_enabled":   flex.BoolToFrameworkLegacy(ctx, apiObject.FlowLogsEnabled),
		"flow_logs_s3_bucket": flex.StringToFrameworkLegacy(ctx, apiObject.FlowLogsS3Bucket),
		"flow_logs_s3_prefix": flex.StringToFrameworkLegacy(ctx, apiObject.FlowLogsS3Prefix),
	}

	return types.ListValueMust(elementType, []attr.Value{types.ObjectValueMust(attributeTypes, attributes)})
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
