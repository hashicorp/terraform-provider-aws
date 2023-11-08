// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const DSNameFoundationModel = "Foundation Model Data Source"

// @FrameworkDataSource(name="Foundation Model")
func newDataSourceFoundationModel(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceFoundationModel{}, nil
}

type dataSourceFoundationModel struct {
	framework.DataSourceWithConfigure
}

func (d *dataSourceFoundationModel) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "aws_bedrock_foundation_model"
}

func (d *dataSourceFoundationModel) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"customizations_supported": schema.SetAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"id": schema.StringAttribute{
				Computed: true,
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
				Required: true,
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
	}
}

func (d *dataSourceFoundationModel) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	conn := d.Meta().BedrockClient(ctx)

	var data foundationModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	input := &bedrock.GetFoundationModelInput{
		ModelIdentifier: data.ModelID.ValueStringPointer(),
	}
	model, err := conn.GetFoundationModel(ctx, input)
	if err != nil {
		response.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Bedrock, create.ErrActionReading, DSNameFoundationModel, data.ModelID.String(), err),
			err.Error(),
		)
		return
	}

	data.refreshFromOutput(ctx, model)
	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type foundationModel struct {
	CustomizationsSupported    types.Set    `tfsdk:"customizations_supported"`
	ID                         types.String `tfsdk:"id"`
	InferenceTypesSupported    types.Set    `tfsdk:"inference_types_supported"`
	InputModalities            types.Set    `tfsdk:"input_modalities"`
	ModelArn                   types.String `tfsdk:"model_arn"`
	ModelID                    types.String `tfsdk:"model_id"`
	ModelName                  types.String `tfsdk:"model_name"`
	OutputModalities           types.Set    `tfsdk:"output_modalities"`
	ProviderName               types.String `tfsdk:"provider_name"`
	ResponseStreamingSupported types.Bool   `tfsdk:"response_streaming_supported"`
}

func (data *foundationModel) refreshFromOutput(ctx context.Context, model *bedrock.GetFoundationModelOutput) {
	if model == nil {
		return
	}

	data.ID = flex.StringToFramework(ctx, model.ModelDetails.ModelId)
	data.ModelArn = flex.StringToFramework(ctx, model.ModelDetails.ModelArn)
	data.ModelID = flex.StringToFramework(ctx, model.ModelDetails.ModelId)
	data.ModelName = flex.StringToFramework(ctx, model.ModelDetails.ModelName)
	data.ProviderName = flex.StringToFramework(ctx, model.ModelDetails.ProviderName)
	data.CustomizationsSupported = flex.FlattenFrameworkStringValueSet(ctx, toStringSlice(model.ModelDetails.CustomizationsSupported))
	data.InferenceTypesSupported = flex.FlattenFrameworkStringValueSet(ctx, toStringSlice(model.ModelDetails.InferenceTypesSupported))
	data.InputModalities = flex.FlattenFrameworkStringValueSet(ctx, toStringSlice(model.ModelDetails.InputModalities))
	data.OutputModalities = flex.FlattenFrameworkStringValueSet(ctx, toStringSlice(model.ModelDetails.OutputModalities))
	data.ResponseStreamingSupported = flex.BoolToFramework(ctx, model.ModelDetails.ResponseStreamingSupported)
}

// toStringSlice converts a slice of custom string types to a slice of strings
func toStringSlice[T ~string](values []T) []string {
	var out []string
	for _, v := range values {
		out = append(out, string(v))
	}
	return out
}
