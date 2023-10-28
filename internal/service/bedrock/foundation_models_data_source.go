// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"

	bedrock_types "github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
)

// @FrameworkDataSource
func newDataSourceFoundationModels(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceFoundationModels{}, nil
}

type dataSourceFoundationModels struct {
	framework.DataSourceWithConfigure
}

// Metadata should return the full name of the data source, such as
// examplecloud_thing.
func (d *dataSourceFoundationModels) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "aws_bedrock_foundation_models"
}

// Schema returns the schema for this data source.
func (d *dataSourceFoundationModels) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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
						"model_arn": schema.StringAttribute{
							Computed: true,
						},
						"model_id": schema.StringAttribute{
							Computed: true,
						},
						"model_name": schema.StringAttribute{
							Computed: true,
						},
						"provider_name": schema.StringAttribute{
							Computed: true,
						},
						"customizations_supported": schema.SetAttribute{
							ElementType: types.StringType,
							Computed:    true,
						},
						"inference_types_supported": schema.SetAttribute{
							ElementType: types.StringType,
							Computed:    true,
						},
						"input_modalities": schema.SetAttribute{
							ElementType: types.StringType,
							Computed:    true,
						},
						"output_modalities": schema.SetAttribute{
							ElementType: types.StringType,
							Computed:    true,
						},
						"response_streaming_supported": schema.BoolAttribute{
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
func (d *dataSourceFoundationModels) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data foundationModels

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().BedrockClient(ctx)

	models, err := conn.ListFoundationModels(ctx, nil)
	if err != nil {
		response.Diagnostics.AddError("reading Bedrock Foundation Models", err.Error())
		return
	}

	data.ID = flex.StringToFramework(ctx, &d.Meta().Region)
	data.ModelSummaries = flattenFoundationModelSummaries(ctx, models.ModelSummaries)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type foundationModels struct {
	ID             types.String `tfsdk:"id"`
	ModelSummaries types.List   `tfsdk:"model_summaries"`
}

type foundationModelSummary struct {
	ModelID                    types.String `tfsdk:"model_id"`
	ModelArn                   types.String `tfsdk:"model_arn"`
	ModelName                  types.String `tfsdk:"model_name"`
	ProviderName               types.String `tfsdk:"provider_name"`
	CustomizationsSupported    types.Set    `tfsdk:"customizations_supported"`
	InferenceTypesSupported    types.Set    `tfsdk:"inference_types_supported"`
	InputModalities            types.Set    `tfsdk:"input_modalities"`
	OutputModalities           types.Set    `tfsdk:"output_modalities"`
	ResponseStreamingSupported types.Bool   `tfsdk:"response_streaming_supported"`
}

func flattenFoundationModelSummaries(ctx context.Context, models []bedrock_types.FoundationModelSummary) types.List {
	attributeTypes := flex.AttributeTypesMust[foundationModelSummary](ctx)

	// HACK: Reflection used above to build the attributeTypes cannot determine the ElemType
	attributeTypes["customizations_supported"] = types.SetType{ElemType: types.StringType}
	attributeTypes["inference_types_supported"] = types.SetType{ElemType: types.StringType}
	attributeTypes["input_modalities"] = types.SetType{ElemType: types.StringType}
	attributeTypes["output_modalities"] = types.SetType{ElemType: types.StringType}

	elemType := types.ObjectType{AttrTypes: attributeTypes}

	if models == nil {
		return types.ListNull(elemType)
	}

	attrs := make([]attr.Value, 0, len(models))
	for _, model := range models {
		attr := map[string]attr.Value{}
		attr["model_arn"] = flex.StringToFramework(ctx, model.ModelArn)
		attr["model_id"] = flex.StringToFramework(ctx, model.ModelId)
		attr["model_name"] = flex.StringToFramework(ctx, model.ModelName)
		attr["provider_name"] = flex.StringToFramework(ctx, model.ProviderName)

		customizationsSupported := make([]string, 0, len(model.CustomizationsSupported))
		for _, r := range model.CustomizationsSupported {
			customizationsSupported = append(customizationsSupported, string(r))
		}
		attr["customizations_supported"] = flex.FlattenFrameworkStringValueSet(ctx, customizationsSupported)

		inferenceTypesSupported := make([]string, 0, len(model.InferenceTypesSupported))
		for _, r := range model.InferenceTypesSupported {
			inferenceTypesSupported = append(inferenceTypesSupported, string(r))
		}
		attr["inference_types_supported"] = flex.FlattenFrameworkStringValueSet(ctx, inferenceTypesSupported)

		inputModalities := make([]string, 0, len(model.InputModalities))
		for _, r := range model.InputModalities {
			inputModalities = append(inputModalities, string(r))
		}
		attr["input_modalities"] = flex.FlattenFrameworkStringValueSet(ctx, inputModalities)

		outputModalities := make([]string, 0, len(model.OutputModalities))
		for _, r := range model.OutputModalities {
			outputModalities = append(outputModalities, string(r))
		}
		attr["output_modalities"] = flex.FlattenFrameworkStringValueSet(ctx, outputModalities)

		attr["response_streaming_supported"] = flex.BoolToFramework(ctx, model.ResponseStreamingSupported)
		val := types.ObjectValueMust(attributeTypes, attr)
		attrs = append(attrs, val)
	}

	return types.ListValueMust(elemType, attrs)
}
