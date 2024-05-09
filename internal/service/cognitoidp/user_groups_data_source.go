// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"context"

	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource(name="User Groups")
func newUserGroupsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &userGroupsDataSource{}, nil
}

const (
	DSNameUserGroups = "User Groups Data Source"
)

type userGroupsDataSource struct {
	framework.DataSourceWithConfigure
}

func (d *userGroupsDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "aws_cognito_user_groups"
}

// Schema returns the schema for this data source.
func (d *userGroupsDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"user_pool_id": schema.StringAttribute{
				Required: true,
			},
		},
		Blocks: map[string]schema.Block{
			"groups": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[dataSourceDataSourceUserGroupsGroups](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrDescription: schema.StringAttribute{
							Computed: true,
						},
						"group_name": schema.StringAttribute{
							Computed: true,
						},
						"precedence": schema.Int64Attribute{
							Computed: true,
						},
						names.AttrRoleARN: schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *userGroupsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	// 🌱 For the person who migrates to sdkv2:
	// this should work by just updating the client, and removing the WithContext method.
	conn := d.Meta().CognitoIDPConn(ctx)

	var data dataSourceDataSourceUserGroupsData
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}
	data.ID = types.StringValue(data.UserPoolID.ValueString())

	resp, err := conn.ListGroupsWithContext(ctx, &cognitoidentityprovider.ListGroupsInput{
		UserPoolId: data.UserPoolID.ValueStringPointer(),
	})
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.CognitoIDP, create.ErrActionReading, DSNameUserGroups, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(flex.Flatten(ctx, resp.Groups, &data.Groups)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type dataSourceDataSourceUserGroupsData struct {
	Groups     fwtypes.ListNestedObjectValueOf[dataSourceDataSourceUserGroupsGroups] `tfsdk:"groups"`
	ID         types.String                                                          `tfsdk:"id"`
	UserPoolID types.String                                                          `tfsdk:"user_pool_id"`
}

type dataSourceDataSourceUserGroupsGroups struct {
	Description types.String `tfsdk:"description"`
	GroupName   types.String `tfsdk:"group_name"`
	Precedence  types.Int64  `tfsdk:"precedence"`
	RoleArn     types.String `tfsdk:"role_arn"`
}
