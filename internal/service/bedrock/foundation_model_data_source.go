// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource(name="Foundation Model")
func newFoundationModelDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &foundationModelDataSource{}, nil
}

type foundationModelDataSource struct {
	framework.DataSourceWithConfigure
}

func (d *foundationModelDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "aws_bedrock_foundation_model"
}

func (d *foundationModelDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"customizations_supported": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
			names.AttrID: framework.IDAttribute(),
			"inference_types_supported": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
			"input_modalities": schema.SetAttribute{
				CustomType:  fwtypes.SetOfStringType,
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
				CustomType:  fwtypes.SetOfStringType,
				ElementType: types.StringType,
				Computed:    true,
			},
			names.AttrProviderName: schema.StringAttribute{
				Computed: true,
			},
			"response_streaming_supported": schema.BoolAttribute{
				Computed: true,
			},
		},
	}
}

func (d *foundationModelDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data foundationModelDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().BedrockClient(ctx)

	input := &bedrock.GetFoundationModelInput{
		ModelIdentifier: fwflex.StringFromFramework(ctx, data.ModelID),
	}

	output, err := conn.GetFoundationModel(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Bedrock Foundation Model (%s)", data.ModelID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output.ModelDetails, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	data.ID = data.ModelID

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type foundationModelDataSourceModel struct {
	CustomizationsSupported    fwtypes.SetValueOf[types.String] `tfsdk:"customizations_supported"`
	ID                         types.String                     `tfsdk:"id"`
	InferenceTypesSupported    fwtypes.SetValueOf[types.String] `tfsdk:"inference_types_supported"`
	InputModalities            fwtypes.SetValueOf[types.String] `tfsdk:"input_modalities"`
	ModelARN                   types.String                     `tfsdk:"model_arn"`
	ModelID                    types.String                     `tfsdk:"model_id"`
	ModelName                  types.String                     `tfsdk:"model_name"`
	OutputModalities           fwtypes.SetValueOf[types.String] `tfsdk:"output_modalities"`
	ProviderName               types.String                     `tfsdk:"provider_name"`
	ResponseStreamingSupported types.Bool                       `tfsdk:"response_streaming_supported"`
}
