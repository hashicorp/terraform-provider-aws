// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package scheduler

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/scheduler/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func expandCapacityProviderStrategyItem(tfMap map[string]interface{}) types.CapacityProviderStrategyItem {
	if tfMap == nil {
		return types.CapacityProviderStrategyItem{}
	}

	a := types.CapacityProviderStrategyItem{}

	if v, ok := tfMap["base"].(int); ok {
		a.Base = int32(v)
	}

	if v, ok := tfMap["capacity_provider"].(string); ok && v != "" {
		a.CapacityProvider = aws.String(v)
	}

	if v, ok := tfMap["weight"].(int); ok {
		a.Weight = int32(v)
	}

	return a
}

func flattenCapacityProviderStrategyItem(apiObject types.CapacityProviderStrategyItem) map[string]interface{} {
	m := map[string]interface{}{}

	m["base"] = apiObject.Base

	if v := apiObject.CapacityProvider; v != nil {
		m["capacity_provider"] = aws.ToString(v)
	}

	m["weight"] = apiObject.Weight

	return m
}

func expandDeadLetterConfig(tfMap map[string]interface{}) *types.DeadLetterConfig {
	if tfMap == nil {
		return nil
	}

	a := &types.DeadLetterConfig{}

	if v, ok := tfMap["arn"].(string); ok && v != "" {
		a.Arn = aws.String(v)
	}

	return a
}

func flattenDeadLetterConfig(apiObject *types.DeadLetterConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.Arn; v != nil {
		m["arn"] = aws.ToString(v)
	}

	return m
}

func expandECSParameters(ctx context.Context, tfMap map[string]interface{}) *types.EcsParameters {
	if tfMap == nil {
		return nil
	}

	a := &types.EcsParameters{}

	if v, ok := tfMap["capacity_provider_strategy"].(*schema.Set); ok && v.Len() > 0 {
		for _, s := range v.List() {
			a.CapacityProviderStrategy = append(a.CapacityProviderStrategy, expandCapacityProviderStrategyItem(s.(map[string]interface{})))
		}
	}

	if v, ok := tfMap["enable_ecs_managed_tags"].(bool); ok {
		a.EnableECSManagedTags = aws.Bool(v)
	}

	if v, ok := tfMap["enable_execute_command"].(bool); ok {
		a.EnableExecuteCommand = aws.Bool(v)
	}

	if v, ok := tfMap["group"].(string); ok && v != "" {
		a.Group = aws.String(v)
	}

	if v, ok := tfMap["launch_type"].(string); ok && v != "" {
		a.LaunchType = types.LaunchType(v)
	}

	if v, ok := tfMap["network_configuration"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.NetworkConfiguration = expandNetworkConfiguration(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["placement_constraints"].(*schema.Set); ok && v.Len() > 0 {
		for _, c := range v.List() {
			a.PlacementConstraints = append(a.PlacementConstraints, expandPlacementConstraint(c.(map[string]interface{})))
		}
	}

	if v, ok := tfMap["placement_strategy"].(*schema.Set); ok && v.Len() > 0 {
		for _, c := range v.List() {
			a.PlacementStrategy = append(a.PlacementStrategy, expandPlacementStrategy(c.(map[string]interface{})))
		}
	}

	if v, ok := tfMap["platform_version"].(string); ok && v != "" {
		a.PlatformVersion = aws.String(v)
	}

	if v, ok := tfMap["propagate_tags"].(string); ok && v != "" {
		a.PropagateTags = types.PropagateTags(v)
	}

	if v, ok := tfMap["reference_id"].(string); ok && v != "" {
		a.ReferenceId = aws.String(v)
	}

	tags := tftags.New(ctx, tfMap["tags"].(map[string]interface{}))

	if len(tags) > 0 {
		for k, v := range tags.IgnoreAWS().Map() {
			a.Tags = append(a.Tags, map[string]string{
				"key":   k,
				"value": v,
			})
		}
	}

	if v, ok := tfMap["task_count"].(int); ok {
		a.TaskCount = aws.Int32(int32(v))
	}

	if v, ok := tfMap["task_definition_arn"].(string); ok && v != "" {
		a.TaskDefinitionArn = aws.String(v)
	}

	return a
}

func flattenECSParameters(ctx context.Context, apiObject *types.EcsParameters) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.CapacityProviderStrategy; v != nil {
		set := schema.NewSet(capacityProviderHash, nil)

		for _, p := range v {
			set.Add(flattenCapacityProviderStrategyItem(p))
		}

		m["capacity_provider_strategy"] = set
	}

	if v := apiObject.EnableECSManagedTags; v != nil {
		m["enable_ecs_managed_tags"] = aws.ToBool(v)
	}

	if v := apiObject.EnableExecuteCommand; v != nil {
		m["enable_execute_command"] = aws.ToBool(v)
	}

	if v := apiObject.Group; v != nil {
		m["group"] = aws.ToString(v)
	}

	if v := string(apiObject.LaunchType); v != "" {
		m["launch_type"] = v
	}

	if v := apiObject.NetworkConfiguration; v != nil {
		m["network_configuration"] = []interface{}{flattenNetworkConfiguration(v)}
	}

	if v := apiObject.PlacementConstraints; len(v) > 0 {
		set := schema.NewSet(placementConstraintHash, nil)

		for _, c := range v {
			set.Add(flattenPlacementConstraint(c))
		}

		m["placement_constraints"] = set
	}

	if v := apiObject.PlacementStrategy; len(v) > 0 {
		set := schema.NewSet(placementStrategyHash, nil)

		for _, s := range v {
			set.Add(flattenPlacementStrategy(s))
		}

		m["placement_strategy"] = set
	}

	if v := apiObject.PlatformVersion; v != nil {
		m["platform_version"] = aws.ToString(v)
	}

	if v := string(apiObject.PropagateTags); v != "" {
		m["propagate_tags"] = v
	}

	if v := apiObject.ReferenceId; v != nil {
		m["reference_id"] = aws.ToString(v)
	}

	if v := apiObject.Tags; len(v) > 0 {
		tags := make(map[string]interface{})

		for _, tagMap := range v {
			key := tagMap["key"]

			// The EventBridge Scheduler API documents raw maps instead of
			// the key-value structure expected by the RunTask API.
			if key == "" {
				continue
			}

			tags[key] = tagMap["value"]
		}

		m["tags"] = tftags.New(ctx, tags).IgnoreAWS().Map()
	}

	if v := apiObject.TaskCount; v != nil {
		m["task_count"] = int(aws.ToInt32(v))
	}

	if v := apiObject.TaskDefinitionArn; v != nil {
		m["task_definition_arn"] = aws.ToString(v)
	}

	return m
}

func expandEventBridgeParameters(tfMap map[string]interface{}) *types.EventBridgeParameters {
	if tfMap == nil {
		return nil
	}

	a := &types.EventBridgeParameters{}

	if v, ok := tfMap["detail_type"].(string); ok && v != "" {
		a.DetailType = aws.String(v)
	}

	if v, ok := tfMap["source"].(string); ok && v != "" {
		a.Source = aws.String(v)
	}

	return a
}

func flattenEventBridgeParameters(apiObject *types.EventBridgeParameters) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.DetailType; v != nil {
		m["detail_type"] = aws.ToString(v)
	}

	if v := apiObject.Source; v != nil {
		m["source"] = aws.ToString(v)
	}

	return m
}

func expandFlexibleTimeWindow(tfMap map[string]interface{}) *types.FlexibleTimeWindow {
	if tfMap == nil {
		return nil
	}

	a := &types.FlexibleTimeWindow{}

	if v, ok := tfMap["maximum_window_in_minutes"].(int); ok && v != 0 {
		a.MaximumWindowInMinutes = aws.Int32(int32(v))
	}

	if v, ok := tfMap["mode"].(string); ok && v != "" {
		a.Mode = types.FlexibleTimeWindowMode(v)
	}

	return a
}

func flattenFlexibleTimeWindow(apiObject *types.FlexibleTimeWindow) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.MaximumWindowInMinutes; v != nil {
		m["maximum_window_in_minutes"] = int(aws.ToInt32(v))
	}

	if v := string(apiObject.Mode); v != "" {
		m["mode"] = v
	}

	return m
}

func expandKinesisParameters(tfMap map[string]interface{}) *types.KinesisParameters {
	if tfMap == nil {
		return nil
	}

	a := &types.KinesisParameters{}

	if v, ok := tfMap["partition_key"].(string); ok && v != "" {
		a.PartitionKey = aws.String(v)
	}

	return a
}

func flattenKinesisParameters(apiObject *types.KinesisParameters) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.PartitionKey; v != nil {
		m["partition_key"] = aws.ToString(v)
	}

	return m
}

func expandNetworkConfiguration(tfMap map[string]interface{}) *types.NetworkConfiguration {
	if tfMap == nil {
		return nil
	}

	awsvpcConfig := &types.AwsVpcConfiguration{}

	if v, ok := tfMap["assign_public_ip"].(bool); ok {
		if v {
			awsvpcConfig.AssignPublicIp = types.AssignPublicIpEnabled
		} else {
			awsvpcConfig.AssignPublicIp = types.AssignPublicIpDisabled
		}
	}

	if v, ok := tfMap["security_groups"].(*schema.Set); ok && v.Len() > 0 {
		awsvpcConfig.SecurityGroups = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["subnets"].(*schema.Set); ok && v.Len() > 0 {
		awsvpcConfig.Subnets = flex.ExpandStringValueSet(v)
	}

	return &types.NetworkConfiguration{
		AwsvpcConfiguration: awsvpcConfig,
	}
}

func flattenNetworkConfiguration(apiObject *types.NetworkConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	// Follow the example of EventBridge targets by flattening out
	// the AWS VPC configuration.

	if v := apiObject.AwsvpcConfiguration.AssignPublicIp; v != "" {
		m["assign_public_ip"] = v == types.AssignPublicIpEnabled
	}

	if v := apiObject.AwsvpcConfiguration.SecurityGroups; v != nil {
		m["security_groups"] = flex.FlattenStringValueSet(v)
	}

	if v := apiObject.AwsvpcConfiguration.Subnets; v != nil {
		m["subnets"] = flex.FlattenStringValueSet(v)
	}

	return m
}

func expandPlacementConstraint(tfMap map[string]interface{}) types.PlacementConstraint {
	if tfMap == nil {
		return types.PlacementConstraint{}
	}

	a := types.PlacementConstraint{}

	if v, ok := tfMap["expression"].(string); ok && v != "" {
		a.Expression = aws.String(v)
	}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		a.Type = types.PlacementConstraintType(v)
	}

	return a
}

func flattenPlacementConstraint(apiObject types.PlacementConstraint) map[string]interface{} {
	m := map[string]interface{}{}

	if v := apiObject.Expression; v != nil {
		m["expression"] = aws.ToString(v)
	}

	if v := string(apiObject.Type); v != "" {
		m["type"] = v
	}

	return m
}

func expandPlacementStrategy(tfMap map[string]interface{}) types.PlacementStrategy {
	if tfMap == nil {
		return types.PlacementStrategy{}
	}

	a := types.PlacementStrategy{}

	if v, ok := tfMap["field"].(string); ok && v != "" {
		a.Field = aws.String(v)
	}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		a.Type = types.PlacementStrategyType(v)
	}

	return a
}

func flattenPlacementStrategy(apiObject types.PlacementStrategy) map[string]interface{} {
	m := map[string]interface{}{}

	if v := apiObject.Field; v != nil {
		m["field"] = aws.ToString(v)
	}

	if v := string(apiObject.Type); v != "" {
		m["type"] = v
	}

	return m
}

func expandRetryPolicy(tfMap map[string]interface{}) *types.RetryPolicy {
	if tfMap == nil {
		return nil
	}

	a := &types.RetryPolicy{}

	if v, ok := tfMap["maximum_event_age_in_seconds"].(int); ok {
		a.MaximumEventAgeInSeconds = aws.Int32(int32(v))
	}

	if v, ok := tfMap["maximum_retry_attempts"].(int); ok {
		a.MaximumRetryAttempts = aws.Int32(int32(v))
	}

	return a
}

func flattenRetryPolicy(apiObject *types.RetryPolicy) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.MaximumEventAgeInSeconds; v != nil {
		m["maximum_event_age_in_seconds"] = int(aws.ToInt32(v))
	}

	if v := apiObject.MaximumRetryAttempts; v != nil {
		m["maximum_retry_attempts"] = int(aws.ToInt32(v))
	}

	return m
}

func expandSageMakerPipelineParameter(tfMap map[string]interface{}) types.SageMakerPipelineParameter {
	if tfMap == nil {
		return types.SageMakerPipelineParameter{}
	}

	a := types.SageMakerPipelineParameter{}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		a.Name = aws.String(v)
	}

	if v, ok := tfMap["value"].(string); ok && v != "" {
		a.Value = aws.String(v)
	}

	return a
}

func flattenSageMakerPipelineParameter(apiObject types.SageMakerPipelineParameter) map[string]interface{} {
	m := map[string]interface{}{}

	if v := apiObject.Name; v != nil {
		m["name"] = aws.ToString(v)
	}

	if v := apiObject.Value; v != nil {
		m["value"] = aws.ToString(v)
	}

	return m
}

func expandSageMakerPipelineParameters(tfMap map[string]interface{}) *types.SageMakerPipelineParameters {
	if tfMap == nil {
		return nil
	}

	a := &types.SageMakerPipelineParameters{}

	if v, ok := tfMap["pipeline_parameter"].(*schema.Set); ok && v.Len() > 0 {
		for _, p := range v.List() {
			a.PipelineParameterList = append(a.PipelineParameterList, expandSageMakerPipelineParameter(p.(map[string]interface{})))
		}
	}

	return a
}

func flattenSageMakerPipelineParameters(apiObject *types.SageMakerPipelineParameters) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.PipelineParameterList; v != nil {
		set := schema.NewSet(sagemakerPipelineParameterHash, nil)

		for _, p := range v {
			set.Add(flattenSageMakerPipelineParameter(p))
		}

		m["pipeline_parameter"] = set
	}

	return m
}

func expandSQSParameters(tfMap map[string]interface{}) *types.SqsParameters {
	if tfMap == nil {
		return nil
	}

	a := &types.SqsParameters{}

	if v, ok := tfMap["message_group_id"].(string); ok && v != "" {
		a.MessageGroupId = aws.String(v)
	}

	return a
}

func flattenSQSParameters(apiObject *types.SqsParameters) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.MessageGroupId; v != nil {
		m["message_group_id"] = aws.ToString(v)
	}

	return m
}

func expandTarget(ctx context.Context, tfMap map[string]interface{}) *types.Target {
	if tfMap == nil {
		return nil
	}

	a := &types.Target{}

	if v, ok := tfMap["arn"].(string); ok && v != "" {
		a.Arn = aws.String(v)
	}

	if v, ok := tfMap["dead_letter_config"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.DeadLetterConfig = expandDeadLetterConfig(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["ecs_parameters"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.EcsParameters = expandECSParameters(ctx, v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["eventbridge_parameters"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.EventBridgeParameters = expandEventBridgeParameters(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["input"].(string); ok && v != "" {
		a.Input = aws.String(v)
	}

	if v, ok := tfMap["kinesis_parameters"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.KinesisParameters = expandKinesisParameters(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["role_arn"].(string); ok && v != "" {
		a.RoleArn = aws.String(v)
	}

	if v, ok := tfMap["retry_policy"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.RetryPolicy = expandRetryPolicy(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["sagemaker_pipeline_parameters"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.SageMakerPipelineParameters = expandSageMakerPipelineParameters(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["sqs_parameters"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.SqsParameters = expandSQSParameters(v[0].(map[string]interface{}))
	}

	return a
}

func flattenTarget(ctx context.Context, apiObject *types.Target) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.Arn; v != nil {
		m["arn"] = aws.ToString(v)
	}

	if v := apiObject.DeadLetterConfig; v != nil {
		m["dead_letter_config"] = []interface{}{flattenDeadLetterConfig(v)}
	}

	if v := apiObject.EcsParameters; v != nil {
		m["ecs_parameters"] = []interface{}{flattenECSParameters(ctx, v)}
	}

	if v := apiObject.EventBridgeParameters; v != nil {
		m["eventbridge_parameters"] = []interface{}{flattenEventBridgeParameters(v)}
	}

	if v := apiObject.Input; v != nil {
		m["input"] = aws.ToString(v)
	}

	if v := apiObject.KinesisParameters; v != nil {
		m["kinesis_parameters"] = []interface{}{flattenKinesisParameters(v)}
	}

	if v := apiObject.RoleArn; v != nil {
		m["role_arn"] = aws.ToString(v)
	}

	if v := apiObject.RetryPolicy; v != nil {
		m["retry_policy"] = []interface{}{flattenRetryPolicy(v)}
	}

	if v := apiObject.SageMakerPipelineParameters; v != nil {
		m["sagemaker_pipeline_parameters"] = []interface{}{flattenSageMakerPipelineParameters(v)}
	}

	if v := apiObject.SqsParameters; v != nil {
		m["sqs_parameters"] = []interface{}{flattenSQSParameters(v)}
	}

	return m
}
