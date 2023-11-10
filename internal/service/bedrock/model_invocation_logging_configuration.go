// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const ResourceNameModelInvocationLoggingConfiguration = "Model Invocation Logging Configuration"

// @FrameworkResource(name="Model Invocation Logging Configuration")
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
					"cloudwatch_config": schema.SingleNestedBlock{
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
	conn := r.Meta().BedrockClient(ctx)

	var data resourceModelInvocationLoggingConfigurationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.ID = flex.StringValueToFramework(ctx, r.Meta().Region)

	loggingConfig := expandLoggingConfig(ctx, data.LoggingConfig, resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	input := bedrock.PutModelInvocationLoggingConfigurationInput{
		LoggingConfig: loggingConfig,
	}

	_, err := conn.PutModelInvocationLoggingConfiguration(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Bedrock, create.ErrActionCreating, ResourceNameModelInvocationLoggingConfiguration, data.ID.String(), err),
			err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceModelInvocationLoggingConfiguration) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().BedrockClient(ctx)

	var state resourceModelInvocationLoggingConfigurationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	output, err := conn.GetModelInvocationLoggingConfiguration(ctx, &bedrock.GetModelInvocationLoggingConfigurationInput{})
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Bedrock, create.ErrActionReading, ResourceNameModelInvocationLoggingConfiguration, state.ID.String(), err),
			err.Error(),
		)
		return
	}

	state.LoggingConfig = flattenLoggingConfig(ctx, output.LoggingConfig)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceModelInvocationLoggingConfiguration) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().BedrockClient(ctx)

	var data resourceModelInvocationLoggingConfigurationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	loggingConfig := expandLoggingConfig(ctx, data.LoggingConfig, resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	input := bedrock.PutModelInvocationLoggingConfigurationInput{
		LoggingConfig: loggingConfig,
	}

	_, err := conn.PutModelInvocationLoggingConfiguration(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Bedrock, create.ErrActionUpdating, ResourceNameModelInvocationLoggingConfiguration, data.ID.String(), err),
			err.Error(),
		)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceModelInvocationLoggingConfiguration) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().BedrockClient(ctx)

	var data resourceModelInvocationLoggingConfigurationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := conn.DeleteModelInvocationLoggingConfiguration(ctx, &bedrock.DeleteModelInvocationLoggingConfigurationInput{})
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.Bedrock, create.ErrActionDeleting, ResourceNameModelInvocationLoggingConfiguration, data.ID.String(), err),
			err.Error(),
		)
	}
}

func (r *resourceModelInvocationLoggingConfiguration) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

type resourceModelInvocationLoggingConfigurationModel struct {
	ID            types.String `tfsdk:"id"`
	LoggingConfig types.Object `tfsdk:"logging_config"`
}

type loggingConfigModel struct {
	EmbeddingDataDeliveryEnabled types.Bool   `tfsdk:"embedding_data_delivery_enabled"`
	ImageDataDeliveryEnabled     types.Bool   `tfsdk:"image_data_delivery_enabled"`
	TextDataDeliveryEnabled      types.Bool   `tfsdk:"text_data_delivery_enabled"`
	CloudWatchConfig             types.Object `tfsdk:"cloudwatch_config"`
	S3Config                     types.Object `tfsdk:"s3_config"`
}

type cloudWatchConfigModel struct {
	LogGroupName              types.String `tfsdk:"log_group_name"`
	RoleArn                   types.String `tfsdk:"role_arn"`
	LargeDataDeliveryS3Config types.Object `tfsdk:"large_data_delivery_s3_config"`
}

type s3ConfigModel struct {
	BucketName types.String `tfsdk:"bucket_name"`
	KeyPrefix  types.String `tfsdk:"key_prefix"`
}

func flattenLoggingConfig(ctx context.Context, apiObject *awstypes.LoggingConfig) types.Object {
	attributeTypes := fwtypes.AttributeTypesMust[loggingConfigModel](ctx)
	// Reflection above cannot determine the nested object attribute types
	cwAttrTypes := fwtypes.AttributeTypesMust[cloudWatchConfigModel](ctx)
	cwAttrTypes["large_data_delivery_s3_config"] = types.ObjectType{AttrTypes: fwtypes.AttributeTypesMust[s3ConfigModel](ctx)}
	attributeTypes["cloudwatch_config"] = types.ObjectType{AttrTypes: cwAttrTypes}
	attributeTypes["s3_config"] = types.ObjectType{AttrTypes: fwtypes.AttributeTypesMust[s3ConfigModel](ctx)}

	if apiObject == nil {
		return types.ObjectNull(attributeTypes)

	}

	attrs := map[string]attr.Value{
		"embedding_data_delivery_enabled": flex.BoolToFramework(ctx, apiObject.EmbeddingDataDeliveryEnabled),
		"image_data_delivery_enabled":     flex.BoolToFramework(ctx, apiObject.ImageDataDeliveryEnabled),
		"text_data_delivery_enabled":      flex.BoolToFramework(ctx, apiObject.TextDataDeliveryEnabled),
		"cloudwatch_config":               flattenCloudWatchConfig(ctx, apiObject.CloudWatchConfig),
		"s3_config":                       flattenS3Config(ctx, apiObject.S3Config),
	}

	return types.ObjectValueMust(attributeTypes, attrs)
}

func flattenCloudWatchConfig(ctx context.Context, apiObject *awstypes.CloudWatchConfig) types.Object {
	attributeTypes := fwtypes.AttributeTypesMust[cloudWatchConfigModel](ctx)
	// Reflection above cannot determine the nested object attribute types
	attributeTypes["large_data_delivery_s3_config"] = types.ObjectType{AttrTypes: fwtypes.AttributeTypesMust[s3ConfigModel](ctx)}

	if apiObject == nil {
		return types.ObjectNull(attributeTypes)
	}

	attrs := map[string]attr.Value{
		"log_group_name":                flex.StringToFramework(ctx, apiObject.LogGroupName),
		"role_arn":                      flex.StringToFramework(ctx, apiObject.RoleArn),
		"large_data_delivery_s3_config": flattenS3Config(ctx, apiObject.LargeDataDeliveryS3Config),
	}

	return types.ObjectValueMust(attributeTypes, attrs)
}

func flattenS3Config(ctx context.Context, apiObject *awstypes.S3Config) types.Object {
	attributeTypes := fwtypes.AttributeTypesMust[s3ConfigModel](ctx)
	if apiObject == nil {
		return types.ObjectNull(attributeTypes)
	}

	attrs := map[string]attr.Value{
		"bucket_name": flex.StringToFramework(ctx, apiObject.BucketName),
		"key_prefix":  flex.StringToFramework(ctx, apiObject.KeyPrefix),
	}

	return types.ObjectValueMust(attributeTypes, attrs)
}

func expandLoggingConfig(ctx context.Context, object types.Object, diags diag.Diagnostics) *awstypes.LoggingConfig {
	if object.IsNull() {
		return nil
	}

	var conf loggingConfigModel
	diags.Append(object.As(ctx, &conf, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}

	apiObject := &awstypes.LoggingConfig{
		EmbeddingDataDeliveryEnabled: conf.EmbeddingDataDeliveryEnabled.ValueBoolPointer(),
		ImageDataDeliveryEnabled:     conf.ImageDataDeliveryEnabled.ValueBoolPointer(),
		TextDataDeliveryEnabled:      conf.TextDataDeliveryEnabled.ValueBoolPointer(),
		CloudWatchConfig:             expandCloudWatchConfig(ctx, conf.CloudWatchConfig, diags),
		S3Config:                     expandS3Config(ctx, conf.S3Config, diags),
	}

	return apiObject
}

func expandCloudWatchConfig(ctx context.Context, object types.Object, diags diag.Diagnostics) *awstypes.CloudWatchConfig {
	if object.IsNull() {
		return nil
	}

	var conf cloudWatchConfigModel
	diags.Append(object.As(ctx, &conf, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}

	return &awstypes.CloudWatchConfig{
		LogGroupName:              conf.LogGroupName.ValueStringPointer(),
		RoleArn:                   conf.RoleArn.ValueStringPointer(),
		LargeDataDeliveryS3Config: expandS3Config(ctx, conf.LargeDataDeliveryS3Config, diags),
	}
}

func expandS3Config(ctx context.Context, object types.Object, diags diag.Diagnostics) *awstypes.S3Config {
	if object.IsNull() {
		return nil
	}

	var conf s3ConfigModel
	diags.Append(object.As(ctx, &conf, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}

	return &awstypes.S3Config{
		BucketName: conf.BucketName.ValueStringPointer(),
		KeyPrefix:  conf.KeyPrefix.ValueStringPointer(),
	}
}
