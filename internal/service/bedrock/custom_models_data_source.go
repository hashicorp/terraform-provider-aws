// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_bedrock_custom_models", name="Custom Models")
func newCustomModelsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &customModelsDataSource{}, nil
}

type customModelsDataSource struct {
	framework.DataSourceWithConfigure
}

func (d *customModelsDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID:      framework.IDAttribute(),
			"model_summaries": framework.DataSourceComputedListOfObjectAttribute[customModelSummaryModel](ctx),
		},
	}
}

func (d *customModelsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data customModelsDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().BedrockClient(ctx)

	input := &bedrock.ListCustomModelsInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	customModel, err := findCustomModels(ctx, conn, input)

	if err != nil {
		response.Diagnostics.AddError("listing Bedrock Custom Models", err.Error())

		return
	}

	output := &bedrock.ListCustomModelsOutput{
		ModelSummaries: customModel,
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringValue(d.Meta().Region(ctx))

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func findCustomModels(ctx context.Context, conn *bedrock.Client, input *bedrock.ListCustomModelsInput) ([]awstypes.CustomModelSummary, error) {
	var output []awstypes.CustomModelSummary

	pages := bedrock.NewListCustomModelsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.ModelSummaries...)
	}

	return output, nil
}

type customModelsDataSourceModel struct {
	ID             types.String                                             `tfsdk:"id"`
	ModelSummaries fwtypes.ListNestedObjectValueOf[customModelSummaryModel] `tfsdk:"model_summaries"`
}

type customModelSummaryModel struct {
	CreationTime timetypes.RFC3339 `tfsdk:"creation_time"`
	ModelARN     fwtypes.ARN       `tfsdk:"model_arn"`
	ModelName    types.String      `tfsdk:"model_name"`
}
