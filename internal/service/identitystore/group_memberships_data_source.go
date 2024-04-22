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

// @FrameworkDataSource(name="Group Memberships")
func newGroupMembershipsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &groupMembershipsDataSource{}, nil
}

type groupMembershipsDataSource struct {
	framework.DataSourceWithConfigure
}

func (*groupMembershipsDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	response.TypeName = "aws_identitystore_group_memberships"
}

func (d *groupMembershipsDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"group_memberships": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[groupMembershipModel](ctx),
				Computed:   true,
				ElementType: types.ObjectType{
					AttrTypes: fwtypes.AttributeTypesMust[groupMembershipModel](ctx),
				},
			},
			"group_id": schema.StringAttribute{
				Required: true,
			},
			"identity_store_id": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (d *groupMembershipsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data groupMembershipsDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().IdentityStoreClient(ctx)

	input := &identitystore.ListGroupMembershipsInput{
		GroupId:         fwflex.StringFromFramework(ctx, data.GroupID),
		IdentityStoreId: fwflex.StringFromFramework(ctx, data.IdentityStoreID),
	}

	var output *identitystore.ListGroupMembershipsOutput
	pages := identitystore.NewListGroupMembershipsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			response.Diagnostics.AddError("listing IdentityStore Group memberships", err.Error())

			return
		}

		if output == nil {
			output = page
		} else {
			output.GroupMemberships = append(output.GroupMemberships, page.GroupMemberships...)
		}
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type groupMembershipsDataSourceModel struct {
	IdentityStoreID  types.String                                          `tfsdk:"identity_store_id"`
	GroupID          types.String                                          `tfsdk:"group_id"`
	GroupMemberships fwtypes.ListNestedObjectValueOf[groupMembershipModel] `tfsdk:"group_memberships"`
}

type groupMembershipModel struct {
	MemberID        types.String `tfsdk:"member_id"`
	MembershipID    types.String `tfsdk:"membership_id"`
	GroupID         types.String `tfsdk:"group_id"`
	IdentityStoreID types.String `tfsdk:"identity_store_id"`
}
