// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"

	bedrock_types "github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

type vpcConfig struct {
	SecurityGroupIds types.Set `tfsdk:"security_group_ids"`
	SubnetIds        types.Set `tfsdk:"subnet_ids"`
}

type validationDataConfig struct {
	Validators types.Set `tfsdk:"validators"`
}

type trainingMetrics struct {
	TrainingLoss types.Float64 `tfsdk:"training_loss"`
}

type validationMetrics struct {
	ValidationLoss types.Float64 `tfsdk:"validation_loss"`
}

func expandVPCConfig(ctx context.Context, tfList []vpcConfig) []bedrock_types.VpcConfig {
	if len(tfList) == 0 {
		return nil
	}
	var vpcConfigs []bedrock_types.VpcConfig

	for _, item := range tfList {
		vpcConfig := bedrock_types.VpcConfig{
			SecurityGroupIds: flex.ExpandFrameworkStringValueSet(ctx, item.SecurityGroupIds),
			SubnetIds:        flex.ExpandFrameworkStringValueSet(ctx, item.SubnetIds),
		}
		vpcConfigs = append(vpcConfigs, vpcConfig)
	}

	return vpcConfigs
}

func expandValidationDataConfig(ctx context.Context, object types.Object, diags diag.Diagnostics) *bedrock_types.ValidationDataConfig {
	if object.IsNull() {
		return nil
	}

	var model validationDataConfig
	diags.Append(object.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil
	}

	apiObject := &bedrock_types.ValidationDataConfig{}
	for _, validator := range flex.ExpandFrameworkStringValueSet(ctx, model.Validators) {
		s3uri := validator
		apiObject.Validators = append(apiObject.Validators, bedrock_types.Validator{
			S3Uri: &s3uri,
		})
	}

	return apiObject
}

func flattenTrainingMetrics(ctx context.Context, apiObject *bedrock_types.TrainingMetrics) types.List {
	attributeTypes := fwtypes.AttributeTypesMust[trainingMetrics](ctx)
	elemType := types.ObjectType{AttrTypes: attributeTypes}

	if apiObject == nil {
		return types.ListNull(elemType)
	}

	attrs := make([]attr.Value, 0, 1)
	attr := map[string]attr.Value{}
	trainingLoss := float64(*apiObject.TrainingLoss)
	attr["training_loss"] = flex.Float64ToFramework(ctx, &trainingLoss)
	val := types.ObjectValueMust(attributeTypes, attr)
	attrs = append(attrs, val)

	return types.ListValueMust(elemType, attrs)
}

func flattenValidationMetrics(ctx context.Context, apiObjects []bedrock_types.ValidatorMetric) []validationMetrics {
	if apiObjects == nil {
		return nil
	}

	var model []validationMetrics
	for _, validationMetric := range apiObjects {
		validationLoss := float64(*validationMetric.ValidationLoss)
		model = append(model, validationMetrics{
			ValidationLoss: flex.Float64ToFramework(ctx, &validationLoss),
		})
	}

	return model
}

func flattenValidationDataConfig(ctx context.Context, apiObject *bedrock_types.ValidationDataConfig) types.Object {
	attributeTypes := fwtypes.AttributeTypesMust[validationDataConfig](ctx)
	attributeTypes["validators"] = types.SetType{ElemType: types.StringType}

	if apiObject == nil {
		return types.ObjectNull(attributeTypes)
	}

	validators := []*string{}
	if apiObject != nil {
		for _, validator := range apiObject.Validators {
			validators = append(validators, validator.S3Uri)
		}
	}
	attrs := map[string]attr.Value{
		"validators": flex.FlattenFrameworkStringSet(ctx, validators),
	}

	return types.ObjectValueMust(attributeTypes, attrs)
}
