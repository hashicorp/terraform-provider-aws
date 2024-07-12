// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cognitoidp

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource(name="User Groups")
func newUserGroupsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &userGroupsDataSource{}, nil
}

type userGroupsDataSource struct {
	framework.DataSourceWithConfigure
}

func (*userGroupsDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "aws_cognito_user_groups"
}

func (d *userGroupsDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"groups": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[groupTypeModel](ctx),
				Computed:   true,
				ElementType: types.ObjectType{
					AttrTypes: fwtypes.AttributeTypesMust[groupTypeModel](ctx),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrUserPoolID: schema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (d *userGroupsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data userGroupsDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().CognitoIDPClient(ctx)

	groups, err := findGroupsByUserPoolID(ctx, conn, data.UserPoolID.ValueString())

	if err != nil {
		response.Diagnostics.AddError("reading Cognito User Groups", err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, groups, &data.Groups)...)
	if response.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringValue(data.UserPoolID.ValueString())

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func findGroupsByUserPoolID(ctx context.Context, conn *cognitoidentityprovider.Client, userPoolID string) ([]awstypes.GroupType, error) {
	input := &cognitoidentityprovider.ListGroupsInput{
		UserPoolId: aws.String(userPoolID),
	}
	var output []awstypes.GroupType

	pages := cognitoidentityprovider.NewListGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.Groups...)
	}

	return output, nil
}

type userGroupsDataSourceModel struct {
	Groups     fwtypes.ListNestedObjectValueOf[groupTypeModel] `tfsdk:"groups"`
	ID         types.String                                    `tfsdk:"id"`
	UserPoolID types.String                                    `tfsdk:"user_pool_id"`
}

type groupTypeModel struct {
	Description types.String `tfsdk:"description"`
	GroupName   types.String `tfsdk:"group_name"`
	Precedence  types.Int64  `tfsdk:"precedence"`
	RoleARN     types.String `tfsdk:"role_arn"`
}
