// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package globalaccelerator

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/globalaccelerator"
	awstypes "github.com/aws/aws-sdk-go-v2/service/globalaccelerator/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_globalaccelerator_accelerator", name="Accelerator")
func newAcceleratorDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	d := &acceleratorDataSource{}

	return d, nil
}

type acceleratorDataSource struct {
	framework.DataSourceWithConfigure
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
				CustomType: fwtypes.NewListNestedObjectTypeOf[acceleratorAttributesModel](ctx),
				Computed:   true,
				ElementType: types.ObjectType{
					AttrTypes: fwtypes.AttributeTypesMust[acceleratorAttributesModel](ctx),
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
				CustomType: fwtypes.NewListNestedObjectTypeOf[ipSetModel](ctx),
				Computed:   true,
				ElementType: types.ObjectType{
					AttrTypes: fwtypes.AttributeTypesMust[ipSetModel](ctx),
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
	ignoreTagsConfig := d.Meta().IgnoreTagsConfig(ctx)

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
	response.Diagnostics.Append(flex.Flatten(ctx, accelerator, &data)...)
	if response.Diagnostics.HasError() {
		return
	}
	data.ID = types.StringValue(acceleratorARN)
	data.HostedZoneID = types.StringValue(d.Meta().GlobalAcceleratorHostedZoneID(ctx))

	attributes, err := findAcceleratorAttributesByARN(ctx, conn, acceleratorARN)
	if err != nil {
		response.Diagnostics.AddError("reading Global Accelerator Accelerator attributes", err.Error())

		return
	}
	response.Diagnostics.Append(flex.Flatten(ctx, attributes, &data.Attributes)...)
	if response.Diagnostics.HasError() {
		return
	}

	tags, err := listTags(ctx, conn, acceleratorARN)

	if err != nil {
		response.Diagnostics.AddError("listing tags for Global Accelerator Accelerator", err.Error())

		return
	}

	data.Tags = tftags.FlattenStringValueMap(ctx, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map())

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type acceleratorDataSourceModel struct {
	ARN              fwtypes.ARN                                                 `tfsdk:"arn"`
	Attributes       fwtypes.ListNestedObjectValueOf[acceleratorAttributesModel] `tfsdk:"attributes"`
	DnsName          types.String                                                `tfsdk:"dns_name" autoflex:",legacy"`
	DualStackDNSName types.String                                                `tfsdk:"dual_stack_dns_name" autoflex:",legacy"`
	Enabled          types.Bool                                                  `tfsdk:"enabled"`
	HostedZoneID     types.String                                                `tfsdk:"hosted_zone_id"`
	ID               types.String                                                `tfsdk:"id"`
	IpAddressType    types.String                                                `tfsdk:"ip_address_type"`
	IpSets           fwtypes.ListNestedObjectValueOf[ipSetModel]                 `tfsdk:"ip_sets"`
	Name             types.String                                                `tfsdk:"name"`
	Tags             tftags.Map                                                  `tfsdk:"tags"`
}

type acceleratorAttributesModel struct {
	FlowLogsEnabled  types.Bool   `tfsdk:"flow_logs_enabled" autoflex:",legacy"`
	FlowLogsS3Bucket types.String `tfsdk:"flow_logs_s3_bucket" autoflex:",legacy"`
	FlowLogsS3Prefix types.String `tfsdk:"flow_logs_s3_prefix" autoflex:",legacy"`
}

type ipSetModel struct {
	IPAddresses fwtypes.ListValueOf[types.String] `tfsdk:"ip_addresses" autoflex:",legacy"`
	IPFamily    types.String                      `tfsdk:"ip_family" autoflex:",legacy"`
}
