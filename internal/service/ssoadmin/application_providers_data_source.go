// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_ssoadmin_application_providers", name="Application Providers")
func newApplicationProvidersDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &applicationProvidersDataSource{}, nil
}

type applicationProvidersDataSource struct {
	framework.DataSourceWithModel[applicationProvidersDataSourceModel]
}

func (d *applicationProvidersDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"application_providers": framework.DataSourceComputedListOfObjectAttribute[applicationProviderModel](ctx),
			names.AttrID:            framework.IDAttribute(),
		},
	}
}

func (d *applicationProvidersDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data applicationProvidersDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().SSOAdminClient(ctx)

	var apiObjects []awstypes.ApplicationProvider
	var input ssoadmin.ListApplicationProvidersInput
	pages := ssoadmin.NewListApplicationProvidersPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			response.Diagnostics.AddError("reading SSO Application Providers", err.Error())

			return
		}

		apiObjects = append(apiObjects, page.ApplicationProviders...)
	}

	data.ID = fwflex.StringValueToFramework(ctx, d.Meta().Region(ctx))

	response.Diagnostics.Append(fwflex.Flatten(ctx, apiObjects, &data.ApplicationProviders)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type applicationProvidersDataSourceModel struct {
	framework.WithRegionModel
	ApplicationProviders fwtypes.ListNestedObjectValueOf[applicationProviderModel] `tfsdk:"application_providers"`
	ID                   types.String                                              `tfsdk:"id"`
}

type applicationProviderModel struct {
	ApplicationProviderARN types.String                                      `tfsdk:"application_provider_arn"`
	DisplayData            fwtypes.ListNestedObjectValueOf[displayDataModel] `tfsdk:"display_data"`
	FederationProtocol     types.String                                      `tfsdk:"federation_protocol"`
}

type displayDataModel struct {
	Description types.String `tfsdk:"description"`
	DisplayName types.String `tfsdk:"display_name"`
	IconURL     types.String `tfsdk:"icon_url"`
}
