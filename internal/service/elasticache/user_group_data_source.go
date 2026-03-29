// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package elasticache

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_elasticache_user_group", name="User Group")
// @Tags(identifierAttribute="arn")
func newUserGroupDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &userGroupDataSource{}, nil
}

type userGroupDataSource struct {
	framework.DataSourceWithModel[userGroupDataSourceModel]
}

func (d *userGroupDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrEngine: schema.StringAttribute{
				Computed: true,
			},
			names.AttrID:   framework.IDAttribute(),
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
			"user_group_id": schema.StringAttribute{
				Required: true,
			},
			"user_ids": schema.ListAttribute{
				CustomType: fwtypes.ListOfStringType,
				Computed:   true,
			},
		},
	}
}

func (d *userGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().ElastiCacheClient(ctx)
	var data userGroupDataSourceModel

	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findUserGroupByID(ctx, conn, data.UserGroupID.ValueString())
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, data.UserGroupID.String())
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &data), smerr.ID, data.UserGroupID.String())
	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = data.UserGroupID

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data), smerr.ID, data.UserGroupID.String())
}

type userGroupDataSourceModel struct {
	framework.WithRegionModel
	ARN         types.String                      `tfsdk:"arn"`
	Engine      types.String                      `tfsdk:"engine"`
	ID          types.String                      `tfsdk:"id"`
	Tags        tftags.Map                        `tfsdk:"tags"`
	UserGroupID types.String                      `tfsdk:"user_group_id"`
	UserIDs     fwtypes.ListValueOf[types.String] `tfsdk:"user_ids"`
}
