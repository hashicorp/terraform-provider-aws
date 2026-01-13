// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3control

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3control"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3control/types"
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

// @FrameworkDataSource("aws_s3control_access_points", name="Access Points")
func newDataSourceAccessPoints(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceAccessPoints{}, nil
}

const (
	DSNameAccessPoints = "Access Points Data Source"
)

type dataSourceAccessPoints struct {
	framework.DataSourceWithModel[dataSourceAccessPointsModel]
}

func (d *dataSourceAccessPoints) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"access_points": framework.DataSourceComputedListOfObjectAttribute[accessPointListDataSourceModel](ctx),
			names.AttrAccountID: schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					fwvalidators.AWSAccountID(),
				},
			},
			names.AttrBucket: schema.StringAttribute{
				Optional: true,
			},
			"data_source_id": schema.StringAttribute{
				Optional: true,
			},
			"data_source_type": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (d *dataSourceAccessPoints) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().S3ControlClient(ctx)

	var data dataSourceAccessPointsModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	var input s3control.ListAccessPointsInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, data, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	if input.AccountId == nil {
		input.AccountId = aws.String(d.Meta().AccountID(ctx))
	}

	out, err := findAccessPoints(ctx, conn, &input)
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

func findAccessPoints(ctx context.Context, conn *s3control.Client, input *s3control.ListAccessPointsInput) ([]awstypes.AccessPoint, error) {
	paginator := s3control.NewListAccessPointsPaginator(conn, input)

	var output []awstypes.AccessPoint
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		output = append(output, page.AccessPointList...)
	}

	return output, nil
}

type dataSourceAccessPointsModel struct {
	framework.WithRegionModel
	AccessPoints   fwtypes.ListNestedObjectValueOf[accessPointListDataSourceModel] `tfsdk:"access_points"`
	AccountID      types.String                                                    `tfsdk:"account_id"`
	Bucket         types.String                                                    `tfsdk:"bucket"`
	DataSourceID   types.String                                                    `tfsdk:"data_source_id"`
	DataSourceType types.String                                                    `tfsdk:"data_source_type"`
}

// A variation on the access points data source model which contains only
// the attributes returned by the ListAccessPoints API response
type accessPointListDataSourceModel struct {
	AccessPointARN   types.String                                           `tfsdk:"access_point_arn"`
	Alias            types.String                                           `tfsdk:"alias"`
	Bucket           types.String                                           `tfsdk:"bucket"`
	BucketAccountID  types.String                                           `tfsdk:"bucket_account_id"`
	DataSourceID     types.String                                           `tfsdk:"data_source_id"`
	DataSourceType   types.String                                           `tfsdk:"data_source_type"`
	Name             types.String                                           `tfsdk:"name"`
	NetworkOrigin    types.String                                           `tfsdk:"network_origin"`
	VPCConfiguration fwtypes.ListNestedObjectValueOf[vpcConfigurationModel] `tfsdk:"vpc_configuration"`
}
