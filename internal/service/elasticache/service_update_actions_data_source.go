// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package elasticache

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
)

// @FrameworkDataSource("aws_elasticache_service_update_actions", name="Service Update Actions")
func newServiceUpdateActionsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &serviceUpdateActionsDataSource{}, nil
}

const (
	DSNameServiceUpdateActions = "Service Update Actions Data Source"
)

type serviceUpdateActionsDataSource struct {
	framework.DataSourceWithModel[serviceUpdateActionsDataSourceModel]
}

func (d *serviceUpdateActionsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"cache_cluster_id": schema.StringAttribute{
				Optional: true,
			},
			"replication_group_id": schema.StringAttribute{
				Optional: true,
			},
			"update_actions": framework.DataSourceComputedListOfObjectAttribute[updateActionModel](ctx),
		},
	}
}

func (d *serviceUpdateActionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().ElastiCacheClient(ctx)

	var data serviceUpdateActionsDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	var input elasticache.DescribeUpdateActionsInput
	resp.Diagnostics.Append(flex.Expand(ctx, data, &input)...)
	if resp.Diagnostics.HasError() {
		return
	}
	input.CacheClusterIds = flex.StringSliceValueFromFramework(ctx, data.CacheClusterID)
	input.ReplicationGroupIds = flex.StringSliceValueFromFramework(ctx, data.ReplicationGroupID)

	updateActions, err := findServiceUpdateActions(ctx, conn, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	if len(updateActions) == 0 {
		v, d := fwtypes.NewListNestedObjectValueOfValueSlice(ctx, []updateActionModel{})
		smerr.AddEnrich(ctx, &resp.Diagnostics, d)
		if resp.Diagnostics.HasError() {
			return
		}
		data.UpdateActions = v
	} else {
		smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, updateActions, &data.UpdateActions))
		if resp.Diagnostics.HasError() {
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}

func (d *serviceUpdateActionsDataSource) ConfigValidators(context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.Conflicting(
			path.MatchRoot("cache_cluster_id"),
			path.MatchRoot("replication_group_id"),
		),
	}
}

type serviceUpdateActionsDataSourceModel struct {
	framework.WithRegionModel
	CacheClusterID     types.String                                       `tfsdk:"cache_cluster_id" autoflex:"-"`
	ReplicationGroupID types.String                                       `tfsdk:"replication_group_id" autoflex:"-"`
	UpdateActions      fwtypes.ListNestedObjectValueOf[updateActionModel] `tfsdk:"update_actions" autoflex:"-"`
}

type updateActionModel struct {
	CacheClusterID      types.String `tfsdk:"cache_cluster_id"`
	Engine              types.String `tfsdk:"engine"`
	EstimatedUpdateTime types.String `tfsdk:"estimated_update_time"`
	ReplicationGroupID  types.String `tfsdk:"replication_group_id"`
	ServiceUpdateName   types.String `tfsdk:"service_update_name"`
}

func findServiceUpdateActions(ctx context.Context, conn *elasticache.Client, input *elasticache.DescribeUpdateActionsInput) ([]awstypes.UpdateAction, error) {
	var output []awstypes.UpdateAction

	pages := elasticache.NewDescribeUpdateActionsPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		output = append(output, page.UpdateActions...)

	}

	return output, nil
}
