// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_rds_global_cluster", name="Global Cluster")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTest=false)
func newDataSourceGlobalCluster(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceGlobalCluster{}, nil
}

type dataSourceGlobalCluster struct {
	framework.DataSourceWithModel[dataSourceGlobalClusterData]
}

func (d *dataSourceGlobalCluster) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrDatabaseName: schema.StringAttribute{
				Computed: true,
			},
			names.AttrDeletionProtection: schema.BoolAttribute{
				Computed: true,
			},
			names.AttrEngine: schema.StringAttribute{
				Computed: true,
			},
			names.AttrEngineVersion: schema.StringAttribute{
				Computed: true,
			},
			names.AttrForceDestroy: schema.BoolAttribute{
				Computed: true,
			},
			names.AttrIdentifier: schema.StringAttribute{
				Required: true,
			},
			"members": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[globalClusterMembersModel](ctx),
				Computed:   true,
			},
			names.AttrResourceID: schema.StringAttribute{
				Computed: true,
			},
			"source_db_cluster_identifier": schema.StringAttribute{
				Computed: true,
			},
			names.AttrStorageEncrypted: schema.BoolAttribute{
				Computed: true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (d *dataSourceGlobalCluster) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data dataSourceGlobalClusterData
	conn := d.Meta().RDSClient(ctx)

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := findGlobalClusterByID(ctx, conn, data.GlobalClusterIdentifier.ValueString())
	if err != nil {
		smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.GlobalClusterIdentifier.String())
		return
	}

	smerr.EnrichAppend(ctx, &response.Diagnostics, flex.Expand(ctx, data, &output))
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type dataSourceGlobalClusterData struct {
	framework.WithRegionModel
	ARN                       types.String                                               `tfsdk:"arn"`
	DatabaseName              types.String                                               `tfsdk:"database_name"`
	DeletionProtection        types.Bool                                                 `tfsdk:"deletion_protection"`
	Engine                    types.String                                               `tfsdk:"engine"`
	EngineVersion             types.String                                               `tfsdk:"engine_version"`
	ForceDestroy              types.Bool                                                 `tfsdk:"force_destroy"`
	GlobalClusterIdentifier   types.String                                               `tfsdk:"identifier"`
	GlobalClusterMembers      fwtypes.ListNestedObjectValueOf[globalClusterMembersModel] `tfsdk:"members"`
	GlobalClusterResourceID   types.String                                               `tfsdk:"resource_id"`
	SourceDbClusterIdentifier types.String                                               `tfsdk:"source_db_cluster_identifier"`
	StorageEncrypted          types.Bool                                                 `tfsdk:"storage_encrypted"`
	Tags                      tftags.Map                                                 `tfsdk:"tags"`
}

type globalClusterMembersModel struct {
	DBClusterARN types.String `tfsdk:"db_cluster_arn"`
	IsWriter     types.Bool   `tfsdk:"is_writer"`
}
