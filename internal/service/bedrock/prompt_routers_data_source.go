// Copyright IBM Corp. 2014, 2025
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
)

// @FrameworkDataSource("aws_bedrock_prompt_routers", name="Prompt Routers")
func newPromptRoutersDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &promptRoutersDataSource{}, nil
}

type promptRoutersDataSource struct {
	framework.DataSourceWithModel[promptRoutersDataSourceModel]
}

func (d *promptRoutersDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"prompt_router_summaries": framework.DataSourceComputedListOfObjectAttribute[promptRouterSummaryModel](ctx),
		},
	}
}

func (d *promptRoutersDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data promptRoutersDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().BedrockClient(ctx)

	input := &bedrock.ListPromptRoutersInput{}

	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	promptRouters, err := findPromptRouters(ctx, conn, input)

	if err != nil {
		response.Diagnostics.AddError("listing Bedrock Prompt Routers", err.Error())
		return
	}

	output := &bedrock.ListPromptRoutersOutput{
		PromptRouterSummaries: promptRouters,
	}
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func findPromptRouters(ctx context.Context, conn *bedrock.Client, input *bedrock.ListPromptRoutersInput) ([]awstypes.PromptRouterSummary, error) {
	var output = make([]awstypes.PromptRouterSummary, 0)

	pages := bedrock.NewListPromptRoutersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		output = append(output, page.PromptRouterSummaries...)
	}

	return output, nil
}

type promptRoutersDataSourceModel struct {
	framework.WithRegionModel
	PromptRouterSummaries fwtypes.ListNestedObjectValueOf[promptRouterSummaryModel] `tfsdk:"prompt_router_summaries"`
}

type promptRouterSummaryModel struct {
	CreatedAt        timetypes.RFC3339                                        `tfsdk:"created_at"`
	Description      types.String                                             `tfsdk:"description"`
	FallbackModel    fwtypes.ListNestedObjectValueOf[promptRouterTargetModel] `tfsdk:"fallback_model"`
	Models           fwtypes.ListNestedObjectValueOf[promptRouterTargetModel] `tfsdk:"models"`
	PromptRouterARN  fwtypes.ARN                                              `tfsdk:"prompt_router_arn"`
	PromptRouterName types.String                                             `tfsdk:"prompt_router_name"`
	RoutingCriteria  fwtypes.ListNestedObjectValueOf[routingCriteriaModel]    `tfsdk:"routing_criteria"`
	Status           fwtypes.StringEnum[awstypes.PromptRouterStatus]          `tfsdk:"status"`
	Type             fwtypes.StringEnum[awstypes.PromptRouterType]            `tfsdk:"type"`
	UpdatedAt        timetypes.RFC3339                                        `tfsdk:"updated_at"`
}
