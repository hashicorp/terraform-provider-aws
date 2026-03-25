// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package s3control

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3control"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3control/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_s3control_multi_region_access_points", name="Multi Region Access Points")
func newMultiRegionAccessPointsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &multiRegionAccessPointsDataSource{}, nil
}

type multiRegionAccessPointsDataSource struct {
	framework.DataSourceWithModel[multiRegionAccessPointsDataSourceModel]
}

func (d *multiRegionAccessPointsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"access_points": framework.DataSourceComputedListOfObjectAttribute[multiRegionAccessPointListDataSourceModel](ctx),
			names.AttrAccountID: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					fwvalidators.AWSAccountID(),
				},
			},
		},
	}
}

func (d *multiRegionAccessPointsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().S3ControlClient(ctx)

	var data multiRegionAccessPointsDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	var input s3control.ListMultiRegionAccessPointsInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, data, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	if input.AccountId == nil {
		input.AccountId = aws.String(d.Meta().AccountID(ctx))
	}

	out, err := findMultiRegionAccessPoints(ctx, conn, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &data.AccessPoints))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

func findMultiRegionAccessPoints(ctx context.Context, conn *s3control.Client, input *s3control.ListMultiRegionAccessPointsInput) ([]awstypes.MultiRegionAccessPointReport, error) {
	paginator := s3control.NewListMultiRegionAccessPointsPaginator(conn, input)

	var output []awstypes.MultiRegionAccessPointReport
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		output = append(output, page.AccessPoints...)
	}

	return output, nil
}

type multiRegionAccessPointsDataSourceModel struct {
	framework.WithRegionModel
	AccessPoints fwtypes.ListNestedObjectValueOf[multiRegionAccessPointListDataSourceModel] `tfsdk:"access_points"`
	AccountID    types.String                                                               `tfsdk:"account_id"`
}

type multiRegionAccessPointListDataSourceModel struct {
	Alias             types.String                                                         `tfsdk:"alias"`
	CreatedAt         timetypes.RFC3339                                                    `tfsdk:"created_at"`
	Name              types.String                                                         `tfsdk:"name"`
	PublicAccessBlock fwtypes.ListNestedObjectValueOf[publicAccessBlockConfigurationModel] `tfsdk:"public_access_block"`
	Regions           fwtypes.ListNestedObjectValueOf[regionReportModel]                   `tfsdk:"regions"`
	Status            types.String                                                         `tfsdk:"status"`
}

type regionReportModel struct {
	Bucket          types.String `tfsdk:"bucket"`
	BucketAccountID types.String `tfsdk:"bucket_account_id"`
	Region          types.String `tfsdk:"region"`
}
