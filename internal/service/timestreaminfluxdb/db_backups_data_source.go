// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package timestreaminfluxdb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/timestreaminfluxdb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/timestreaminfluxdb/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
)

// @FrameworkDataSource("aws_timestreaminfluxdb_db_backups", name="DB Backups")
func newDBBackupsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dbBackupsDataSource{}, nil
}

type dbBackupsDataSource struct {
	framework.DataSourceWithModel[dbBackupsDataSourceModel]
}

func (d *dbBackupsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"backups": framework.DataSourceComputedListOfObjectAttribute[dbBackupSummaryModel](ctx),
			"db_resource_id": schema.StringAttribute{
				Optional:    true,
				Description: `The id of the DB instance or DB cluster to list backups for. If omitted, all backups in the account and Region are returned.`,
			},
		},
	}
}

func (d *dbBackupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dbBackupsDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().TimestreamInfluxDBClient(ctx)

	var input timestreaminfluxdb.ListDbBackupsInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, data, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	output, err := findDBBackups(ctx, conn, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, output, &data.Backups, fwflex.WithNoIgnoredFieldNames()))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

func findDBBackups(ctx context.Context, conn *timestreaminfluxdb.Client, input *timestreaminfluxdb.ListDbBackupsInput) ([]awstypes.DbBackupSummary, error) {
	var output []awstypes.DbBackupSummary

	pages := timestreaminfluxdb.NewListDbBackupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		output = append(output, page.Items...)
	}

	return output, nil
}

type dbBackupsDataSourceModel struct {
	framework.WithRegionModel
	Backups      fwtypes.ListNestedObjectValueOf[dbBackupSummaryModel] `tfsdk:"backups"`
	DBResourceID types.String                                          `tfsdk:"db_resource_id"`
}

type dbBackupSummaryModel struct {
	ARN            types.String      `tfsdk:"arn"`
	CreatedAt      timetypes.RFC3339 `tfsdk:"created_at"`
	DBResourceID   types.String      `tfsdk:"db_resource_id"`
	DeploymentType types.String      `tfsdk:"deployment_type"`
	EngineType     types.String      `tfsdk:"engine_type"`
	ExpiresAfter   types.String      `tfsdk:"expires_after"`
	ID             types.String      `tfsdk:"id"`
	KMSKeyID       types.String      `tfsdk:"kms_key_id"`
	Name           types.String      `tfsdk:"name"`
	Status         types.String      `tfsdk:"status"`
	Type           types.String      `tfsdk:"type"`
}
