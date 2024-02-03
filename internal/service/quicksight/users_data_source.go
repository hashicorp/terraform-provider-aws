// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/quicksight"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
)

// @FrameworkDataSource(name="Users")
func newDataSourceQuicksightUsers(_ context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceQuicksightUsers{}, nil
}

type dataSourceQuicksightUsers struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceQuicksightUsers) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "aws_quicksight_users"
}

type dataSourceQuicksightUsersData struct {
	AwsAccountID types.String                    `tfsdk:"aws_account_id"`
	Namespace    types.String                    `tfsdk:"namespace"`
	Users        []dataSourceQuicksightUserData  `tfsdk:"users"`
	Filter       *dataSourceQuicksightUserFilter `tfsdk:"filter"`
}

type dataSourceQuicksightUserData struct {
	Active       types.Bool   `tfsdk:"active"`
	Arn          types.String `tfsdk:"arn"`
	Email        types.String `tfsdk:"email"`
	IdentityType types.String `tfsdk:"identity_type"`
	PrincipalID  types.String `tfsdk:"principal_id"`
	UserName     types.String `tfsdk:"user_name"`
	UserRole     types.String `tfsdk:"user_role"`
}

type dataSourceQuicksightUserFilter struct {
	Active        types.Bool   `tfsdk:"active"`
	EmailRegex    types.String `tfsdk:"email_regex"`
	IdentityType  types.String `tfsdk:"identity_type"`
	UserNameRegex types.String `tfsdk:"user_name_regex"`
	UserRole      types.String `tfsdk:"user_role"`
}

func (d *dataSourceQuicksightUsers) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"aws_account_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
			"namespace": schema.StringAttribute{
				Optional: true,
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"filter": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"active": schema.BoolAttribute{
						Optional: true,
					},
					"email_regex": schema.StringAttribute{
						Optional: true,
					},
					"identity_type": schema.StringAttribute{
						Optional: true,
						Validators: []validator.String{
							stringvalidator.OneOf(
								quicksight.IdentityTypeIam,
								quicksight.IdentityTypeQuicksight,
							),
						},
					},
					"user_name_regex": schema.StringAttribute{
						Optional: true,
					},
					"user_role": schema.StringAttribute{
						Optional: true,
						Validators: []validator.String{
							stringvalidator.OneOf(
								quicksight.RoleAdmin,
								quicksight.RoleAuthor,
								quicksight.RoleReader,
							),
						},
					},
				},
			},
			"users": schema.SetNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"active": schema.BoolAttribute{
							Computed: true,
						},
						"arn": schema.StringAttribute{
							Computed: true,
						},
						"email": schema.StringAttribute{
							Computed: true,
						},
						"identity_type": schema.StringAttribute{
							Computed: true,
						},
						"principal_id": schema.StringAttribute{
							Computed: true,
						},
						"user_name": schema.StringAttribute{
							Computed: true,
						},
						"user_role": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *dataSourceQuicksightUsers) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data dataSourceQuicksightUsersData

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().QuickSightConn(ctx)

	defaultNS := "default"
	in := &quicksight.ListUsersInput{
		AwsAccountId: &d.Meta().AccountID,
		Namespace:    aws.String(defaultNS)}

	out, err := conn.ListUsersWithContext(ctx, in)

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to retrieve list of all Quicksight users ",
			err.Error(),
		)
		return
	}

	if data.Filter == nil {
		data.Filter = &dataSourceQuicksightUserFilter{}
	}

	for _, user := range out.UserList {
		skip, err := filterUser(user, data.Filter)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to parse filters on user",
				err.Error(),
			)
			return
		}

		if skip {
			continue
		}

		var userData dataSourceQuicksightUserData
		userData.Active = types.BoolPointerValue(user.Active)
		userData.Arn = types.StringPointerValue(user.Arn)
		userData.Email = types.StringPointerValue(user.Email)
		userData.IdentityType = types.StringPointerValue(user.IdentityType)
		userData.PrincipalID = types.StringPointerValue(user.PrincipalId)
		userData.UserName = types.StringPointerValue(user.UserName)
		userData.UserRole = types.StringPointerValue(user.Role)

		data.Users = append(data.Users, userData)
	}

	data.AwsAccountID = types.StringPointerValue(&d.Meta().AccountID)
	data.Namespace = types.StringValue(defaultNS)

	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
