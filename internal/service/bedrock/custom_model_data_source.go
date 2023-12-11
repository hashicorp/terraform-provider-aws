// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkDataSource
// @Tags(identifierAttribute="model_arn")
func newDataSourceCustomModel(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &dataSourceCustomModel{}, nil
}

type dataSourceCustomModel struct {
	framework.DataSourceWithConfigure
}

// Metadata should return the full name of the data source, such as
// examplecloud_thing.
func (d *dataSourceCustomModel) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "aws_bedrock_custom_model"
}

// Schema returns the schema for this data source.
func (d *dataSourceCustomModel) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"model_id": schema.StringAttribute{
				Required: true,
			},
			"base_model_arn": schema.StringAttribute{
				Computed: true,
			},
			"creation_time": schema.StringAttribute{
				Computed: true,
			},
			"hyper_parameters": schema.MapAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
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
			"model_kms_key_arn": schema.StringAttribute{
				Computed: true,
			},
			"model_name": schema.StringAttribute{
				Computed: true,
			},
			"training_data_config": schema.StringAttribute{
				Computed: true,
			},
			"output_data_config": schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"training_metrics": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"training_loss": schema.Float64Attribute{
						Computed: true,
					},
				},
			},
			"validation_data_config": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"validators": schema.ListAttribute{
						ElementType: types.StringType,
						Computed:    true,
					},
				},
			},
			"validation_metrics": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"validation_loss": schema.Float64Attribute{
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
func (d *dataSourceCustomModel) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var data customModel

	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := d.Meta().BedrockClient(ctx)

	input := &bedrock.GetCustomModelInput{
		ModelIdentifier: data.ModelID.ValueStringPointer(),
	}
	model, err := conn.GetCustomModel(ctx, input)
	if err != nil {
		response.Diagnostics.AddError("reading Bedrock Custom Model", err.Error())
		return
	}

	jobTags, err := listTags(ctx, conn, *model.JobArn)
	if err != nil {
		response.Diagnostics.AddError("reading Tags for Job", err.Error())
	}
	data.JobTags = flex.FlattenFrameworkStringValueMap(ctx, jobTags.IgnoreAWS().Map())

	data.refreshFromOutput(ctx, model)
	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type customModel struct {
	ID                   types.String          `tfsdk:"id"`
	ModelID              types.String          `tfsdk:"model_id"`
	BaseModelArn         types.String          `tfsdk:"base_model_arn"`
	CreationTime         types.String          `tfsdk:"creation_time"`
	JobArn               types.String          `tfsdk:"job_arn"`
	ModelArn             types.String          `tfsdk:"model_arn"`
	ModelName            types.String          `tfsdk:"model_name"`
	JobName              types.String          `tfsdk:"job_name"`
	ModelKmsKeyArn       types.String          `tfsdk:"model_kms_key_arn"`
	HyperParameters      types.Map             `tfsdk:"hyper_parameters"`
	TrainingDataConfig   types.String          `tfsdk:"training_data_config"`
	TrainingMetrics      types.List            `tfsdk:"training_metrics"`
	ValidationDataConfig *validationDataConfig `tfsdk:"validation_data_config"`
	ValidationMetrics    []validationMetrics   `tfsdk:"validation_metrics"`
	OutputDataConfig     types.String          `tfsdk:"output_data_config"`
	JobTags              types.Map             `tfsdk:"job_tags"`
	Tags                 types.Map             `tfsdk:"tags"`
}

func (data *customModel) refreshFromOutput(ctx context.Context, model *bedrock.GetCustomModelOutput) {
	if model == nil {
		return
	}

	data.ID = flex.StringToFramework(ctx, model.ModelArn)
	data.BaseModelArn = flex.StringToFramework(ctx, model.BaseModelArn)
	data.CreationTime = flex.StringValueToFramework[string](ctx, model.CreationTime.Format(time.RFC3339))
	data.JobArn = flex.StringToFramework(ctx, model.JobArn)
	data.ModelArn = flex.StringToFramework(ctx, model.ModelArn)
	data.ModelName = flex.StringToFramework(ctx, model.ModelName)
	data.JobName = flex.StringToFramework(ctx, model.JobName)
	data.ModelKmsKeyArn = flex.StringToFramework(ctx, model.ModelKmsKeyArn)
	data.HyperParameters = flex.FlattenFrameworkStringValueMap(ctx, model.HyperParameters)
	data.TrainingDataConfig = flex.StringToFramework(ctx, model.TrainingDataConfig.S3Uri)
	data.TrainingMetrics = flattenTrainingMetrics(ctx, model.TrainingMetrics)
	data.ValidationMetrics = flattenValidationMetrics(ctx, model.ValidationMetrics)
	data.OutputDataConfig = flex.StringToFramework(ctx, model.OutputDataConfig.S3Uri)
}
