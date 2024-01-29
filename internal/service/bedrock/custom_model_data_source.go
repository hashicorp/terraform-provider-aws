// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
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
// @Tags(identifierAttribute="model_arn")
func newCustomModelDataSource(context.Context) (datasource.DataSourceWithConfigure, error) {
	return &customModelDataSource{}, nil
}

type customModelDataSource struct {
	framework.DataSourceWithConfigure
}

func (d *customModelDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "aws_bedrock_custom_model"
}

// Schema returns the schema for this data source.
func (d *customModelDataSource) Schema(ctx context.Context, request datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"base_model_arn": schema.StringAttribute{
				Computed: true,
			},
			"creation_time": schema.StringAttribute{
				Computed: true,
			},
			"hyper_parameters": schema.MapAttribute{
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
			"output_data_config": schema.ObjectAttribute{
				CustomType: fwtypes.NewObjectTypeOf[customModelOutputDataConfigModel](ctx),
				Computed:   true,
				AttributeTypes: map[string]attr.Type{
					"s3_uri": types.StringType,
				},
			},
			names.AttrTags: tftags.TagsAttributeComputedOnly(),
			"training_data_config": schema.ObjectAttribute{
				CustomType: fwtypes.NewObjectTypeOf[customModelTrainingDataConfigModel](ctx),
				Computed:   true,
				AttributeTypes: map[string]attr.Type{
					"s3_uri": types.StringType,
				},
			},
			"training_metrics": schema.ObjectAttribute{
				CustomType: fwtypes.NewObjectTypeOf[customModelTrainingMetricsModel](ctx),
				Computed:   true,
				AttributeTypes: map[string]attr.Type{
					"training_loss": types.Float64Type,
				},
			},
			"validation_data_config": schema.ObjectAttribute{
				CustomType: fwtypes.NewObjectTypeOf[customModelValidationDataConfigModel](ctx),
				Computed:   true,
				AttributeTypes: map[string]attr.Type{
					"validators": types.ListType{ElemType: fwtypes.NewObjectTypeOf[customModelValidatorConfigModel](ctx)},
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

	output, err := findCustomModelByID(ctx, conn, data.ModelID.ValueString())

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Bedrock Custom Model (%s)", data.ModelID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	data.ID = data.ModelID

	jobARN := aws.ToString(output.JobArn)
	jobTags, err := listTags(ctx, conn, jobARN)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Bedrock Custom Model Job (%s) tags", jobARN), err.Error())

		return
	}

	data.JobTags = fwflex.FlattenFrameworkStringValueMap(ctx, jobTags.IgnoreAWS().Map())

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

type customModelDataSourceModel struct {
	BaseModelARN         types.String                                                       `tfsdk:"base_model_arn"`
	CreationTime         types.String                                                       `tfsdk:"creation_time"`
	HyperParameters      fwtypes.MapValueOf[types.String]                                   `tfsdk:"hyper_parameters"`
	ID                   types.String                                                       `tfsdk:"id"`
	JobARN               types.String                                                       `tfsdk:"job_arn"`
	JobName              types.String                                                       `tfsdk:"job_name"`
	JobTags              types.Map                                                          `tfsdk:"job_tags"`
	ModelARN             types.String                                                       `tfsdk:"model_arn"`
	ModelID              types.String                                                       `tfsdk:"model_id"`
	ModelKMSKeyARN       types.String                                                       `tfsdk:"model_kms_key_arn"`
	ModelName            types.String                                                       `tfsdk:"model_name"`
	OutputDataConfig     types.String                                                       `tfsdk:"output_data_config"`
	Tags                 types.Map                                                          `tfsdk:"tags"`
	TrainingDataConfig   types.String                                                       `tfsdk:"training_data_config"`
	TrainingMetrics      fwtypes.ListNestedObjectValueOf[customModelTrainingMetricsModel]   `tfsdk:"training_metrics"`
	ValidationDataConfig fwtypes.ObjectValueOf[customModelValidationDataConfigModel]        `tfsdk:"validation_data_config"`
	ValidationMetrics    fwtypes.ListNestedObjectValueOf[customModelValidationMetricsModel] `tfsdk:"validation_metrics"`
}

type customModelOutputDataConfigModel struct {
	S3URI types.String `tfsdk:"s3_uri"`
}

type customModelTrainingDataConfigModel struct {
	S3URI types.String `tfsdk:"s3_uri"`
}

type customModelTrainingMetricsModel struct {
	TrainingLoss types.Float64 `tfsdk:"training_loss"`
}

type customModelValidationDataConfigModel struct {
	Validators fwtypes.ListValueOf[types.String] `tfsdk:"validators"`
}

type customModelValidationMetricsModel struct {
	ValidationLoss types.Float64 `tfsdk:"validation_loss"`
}

type customModelValidatorConfigModel struct {
	S3URI types.String `tfsdk:"s3_uri"`
}

type customModelVPCConfigModel struct {
	SecurityGroupIDs fwtypes.SetValueOf[types.String] `tfsdk:"security_group_ids"`
	SubnetIDs        fwtypes.SetValueOf[types.String] `tfsdk:"subnet_ids"`
}
