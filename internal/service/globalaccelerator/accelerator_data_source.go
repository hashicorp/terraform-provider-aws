// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package globalaccelerator

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/globalaccelerator"
	awstypes "github.com/aws/aws-sdk-go-v2/service/globalaccelerator/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource(name="Accelerator")
func newAcceleratorDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	d := &acceleratorDataSource{}

	return d, nil
}

type acceleratorDataSource struct {
	framework.DataSourceWithConfigure
}

func (*acceleratorDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "aws_globalaccelerator_accelerator"
}

func (d *acceleratorDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
				Computed:   true,
			},
			names.AttrAttributes: schema.ListAttribute{
				Computed: true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"flow_logs_enabled":   types.BoolType,
						"flow_logs_s3_bucket": types.StringType,
						"flow_logs_s3_prefix": types.StringType,
					},
				},
			},
			names.AttrDNSName: schema.StringAttribute{
				Computed: true,
			},
			"dual_stack_dns_name": schema.StringAttribute{
				Computed: true,
			},
			names.AttrEnabled: schema.BoolAttribute{
				Computed: true,
			},
			names.AttrHostedZoneID: schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			names.AttrIPAddressType: schema.StringAttribute{
				Computed: true,
			},
			"ip_sets": schema.ListAttribute{
				Computed: true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						names.AttrIPAddresses: types.ListType{ElemType: types.StringType},
						"ip_family":           types.StringType,
					},
				},
			},
			names.AttrName: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (d *acceleratorDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data acceleratorDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().GlobalAcceleratorClient(ctx)
	ignoreTagsConfig := d.Meta().IgnoreTagsConfig

	var results []awstypes.Accelerator
	pages := globalaccelerator.NewListAcceleratorsPaginator(conn, &globalaccelerator.ListAcceleratorsInput{})
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			response.Diagnostics.AddError("listing Global Accelerator Accelerators", err.Error())

			return
		}

		for _, v := range page.Accelerators {
			if !data.ARN.IsNull() && data.ARN.ValueString() != aws.ToString(v.AcceleratorArn) {
				continue
			}

			if !data.Name.IsNull() && data.Name.ValueString() != aws.ToString(v.Name) {
				continue
			}

			results = append(results, v)
		}
	}

	if n := len(results); n == 0 {
		response.Diagnostics.AddError("no matching Global Accelerator Accelerator found", "")

		return
	} else if n > 1 {
		response.Diagnostics.AddError("multiple Global Accelerator Accelerators matched; use additional constraints to reduce matches to a single Global Accelerator Accelerator", "")

		return
	}

	accelerator := results[0]
	acceleratorARN := aws.ToString(accelerator.AcceleratorArn)
	data.ARN = flex.StringToFrameworkARN(ctx, accelerator.AcceleratorArn)
	data.DnsName = flex.StringToFrameworkLegacy(ctx, accelerator.DnsName)
	data.DualStackDNSName = flex.StringToFrameworkLegacy(ctx, accelerator.DualStackDnsName)
	data.Enabled = flex.BoolToFrameworkLegacy(ctx, accelerator.Enabled)
	data.HostedZoneID = types.StringValue(d.Meta().GlobalAcceleratorHostedZoneID(ctx))
	data.ID = types.StringValue(acceleratorARN)
	data.IpAddressType = flex.StringValueToFrameworkLegacy(ctx, accelerator.IpAddressType)
	data.IpSets = d.flattenIPSetsFramework(ctx, accelerator.IpSets)
	data.Name = flex.StringToFrameworkLegacy(ctx, accelerator.Name)

	attributes, err := findAcceleratorAttributesByARN(ctx, conn, acceleratorARN)

	if err != nil {
		response.Diagnostics.AddError("reading Global Accelerator Accelerator attributes", err.Error())

		return
	}

	data.Attributes = d.flattenAcceleratorAttributesFramework(ctx, attributes)

	tags, err := listTags(ctx, conn, acceleratorARN)

	if err != nil {
		response.Diagnostics.AddError("listing tags for Global Accelerator Accelerator", err.Error())

		return
	}

	data.Tags = flex.FlattenFrameworkStringValueMapLegacy(ctx, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map())

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (d *acceleratorDataSource) flattenIPSetFramework(ctx context.Context, apiObject *awstypes.IpSet) types.Object {
	attributeTypes := map[string]attr.Type{
		names.AttrIPAddresses: types.ListType{ElemType: types.StringType},
		"ip_family":           types.StringType,
	}

	if apiObject == nil {
		return types.ObjectNull(attributeTypes)
	}

	attributes := map[string]attr.Value{
		names.AttrIPAddresses: flex.FlattenFrameworkStringValueListLegacy(ctx, apiObject.IpAddresses),
		"ip_family":           flex.StringToFrameworkLegacy(ctx, apiObject.IpFamily),
	}

	return types.ObjectValueMust(attributeTypes, attributes)
}

func (d *acceleratorDataSource) flattenIPSetsFramework(ctx context.Context, apiObjects []awstypes.IpSet) types.List {
	elementType := types.ObjectType{AttrTypes: map[string]attr.Type{
		names.AttrIPAddresses: types.ListType{ElemType: types.StringType},
		"ip_family":           types.StringType,
	}}
	var elements []attr.Value

	for _, apiObject := range apiObjects {
		elements = append(elements, d.flattenIPSetFramework(ctx, &apiObject))
	}

	return types.ListValueMust(elementType, elements)
}

func (d *acceleratorDataSource) flattenAcceleratorAttributesFramework(ctx context.Context, apiObject *awstypes.AcceleratorAttributes) types.List {
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

type acceleratorDataSourceModel struct {
	ARN              fwtypes.ARN  `tfsdk:"arn"`
	Attributes       types.List   `tfsdk:"attributes"`
	DnsName          types.String `tfsdk:"dns_name"`
	DualStackDNSName types.String `tfsdk:"dual_stack_dns_name"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	HostedZoneID     types.String `tfsdk:"hosted_zone_id"`
	ID               types.String `tfsdk:"id"`
	IpAddressType    types.String `tfsdk:"ip_address_type"`
	IpSets           types.List   `tfsdk:"ip_sets"`
	Name             types.String `tfsdk:"name"`
	Tags             types.Map    `tfsdk:"tags"`
}
