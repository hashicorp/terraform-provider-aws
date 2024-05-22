// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource(name="User Group")
func newUserGroupDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &userGroupDataSource{}, nil
}

const (
	DSNameUserGroup = "User Group Data Source"
)

type userGroupDataSource struct {
	framework.DataSourceWithConfigure
}

func (d *userGroupDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
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
			"user_pool_id": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (d *userGroupDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data dataSourceDataSourceUserGroupData

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	parts := []string{
		data.Name.ValueString(),
		data.UserPoolID.ValueString(),
	}
	partCount := 2
	id, err := intflex.FlattenResourceId(parts, partCount, false)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CognitoIDP, create.ErrActionFlatteningResourceId, DSNameUserGroup, data.Name.String(), err),
			err.Error(),
		)
		return
	}
	data.ID = types.StringValue(id)

	params := &cognitoidentityprovider.GetGroupInput{
		GroupName:  data.Name.ValueStringPointer(),
		UserPoolId: data.UserPoolID.ValueStringPointer(),
	}
	// 🌱 For the person who migrates to sdkv2:
	// this should work by just updating the client, and removing the WithContext method.
	conn := d.Meta().CognitoIDPConn(ctx)
	resp, err := conn.GetGroupWithContext(ctx, params)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CognitoIDP, create.ErrActionReading, DSNameUserGroup, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(flex.Flatten(ctx, resp.Group, &data)...)
	if response.Diagnostics.HasError() {
		return
	}
	data.Name = types.StringValue(aws.StringValue(resp.Group.GroupName))

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type dataSourceDataSourceUserGroupData struct {
	Description types.String `tfsdk:"description"`
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Precedence  types.Int64  `tfsdk:"precedence"`
	RoleARN     types.String `tfsdk:"role_arn"`
	UserPoolID  types.String `tfsdk:"user_pool_id"`
}
