// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource("aws_bedrock_custom_model", name="Custom Model")
func newCustomModelDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &customModelDataSource{}, nil
}

type customModelDataSource struct {
	framework.DataSourceWithConfigure
}

func (d *customModelDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"base_model_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
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
				CustomType: fwtypes.ARNType,
				Computed:   true,
			},
			"job_name": schema.StringAttribute{
				Computed: true,
			},
			"job_tags": tftags.TagsAttributeComputedOnly(),
			"model_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
			},
			"model_id": schema.StringAttribute{
				Required: true,
			},
			"model_kms_key_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Computed:   true,
			},
			"model_name": schema.StringAttribute{
				Computed: true,
			},
			"model_tags":             tftags.TagsAttributeComputedOnly(),
			"output_data_config":     framework.DataSourceComputedListOfObjectAttribute[outputDataConfigModel](ctx),
			"training_data_config":   framework.DataSourceComputedListOfObjectAttribute[trainingDataConfigModel](ctx),
			"training_metrics":       framework.DataSourceComputedListOfObjectAttribute[trainingMetricsModel](ctx),
			"validation_data_config": framework.DataSourceComputedListOfObjectAttribute[validationDataConfigModel](ctx),
			"validation_metrics":     framework.DataSourceComputedListOfObjectAttribute[validatorMetricModel](ctx),
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

	data.JobTags = tftags.FlattenStringValueMap(ctx, jobTags.IgnoreAWS().Map())

	modelARN := aws.ToString(outputGM.ModelArn)
	modelTags, err := listTags(ctx, conn, modelARN)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Bedrock Custom Model (%s) tags", modelARN), err.Error())

		return
	}

	data.ModelTags = tftags.FlattenStringValueMap(ctx, modelTags.IgnoreAWS().Map())

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type customModelDataSourceModel struct {
	BaseModelARN         fwtypes.ARN                                                `tfsdk:"base_model_arn"`
	CreationTime         timetypes.RFC3339                                          `tfsdk:"creation_time"`
	HyperParameters      fwtypes.MapOfString                                        `tfsdk:"hyperparameters"`
	ID                   types.String                                               `tfsdk:"id"`
	JobARN               fwtypes.ARN                                                `tfsdk:"job_arn"`
	JobName              types.String                                               `tfsdk:"job_name"`
	JobTags              tftags.Map                                                 `tfsdk:"job_tags"`
	ModelARN             fwtypes.ARN                                                `tfsdk:"model_arn"`
	ModelID              types.String                                               `tfsdk:"model_id"`
	ModelKMSKeyARN       fwtypes.ARN                                                `tfsdk:"model_kms_key_arn"`
	ModelName            types.String                                               `tfsdk:"model_name"`
	ModelTags            tftags.Map                                                 `tfsdk:"model_tags"`
	OutputDataConfig     fwtypes.ListNestedObjectValueOf[outputDataConfigModel]     `tfsdk:"output_data_config"`
	TrainingDataConfig   fwtypes.ListNestedObjectValueOf[trainingDataConfigModel]   `tfsdk:"training_data_config"`
	TrainingMetrics      fwtypes.ListNestedObjectValueOf[trainingMetricsModel]      `tfsdk:"training_metrics"`
	ValidationDataConfig fwtypes.ListNestedObjectValueOf[validationDataConfigModel] `tfsdk:"validation_data_config"`
	ValidationMetrics    fwtypes.ListNestedObjectValueOf[validatorMetricModel]      `tfsdk:"validation_metrics"`
}
