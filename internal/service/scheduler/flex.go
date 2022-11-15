package scheduler

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/scheduler/types"
)

func expandFlexibleTimeWindow(tfMap map[string]interface{}) *types.FlexibleTimeWindow {
	if tfMap == nil {
		return nil
	}

	a := &types.FlexibleTimeWindow{}

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

	if v := apiObject.Mode; v != "" {
		m["mode"] = string(v)
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

	if v, ok := tfMap["role_arn"].(string); ok && v != "" {
		a.RoleArn = aws.String(v)
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

	if v := apiObject.RoleArn; v != nil && len(*v) > 0 {
		m["role_arn"] = aws.ToString(v)
	}

	return m
}
