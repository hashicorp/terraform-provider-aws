// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
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

// @FrameworkDataSource(name="Producer Data Shares")
func newDataSourceProducerDataShares(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceProducerDataShares{}, nil
}

const (
	DSNameProducerDataShares = "Producer Data Shares Data Source"
)

type dataSourceProducerDataShares struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceProducerDataShares) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	resp.TypeName = "aws_redshift_producer_data_shares"
}

func (d *dataSourceProducerDataShares) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"producer_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.DataShareStatusForProducer](),
				Optional:   true,
			},
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
func (d *dataSourceProducerDataShares) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().RedshiftClient(ctx)

	var data dataSourceProducerDataSharesData
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ID = types.StringValue(data.ProducerARN.ValueString())

	input := &redshift.DescribeDataSharesForProducerInput{
		ProducerArn: aws.String(data.ProducerARN.ValueString()),
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

type dataSourceProducerDataSharesData struct {
	DataShares  fwtypes.ListNestedObjectValueOf[dataSharesData]         `tfsdk:"data_shares"`
	ID          types.String                                            `tfsdk:"id"`
	Status      fwtypes.StringEnum[awstypes.DataShareStatusForProducer] `tfsdk:"status"`
	ProducerARN fwtypes.ARN                                             `tfsdk:"producer_arn"`
}
