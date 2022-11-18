package scheduler

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/scheduler/types"
)

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

	if v, ok := tfMap["input"].(string); ok && v != "" {
		a.Input = aws.String(v)
	}

	if v, ok := tfMap["role_arn"].(string); ok && v != "" {
		a.RoleArn = aws.String(v)
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

	if v := apiObject.Input; v != nil && len(*v) > 0 {
		m["input"] = aws.ToString(v)
	}

	if v := apiObject.RoleArn; v != nil && len(*v) > 0 {
		m["role_arn"] = aws.ToString(v)
	}

	if v := apiObject.SqsParameters; v != nil {
		m["sqs_parameters"] = []interface{}{flattenSqsParameters(v)}
	}

	return m
}
