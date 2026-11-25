// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mockframeworkdatasource

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mockframeworkdatasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_mockframeworkdatasource_baz", name="Baz")
func newBazDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	var d bazDataSource

	return &d, nil
}

type bazDataSource struct {
	bazDataSourceBase
}

// bazDataSourceBase implements Read, which is promoted to bazDataSource.
type bazDataSourceBase struct {
	framework.DataSourceWithModel[bazDataSourceModel]
}

func (d *bazDataSourceBase) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data bazDataSourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().MockFrameworkDataSourceClient(ctx)

	input := mockframeworkdatasource.GetBazInput{
		BazId: aws.String(data.ID.ValueString()),
	}

	output, err := conn.GetBaz(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, names.AttrID, data.ID.ValueString())
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, fwflex.Flatten(ctx, output.Baz, &data))
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data))
}

type bazDataSourceModel struct {
	framework.WithRegionModel
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}
