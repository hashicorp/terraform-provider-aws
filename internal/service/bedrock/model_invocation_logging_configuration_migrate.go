// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func modelInvocationLoggingConfigurationSchemaV0(ctx context.Context) schema.Schema {
	return schema.Schema{
		Version: 0,
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"logging_config": schema.SingleNestedBlock{ // nosemgrep:ci.avoid-SingleNestedBlock pre-existing, will be converted
				CustomType: fwtypes.NewObjectTypeOf[loggingConfigModelV0](ctx),
				Attributes: map[string]schema.Attribute{
					"embedding_data_delivery_enabled": schema.BoolAttribute{
						Optional: true,
						Computed: true,
						Default:  booldefault.StaticBool(true),
					},
					"image_data_delivery_enabled": schema.BoolAttribute{
						Optional: true,
						Computed: true,
						Default:  booldefault.StaticBool(true),
					},
					"text_data_delivery_enabled": schema.BoolAttribute{
						Optional: true,
						Computed: true,
						Default:  booldefault.StaticBool(true),
					},
					"video_data_delivery_enabled": schema.BoolAttribute{
						Optional: true,
						Computed: true,
						Default:  booldefault.StaticBool(true),
					},
				},
				Blocks: map[string]schema.Block{
					"cloudwatch_config": schema.SingleNestedBlock{ // nosemgrep:ci.avoid-SingleNestedBlock pre-existing, will be converted
						CustomType: fwtypes.NewObjectTypeOf[cloudWatchConfigModelV0](ctx),
						Attributes: map[string]schema.Attribute{
							names.AttrLogGroupName: schema.StringAttribute{
								Optional: true,
							},
							names.AttrRoleARN: schema.StringAttribute{
								CustomType: fwtypes.ARNType,
								Optional:   true,
							},
						},
						Blocks: map[string]schema.Block{
							"large_data_delivery_s3_config": schema.SingleNestedBlock{ // nosemgrep:ci.avoid-SingleNestedBlock pre-existing, will be converted
								CustomType: fwtypes.NewObjectTypeOf[s3ConfigModel](ctx),
								Attributes: map[string]schema.Attribute{
									names.AttrBucketName: schema.StringAttribute{
										Optional: true,
									},
									"key_prefix": schema.StringAttribute{
										Optional: true,
									},
								},
							},
						},
					},
					"s3_config": schema.SingleNestedBlock{ // nosemgrep:ci.avoid-SingleNestedBlock pre-existing, will be converted
						CustomType: fwtypes.NewObjectTypeOf[s3ConfigModel](ctx),
						Attributes: map[string]schema.Attribute{
							names.AttrBucketName: schema.StringAttribute{
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

type modelInvocationLoggingConfigurationResourceModelV0 struct {
	ID            types.String                                `tfsdk:"id"`
	LoggingConfig fwtypes.ObjectValueOf[loggingConfigModelV0] `tfsdk:"logging_config"`
}

type loggingConfigModelV0 struct {
	CloudWatchConfig             fwtypes.ObjectValueOf[cloudWatchConfigModelV0] `tfsdk:"cloudwatch_config"`
	EmbeddingDataDeliveryEnabled types.Bool                                     `tfsdk:"embedding_data_delivery_enabled"`
	ImageDataDeliveryEnabled     types.Bool                                     `tfsdk:"image_data_delivery_enabled"`
	S3Config                     fwtypes.ObjectValueOf[s3ConfigModel]           `tfsdk:"s3_config"`
	TextDataDeliveryEnabled      types.Bool                                     `tfsdk:"text_data_delivery_enabled"`
	VideoDataDeliveryEnabled     types.Bool                                     `tfsdk:"video_data_delivery_enabled"`
}

type cloudWatchConfigModelV0 struct {
	LargeDataDeliveryS3Config fwtypes.ObjectValueOf[s3ConfigModel] `tfsdk:"large_data_delivery_s3_config"`
	LogGroupName              types.String                         `tfsdk:"log_group_name"`
	RoleArn                   fwtypes.ARN                          `tfsdk:"role_arn"`
}

func upgradeModelInvocationLoggingConfigurationFromV0(ctx context.Context, request resource.UpgradeStateRequest, response *resource.UpgradeStateResponse) {
	var schemaDataV0 modelInvocationLoggingConfigurationResourceModelV0
	response.Diagnostics.Append(request.State.Get(ctx, &schemaDataV0)...)
	if response.Diagnostics.HasError() {
		return
	}

	schemaDataV1 := modelInvocationLoggingConfigurationResourceModel{
		ID:            schemaDataV0.ID,
		LoggingConfig: upgradeLoggingConfigModelFromV0(ctx, schemaDataV0.LoggingConfig, &response.Diagnostics),
	}

	response.Diagnostics.Append(response.State.Set(ctx, schemaDataV1)...)
}

func upgradeLoggingConfigModelFromV0(ctx context.Context, old fwtypes.ObjectValueOf[loggingConfigModelV0], diags *diag.Diagnostics) fwtypes.ListNestedObjectValueOf[loggingConfigModel] {
	if old.IsNull() {
		return fwtypes.NewListNestedObjectValueOfNull[loggingConfigModel](ctx)
	}

	var loggingConfigV0 loggingConfigModelV0
	diags.Append(old.As(ctx, &loggingConfigV0, basetypes.ObjectAsOptions{})...)

	newList := []loggingConfigModel{
		{
			CloudWatchConfig:             upgradeCloudWatchConfigModelFromV0(ctx, loggingConfigV0.CloudWatchConfig, diags),
			EmbeddingDataDeliveryEnabled: loggingConfigV0.EmbeddingDataDeliveryEnabled,
			ImageDataDeliveryEnabled:     loggingConfigV0.ImageDataDeliveryEnabled,
			S3Config:                     upgradeS3ConfigModel(ctx, loggingConfigV0.S3Config, diags),
			TextDataDeliveryEnabled:      loggingConfigV0.TextDataDeliveryEnabled,
			VideoDataDeliveryEnabled:     loggingConfigV0.VideoDataDeliveryEnabled,
		},
	}

	result, d := fwtypes.NewListNestedObjectValueOfValueSlice(ctx, newList)
	diags.Append(d...)

	return result
}

func upgradeCloudWatchConfigModelFromV0(ctx context.Context, old fwtypes.ObjectValueOf[cloudWatchConfigModelV0], diags *diag.Diagnostics) fwtypes.ListNestedObjectValueOf[cloudWatchConfigModel] {
	if old.IsNull() {
		return fwtypes.NewListNestedObjectValueOfNull[cloudWatchConfigModel](ctx)
	}

	var cloudWatchConfigV0 cloudWatchConfigModelV0
	diags.Append(old.As(ctx, &cloudWatchConfigV0, basetypes.ObjectAsOptions{})...)

	newList := []cloudWatchConfigModel{
		{
			LargeDataDeliveryS3Config: upgradeS3ConfigModel(ctx, cloudWatchConfigV0.LargeDataDeliveryS3Config, diags),
			LogGroupName:              cloudWatchConfigV0.LogGroupName,
			RoleArn:                   cloudWatchConfigV0.RoleArn,
		},
	}

	result, d := fwtypes.NewListNestedObjectValueOfValueSlice(ctx, newList)
	diags.Append(d...)

	return result
}

func upgradeS3ConfigModel(ctx context.Context, old fwtypes.ObjectValueOf[s3ConfigModel], diags *diag.Diagnostics) fwtypes.ListNestedObjectValueOf[s3ConfigModel] {
	if old.IsNull() {
		return fwtypes.NewListNestedObjectValueOfNull[s3ConfigModel](ctx)
	}

	var s3ConfigV0 s3ConfigModel
	diags.Append(old.As(ctx, &s3ConfigV0, basetypes.ObjectAsOptions{UnhandledUnknownAsEmpty: true})...)

	newList := []s3ConfigModel{
		{
			BucketName: s3ConfigV0.BucketName,
			KeyPrefix:  s3ConfigV0.KeyPrefix,
		},
	}

	result, d := fwtypes.NewListNestedObjectValueOfValueSlice(ctx, newList)
	diags.Append(d...)

	return result
}
