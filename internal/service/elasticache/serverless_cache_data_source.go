// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache

import (
	"context"

	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_elasticache_serverless_cache", name="Serverless Cache")
func newDataSourceServerlessCache(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceServerlessCache{}, nil
}

type dataSourceServerlessCache struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceServerlessCache) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
			},
			"cache_usage_limits": schema.ObjectAttribute{
				CustomType: fwtypes.NewObjectTypeOf[dsCacheUsageLimits](ctx),
				Computed:   true,
			},
			names.AttrCreateTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"daily_snapshot_time": schema.StringAttribute{
				Computed: true,
			},
			names.AttrDescription: schema.StringAttribute{
				Computed: true,
			},
			names.AttrEndpoint: schema.ObjectAttribute{
				CustomType: fwtypes.NewObjectTypeOf[dsEndpoint](ctx),
				Computed:   true,
			},
			names.AttrEngine: schema.StringAttribute{
				Computed: true,
			},
			"full_engine_version": schema.StringAttribute{
				Computed: true,
			},
			names.AttrKMSKeyID: schema.StringAttribute{
				Computed: true,
			},
			"major_engine_version": schema.StringAttribute{
				Computed: true,
			},
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			"reader_endpoint": schema.ObjectAttribute{
				CustomType: fwtypes.NewObjectTypeOf[dsEndpoint](ctx),
				Computed:   true,
			},
			names.AttrSecurityGroupIDs: schema.ListAttribute{
				CustomType: fwtypes.ListOfStringType,
				Computed:   true,
			},
			"snapshot_retention_limit": schema.Int64Attribute{
				Computed: true,
			},
			names.AttrStatus: schema.StringAttribute{
				Computed: true,
			},
			names.AttrSubnetIDs: schema.ListAttribute{
				CustomType: fwtypes.ListOfStringType,
				Computed:   true,
			},
			"user_group_id": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *dataSourceServerlessCache) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data dsServerlessCache
	conn := d.Meta().ElastiCacheClient(ctx)

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := findServerlessCacheByID(ctx, conn, data.Name.ValueString())

	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.ElastiCache, create.ErrActionReading, "Serverless Cache", data.Name.ValueString(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(flex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type dsServerlessCache struct {
	ARN                    fwtypes.ARN                               `tfsdk:"arn"`
	CacheUsageLimits       fwtypes.ObjectValueOf[dsCacheUsageLimits] `tfsdk:"cache_usage_limits"`
	CreateTime             timetypes.RFC3339                         `tfsdk:"create_time"`
	DailySnapshotTime      types.String                              `tfsdk:"daily_snapshot_time"`
	Description            types.String                              `tfsdk:"description"`
	Endpoint               fwtypes.ObjectValueOf[dsEndpoint]         `tfsdk:"endpoint"`
	Engine                 types.String                              `tfsdk:"engine"`
	FullEngineVersion      types.String                              `tfsdk:"full_engine_version"`
	KmsKeyID               types.String                              `tfsdk:"kms_key_id"`
	MajorEngineVersion     types.String                              `tfsdk:"major_engine_version"`
	Name                   types.String                              `tfsdk:"name"`
	ReaderEndpoint         fwtypes.ObjectValueOf[dsEndpoint]         `tfsdk:"reader_endpoint"`
	SecurityGroupIDs       fwtypes.ListValueOf[types.String]         `tfsdk:"security_group_ids"`
	SnapshotRetentionLimit types.Int64                               `tfsdk:"snapshot_retention_limit"`
	Status                 types.String                              `tfsdk:"status"`
	SubnetIDs              fwtypes.ListValueOf[types.String]         `tfsdk:"subnet_ids"`
	UserGroupID            types.String                              `tfsdk:"user_group_id"`
}

type dsCacheUsageLimits struct {
	DataStorage   fwtypes.ObjectValueOf[dsDataStorage]   `tfsdk:"data_storage"`
	ECPUPerSecond fwtypes.ObjectValueOf[dsECPUPerSecond] `tfsdk:"ecpu_per_second"`
}

type dsDataStorage struct {
	Maximum types.Int64                                  `tfsdk:"maximum"`
	Minimum types.Int64                                  `tfsdk:"minimum"`
	Unit    fwtypes.StringEnum[awstypes.DataStorageUnit] `tfsdk:"unit"`
}

type dsECPUPerSecond struct {
	Maximum types.Int64 `tfsdk:"maximum"`
	Minimum types.Int64 `tfsdk:"minimum"`
}

type dsEndpoint struct {
	Address types.String `tfsdk:"address"`
	Port    types.Int64  `tfsdk:"port"`
}
