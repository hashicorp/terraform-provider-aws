// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ce

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

type costAllocationTagsDataSource struct {
	framework.DataSourceWithModel[costAllocationTagsDataSourceModel]
}

// @FrameworkDataSource("aws_ce_cost_allocation_tags", name="Cost Allocation Tags")
func newCostAllocationTagsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &costAllocationTagsDataSource{}, nil
}

func (d *costAllocationTagsDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrStatus: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			names.AttrType: schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"tag_keys": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Optional:    true,
			},
			names.AttrTags: framework.DataSourceComputedListOfObjectAttribute[costAllocationTagModel](ctx),
		},
	}
}

func (d *costAllocationTagsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data costAllocationTagsDataSourceModel

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().CEClient(ctx)

	input := &costexplorer.ListCostAllocationTagsInput{}
	response.Diagnostics.Append(flex.Expand(ctx, data, input)...)

	output, err := conn.ListCostAllocationTags(ctx, input)
	if err != nil {
		response.Diagnostics.AddError("listing Cost Allocation Tags", err.Error())
		return
	}

	response.Diagnostics.Append(flex.Flatten(ctx, output.CostAllocationTags, &data.Tags)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type costAllocationTagsDataSourceModel struct {
	Status  types.String                                            `tfsdk:"status"`
	Type    types.String                                            `tfsdk:"type"`
	TagKeys fwtypes.ListOfString                                    `tfsdk:"tag_keys"`
	Tags    fwtypes.ListNestedObjectValueOf[costAllocationTagModel] `tfsdk:"tags"`
}

type costAllocationTagModel struct {
	TagKey          types.String      `tfsdk:"tag_key"`
	Status          types.String      `tfsdk:"status"`
	Type            types.String      `tfsdk:"type"`
	LastUpdatedDate timetypes.RFC3339 `tfsdk:"last_updated_date"`
	LastUsedDate    timetypes.RFC3339 `tfsdk:"last_used_date"`
}
