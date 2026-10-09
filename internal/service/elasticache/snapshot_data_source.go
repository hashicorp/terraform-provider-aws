// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package elasticache

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_elasticache_snapshot", name="Snapshot")
func newSnapshotDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &snapshotDataSource{}, nil
}

type snapshotDataSource struct {
	framework.DataSourceWithModel[snapshotDataSourceModel]
}

func (d *snapshotDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
			},
			names.AttrAutoMinorVersionUpgrade: schema.BoolAttribute{
				Computed: true,
			},
			"automatic_failover": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.AutomaticFailoverStatus](),
				Computed:   true,
			},
			"cache_cluster_create_time": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"cluster_id": schema.StringAttribute{
				Required: true,
			},
			"data_tiering": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.DataTieringStatus](),
				Computed:   true,
			},
			"durability": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.Durability](),
				Computed:   true,
			},
			names.AttrEngine: schema.StringAttribute{
				Computed: true,
			},
			names.AttrEngineVersion: schema.StringAttribute{
				Computed: true,
			},
			names.AttrKMSKeyID: schema.StringAttribute{
				Computed: true,
			},
			"maintenance_window": schema.StringAttribute{
				Computed: true,
			},
			names.AttrMostRecent: schema.BoolAttribute{
				Optional: true,
			},
			"node_snapshots": framework.DataSourceComputedListOfObjectAttribute[nodeSnapshotModel](ctx),
			"node_type": schema.StringAttribute{
				Computed: true,
			},
			"num_cache_nodes": schema.Int32Attribute{
				Computed: true,
			},
			"num_node_groups": schema.Int32Attribute{
				Computed: true,
			},
			names.AttrParameterGroupName: schema.StringAttribute{
				Computed: true,
			},
			names.AttrPort: schema.Int32Attribute{
				Computed: true,
			},
			"preferred_availability_zone": schema.StringAttribute{
				Computed: true,
			},
			"preferred_outpost_arn": schema.StringAttribute{
				Computed: true,
			},
			"replication_group_description": schema.StringAttribute{
				Computed: true,
			},
			"replication_group_id": schema.StringAttribute{
				Computed: true,
			},
			"snapshot_name": schema.StringAttribute{
				Computed: true,
			},
			"snapshot_retention_limit": schema.Int32Attribute{
				Computed: true,
			},
			"snapshot_source": schema.StringAttribute{
				Computed: true,
			},
			"snapshot_window": schema.StringAttribute{
				Computed: true,
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
			},
			"subnet_group_name": schema.StringAttribute{
				Computed: true,
			},
			names.AttrTopicARN: schema.StringAttribute{
				Computed: true,
			},
			names.AttrVPCID: schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *snapshotDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data snapshotDataSourceModel
	smerr.AddEnrich(ctx, &response.Diagnostics, request.Config.Get(ctx, &data))
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().ElastiCacheClient(ctx)

	clusterID := data.CacheClusterID.ValueString()
	input := elasticache.DescribeSnapshotsInput{
		CacheClusterId:      aws.String(clusterID),
		ShowNodeGroupConfig: aws.Bool(true),
	}

	snapshots, err := findSnapshots(ctx, conn, &input)
	if err != nil && !retry.NotFound(err) {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, clusterID)
		return
	}

	if len(snapshots) == 0 {
		smerr.AddError(ctx, &response.Diagnostics, fmt.Errorf("no ElastiCache Snapshot found matching cluster ID: %s", clusterID), smerr.ID, clusterID)
		return
	}

	if len(snapshots) > 1 && !data.MostRecent.ValueBool() {
		smerr.AddError(ctx, &response.Diagnostics, fmt.Errorf("multiple ElastiCache Snapshots matched cluster ID %s; set most_recent to true to select the newest", clusterID), smerr.ID, clusterID)
		return
	}

	snapshot := mostRecentSnapshot(snapshots)

	smerr.AddEnrich(ctx, &response.Diagnostics, flex.Flatten(ctx, &snapshot, &data), smerr.ID, clusterID)
	if response.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &response.Diagnostics, response.State.Set(ctx, &data), smerr.ID, clusterID)
}

type snapshotDataSourceModel struct {
	framework.WithRegionModel
	ARN                         fwtypes.ARN                                          `tfsdk:"arn"`
	AutoMinorVersionUpgrade     types.Bool                                           `tfsdk:"auto_minor_version_upgrade"`
	AutomaticFailover           fwtypes.StringEnum[awstypes.AutomaticFailoverStatus] `tfsdk:"automatic_failover"`
	CacheClusterCreateTime      timetypes.RFC3339                                    `tfsdk:"cache_cluster_create_time"`
	CacheClusterID              types.String                                         `tfsdk:"cluster_id"`
	CacheNodeType               types.String                                         `tfsdk:"node_type"`
	CacheParameterGroupName     types.String                                         `tfsdk:"parameter_group_name"`
	CacheSubnetGroupName        types.String                                         `tfsdk:"subnet_group_name"`
	DataTiering                 fwtypes.StringEnum[awstypes.DataTieringStatus]       `tfsdk:"data_tiering"`
	Durability                  fwtypes.StringEnum[awstypes.Durability]              `tfsdk:"durability"`
	Engine                      types.String                                         `tfsdk:"engine"`
	EngineVersion               types.String                                         `tfsdk:"engine_version"`
	KmsKeyID                    types.String                                         `tfsdk:"kms_key_id"`
	MostRecent                  types.Bool                                           `tfsdk:"most_recent" autoflex:"-"`
	NodeSnapshots               fwtypes.ListNestedObjectValueOf[nodeSnapshotModel]   `tfsdk:"node_snapshots"`
	NumCacheNodes               types.Int32                                          `tfsdk:"num_cache_nodes"`
	NumNodeGroups               types.Int32                                          `tfsdk:"num_node_groups"`
	Port                        types.Int32                                          `tfsdk:"port"`
	PreferredAvailabilityZone   types.String                                         `tfsdk:"preferred_availability_zone"`
	PreferredMaintenanceWindow  types.String                                         `tfsdk:"maintenance_window"`
	PreferredOutpostARN         types.String                                         `tfsdk:"preferred_outpost_arn"`
	ReplicationGroupDescription types.String                                         `tfsdk:"replication_group_description"`
	ReplicationGroupID          types.String                                         `tfsdk:"replication_group_id"`
	SnapshotName                types.String                                         `tfsdk:"snapshot_name"`
	SnapshotRetentionLimit      types.Int32                                          `tfsdk:"snapshot_retention_limit"`
	SnapshotSource              types.String                                         `tfsdk:"snapshot_source"`
	SnapshotStatus              types.String                                         `tfsdk:"status"`
	SnapshotWindow              types.String                                         `tfsdk:"snapshot_window"`
	TopicARN                    types.String                                         `tfsdk:"topic_arn"`
	VpcID                       types.String                                         `tfsdk:"vpc_id"`
}

type nodeSnapshotModel struct {
	CacheClusterID         types.String                                                 `tfsdk:"cache_cluster_id"`
	CacheNodeCreateTime    timetypes.RFC3339                                            `tfsdk:"cache_node_create_time"`
	CacheNodeID            types.String                                                 `tfsdk:"cache_node_id"`
	CacheSize              types.String                                                 `tfsdk:"cache_size"`
	NodeGroupConfiguration fwtypes.ListNestedObjectValueOf[nodeGroupConfigurationModel] `tfsdk:"node_group_configuration"`
	NodeGroupID            types.String                                                 `tfsdk:"node_group_id"`
	SnapshotCreateTime     timetypes.RFC3339                                            `tfsdk:"snapshot_create_time"`
}

type nodeGroupConfigurationModel struct {
	NodeGroupID              types.String                      `tfsdk:"node_group_id"`
	PrimaryAvailabilityZone  types.String                      `tfsdk:"primary_availability_zone"`
	PrimaryOutpostARN        types.String                      `tfsdk:"primary_outpost_arn"`
	ReplicaAvailabilityZones fwtypes.ListValueOf[types.String] `tfsdk:"replica_availability_zones"`
	ReplicaCount             types.Int32                       `tfsdk:"replica_count"`
	ReplicaOutpostARNs       fwtypes.ListValueOf[types.String] `tfsdk:"replica_outpost_arns"`
	Slots                    types.String                      `tfsdk:"slots"`
}

func findSnapshots(ctx context.Context, conn *elasticache.Client, input *elasticache.DescribeSnapshotsInput) ([]awstypes.Snapshot, error) {
	var output []awstypes.Snapshot

	pages := elasticache.NewDescribeSnapshotsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.SnapshotNotFoundFault](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Snapshots...)
	}

	return output, nil
}

// snapshotCreateTime returns the latest node snapshot create time (Snapshot has no top-level one), falling back to the cluster create time.
func snapshotCreateTime(snapshot awstypes.Snapshot) time.Time {
	var latest time.Time
	for _, nodeSnapshot := range snapshot.NodeSnapshots {
		if t := aws.ToTime(nodeSnapshot.SnapshotCreateTime); t.After(latest) {
			latest = t
		}
	}

	if latest.IsZero() {
		latest = aws.ToTime(snapshot.CacheClusterCreateTime)
	}

	return latest
}

// mostRecentSnapshot returns the newest snapshot by creation time; the caller must ensure the slice is non-empty.
func mostRecentSnapshot(snapshots []awstypes.Snapshot) awstypes.Snapshot {
	return slices.MaxFunc(snapshots, func(a, b awstypes.Snapshot) int {
		return snapshotCreateTime(a).Compare(snapshotCreateTime(b))
	})
}
