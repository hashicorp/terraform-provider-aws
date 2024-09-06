// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource(name="Custom Model")
func newCustomModelDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &customModelDataSource{}, nil
}

type customModelDataSource struct {
	framework.DataSourceWithConfigure
}

func (d *customModelDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "aws_bedrock_custom_model"
}

func (d *customModelDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"base_model_arn": schema.StringAttribute{
				Computed: true,
			},
			names.AttrCreationTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			"hyperparameters": schema.MapAttribute{
				CustomType:  fwtypes.MapOfStringType,
				Computed:    true,
				ElementType: types.StringType,
			},
			names.AttrID: framework.IDAttribute(),
			"job_arn": schema.StringAttribute{
				Computed: true,
			},
			"job_name": schema.StringAttribute{
				Computed: true,
			},
			"job_tags": tftags.TagsAttributeComputedOnly(),
			"model_arn": schema.StringAttribute{
				Computed: true,
			},
			"model_id": schema.StringAttribute{
				Required: true,
			},
			"model_kms_key_arn": schema.StringAttribute{
				Computed: true,
			},
			"model_name": schema.StringAttribute{
				Computed: true,
			},
			"model_tags": tftags.TagsAttributeComputedOnly(),
			"output_data_config": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[customModelOutputDataConfigModel](ctx),
				Computed:   true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"s3_uri": types.StringType,
					},
				},
			},
			"training_data_config": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[customModelTrainingDataConfigModel](ctx),
				Computed:   true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"s3_uri": types.StringType,
					},
				},
			},
			"training_metrics": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[customModelTrainingMetricsModel](ctx),
				Computed:   true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"training_loss": types.Float64Type,
					},
				},
			},
			"validation_data_config": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[customModelValidationDataConfigModel](ctx),
				Computed:   true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"validator": fwtypes.NewListNestedObjectTypeOf[customModelValidatorConfigModel](ctx),
					},
				},
			},
			"validation_metrics": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[customModelValidationMetricsModel](ctx),
				Computed:   true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"validation_loss": types.Float64Type,
					},
				},
			},
		},
	}
}

func (d *customModelDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data customModelDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().BedrockClient(ctx)

	modelID := data.ModelID.ValueString()
	outputGM, err := findCustomModelByID(ctx, conn, modelID)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Bedrock Custom Model (%s)", modelID), err.Error())

		return
	}

	jobARN := aws.ToString(outputGM.JobArn)
	outputGJ, err := findModelCustomizationJobByID(ctx, conn, jobARN)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Bedrock Custom Model customization job (%s)", jobARN), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, outputGM, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Some fields are only available in GetModelCustomizationJobOutput.
	var dataFromGetModelCustomizationJob customModelResourceModel
	response.Diagnostics.Append(fwflex.Flatten(ctx, outputGJ, &dataFromGetModelCustomizationJob)...)
	if response.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringValue(modelID)
	data.JobName = dataFromGetModelCustomizationJob.JobName
	data.ValidationDataConfig = dataFromGetModelCustomizationJob.ValidationDataConfig

	jobTags, err := listTags(ctx, conn, jobARN)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Bedrock Custom Model customization job (%s) tags", jobARN), err.Error())

		return
	}

	data.JobTags = fwflex.FlattenFrameworkStringValueMap(ctx, jobTags.IgnoreAWS().Map())

	modelARN := aws.ToString(outputGM.ModelArn)
	modelTags, err := listTags(ctx, conn, modelARN)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Bedrock Custom Model (%s) tags", modelARN), err.Error())

		return
	}

	data.ModelTags = fwflex.FlattenFrameworkStringValueMap(ctx, modelTags.IgnoreAWS().Map())

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type customModelDataSourceModel struct {
	BaseModelARN         types.String                                                          `tfsdk:"base_model_arn"`
	CreationTime         timetypes.RFC3339                                                     `tfsdk:"creation_time"`
	HyperParameters      fwtypes.MapValueOf[types.String]                                      `tfsdk:"hyperparameters"`
	ID                   types.String                                                          `tfsdk:"id"`
	JobARN               types.String                                                          `tfsdk:"job_arn"`
	JobName              types.String                                                          `tfsdk:"job_name"`
	JobTags              types.Map                                                             `tfsdk:"job_tags"`
	ModelARN             types.String                                                          `tfsdk:"model_arn"`
	ModelID              types.String                                                          `tfsdk:"model_id"`
	ModelKMSKeyARN       types.String                                                          `tfsdk:"model_kms_key_arn"`
	ModelName            types.String                                                          `tfsdk:"model_name"`
	ModelTags            types.Map                                                             `tfsdk:"model_tags"`
	OutputDataConfig     fwtypes.ListNestedObjectValueOf[customModelOutputDataConfigModel]     `tfsdk:"output_data_config"`
	TrainingDataConfig   fwtypes.ListNestedObjectValueOf[customModelTrainingDataConfigModel]   `tfsdk:"training_data_config"`
	TrainingMetrics      fwtypes.ListNestedObjectValueOf[customModelTrainingMetricsModel]      `tfsdk:"training_metrics"`
	ValidationDataConfig fwtypes.ListNestedObjectValueOf[customModelValidationDataConfigModel] `tfsdk:"validation_data_config"`
	ValidationMetrics    fwtypes.ListNestedObjectValueOf[customModelValidationMetricsModel]    `tfsdk:"validation_metrics"`
}
