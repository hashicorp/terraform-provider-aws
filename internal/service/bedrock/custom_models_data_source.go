// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
)

// @FrameworkDataSource
func newDataSourceCustomModels(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceCustomModels{}, nil
}

type dataSourceCustomModels struct {
	framework.DataSourceWithConfigure
}

// Metadata should return the full name of the data source, such as
// examplecloud_thing.
func (d *dataSourceCustomModels) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "aws_bedrock_custom_models"
}

// Schema returns the schema for this data source.
func (d *dataSourceCustomModels) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"model_summaries": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"base_model_arn": schema.StringAttribute{
							Computed: true,
						},
						"base_model_name": schema.StringAttribute{
							Computed: true,
						},
						"model_arn": schema.StringAttribute{
							Computed: true,
						},
						"model_name": schema.StringAttribute{
							Computed: true,
						},
						"creation_time": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

// Read is called when the provider must read data source values in order to update state.
// Config values should be read from the ReadRequest and new state values set on the ReadResponse.
func (d *dataSourceCustomModels) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data customModels

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().BedrockClient(ctx)

	models, err := conn.ListCustomModels(ctx, nil)
	if err != nil {
		response.Diagnostics.AddError("reading Bedrock Custom Models", err.Error())
		return
	}

	data.ID = flex.StringToFramework(ctx, &d.Meta().Region)
	response.Diagnostics.Append(data.refreshFromOutput(ctx, models)...)
	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type customModels struct {
	ID             types.String `tfsdk:"id"`
	ModelSummaries types.List   `tfsdk:"model_summaries"`
}

func (cms *customModels) refreshFromOutput(ctx context.Context, out *bedrock.ListCustomModelsOutput) diag.Diagnostics {
	var diags diag.Diagnostics

	if out == nil {
		return diags
	}

	// cms.ModelSummaries = flattenCustomModelSummaries(ctx, out.ModelSummaries)

	return diags
}

// func flattenCustomModelSummaries(ctx context.Context, models []bedrock_types.CustomModelSummary) types.List {
// 	attributeTypes := flex.AttributeTypesMust[customModels](ctx)
// 	elemType := types.ObjectType{AttrTypes: attributeTypes}

// 	if models == nil {
// 		return types.ListNull(elemType)
// 	}

// 	attrs := make([]attr.Value, 0, len(models))
// 	for _, model := range models {
// 		attr := map[string]attr.Value{}
// 		attr["base_model_arn"] = flex.StringToFramework(ctx, model.BaseModelArn)
// 		attr["base_model_name"] = flex.StringToFramework(ctx, model.BaseModelName)
// 		attr["model_arn"] = flex.StringToFramework(ctx, model.ModelArn)
// 		attr["model_name"] = flex.StringToFramework(ctx, model.ModelName)
// 		attr["creation_time"] = flex.StringValueToFramework[string](ctx, model.CreationTime.Format(time.RFC3339))
// 		val := types.ObjectValueMust(attributeTypes, attr)
// 		attrs = append(attrs, val)
// 	}

// 	return types.ListValueMust(elemType, attrs)
// }
