// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_elasticache_reserved_cache_node_offering", name="Reserved Cache Node Offering")
func newDataSourceReservedCacheNodeOffering(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceReservedCacheNodeOffering{}, nil
}

type dataSourceReservedCacheNodeOffering struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceReservedCacheNodeOffering) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cache_node_type": schema.StringAttribute{
				Required: true,
			},
			names.AttrDuration: schema.StringAttribute{
				CustomType: fwtypes.RFC3339DurationType,
				Required:   true,
				Validators: []validator.String{
					stringvalidator.OneOf("P1Y", "P3Y"),
				},
			},
			"fixed_price": schema.Float64Attribute{
				Computed: true,
			},
			"offering_id": schema.StringAttribute{
				Computed: true,
			},
			"offering_type": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"Light Utilization",
						"Medium Utilization",
						"Heavy Utilization",
						"Partial Upfront",
						"All Upfront",
						"No Upfront",
					),
				},
			},
			"product_description": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf(engine_Values()...),
				},
			},
		},
	}
}

// Read is called when the provider must read data source values in order to update state.
// Config values should be read from the ReadRequest and new state values set on the ReadResponse.
func (d *dataSourceReservedCacheNodeOffering) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data dataSourceReservedCacheNodeOfferingModel

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().ElastiCacheClient(ctx)

	flexOpt := flex.WithFieldNamePrefix("ReservedCacheNodes")

	var input elasticache.DescribeReservedCacheNodesOfferingsInput
	response.Diagnostics.Append(flex.Expand(ctx, data, &input, flexOpt)...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, err := conn.DescribeReservedCacheNodesOfferings(ctx, &input)
	if err != nil {
		response.Diagnostics.AddError("reading ElastiCache Reserved Cache Node Offering", err.Error())
		return
	}

	offering, err := tfresource.AssertSingleValueResult(resp.ReservedCacheNodesOfferings)
	if err != nil {
		response.Diagnostics.AddError("reading ElastiCache Reserved Cache Node Offering", err.Error())
		return
	}

	response.Diagnostics.Append(flex.Flatten(ctx, offering, &data, flexOpt)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type dataSourceReservedCacheNodeOfferingModel struct {
	CacheNodeType      types.String            `tfsdk:"cache_node_type"`
	Duration           fwtypes.RFC3339Duration `tfsdk:"duration" autoflex:",noflatten"`
	FixedPrice         types.Float64           `tfsdk:"fixed_price"`
	OfferingID         types.String            `tfsdk:"offering_id"`
	OfferingType       types.String            `tfsdk:"offering_type"`
	ProductDescription types.String            `tfsdk:"product_description"`
}
