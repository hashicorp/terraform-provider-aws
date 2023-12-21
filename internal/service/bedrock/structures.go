// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package bedrock

import (
	"context"

	awstypes "github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

type vpcConfig struct {
	SecurityGroupIds types.Set `tfsdk:"security_group_ids"`
	SubnetIds        types.Set `tfsdk:"subnet_ids"`
}

func expandVPCConfig(ctx context.Context, l []vpcConfig) *awstypes.VpcConfig {
	if len(l) == 0 {
		return nil
	}
	vpcConfigs := make([]*awstypes.VpcConfig, len(l))
	for i, item := range l {
		vpcConfigs[i] = item.expand(ctx)
	}
	return vpcConfigs[0] // return single object, not list
}

func (v vpcConfig) expand(ctx context.Context) *awstypes.VpcConfig {
	return &awstypes.VpcConfig{
		SecurityGroupIds: flex.ExpandFrameworkStringValueSet(ctx, v.SecurityGroupIds),
		SubnetIds:        flex.ExpandFrameworkStringValueSet(ctx, v.SubnetIds),
	}
}

type outputDataConfig struct {
	S3Uri types.String `tfsdk:"s3_uri"`
}

func expandOutputDataConfig(ctx context.Context, l []outputDataConfig) *awstypes.OutputDataConfig {
	if len(l) == 0 {
		return nil
	}
	outputDataConfigs := make([]*awstypes.OutputDataConfig, len(l))
	for i, item := range l {
		outputDataConfigs[i] = item.expand(ctx)
	}
	return outputDataConfigs[0] // return single object, not list
}

func (o outputDataConfig) expand(ctx context.Context) *awstypes.OutputDataConfig {
	return &awstypes.OutputDataConfig{
		S3Uri: flex.StringFromFramework(ctx, o.S3Uri),
	}
}

type trainingDataConfig struct {
	S3Uri types.String `tfsdk:"s3_uri"`
}

func expandTrainingDataConfig(ctx context.Context, l []trainingDataConfig) *awstypes.TrainingDataConfig {
	if len(l) == 0 {
		return nil
	}
	trainingDataConfigs := make([]*awstypes.TrainingDataConfig, len(l))
	for i, item := range l {
		trainingDataConfigs[i] = item.expand(ctx)
	}
	return trainingDataConfigs[0] // return single object, not list
}

func (t trainingDataConfig) expand(ctx context.Context) *awstypes.TrainingDataConfig {
	return &awstypes.TrainingDataConfig{
		S3Uri: flex.StringFromFramework(ctx, t.S3Uri),
	}
}

type trainingMetrics struct {
	TrainingLoss types.Float64 `tfsdk:"training_loss"`
}

func expandTrainingMetrics(ctx context.Context, l []trainingMetrics) *awstypes.TrainingMetrics {
	if len(l) == 0 {
		return nil
	}
	trainingMetricsObject := make([]*awstypes.TrainingMetrics, len(l))
	for i, item := range l {
		trainingMetricsObject[i] = item.expand(ctx)
	}
	return trainingMetricsObject[0] // return single object, not list
}

func (t trainingMetrics) expand(ctx context.Context) *awstypes.TrainingMetrics {
	if t.TrainingLoss.IsNull() || t.TrainingLoss.IsUnknown() {
		return nil
	}
	return &awstypes.TrainingMetrics{
		TrainingLoss: flex.Float32FromFrameworkFloat64(ctx, t.TrainingLoss),
	}
}

type validationDataConfig struct {
	Validators types.Set `tfsdk:"validator"`
}

func expandValidationDataConfig(ctx context.Context, l []validationDataConfig, diags diag.Diagnostics) *awstypes.ValidationDataConfig {
	if len(l) == 0 {
		return nil
	}
	validationDataConfigs := make([]*awstypes.ValidationDataConfig, len(l))
	for i, item := range l {
		validationDataConfigs[i] = item.expand(ctx, diags)
	}
	return validationDataConfigs[0] // return single object, not list
}

func (v validationDataConfig) expand(ctx context.Context, diags diag.Diagnostics) *awstypes.ValidationDataConfig {
	var validators []validatorConfig
	diags.Append(v.Validators.ElementsAs(ctx, &validators, false)...)
	if diags.HasError() {
		return nil
	}
	return &awstypes.ValidationDataConfig{
		Validators: expandValidators(ctx, validators),
	}
}

type validatorConfig struct {
	S3Uri types.String `tfsdk:"s3_uri"`
}

func expandValidators(ctx context.Context, l []validatorConfig) []awstypes.Validator {
	validators := make([]awstypes.Validator, len(l))
	for i, item := range l {
		validators[i] = item.expand(ctx)
	}
	return validators
}

func (v validatorConfig) expand(ctx context.Context) awstypes.Validator {
	return awstypes.Validator{
		S3Uri: flex.StringFromFramework(ctx, v.S3Uri),
	}
}

type validationMetrics struct {
	ValidationLoss types.Float64 `tfsdk:"validation_loss"`
}

func flattenTrainingMetrics(ctx context.Context, apiObject *awstypes.TrainingMetrics) types.List {
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

func flattenValidationMetrics(ctx context.Context, apiObjects []awstypes.ValidatorMetric) []validationMetrics {
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

func flattenValidationDataConfig(ctx context.Context, apiObject *awstypes.ValidationDataConfig) types.Object {
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
