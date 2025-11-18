// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"
	"fmt"

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

// @FrameworkDataSource("aws_bedrock_prompt_router", name="Prompt Router")
func newPromptRouterDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &promptRouterDataSource{}, nil
}

type promptRouterDataSource struct {
	framework.DataSourceWithModel[promptRouterDataSourceModel]
}

func (d *promptRouterDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrCreatedAt: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrDescription: schema.StringAttribute{
				Computed: true,
			},
			"fallback_model": framework.DataSourceComputedListOfObjectAttribute[promptRouterTargetModel](ctx),
			"models":         framework.DataSourceComputedListOfObjectAttribute[promptRouterTargetModel](ctx),
			"prompt_router_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			"prompt_router_name": schema.StringAttribute{
				Computed: true,
			},
			"routing_criteria": framework.DataSourceComputedListOfObjectAttribute[routingCriteriaModel](ctx),
			names.AttrStatus: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.PromptRouterStatus](),
				Computed:   true,
			},
			names.AttrType: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.PromptRouterType](),
				Computed:   true,
			},
			"updated_at": schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
		},
	}
}

func (d *promptRouterDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data promptRouterDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().BedrockClient(ctx)

	output, err := findPromptRouterByARN(ctx, conn, data.PromptRouterARN.ValueString())

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Bedrock Prompt Router (%s)", data.PromptRouterARN.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type promptRouterDataSourceModel struct {
	framework.WithRegionModel
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
