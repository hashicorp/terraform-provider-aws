// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

type dataSourceDataSourceUserGroups struct {
	framework.DataSourceWithConfigure
}

// @FrameworkDataSource(name="User Groups")
func newDataSourceDataSourceUserGroups(context.Context) (datasource.DataSourceWithConfigure, error) {
	d := &dataSourceDataSourceUserGroups{}
	d.SetMigratedFromPluginSDK(true)

	return d, nil
}

func (d *dataSourceDataSourceUserGroups) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "aws_cognito_user_groups"
}

// Schema returns the schema for this data source.
func (d *dataSourceDataSourceUserGroups) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"user_pool_id": schema.StringAttribute{
				Required: true,
			},
		},
		Blocks: map[string]schema.Block{
			"groups": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[dataSourceDataSourceUserGroupsGroups](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"description": schema.StringAttribute{
							Computed: true,
						},
						"group_name": schema.StringAttribute{
							Computed: true,
						},
						"precedence": schema.Int64Attribute{
							Computed: true,
						},
						"role_arn": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *dataSourceDataSourceUserGroups) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	// 🌱 For the person who migrates to sdkv2:
	// this should work by just updating the client, and removing the WithContext method.
	conn := d.Meta().CognitoIDPConn(ctx)

	var data dataSourceDataSourceUserGroupsData
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	resp, err := conn.ListGroupsWithContext(ctx, &cognitoidentityprovider.ListGroupsInput{
		UserPoolId: data.UserPoolID.ValueStringPointer(),
	})
	if err != nil {
		response.Diagnostics.AddError(
			fmt.Sprintf("Error reading aws_cognito_user_groups: %s", data.UserPoolID),
			err.Error(),
		)
		return
	}
	response.Diagnostics.Append(flex.Flatten(ctx, resp.Groups, &data.Groups)...)
	if response.Diagnostics.HasError() {
		return
	}
	data.ID = types.StringValue(data.UserPoolID.String())
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
