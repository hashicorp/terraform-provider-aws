// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const DSNameFoundationModels = "Foundation Models Data Source"

// @FrameworkDataSource(name="Foundation Models")
func newDataSourceFoundationModels(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceFoundationModels{}, nil
}

type dataSourceFoundationModels struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceFoundationModels) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "aws_bedrock_foundation_models"
}

func (d *dataSourceFoundationModels) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"by_customization_type": schema.StringAttribute{
				Optional: true,
			},
			"by_inference_type": schema.StringAttribute{
				Optional: true,
			},
			"by_output_modality": schema.StringAttribute{
				Optional: true,
			},
			"by_provider": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexache.MustCompile(`^[a-z0-9-]{1,63}$`), ""),
				},
			},
			"id": schema.StringAttribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"model_summaries": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
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
						"model_arn": schema.StringAttribute{
							Computed: true,
						},
						"model_id": schema.StringAttribute{
							Computed: true,
						},
						"model_name": schema.StringAttribute{
							Computed: true,
						},
						"output_modalities": schema.SetAttribute{
							ElementType: types.StringType,
							Computed:    true,
						},
						"provider_name": schema.StringAttribute{
							Computed: true,
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

func (d *dataSourceFoundationModels) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	conn := d.Meta().BedrockClient(ctx)

	var data foundationModels
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}
	data.ID = types.StringValue(d.Meta().Region)

	input := &bedrock.ListFoundationModelsInput{}
	if !data.ByCustomizationType.IsNull() {
		input.ByCustomizationType = awstypes.ModelCustomization(data.ByCustomizationType.ValueString())
	}
	if !data.ByInferenceType.IsNull() {
		input.ByInferenceType = awstypes.InferenceType(data.ByInferenceType.ValueString())
	}
	if !data.ByOutputModality.IsNull() {
		input.ByOutputModality = awstypes.ModelModality(data.ByOutputModality.ValueString())
	}
	if !data.ByProvider.IsNull() {
		input.ByProvider = aws.String(data.ByProvider.ValueString())
	}

	models, err := conn.ListFoundationModels(ctx, input)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Bedrock, create.ErrActionReading, DSNameFoundationModels, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	data.ModelSummaries = flattenFoundationModelSummaries(ctx, models.ModelSummaries)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type foundationModels struct {
	ByCustomizationType types.String `tfsdk:"by_customization_type"`
	ByInferenceType     types.String `tfsdk:"by_inference_type"`
	ByOutputModality    types.String `tfsdk:"by_output_modality"`
	ByProvider          types.String `tfsdk:"by_provider"`
	ID                  types.String `tfsdk:"id"`
	ModelSummaries      types.List   `tfsdk:"model_summaries"`
}

type foundationModelSummary struct {
	CustomizationsSupported    types.Set    `tfsdk:"customizations_supported"`
	InferenceTypesSupported    types.Set    `tfsdk:"inference_types_supported"`
	InputModalities            types.Set    `tfsdk:"input_modalities"`
	ModelID                    types.String `tfsdk:"model_id"`
	ModelArn                   types.String `tfsdk:"model_arn"`
	ModelName                  types.String `tfsdk:"model_name"`
	OutputModalities           types.Set    `tfsdk:"output_modalities"`
	ProviderName               types.String `tfsdk:"provider_name"`
	ResponseStreamingSupported types.Bool   `tfsdk:"response_streaming_supported"`
}

func flattenFoundationModelSummaries(ctx context.Context, models []awstypes.FoundationModelSummary) types.List {
	attributeTypes := fwtypes.AttributeTypesMust[foundationModelSummary](ctx)

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

		attr["customizations_supported"] = flex.FlattenFrameworkStringValueSet(ctx, toStringSlice(model.CustomizationsSupported))
		attr["inference_types_supported"] = flex.FlattenFrameworkStringValueSet(ctx, toStringSlice(model.InferenceTypesSupported))
		attr["input_modalities"] = flex.FlattenFrameworkStringValueSet(ctx, toStringSlice(model.InputModalities))
		attr["output_modalities"] = flex.FlattenFrameworkStringValueSet(ctx, toStringSlice(model.OutputModalities))
		attr["response_streaming_supported"] = flex.BoolToFramework(ctx, model.ResponseStreamingSupported)

		val := types.ObjectValueMust(attributeTypes, attr)
		attrs = append(attrs, val)
	}

	return types.ListValueMust(elemType, attrs)
}
