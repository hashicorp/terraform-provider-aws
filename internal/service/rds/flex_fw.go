package rds

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

func expandScalingConfigurationFramework(tfList []scalingConfiguration) *rds.ScalingConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	m := tfList[0]
	apiObject := &rds.ScalingConfiguration{}

	if !m.AutoPause.IsNull() {
		apiObject.AutoPause = aws.Bool(m.AutoPause.ValueBool())
	}

	if !m.MaxCapacity.IsNull() {
		apiObject.MaxCapacity = aws.Int64(m.MaxCapacity.ValueInt64())
	}

	if !m.MinCapacity.IsNull() {
		apiObject.MinCapacity = aws.Int64(m.MinCapacity.ValueInt64())
	}

	if !m.SecondsUntilAutoPause.IsNull() {
		apiObject.SecondsUntilAutoPause = aws.Int64(m.SecondsUntilAutoPause.ValueInt64())
	}

	if !m.TimeoutAction.IsNull() {
		apiObject.TimeoutAction = aws.String(m.TimeoutAction.ValueString())
	}

	return apiObject
}

func flattenScalingConfigurationFramework(ctx context.Context, apiObject *rds.ScalingConfiguration) types.List {
	elemType := types.ObjectType{AttrTypes: scalingConfigurationAttrTypes}

	if apiObject == nil {
		return types.ListValueMust(elemType, []attr.Value{})
	}

	attrs := map[string]attr.Value{
		"auto_pause":               flex.BoolToFrameworkLegacy(ctx, apiObject.AutoPause),
		"max_capacity":             flex.Int64ToFrameworkLegacy(ctx, apiObject.MaxCapacity),
		"min_capacity":             flex.Int64ToFrameworkLegacy(ctx, apiObject.MinCapacity),
		"seconds_until_autl_pause": flex.Int64ToFrameworkLegacy(ctx, apiObject.SecondsUntilAutoPause),
		"timeout_action":           flex.StringToFrameworkLegacy(ctx, apiObject.TimeoutAction),
	}

	vals := types.ObjectValueMust(scalingConfigurationAttrTypes, attrs)

	return types.ListValueMust(elemType, []attr.Value{vals})
}

func expandServerlessV2ScalingConfigurationFramework(tfList []serverlessV2ScalingConfiguration) *rds.ServerlessV2ScalingConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	m := tfList[0]
	apiObject := &rds.ServerlessV2ScalingConfiguration{
		MaxCapacity: aws.Float64(m.MaxCapacity.ValueFloat64()),
		MinCapacity: aws.Float64(m.MinCapacity.ValueFloat64()),
	}

	return apiObject
}

func flattenServerlessV2ScalingConfigurationFramework(ctx context.Context, apiObject *rds.ServerlessV2ScalingConfigurationInfo) types.List {
	elemType := types.ObjectType{AttrTypes: serverlessV2ScalingConfigurationAttrTypes}

	if apiObject == nil {
		return types.ListValueMust(elemType, []attr.Value{})
	}

	attrs := map[string]attr.Value{

		"max_capacity": flex.Float64ToFrameworkLegacy(ctx, apiObject.MaxCapacity),
		"min_capacity": flex.Float64ToFrameworkLegacy(ctx, apiObject.MinCapacity),
	}

	vals := types.ObjectValueMust(serverlessV2ScalingConfigurationAttrTypes, attrs)

	return types.ListValueMust(elemType, []attr.Value{vals})
}
