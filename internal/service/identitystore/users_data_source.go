// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package identitystore

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/identitystore"
	awstypes "github.com/aws/aws-sdk-go-v2/service/identitystore/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_identitystore_users", name="Users")
func newUsersDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &UsersDataSource{}, nil
}

const (
	DSNameUsers = "Users Data Source"
)

type UsersDataSource struct {
	framework.DataSourceWithConfigure
}

func (d *UsersDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"identity_store_id": schema.StringAttribute{
				Required: true,
			},
			"users": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[userModel](ctx),
				Computed:   true,
			},
		},
	}
}

func (d *UsersDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data usersDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().IdentityStoreClient(ctx)
	input := identitystore.ListUsersInput{
		IdentityStoreId: fwflex.StringFromFramework(ctx, data.IdentityStoreID),
	}

	output, err := findUsers(ctx, conn, &input)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.IdentityStore, create.ErrActionReading, DSNameUsers, data.IdentityStoreID.String(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data.Users)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func findUsers(ctx context.Context, conn *identitystore.Client, input *identitystore.ListUsersInput) ([]awstypes.User, error) {
	var output []awstypes.User

	paginator := identitystore.NewListUsersPaginator(conn, input)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		output = append(output, page.Users...)
	}

	return output, nil
}

type usersDataSourceModel struct {
	IdentityStoreID types.String                               `tfsdk:"identity_store_id"`
	Users           fwtypes.ListNestedObjectValueOf[userModel] `tfsdk:"users"`
}

type userModel struct {
	Addresses         fwtypes.ListNestedObjectValueOf[addressesModel]    `tfsdk:"addresses"`
	DisplayName       types.String                                       `tfsdk:"display_name"`
	Emails            fwtypes.ListNestedObjectValueOf[emailsModel]       `tfsdk:"emails"`
	ExternalIDs       fwtypes.ListNestedObjectValueOf[externalIDModel]   `tfsdk:"external_ids"`
	IdentityStoreID   types.String                                       `tfsdk:"identity_store_id"`
	Locale            types.String                                       `tfsdk:"locale"`
	Name              fwtypes.ListNestedObjectValueOf[nameModel]         `tfsdk:"name"`
	Nickname          types.String                                       `tfsdk:"nickname"`
	PhoneNumbers      fwtypes.ListNestedObjectValueOf[phoneNumbersModel] `tfsdk:"phone_numbers"`
	PreferredLanguage types.String                                       `tfsdk:"preferred_language"`
	ProfileUrl        types.String                                       `tfsdk:"profile_url"`
	Timezone          types.String                                       `tfsdk:"timezone"`
	Title             types.String                                       `tfsdk:"title"`
	UserID            types.String                                       `tfsdk:"user_id"`
	UserName          types.String                                       `tfsdk:"user_name"`
	UserType          types.String                                       `tfsdk:"user_type"`
}

type addressesModel struct {
	Country       types.String `tfsdk:"country"`
	Formatted     types.String `tfsdk:"formatted"`
	Locality      types.String `tfsdk:"locality"`
	PostalCode    types.String `tfsdk:"postal_code"`
	Primary       types.Bool   `tfsdk:"primary"`
	Region        types.String `tfsdk:"region"`
	StreetAddress types.String `tfsdk:"street_address"`
	Type          types.String `tfsdk:"type"`
}

type emailsModel struct {
	Primary types.Bool   `tfsdk:"primary"`
	Type    types.String `tfsdk:"type"`
	Value   types.String `tfsdk:"value"`
}

type nameModel struct {
	FamilyName      types.String `tfsdk:"family_name"`
	Formatted       types.String `tfsdk:"formatted"`
	GivenName       types.String `tfsdk:"given_name"`
	HonorificPrefix types.String `tfsdk:"honorific_prefix"`
	HonorificSuffix types.String `tfsdk:"honorific_suffix"`
	MiddleName      types.String `tfsdk:"middle_name"`
}

type phoneNumbersModel struct {
	Primary types.Bool   `tfsdk:"primary"`
	Type    types.String `tfsdk:"type"`
	Value   types.String `tfsdk:"value"`
}
