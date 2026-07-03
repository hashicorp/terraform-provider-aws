// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package elasticache

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_elasticache_replication_groups", name="Replication Groups")
func newReplicationGroupsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &replicationGroupsDataSource{}, nil
}

type replicationGroupsDataSource struct {
	framework.DataSourceWithModel[replicationGroupsDataSourceModel]
}

func (d *replicationGroupsDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"replication_group_ids": schema.ListAttribute{
				CustomType:  fwtypes.ListOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
		},
	}
}

// Read is called when the provider must read data source values in order to update state.
// Config values should be read from the ReadRequest and new state values set on the ReadResponse.
func (d *replicationGroupsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data replicationGroupsDataSourceModel

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().ElastiCacheClient(ctx)

	input := elasticache.DescribeReplicationGroupsInput{}
	output, err := findReplicationGroups(ctx, conn, &input, tfslices.PredicateTrue[*awstypes.ReplicationGroup]())
	if err != nil {
		response.Diagnostics.AddError("reading ElastiCache Replication Groups", err.Error())
		return
	}

	replicationGroupIDs := tfslices.ApplyToAll(output, func(v awstypes.ReplicationGroup) string {
		return aws.ToString(v.ReplicationGroupId)
	})

	data.ID = types.StringValue(d.Meta().Region(ctx))
	data.ReplicationGroupIDs = fwflex.FlattenFrameworkStringValueListOfString(ctx, replicationGroupIDs)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type replicationGroupsDataSourceModel struct {
	framework.WithRegionModel
	ID                  types.String         `tfsdk:"id"`
	ReplicationGroupIDs fwtypes.ListOfString `tfsdk:"replication_group_ids"`
}
