// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package imagebuilder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func flattenWorkflowParameter(apiObject *imagebuilder.WorkflowParameter) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = aws.StringValue(v)
	}

	if v := apiObject.Value; v != nil {
		// ImageBuilder API quirk
		// Even though Value is a slice, only one element is accepted.
		tfMap[names.AttrValue] = aws.StringValueSlice(v)[0]
	}

	return tfMap
}

func flattenWorkflowParameters(apiObjects []*imagebuilder.WorkflowParameter) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenWorkflowParameter(apiObject))
	}

	return tfList
}

func flattenWorkflowConfiguration(apiObject *imagebuilder.WorkflowConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.OnFailure; v != nil {
		tfMap["on_failure"] = aws.String(*v)
	}

	if v := apiObject.ParallelGroup; v != nil {
		tfMap["parallel_group"] = aws.String(*v)
	}

	if v := apiObject.Parameters; v != nil {
		tfMap[names.AttrParameter] = flattenWorkflowParameters(v)
	}

	if v := apiObject.WorkflowArn; v != nil {
		tfMap["workflow_arn"] = aws.StringValue(v)
	}

	return tfMap
}

func flattenWorkflowConfigurations(apiObjects []*imagebuilder.WorkflowConfiguration) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		if apiObjects == nil {
			continue
		}

		tfList = append(tfList, flattenWorkflowConfiguration(apiObject))
	}

	return tfList
}

func expandWorkflowParameter(tfMap map[string]interface{}) *imagebuilder.WorkflowParameter {
	if tfMap == nil {
		return nil
	}

	apiObject := &imagebuilder.WorkflowParameter{}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap[names.AttrValue].(string); ok && v != "" {
		// ImageBuilder API quirk
		// Even though Value is a slice, only one element is accepted.
		apiObject.Value = aws.StringSlice([]string{v})
	}

	return apiObject
}

func expandWorkflowParameters(tfList []interface{}) []*imagebuilder.WorkflowParameter {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*imagebuilder.WorkflowParameter

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandWorkflowParameter(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandWorkflowConfiguration(tfMap map[string]interface{}) *imagebuilder.WorkflowConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &imagebuilder.WorkflowConfiguration{}

	if v, ok := tfMap["on_failure"].(string); ok && v != "" {
		apiObject.OnFailure = aws.String(v)
	}

	if v, ok := tfMap["parallel_group"].(string); ok && v != "" {
		apiObject.ParallelGroup = aws.String(v)
	}

	if v, ok := tfMap[names.AttrParameter].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Parameters = expandWorkflowParameters(v.List())
	}

	if v, ok := tfMap["workflow_arn"].(string); ok && v != "" {
		apiObject.WorkflowArn = aws.String(v)
	}

	return apiObject
}

func expandWorkflowConfigurations(tfList []interface{}) []*imagebuilder.WorkflowConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*imagebuilder.WorkflowConfiguration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandWorkflowConfiguration(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}
