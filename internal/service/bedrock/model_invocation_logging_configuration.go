// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	bedrock_types "github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
)

type resourceModelInvocationLoggingConfigurationModel struct {
	ID            types.String        `tfsdk:"id"`
	LoggingConfig *loggingConfigModel `tfsdk:"logging_config"`
}

type loggingConfigModel struct {
	EmbeddingDataDeliveryEnabled types.Bool             `tfsdk:"embedding_data_delivery_enabled"`
	ImageDataDeliveryEnabled     types.Bool             `tfsdk:"image_data_delivery_enabled"`
	TextDataDeliveryEnabled      types.Bool             `tfsdk:"text_data_delivery_enabled"`
	CloudWatchConfig             *cloudWatchConfigModel `tfsdk:"cloud_watch_config"`
	S3Config                     *s3ConfigModel         `tfsdk:"s3_config"`
}

type cloudWatchConfigModel struct {
	LogGroupName              types.String   `tfsdk:"log_group_name"`
	RoleArn                   types.String   `tfsdk:"role_arn"`
	LargeDataDeliveryS3Config *s3ConfigModel `tfsdk:"large_data_delivery_s3_config"`
}

type s3ConfigModel struct {
	BucketName types.String `tfsdk:"bucket_name"`
	KeyPrefix  types.String `tfsdk:"key_prefix"`
}

// @FrameworkResource
func newResourceModelInvocationLoggingConfiguration(context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceModelInvocationLoggingConfiguration{}, nil
}

type resourceModelInvocationLoggingConfiguration struct {
	framework.ResourceWithConfigure
}

func (r *resourceModelInvocationLoggingConfiguration) Metadata(_ context.Context, request resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_bedrock_model_invocation_logging_configuration"
}

func (r *resourceModelInvocationLoggingConfiguration) Schema(ctx context.Context, request resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": framework.IDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"logging_config": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"embedding_data_delivery_enabled": schema.BoolAttribute{
						Required: true,
					},
					"image_data_delivery_enabled": schema.BoolAttribute{
						Required: true,
					},
					"text_data_delivery_enabled": schema.BoolAttribute{
						Required: true,
					},
				},
				Blocks: map[string]schema.Block{
					"cloud_watch_config": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"log_group_name": schema.StringAttribute{
								Optional: true,
							},
							"role_arn": schema.StringAttribute{
								Optional: true,
							},
						},
						Blocks: map[string]schema.Block{
							"large_data_delivery_s3_config": schema.SingleNestedBlock{
								Attributes: map[string]schema.Attribute{
									"bucket_name": schema.StringAttribute{
										Optional: true,
									},
									"key_prefix": schema.StringAttribute{
										Optional: true,
									},
								},
							},
						},
					},
					"s3_config": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{
							"bucket_name": schema.StringAttribute{
								Optional: true,
							},
							"key_prefix": schema.StringAttribute{
								Optional: true,
							},
						},
					},
				},
			},
		},
	}
}

func (r *resourceModelInvocationLoggingConfiguration) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resourceModelInvocationLoggingConfigurationModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	lc := expandLoggingConfig(data.LoggingConfig)

	conn := r.Meta().BedrockClient(ctx)
	input := bedrock.PutModelInvocationLoggingConfigurationInput{
		LoggingConfig: lc,
	}

	_, err := conn.PutModelInvocationLoggingConfiguration(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError("failed to put model invocation logging configuration", err.Error())
		return
	}

	data.ID = flex.StringValueToFramework(ctx, r.Meta().Region)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceModelInvocationLoggingConfiguration) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceModelInvocationLoggingConfigurationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().BedrockClient(ctx)
	output, err := conn.GetModelInvocationLoggingConfiguration(ctx, &bedrock.GetModelInvocationLoggingConfigurationInput{})
	if err != nil {
		resp.Diagnostics.AddError("failed to get model invocation logging configuration", err.Error())
		return
	}

	state.ID = flex.StringValueToFramework(ctx, r.Meta().Region)
	state.LoggingConfig = flattenLoggingConfig(ctx, output.LoggingConfig)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceModelInvocationLoggingConfiguration) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data resourceModelInvocationLoggingConfigurationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	lc := expandLoggingConfig(data.LoggingConfig)

	conn := r.Meta().BedrockClient(ctx)
	input := bedrock.PutModelInvocationLoggingConfigurationInput{
		LoggingConfig: lc,
	}

	_, err := conn.PutModelInvocationLoggingConfiguration(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError("failed to put model invocation logging configuration", err.Error())
		return
	}

	data.ID = flex.StringValueToFramework(ctx, r.Meta().Region)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceModelInvocationLoggingConfiguration) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().BedrockClient(ctx)

	_, err := conn.DeleteModelInvocationLoggingConfiguration(ctx, &bedrock.DeleteModelInvocationLoggingConfigurationInput{})
	if err != nil {
		resp.Diagnostics.AddError("failed to delete model invocation logging configuration", err.Error())
		return
	}
}

func (r *resourceModelInvocationLoggingConfiguration) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func flattenLoggingConfig(ctx context.Context, apiObject *bedrock_types.LoggingConfig) *loggingConfigModel {
	if apiObject == nil {
		return nil
	}

	return &loggingConfigModel{
		EmbeddingDataDeliveryEnabled: flex.BoolToFramework(ctx, apiObject.EmbeddingDataDeliveryEnabled),
		ImageDataDeliveryEnabled:     flex.BoolToFramework(ctx, apiObject.ImageDataDeliveryEnabled),
		TextDataDeliveryEnabled:      flex.BoolToFramework(ctx, apiObject.TextDataDeliveryEnabled),
		CloudWatchConfig:             flattenCloudWatchConfig(ctx, apiObject.CloudWatchConfig),
		S3Config:                     flattenS3Config(ctx, apiObject.S3Config),
	}
}

func flattenCloudWatchConfig(ctx context.Context, apiObject *bedrock_types.CloudWatchConfig) *cloudWatchConfigModel {
	if apiObject == nil {
		return nil
	}

	return &cloudWatchConfigModel{
		LogGroupName:              flex.StringToFramework(ctx, apiObject.LogGroupName),
		RoleArn:                   flex.StringToFramework(ctx, apiObject.RoleArn),
		LargeDataDeliveryS3Config: flattenS3Config(ctx, apiObject.LargeDataDeliveryS3Config),
	}
}

func flattenS3Config(ctx context.Context, apiObject *bedrock_types.S3Config) *s3ConfigModel {
	if apiObject == nil {
		return nil
	}

	return &s3ConfigModel{
		BucketName: flex.StringToFramework(ctx, apiObject.BucketName),
		KeyPrefix:  flex.StringToFramework(ctx, apiObject.KeyPrefix),
	}
}

func expandLoggingConfig(model *loggingConfigModel) *bedrock_types.LoggingConfig {
	if model == nil {
		return nil
	}

	apiObject := &bedrock_types.LoggingConfig{
		EmbeddingDataDeliveryEnabled: model.EmbeddingDataDeliveryEnabled.ValueBoolPointer(),
		ImageDataDeliveryEnabled:     model.ImageDataDeliveryEnabled.ValueBoolPointer(),
		TextDataDeliveryEnabled:      model.TextDataDeliveryEnabled.ValueBoolPointer(),
	}
	if model.CloudWatchConfig != nil {
		apiObject.CloudWatchConfig = expandCloudWatchConfig(model.CloudWatchConfig)
	}
	if model.S3Config != nil {
		apiObject.S3Config = expandS3Config(model.S3Config)
	}

	return apiObject
}

func expandCloudWatchConfig(model *cloudWatchConfigModel) *bedrock_types.CloudWatchConfig {
	if model == nil {
		return nil
	}

	apiObject := &bedrock_types.CloudWatchConfig{
		LogGroupName:              model.LogGroupName.ValueStringPointer(),
		RoleArn:                   model.RoleArn.ValueStringPointer(),
		LargeDataDeliveryS3Config: expandS3Config(model.LargeDataDeliveryS3Config),
	}

	return apiObject
}

func expandS3Config(model *s3ConfigModel) *bedrock_types.S3Config {
	if model == nil {
		return nil
	}

	apiObject := &bedrock_types.S3Config{
		BucketName: model.BucketName.ValueStringPointer(),
		KeyPrefix:  model.KeyPrefix.ValueStringPointer(),
	}

	return apiObject
}
