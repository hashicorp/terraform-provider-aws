// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package timestreaminfluxdb

import (
	"context"
	"fmt"
	"iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/timestreaminfluxdb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/timestreaminfluxdb/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/logging"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkListResource("aws_timestreaminfluxdb_db_backup")
func newDBBackupResourceAsListResource() list.ListResourceWithConfigure {
	return &dbBackupListResource{}
}

var _ list.ListResource = &dbBackupListResource{}

type dbBackupListResource struct {
	dbBackupResource
	framework.WithList
}

func (l *dbBackupListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"db_resource_id": listschema.StringAttribute{
				Optional:    true,
				Description: "ID of the DB instance or DB cluster to list backups for.",
			},
		},
	}
}

func (l *dbBackupListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	conn := l.Meta().TimestreamInfluxDBClient(ctx)

	var query listDBBackupModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = smerr.ListStreamEnrich(ctx, diags)
			return
		}
	}

	var input timestreaminfluxdb.ListDbBackupsInput
	if diags := fwflex.Expand(ctx, query, &input); diags.HasError() {
		stream.Results = smerr.ListStreamEnrich(ctx, diags)
		return
	}

	tflog.Info(ctx, "Listing Timestream for InfluxDB DB Backups", map[string]any{
		logging.ResourceAttributeKey("db_resource_id"): aws.ToString(input.DbResourceId),
	})

	stream.Results = func(yield func(list.ListResult) bool) {
		for item, err := range listDBBackups(ctx, conn, &input) {
			if err != nil {
				yield(smerr.NewListResultError(ctx, err))
				return
			}

			// Only on-demand backups are managed by aws_timestreaminfluxdb_db_backup.
			if item.Type != awstypes.DbBackupTypeOnDemand || item.Status == awstypes.DbBackupStatusDeleting {
				continue
			}

			id := aws.ToString(item.Id)
			ctx := tflog.SetField(ctx, logging.ResourceAttributeKey(names.AttrID), id)

			// The list API returns a summary, so fetch the full backup for the resource data.
			var output *timestreaminfluxdb.GetDbBackupOutput
			if request.IncludeResource {
				var err error
				output, err = findDBBackupByID(ctx, conn, id)
				if retry.NotFound(err) {
					continue
				}
				if err != nil {
					yield(smerr.NewListResultError(ctx, err, smerr.ID, id))
					return
				}
			}

			result := request.NewListResult(ctx)

			var data dbBackupResourceModel
			l.SetResult(ctx, l.Meta(), request.IncludeResource, &data, &result, func() {
				data.ID = fwflex.StringValueToFramework(ctx, id)

				if request.IncludeResource {
					smerr.AddEnrich(ctx, &result.Diagnostics, l.flatten(ctx, output, &data), smerr.ID, id)
					if result.Diagnostics.HasError() {
						return
					}
				}

				result.DisplayName = aws.ToString(item.Name)
			})

			if result.Diagnostics.HasError() {
				yield(list.ListResult{Diagnostics: result.Diagnostics})
				return
			}

			if !yield(result) {
				return
			}
		}
	}
}

type listDBBackupModel struct {
	framework.WithRegionModel
	DBResourceID types.String `tfsdk:"db_resource_id"`
}

func listDBBackups(ctx context.Context, conn *timestreaminfluxdb.Client, input *timestreaminfluxdb.ListDbBackupsInput) iter.Seq2[awstypes.DbBackupSummary, error] {
	return func(yield func(awstypes.DbBackupSummary, error) bool) {
		pages := timestreaminfluxdb.NewListDbBackupsPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				yield(awstypes.DbBackupSummary{}, fmt.Errorf("listing Timestream for InfluxDB DB Backups: %w", err))
				return
			}

			for _, item := range page.Items {
				if !yield(item, nil) {
					return
				}
			}
		}
	}
}

func (r *dbBackupResource) flatten(ctx context.Context, output *timestreaminfluxdb.GetDbBackupOutput, data *dbBackupResourceModel) diag.Diagnostics {
	return fwflex.Flatten(ctx, output, data)
}
