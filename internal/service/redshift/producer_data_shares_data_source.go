// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/redshift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/redshift/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_redshift_producer_data_shares", name="Producer Data Shares")
func newProducerDataSharesDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &producerDataSharesDataSource{}, nil
}

const (
	DSNameProducerDataShares = "Producer Data Shares Data Source"
)

type producerDataSharesDataSource struct {
	framework.DataSourceWithModel[producerDataSharesDataSourceModel]
}

func (d *producerDataSharesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"data_shares": framework.DataSourceComputedListOfObjectAttribute[dataSharesData](ctx),
			names.AttrID:  framework.IDAttribute(),
			"producer_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.DataShareStatusForProducer](),
				Optional:   true,
			},
		},
	}
}
func (d *producerDataSharesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().RedshiftClient(ctx)

	var data producerDataSharesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ID = types.StringValue(data.ProducerARN.ValueString())

	input := &redshift.DescribeDataSharesForProducerInput{
		ProducerArn: data.ProducerARN.ValueStringPointer(),
	}
	if !data.Status.IsNull() {
		input.Status = awstypes.DataShareStatusForProducer(data.Status.ValueString())
	}

	paginator := redshift.NewDescribeDataSharesForProducerPaginator(conn, input)

	var out redshift.DescribeDataSharesForProducerOutput
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			resp.Diagnostics.AddError(
				create.ProblemStandardMessage(names.Redshift, create.ErrActionReading, DSNameProducerDataShares, data.ID.String(), err),
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

type producerDataSharesDataSourceModel struct {
	framework.WithRegionModel
	DataShares  fwtypes.ListNestedObjectValueOf[dataSharesData]         `tfsdk:"data_shares"`
	ID          types.String                                            `tfsdk:"id"`
	Status      fwtypes.StringEnum[awstypes.DataShareStatusForProducer] `tfsdk:"status"`
	ProducerARN fwtypes.ARN                                             `tfsdk:"producer_arn"`
}
