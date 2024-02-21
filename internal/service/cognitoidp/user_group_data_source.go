// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
)

// @FrameworkDataSource
func newDataSourceDataSourceUserGroup(context.Context) (datasource.DataSourceWithConfigure, error) {
	d := &dataSourceDataSourceUserGroup{}
	d.SetMigratedFromPluginSDK(true)

	return d, nil
}

type dataSourceDataSourceUserGroup struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceDataSourceUserGroup) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "aws_cognito_user_group"
}

func (d *dataSourceDataSourceUserGroup) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"description": schema.StringAttribute{
				Computed: true,
			},
			"id": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"precedence": schema.Int64Attribute{
				Computed: true,
			},
			"role_arn": schema.StringAttribute{
				Computed: true,
			},
			"user_pool_id": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (d *dataSourceDataSourceUserGroup) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data dataSourceDataSourceUserGroupData

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

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
			fmt.Sprintf("reading Cognito User Group (%s) for UserPool (%s)", data.Name.ValueString(), data.UserPoolID.ValueString()),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(flex.Flatten(ctx, resp.Group, &data)...)
	if response.Diagnostics.HasError() {
		return
	}
	id := fmt.Sprintf("%s/%s", data.Name.ValueString(), data.UserPoolID.ValueString())
	data.ID = types.StringValue(id)
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
