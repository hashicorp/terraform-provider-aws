// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"

	bedrock_types "github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
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

func expandVpcConfig(ctx context.Context, tfList []vpcConfig) []bedrock_types.VpcConfig {
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

func expandValidationDataConfig(ctx context.Context, model *validationDataConfig) *bedrock_types.ValidationDataConfig {
	if model == nil {
		return nil
	}

	apiObject := &bedrock_types.ValidationDataConfig{}
	for _, validator := range flex.ExpandFrameworkStringValueSet(ctx, model.Validators) {
		apiObject.Validators = append(apiObject.Validators, bedrock_types.Validator{
			S3Uri: &validator,
		})
	}

	return apiObject
}

func flattenTrainingMetrics(ctx context.Context, apiObject *bedrock_types.TrainingMetrics) *trainingMetrics {
	model := &trainingMetrics{}

	if apiObject != nil {
		trainingLoss := float64(*apiObject.TrainingLoss)
		model.TrainingLoss = flex.Float64ToFramework(ctx, &trainingLoss)
	}

	return model
}

func flattenValidationMetrics(ctx context.Context, apiObjects []bedrock_types.ValidatorMetric) []validationMetrics {
	if apiObjects != nil {
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

func flattenValidationDataConfig(ctx context.Context, apiObject *bedrock_types.ValidationDataConfig) *validationDataConfig {
	model := &validationDataConfig{}

	validators := []*string{}
	if apiObject != nil {
		for _, validator := range apiObject.Validators {
			validators = append(validators, validator.S3Uri)
		}
		model.Validators = flex.FlattenFrameworkStringSet(ctx, validators)
	}

	return model
}
