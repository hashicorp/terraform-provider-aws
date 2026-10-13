// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package elasticache

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
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

// @FrameworkDataSource("aws_elasticache_service_updates", name="Service Updates")
func newServiceUpdatesDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &serviceUpdatesDataSource{}, nil
}

const (
	DSNameServiceUpdates = "Service Updates Data Source"
)

type serviceUpdatesDataSource struct {
	framework.DataSourceWithModel[serviceUpdatesDataSourceModel]
}

func (d *serviceUpdatesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrStatus: schema.SetAttribute{
				Optional:    true,
				CustomType:  fwtypes.SetOfStringEnumType[awstypes.ServiceUpdateStatus](),
				ElementType: fwtypes.StringEnumType[awstypes.ServiceUpdateStatus](),
			},
			"service_updates": framework.DataSourceComputedListOfObjectAttribute[serviceUpdateModel](ctx),
		},
	}
}

func (d *serviceUpdatesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().ElastiCacheClient(ctx)

	var data serviceUpdatesDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	var input elasticache.DescribeServiceUpdatesInput
	resp.Diagnostics.Append(flex.Expand(ctx, data, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serviceUpdates, err := findServiceUpdates(ctx, conn, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, serviceUpdates, &data.ServiceUpdates))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

type serviceUpdatesDataSourceModel struct {
	framework.WithRegionModel
	ServiceUpdates      fwtypes.ListNestedObjectValueOf[serviceUpdateModel]   `tfsdk:"service_updates"`
	ServiceUpdateStatus fwtypes.SetOfStringEnum[awstypes.ServiceUpdateStatus] `tfsdk:"status"`
}

type serviceUpdateModel struct {
	AutoUpdateAfterRecommendedApplyByDate types.Bool        `tfsdk:"auto_update_after_recommended_apply_by_date"`
	Engine                                types.String      `tfsdk:"engine"`
	EngineVersion                         types.String      `tfsdk:"engine_version"`
	EstimatedUpdateTime                   types.String      `tfsdk:"estimated_update_time"`
	ServiceUpdateDescription              types.String      `tfsdk:"description"`
	ServiceUpdateEndDate                  timetypes.RFC3339 `tfsdk:"end_date"`
	ServiceUpdateName                     types.String      `tfsdk:"name"`
	ServiceUpdateRecommendedApplyByDate   timetypes.RFC3339 `tfsdk:"recommended_apply_by_date"`
	ServiceUpdateReleaseDate              timetypes.RFC3339 `tfsdk:"release_date"`
	ServiceUpdateSeverity                 types.String      `tfsdk:"severity"`
	ServiceUpdateStatus                   types.String      `tfsdk:"status"`
	ServiceUpdateType                     types.String      `tfsdk:"type"`
}

func findServiceUpdates(ctx context.Context, conn *elasticache.Client, input *elasticache.DescribeServiceUpdatesInput) ([]awstypes.ServiceUpdate, error) {
	// Ensure that if no results are returned, an empty slice is returned instead of nil
	output := make([]awstypes.ServiceUpdate, 0)

	pages := elasticache.NewDescribeServiceUpdatesPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		output = append(output, page.ServiceUpdates...)
	}

	return output, nil
}
