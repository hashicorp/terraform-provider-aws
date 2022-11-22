package scheduler

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/scheduler/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

func expandAwsvpcConfiguration(tfMap map[string]interface{}) *types.AwsVpcConfiguration {
	if tfMap == nil {
		return nil
	}

	a := &types.AwsVpcConfiguration{}

	if v, ok := tfMap["assign_public_ip"].(string); ok && v != "" {
		a.AssignPublicIp = types.AssignPublicIp(v)
	}

	if v, ok := tfMap["security_groups"].(*schema.Set); ok && v.Len() > 0 {
		a.SecurityGroups = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["subnets"].(*schema.Set); ok && v.Len() > 0 {
		a.Subnets = flex.ExpandStringValueSet(v)
	}

	return a
}

func flattenAwsvpcConfiguration(apiObject *types.AwsVpcConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.AssignPublicIp; string(v) != "" {
		m["assign_public_ip"] = string(v)
	}

	m["security_groups"] = flex.FlattenStringValueSet(apiObject.SecurityGroups)
	m["subnets"] = flex.FlattenStringValueSet(apiObject.Subnets)

	return m
}

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

	if v := aws.ToString(apiObject.CapacityProvider); v != "" {
		m["capacity_provider"] = v
	}

	m["weight"] = apiObject.Weight

	return m
}

func expandDeadLetterConfig(tfMap map[string]interface{}) *types.DeadLetterConfig {
	if tfMap == nil {
		return nil
	}

	a := &types.DeadLetterConfig{}

	if v, ok := tfMap["arn"]; ok && v.(string) != "" {
		a.Arn = aws.String(v.(string))
	}

	return a
}

func flattenDeadLetterConfig(apiObject *types.DeadLetterConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := aws.ToString(apiObject.Arn); v != "" {
		m["arn"] = v
	}

	return m
}

func expandEcsParameters(tfMap map[string]interface{}) *types.EcsParameters {
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

	if v, ok := tfMap["network_configuration"].([]interface{}); ok && len(v) > 0 {
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

	if v, ok := tfMap["tags"].(map[string]interface{}); ok && v != nil {
		for k, v := range v {
			a.Tags = append(a.Tags, map[string]string{
				"key":   k,
				"value": v.(string),
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

func flattenEcsParameters(apiObject *types.EcsParameters) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.CapacityProviderStrategy; v != nil {
		var l []interface{}

		for _, p := range v {
			l = append(l, flattenCapacityProviderStrategyItem(p))
		}

		m["capacity_provider_strategy"] = schema.NewSet(capacityProviderHash, l)
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

	if v := apiObject.LaunchType; string(v) != "" {
		m["launch_type"] = v
	}

	if v := apiObject.NetworkConfiguration; v != nil {
		m["network_configuration"] = []interface{}{flattenNetworkConfiguration(v)}
	}

	if v := apiObject.PlacementConstraints; len(v) > 0 {
		var l []interface{}

		for _, c := range v {
			l = append(l, flattenPlacementConstraint(c))
		}

		m["placement_constraints"] = schema.NewSet(placementConstraintHash, l)
	}

	if v := apiObject.PlacementStrategy; len(v) > 0 {
		var l []interface{}

		for _, s := range v {
			l = append(l, flattenPlacementStrategy(s))
		}

		m["placement_strategy"] = schema.NewSet(placementStrategyHash, l)
	}

	if v := apiObject.PlatformVersion; v != nil {
		m["platform_version"] = aws.ToString(v)
	}

	if v := apiObject.PropagateTags; string(v) != "" {
		m["propagate_tags"] = v
	}

	if v := apiObject.ReferenceId; v != nil {
		m["reference_id"] = aws.ToString(v)
	}

	if v := apiObject.Tags; len(v) > 0 {
		result := make(map[string]interface{})

		for _, tagMap := range v {
			key := tagMap["key"]

			// The EventBridge Scheduler API documents raw maps instead of
			// the key-value structure expected by the RunTask API.
			if key == "" {
				continue
			}

			result[key] = tagMap["value"]
		}

		m["tags"] = result
	}

	if v := apiObject.TaskCount; v != nil {
		m["task_count"] = int(aws.ToInt32(v))
	}

	if v := aws.ToString(apiObject.TaskDefinitionArn); v != "" {
		m["task_definition_arn"] = v
	}

	return m
}

func expandEventBridgeParameters(tfMap map[string]interface{}) *types.EventBridgeParameters {
	if tfMap == nil {
		return nil
	}

	a := &types.EventBridgeParameters{}

	if v, ok := tfMap["detail_type"]; ok && v.(string) != "" {
		a.DetailType = aws.String(v.(string))
	}

	if v, ok := tfMap["source"]; ok && v.(string) != "" {
		a.Source = aws.String(v.(string))
	}

	return a
}

func flattenEventBridgeParameters(apiObject *types.EventBridgeParameters) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := aws.ToString(apiObject.DetailType); v != "" {
		m["detail_type"] = v
	}

	if v := aws.ToString(apiObject.Source); v != "" {
		m["source"] = v
	}

	return m
}

func expandFlexibleTimeWindow(tfMap map[string]interface{}) *types.FlexibleTimeWindow {
	if tfMap == nil {
		return nil
	}

	a := &types.FlexibleTimeWindow{}

	if v, ok := tfMap["maximum_window_in_minutes"]; ok && v.(int) != 0 {
		a.MaximumWindowInMinutes = aws.Int32(int32(v.(int)))
	}

	if v, ok := tfMap["mode"]; ok && v.(string) != "" {
		a.Mode = types.FlexibleTimeWindowMode(v.(string))
	}

	return a
}

func flattenFlexibleTimeWindow(apiObject *types.FlexibleTimeWindow) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := aws.ToInt32(apiObject.MaximumWindowInMinutes); v != 0 {
		m["maximum_window_in_minutes"] = int(v)
	}

	if v := apiObject.Mode; v != "" {
		m["mode"] = string(v)
	}

	return m
}

func expandKinesisParameters(tfMap map[string]interface{}) *types.KinesisParameters {
	if tfMap == nil {
		return nil
	}

	a := &types.KinesisParameters{}

	if v, ok := tfMap["partition_key"]; ok && v.(string) != "" {
		a.PartitionKey = aws.String(v.(string))
	}

	return a
}

func flattenKinesisParameters(apiObject *types.KinesisParameters) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := aws.ToString(apiObject.PartitionKey); v != "" {
		m["partition_key"] = v
	}

	return m
}

func expandNetworkConfiguration(tfMap map[string]interface{}) *types.NetworkConfiguration {
	if tfMap == nil {
		return nil
	}

	a := &types.NetworkConfiguration{}

	if v, ok := tfMap["awsvpc_configuration"].([]interface{}); ok && len(v) > 0 {
		a.AwsvpcConfiguration = expandAwsvpcConfiguration(v[0].(map[string]interface{}))
	}

	return a
}

func flattenNetworkConfiguration(apiObject *types.NetworkConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.AwsvpcConfiguration; v != nil {
		m["awsvpc_configuration"] = []interface{}{flattenAwsvpcConfiguration(v)}
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

	if v := aws.ToString(apiObject.Expression); v != "" {
		m["expression"] = v
	}

	if v := apiObject.Type; v != "" {
		m["type"] = string(v)
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

	if v := aws.ToString(apiObject.Field); v != "" {
		m["field"] = v
	}

	if v := apiObject.Type; v != "" {
		m["type"] = string(v)
	}

	return m
}

func expandRetryPolicy(tfMap map[string]interface{}) *types.RetryPolicy {
	if tfMap == nil {
		return nil
	}

	a := &types.RetryPolicy{}

	if v, ok := tfMap["maximum_event_age_in_seconds"]; ok {
		a.MaximumEventAgeInSeconds = aws.Int32(int32(v.(int)))
	}

	if v, ok := tfMap["maximum_retry_attempts"]; ok {
		a.MaximumRetryAttempts = aws.Int32(int32(v.(int)))
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

func expandSqsParameters(tfMap map[string]interface{}) *types.SqsParameters {
	if tfMap == nil {
		return nil
	}

	a := &types.SqsParameters{}

	if v, ok := tfMap["message_group_id"]; ok && v.(string) != "" {
		a.MessageGroupId = aws.String(v.(string))
	}

	return a
}

func expandSageMakerPipelineParameter(tfMap map[string]interface{}) types.SageMakerPipelineParameter {
	if tfMap == nil {
		return types.SageMakerPipelineParameter{}
	}

	a := types.SageMakerPipelineParameter{}

	if v, ok := tfMap["name"]; ok && v.(string) != "" {
		a.Name = aws.String(v.(string))
	}

	if v, ok := tfMap["value"]; ok && v.(string) != "" {
		a.Value = aws.String(v.(string))
	}

	return a
}

func flattenSageMakerPipelineParameter(apiObject types.SageMakerPipelineParameter) map[string]interface{} {
	m := map[string]interface{}{}

	if v := aws.ToString(apiObject.Name); v != "" {
		m["name"] = v
	}

	if v := aws.ToString(apiObject.Value); v != "" {
		m["value"] = v
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
		var l []interface{}

		for _, p := range v {
			l = append(l, flattenSageMakerPipelineParameter(p))
		}

		m["pipeline_parameter"] = schema.NewSet(sagemakerPipelineParameterHash, l)
	}

	return m
}

func flattenSqsParameters(apiObject *types.SqsParameters) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := aws.ToString(apiObject.MessageGroupId); v != "" {
		m["message_group_id"] = v
	}

	return m
}

func expandTarget(tfMap map[string]interface{}) *types.Target {
	if tfMap == nil {
		return nil
	}

	a := &types.Target{}

	if v, ok := tfMap["arn"].(string); ok && v != "" {
		a.Arn = aws.String(v)
	}

	if v, ok := tfMap["dead_letter_config"]; ok && len(v.([]interface{})) > 0 {
		a.DeadLetterConfig = expandDeadLetterConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := tfMap["ecs_parameters"]; ok && len(v.([]interface{})) > 0 {
		a.EcsParameters = expandEcsParameters(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := tfMap["eventbridge_parameters"]; ok && len(v.([]interface{})) > 0 {
		a.EventBridgeParameters = expandEventBridgeParameters(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := tfMap["input"].(string); ok && v != "" {
		a.Input = aws.String(v)
	}

	if v, ok := tfMap["kinesis_parameters"]; ok && len(v.([]interface{})) > 0 {
		a.KinesisParameters = expandKinesisParameters(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := tfMap["role_arn"].(string); ok && v != "" {
		a.RoleArn = aws.String(v)
	}

	if v, ok := tfMap["retry_policy"]; ok && len(v.([]interface{})) > 0 {
		a.RetryPolicy = expandRetryPolicy(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := tfMap["sagemaker_pipeline_parameters"]; ok && len(v.([]interface{})) > 0 {
		a.SageMakerPipelineParameters = expandSageMakerPipelineParameters(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := tfMap["sqs_parameters"]; ok && len(v.([]interface{})) > 0 {
		a.SqsParameters = expandSqsParameters(v.([]interface{})[0].(map[string]interface{}))
	}

	return a
}

func flattenTarget(apiObject *types.Target) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.Arn; v != nil && len(*v) > 0 {
		m["arn"] = aws.ToString(v)
	}

	if v := apiObject.DeadLetterConfig; v != nil {
		m["dead_letter_config"] = []interface{}{flattenDeadLetterConfig(v)}
	}

	if v := apiObject.EcsParameters; v != nil {
		m["ecs_parameters"] = []interface{}{flattenEcsParameters(v)}
	}

	if v := apiObject.EventBridgeParameters; v != nil {
		m["eventbridge_parameters"] = []interface{}{flattenEventBridgeParameters(v)}
	}

	if v := apiObject.Input; v != nil && len(*v) > 0 {
		m["input"] = aws.ToString(v)
	}

	if v := apiObject.KinesisParameters; v != nil {
		m["kinesis_parameters"] = []interface{}{flattenKinesisParameters(v)}
	}

	if v := apiObject.RoleArn; v != nil && len(*v) > 0 {
		m["role_arn"] = aws.ToString(v)
	}

	if v := apiObject.RetryPolicy; v != nil {
		m["retry_policy"] = []interface{}{flattenRetryPolicy(v)}
	}

	if v := apiObject.SageMakerPipelineParameters; v != nil {
		m["sagemaker_pipeline_parameters"] = []interface{}{flattenSageMakerPipelineParameters(v)}
	}

	if v := apiObject.SqsParameters; v != nil {
		m["sqs_parameters"] = []interface{}{flattenSqsParameters(v)}
	}

	return m
}
