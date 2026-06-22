// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_dynamodb_backups", name="Backups")
func newBackupsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &backupsDataSource{}, nil
}

type backupsDataSource struct {
	framework.DataSourceWithModel[backupsDataSourceModel]
}

func (d *backupsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"backup_summaries": framework.DataSourceComputedListOfObjectAttribute[backupSummaryModel](ctx),
			"backup_type": schema.StringAttribute{
				Optional:   true,
				CustomType: fwtypes.StringEnumType[awstypes.BackupTypeFilter](),
			},
			names.AttrTableName: schema.StringAttribute{
				Optional: true,
			},
			"time_range_lower_bound": schema.StringAttribute{
				Optional:   true,
				CustomType: timetypes.RFC3339Type{},
			},
			"time_range_upper_bound": schema.StringAttribute{
				Optional:   true,
				CustomType: timetypes.RFC3339Type{},
			},
		},
	}
}

func (d *backupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().DynamoDBClient(ctx)

	var data backupsDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	var input dynamodb.ListBackupsInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, data, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findBackups(ctx, conn, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &data.BackupSummaries))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

func findBackups(ctx context.Context, conn *dynamodb.Client, input *dynamodb.ListBackupsInput) ([]awstypes.BackupSummary, error) {
	var output []awstypes.BackupSummary

	err := listBackupsPages(ctx, conn, input, func(page *dynamodb.ListBackupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		output = append(output, page.BackupSummaries...)
		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

type backupsDataSourceModel struct {
	framework.WithRegionModel
	BackupSummaries     fwtypes.ListNestedObjectValueOf[backupSummaryModel] `tfsdk:"backup_summaries"`
	BackupType          fwtypes.StringEnum[awstypes.BackupTypeFilter]       `tfsdk:"backup_type"`
	TableName           types.String                                        `tfsdk:"table_name"`
	TimeRangeLowerBound timetypes.RFC3339                                   `tfsdk:"time_range_lower_bound"`
	TimeRangeUpperBound timetypes.RFC3339                                   `tfsdk:"time_range_upper_bound"`
}

type backupSummaryModel struct {
	BackupARN              types.String                              `tfsdk:"backup_arn"`
	BackupCreationDateTime timetypes.RFC3339                         `tfsdk:"backup_creation_date_time"`
	BackupExpiryDateTime   timetypes.RFC3339                         `tfsdk:"backup_expiry_date_time"`
	BackupName             types.String                              `tfsdk:"backup_name"`
	BackupSizeBytes        types.Int64                               `tfsdk:"backup_size_bytes"`
	BackupStatus           fwtypes.StringEnum[awstypes.BackupStatus] `tfsdk:"backup_status"`
	BackupType             fwtypes.StringEnum[awstypes.BackupType]   `tfsdk:"backup_type"`
	TableARN               types.String                              `tfsdk:"table_arn"`
	TableID                types.String                              `tfsdk:"table_id"`
	TableName              types.String                              `tfsdk:"table_name"`
}
