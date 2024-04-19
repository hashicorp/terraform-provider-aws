// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package identitystore

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/identitystore"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource(name="Groups")
func newDataSourceGroups(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceGroups{}, nil
}

const (
	DSNameGroups = "Groups Data Source"
)

type dataSourceGroups struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceGroups) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) { // nosemgrep:ci.meta-in-func-name
	response.TypeName = "aws_identitystore_groups"
}

func (d *dataSourceGroups) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"identity_store_id": schema.StringAttribute{
				Required: true,
			},
		},
		Blocks: map[string]schema.Block{
			"groups": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[groupsData](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"group_id": schema.StringAttribute{
							Computed: true,
						},
						"description": schema.StringAttribute{
							Computed: true,
						},
						"display_name": schema.StringAttribute{
							Computed: true,
						},
					},
					Blocks: map[string]schema.Block{
						"external_ids": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[externalIdsData](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										Computed: true,
									},
									"issuer": schema.StringAttribute{
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *dataSourceGroups) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	conn := d.Meta().IdentityStoreClient(ctx)

	var data dataSourceGroupsData
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	input := &identitystore.ListGroupsInput{}
	response.Diagnostics.Append(flex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	paginator := identitystore.NewListGroupsPaginator(conn, input)
	var out identitystore.ListGroupsOutput

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.IdentityStore, create.ErrActionReading, DSNameGroups, data.IdentityStoreId.String(), err),
				err.Error(),
			)
			return
		}

		if page != nil && len(page.Groups) > 0 {
			out.Groups = append(out.Groups, page.Groups...)
		}
	}

	response.Diagnostics.Append(flex.Flatten(ctx, out, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type dataSourceGroupsData struct {
	IdentityStoreId types.String                                `tfsdk:"identity_store_id"`
	Groups          fwtypes.ListNestedObjectValueOf[groupsData] `tfsdk:"groups"`
}

type groupsData struct {
	GroupId     types.String                                     `tfsdk:"group_id"`
	Description types.String                                     `tfsdk:"description"`
	DisplayName types.String                                     `tfsdk:"display_name"`
	ExternalIds fwtypes.ListNestedObjectValueOf[externalIdsData] `tfsdk:"external_ids"`
}

type externalIdsData struct {
	Id     types.String `tfsdk:"id"`
	Issuer types.String `tfsdk:"issuer"`
}
