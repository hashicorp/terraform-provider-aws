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

// @FrameworkDataSource(name="Groups")
func newGroupsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &groupsDataSource{}, nil
}

type groupsDataSource struct {
	framework.DataSourceWithConfigure
}

func (*groupsDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	response.TypeName = "aws_identitystore_groups"
}

func (d *groupsDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"groups": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[groupModel](ctx),
				Computed:   true,
				ElementType: types.ObjectType{
					AttrTypes: fwtypes.AttributeTypesMust[groupModel](ctx),
				},
			},
			"identity_store_id": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (d *groupsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data groupsDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().IdentityStoreClient(ctx)

	input := &identitystore.ListGroupsInput{
		IdentityStoreId: fwflex.StringFromFramework(ctx, data.IdentityStoreID),
	}

	var output *identitystore.ListGroupsOutput
	pages := identitystore.NewListGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			response.Diagnostics.AddError("listing IdentityStore Groups", err.Error())

			return
		}

		if output == nil {
			output = page
		} else {
			output.Groups = append(output.Groups, page.Groups...)
		}
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type groupsDataSourceModel struct {
	IdentityStoreID types.String                                `tfsdk:"identity_store_id"`
	Groups          fwtypes.ListNestedObjectValueOf[groupModel] `tfsdk:"groups"`
}

type groupModel struct {
	Description     types.String                                     `tfsdk:"description"`
	DisplayName     types.String                                     `tfsdk:"display_name"`
	ExternalIDs     fwtypes.ListNestedObjectValueOf[externalIDModel] `tfsdk:"external_ids"`
	GroupID         types.String                                     `tfsdk:"group_id"`
	IdentityStoreID types.String                                     `tfsdk:"identity_store_id"`
}

type externalIDModel struct {
	ID     types.String `tfsdk:"id"`
	Issuer types.String `tfsdk:"issuer"`
}
