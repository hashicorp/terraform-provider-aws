// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/redshift"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource(name="Data Shares")
func newDataSourceDataShares(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceDataShares{}, nil
}

const (
	DSNameDataShares = "Data Shares Data Source"
)

type dataSourceDataShares struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceDataShares) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	resp.TypeName = "aws_redshift_data_shares"
}

func (d *dataSourceDataShares) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"data_shares": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[dataSharesData](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"data_share_arn": schema.StringAttribute{
							Computed: true,
						},
						"managed_by": schema.StringAttribute{
							Computed: true,
						},
						"producer_arn": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}
func (d *dataSourceDataShares) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().RedshiftClient(ctx)

	var data dataSourceDataSharesData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ID = types.StringValue(d.Meta().Region)

	paginator := redshift.NewDescribeDataSharesPaginator(conn, &redshift.DescribeDataSharesInput{})

	var out redshift.DescribeDataSharesOutput
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Redshift, create.ErrActionReading, DSNameDataShares, data.ID.String(), err),
				err.Error(),
			)
			return
		}

		if page != nil && len(page.DataShares) > 0 {
			out.DataShares = append(out.DataShares, page.DataShares...)
		}
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, out, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

type dataSourceDataSharesData struct {
	DataShares fwtypes.ListNestedObjectValueOf[dataSharesData] `tfsdk:"data_shares"`
	ID         types.String                                    `tfsdk:"id"`
}

type dataSharesData struct {
	DataShareARN types.String `tfsdk:"data_share_arn"`
	ManagedBy    types.String `tfsdk:"managed_by"`
	ProducerARN  types.String `tfsdk:"producer_arn"`
}
