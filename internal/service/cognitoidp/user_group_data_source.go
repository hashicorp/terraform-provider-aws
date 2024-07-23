// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource(name="User Group")
func newUserGroupDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &userGroupDataSource{}, nil
}

type userGroupDataSource struct {
	framework.DataSourceWithConfigure
}

func (*userGroupDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "aws_cognito_user_group"
}

func (d *userGroupDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrDescription: schema.StringAttribute{
				Computed: true,
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Required: true,
			},
			"precedence": schema.Int64Attribute{
				Computed: true,
			},
			names.AttrRoleARN: schema.StringAttribute{
				Computed: true,
			},
			names.AttrUserPoolID: schema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (d *userGroupDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data userGroupDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().CognitoIDPClient(ctx)

	id := userGroupCreateResourceID(data.UserPoolID.ValueString(), data.GroupName.ValueString())
	group, err := findGroupByTwoPartKey(ctx, conn, data.UserPoolID.ValueString(), data.GroupName.ValueString())

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Cognito User Group (%s)", id), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, group, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	data.ID = fwflex.StringValueToFramework(ctx, id)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type userGroupDataSourceModel struct {
	Description types.String `tfsdk:"description"`
	GroupName   types.String `tfsdk:"name"`
	ID          types.String `tfsdk:"id"`
	Precedence  types.Int64  `tfsdk:"precedence"`
	RoleARN     types.String `tfsdk:"role_arn"`
	UserPoolID  types.String `tfsdk:"user_pool_id"`
}
