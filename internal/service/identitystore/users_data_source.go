// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package identitystore

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/identitystore"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

// @FrameworkDataSource(name="Users")
func newUsersDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &UsersDataSource{}, nil
}

type UsersDataSource struct {
	framework.DataSourceWithConfigure
}

func (*UsersDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	response.TypeName = "aws_identitystore_users"
}

func (d *UsersDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"users": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[UserModel](ctx),
				Computed:   true,
				ElementType: types.ObjectType{
					AttrTypes: fwtypes.AttributeTypesMust[UserModel](ctx),
				},
			},
			"identity_store_id": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (d *UsersDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data UsersDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().IdentityStoreClient(ctx)

	input := &identitystore.ListUsersInput{
		IdentityStoreId: fwflex.StringFromFramework(ctx, data.IdentityStoreID),
	}

	var output *identitystore.ListUsersOutput
	pages := identitystore.NewListUsersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			response.Diagnostics.AddError("listing IdentityStore Users", err.Error())

			return
		}

		if output == nil {
			output = page
		} else {
			output.Users = append(output.Users, page.Users...)
		}
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type UsersDataSourceModel struct {
	IdentityStoreID types.String                               `tfsdk:"identity_store_id"`
	Users           fwtypes.ListNestedObjectValueOf[UserModel] `tfsdk:"users"`
}

type UserModel struct {
	Addresses         fwtypes.ListNestedObjectValueOf[externalIDModel] `tfsdk:"addresses"`
	DisplayName       types.String                                     `tfsdk:"display_name"`
	Emails            fwtypes.ListNestedObjectValueOf[externalIDModel] `tfsdk:"emails"`
	ExternalIDs       fwtypes.ListNestedObjectValueOf[externalIDModel] `tfsdk:"external_ids"`
	IdentityStoreID   types.String                                     `tfsdk:"identity_store_id"`
	Locale            types.String                                     `tfsdk:"locale"`
	Name              fwtypes.ListNestedObjectValueOf[externalIDModel] `tfsdk:"name"`
	NickName          types.String                                     `tfsdk:"nickname"`
	PhoneNumbers      fwtypes.ListNestedObjectValueOf[externalIDModel] `tfsdk:"phone_numbers"`
	PreferredLanguage types.String                                     `tfsdk:"preferred_language"`
	ProfileUrl        types.String                                     `tfsdk:"profile_url"`
	Timezone          types.String                                     `tfsdk:"timezone"`
	Title             types.String                                     `tfsdk:"title"`
	UserID            types.String                                     `tfsdk:"user_id"`
	UserName          types.String                                     `tfsdk:"user_name"`
	UserType          types.String                                     `tfsdk:"user_type"`
}
