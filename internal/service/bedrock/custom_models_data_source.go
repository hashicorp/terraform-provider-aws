// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource(name="Custom Models")
func newCustomModelsDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &customModelsDataSource{}, nil
}

type customModelsDataSource struct {
	framework.DataSourceWithConfigure
}

func (d *customModelsDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "aws_bedrock_custom_models"
}

func (d *customModelsDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			"model_summaries": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[customModelSummaryModel](ctx),
				Computed:   true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						names.AttrCreationTime: timetypes.RFC3339Type{},
						"model_arn":            types.StringType,
						"model_name":           types.StringType,
					},
				},
			},
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

	output, err := conn.ListCustomModels(ctx, input)

	if err != nil {
		response.Diagnostics.AddError("listing Bedrock Custom Models", err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringValue(d.Meta().Region)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type customModelsDataSourceModel struct {
	ID             types.String                                             `tfsdk:"id"`
	ModelSummaries fwtypes.ListNestedObjectValueOf[customModelSummaryModel] `tfsdk:"model_summaries"`
}

type customModelSummaryModel struct {
	CreationTime timetypes.RFC3339 `tfsdk:"creation_time"`
	ModelARN     types.String      `tfsdk:"model_arn"`
	ModelName    types.String      `tfsdk:"model_name"`
}
