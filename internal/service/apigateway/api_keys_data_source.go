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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
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

func (d *dataSourceAPIKeys) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "aws_api_gateway_api_keys"
}

func (d *dataSourceAPIKeys) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"customer_id": schema.StringAttribute{
				Optional: true,
			},
			"id": framework.IDAttribute(),
			"include_values": schema.BoolAttribute{
				Optional: true,
			},
		},
		Blocks: map[string]schema.Block{
			"items": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						names.AttrCreatedDate: schema.StringAttribute{
							Computed: true,
						},
						"customer_id": schema.StringAttribute{
							Computed: true,
						},
						names.AttrDescription: schema.StringAttribute{
							Computed: true,
						},
						names.AttrEnabled: schema.BoolAttribute{
							Computed: true,
						},
						names.AttrID: framework.IDAttribute(),
						names.AttrLastUpdatedDate: schema.StringAttribute{
							Computed: true,
						},
						names.AttrName: schema.StringAttribute{
							Computed: true,
						},
						"stage_keys": schema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
						},
						names.AttrTags: tftags.TagsAttribute(),
						names.AttrValue: schema.StringAttribute{
							Computed:  true,
							Sensitive: true,
						},
					},
				},
			},
		},
	}
}

func (d *dataSourceAPIKeys) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data dataSourceAPIKeysModel

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	data.ID = flex.StringValueToFramework(ctx, d.Meta().Region)

	if response.Diagnostics.HasError() {
		return
	}

	var apiKeyItems []awstypes.ApiKey

	conn := d.Meta().APIGatewayClient(ctx)
	input := &apigateway.GetApiKeysInput{
		IncludeValues: flex.BoolFromFramework(ctx, data.IncludeValues),
		CustomerId:    flex.StringFromFramework(ctx, data.CustomerID),
	}

	pages := apigateway.NewGetApiKeysPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			response.Diagnostics.AddError(
				create.ProblemStandardMessage(names.APIGateway, create.ErrActionReading, DSNameAPIKeys, data.ID.ValueString(), err),
				err.Error(),
			)
			return
		}

		apiKeyItems = append(apiKeyItems, page.Items...)
	}

	data.Items = flattenAPIKeyItems(ctx, apiKeyItems)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func flattenAPIKeyItems(ctx context.Context, apiKeyItems []awstypes.ApiKey) []apiKeyModel {
	apiKeys := []apiKeyModel{}
	for _, apiKeyItem := range apiKeyItems {
		keyItem := apiKeyModel{
			CreatedDate:    flex.TimeToFramework(ctx, apiKeyItem.CreatedDate),
			CustomerID:     flex.StringToFramework(ctx, apiKeyItem.CustomerId),
			Description:    flex.StringToFramework(ctx, apiKeyItem.Description),
			Name:           flex.StringToFramework(ctx, apiKeyItem.Name),
			Enabled:        flex.BoolToFramework(ctx, &apiKeyItem.Enabled),
			LastUpdateDate: flex.TimeToFramework(ctx, apiKeyItem.LastUpdatedDate),
			ID:             flex.StringToFramework(ctx, apiKeyItem.Id),
			StageKeys:      flex.FlattenFrameworkStringValueList(ctx, apiKeyItem.StageKeys),
			Tags:           flex.FlattenFrameworkStringValueMap(ctx, apiKeyItem.Tags),
			Value:          flex.StringToFramework(ctx, apiKeyItem.Value),
		}

		apiKeys = append(apiKeys, keyItem)
	}

	return apiKeys
}

type dataSourceAPIKeysModel struct {
	CustomerID    types.String  `tfsdk:"customer_id"`
	ID            types.String  `tfsdk:"id"`
	IncludeValues types.Bool    `tfsdk:"include_values"`
	Items         []apiKeyModel `tfsdk:"items"`
}

type apiKeyModel struct {
	CreatedDate    timetypes.RFC3339 `tfsdk:"created_date"`
	CustomerID     types.String      `tfsdk:"customer_id"`
	Description    types.String      `tfsdk:"description"`
	Enabled        types.Bool        `tfsdk:"enabled"`
	ID             types.String      `tfsdk:"id"`
	LastUpdateDate timetypes.RFC3339 `tfsdk:"last_updated_date"`
	Name           types.String      `tfsdk:"name"`
	StageKeys      types.List        `tfsdk:"stage_keys"`
	Tags           types.Map         `tfsdk:"tags"`
	Value          types.String      `tfsdk:"value"`
}
