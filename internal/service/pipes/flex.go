package pipes

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/pipes/types"
)

func expandFilter(tfMap map[string]interface{}) *types.Filter {
	if tfMap == nil {
		return nil
	}

	output := &types.Filter{}

	if v, ok := tfMap["pattern"].(string); ok && len(v) > 0 {
		output.Pattern = aws.String(v)
	}

	return output
}

func flattenFilter(apiObject types.Filter) map[string]interface{} {
	m := map[string]interface{}{}

	if v := apiObject.Pattern; v != nil {
		m["pattern"] = aws.ToString(v)
	}

	return m
}

func expandFilters(tfList []interface{}) []types.Filter {
	if len(tfList) == 0 {
		return nil
	}

	var s []types.Filter

	for _, v := range tfList {
		a := expandFilter(v.(map[string]interface{}))

		if a == nil {
			continue
		}

		s = append(s, *a)
	}

	return s
}

func flattenFilters(apiObjects []types.Filter) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var l []interface{}

	for _, apiObject := range apiObjects {
		l = append(l, flattenFilter(apiObject))
	}

	return l
}

func expandFilterCriteria(tfMap map[string]interface{}) *types.FilterCriteria {
	if tfMap == nil {
		return nil
	}

	output := &types.FilterCriteria{}

	if v, ok := tfMap["filter"].([]interface{}); ok && len(v) > 0 {
		output.Filters = expandFilters(v)
	}

	return output
}

func flattenFilterCriteria(apiObject *types.FilterCriteria) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	m["filter"] = flattenFilters(apiObject.Filters)

	return m
}

func expandPipeSourceParameters(tfMap map[string]interface{}) *types.PipeSourceParameters {
	if tfMap == nil {
		return nil
	}

	a := &types.PipeSourceParameters{}

	if v, ok := tfMap["filter_criteria"].([]interface{}); ok {
		a.FilterCriteria = expandFilterCriteria(v[0].(map[string]interface{}))
	}

	return a
}

func flattenPipeSourceParameters(apiObject *types.PipeSourceParameters) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.FilterCriteria; v != nil {
		m["filter_criteria"] = []interface{}{flattenFilterCriteria(v)}
	}

	return m
}

func expandPipeTargetParameters(tfMap map[string]interface{}) *types.PipeTargetParameters {
	if tfMap == nil {
		return nil
	}

	a := &types.PipeTargetParameters{}

	if v, ok := tfMap["input_template"].(string); ok {
		a.InputTemplate = aws.String(v)
	}

	return a
}

func flattenPipeTargetParameters(apiObject *types.PipeTargetParameters) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.InputTemplate; v != nil {
		m["input_template"] = aws.ToString(v)
	}

	return m
}
