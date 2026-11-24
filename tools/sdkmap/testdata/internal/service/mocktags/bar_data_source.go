// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mocktags

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_mocktags_bar", name="Bar")
// @Tags(identifierAttribute="arn", resourceType="Bar")
func newBarDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &barDataSource{}, nil
}

type barDataSource struct {
	framework.DataSourceWithModel[barDataSourceModel]
}

func (d *barDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data barDataSourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().MockTagsClient(ctx)

	bar, err := findBarByID(ctx, conn, data.BarID.ValueString())
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, names.AttrID, data.BarID.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, bar, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

type barDataSourceModel struct {
	framework.WithRegionModel
	ARN   types.String `tfsdk:"arn"`
	BarID types.String `tfsdk:"bar_id"`
	Tags  tftags.Map   `tfsdk:"tags"`
}
