// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	awstypes "github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_api_gateway_api_keys", name="API Keys")
func newDataSourceAPIKeys(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceAPIKeys{}, nil
}

const (
	DSNameAPIKeys = "API Keys"
)

type dataSourceAPIKeys struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceAPIKeys) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"customer_id": schema.StringAttribute{
				Optional: true,
			},
			names.AttrID: framework.IDAttribute(),
			"include_values": schema.BoolAttribute{
				Optional: true,
			},
			"items": framework.DataSourceComputedListOfObjectAttribute[apiKeyModel](ctx),
		},
	}
}

func (d *dataSourceAPIKeys) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data dataSourceAPIKeysModel

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}
	data.ID = flex.StringValueToFramework(ctx, d.Meta().Region(ctx))

	conn := d.Meta().APIGatewayClient(ctx)
	input := apigateway.GetApiKeysInput{
		IncludeValues: flex.BoolFromFramework(ctx, data.IncludeValues),
		CustomerId:    flex.StringFromFramework(ctx, data.CustomerID),
	}

	items, err := findAPIKeys(ctx, conn, &input)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.APIGateway, create.ErrActionReading, DSNameAPIKeys, data.ID.ValueString(), err),
			err.Error(),
		)
		return
	}

	response.Diagnostics.Append(flex.Flatten(ctx, items, &data.Items)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func findAPIKeys(ctx context.Context, conn *apigateway.Client, input *apigateway.GetApiKeysInput) ([]awstypes.ApiKey, error) {
	var items []awstypes.ApiKey

	pages := apigateway.NewGetApiKeysPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		items = append(items, page.Items...)
	}

	return items, nil
}

type dataSourceAPIKeysModel struct {
	CustomerID    types.String                                 `tfsdk:"customer_id"`
	ID            types.String                                 `tfsdk:"id"`
	IncludeValues types.Bool                                   `tfsdk:"include_values"`
	Items         fwtypes.ListNestedObjectValueOf[apiKeyModel] `tfsdk:"items"`
}

type apiKeyModel struct {
	CreatedDate     timetypes.RFC3339    `tfsdk:"created_date"`
	CustomerID      types.String         `tfsdk:"customer_id"`
	Description     types.String         `tfsdk:"description"`
	Enabled         types.Bool           `tfsdk:"enabled"`
	ID              types.String         `tfsdk:"id"`
	LastUpdatedDate timetypes.RFC3339    `tfsdk:"last_updated_date"`
	Name            types.String         `tfsdk:"name"`
	StageKeys       fwtypes.ListOfString `tfsdk:"stage_keys"`
	Tags            fwtypes.MapOfString  `tfsdk:"tags"`
	Value           types.String         `tfsdk:"value"`
}
