// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mockframeworkdatasource

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
)

// @FrameworkDataSource("aws_mockframeworkdatasource_bar", name="Bar")
func newBarDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &barDataSource{}, nil
}

type barDataSource struct {
	framework.DataSourceWithModel[barDataSourceModel]
}

func (d barDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data barDataSourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	partition := d.Meta().Partition(ctx)
	data.ID = types.StringValue(partition)
	data.Partition = types.StringValue(partition)

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

type barDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Partition types.String `tfsdk:"partition"`
}
