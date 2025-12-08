// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package imagebuilder

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/imagebuilder/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func flattenWorkflowParameter(apiObject *awstypes.WorkflowParameter) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.Name; v != nil {
		tfMap[names.AttrName] = aws.ToString(v)
	}

	if v := apiObject.Value; len(v) > 0 {
		// ImageBuilder API quirk
		// Even though Value is a slice, only one element is accepted.
		tfMap[names.AttrValue] = v[0]
	}

	return tfMap
}

func flattenWorkflowParameters(apiObjects []awstypes.WorkflowParameter) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenWorkflowParameter(&apiObject))
	}

	return tfList
}

func flattenWorkflowConfiguration(apiObject *awstypes.WorkflowConfiguration) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{
		"on_failure": apiObject.OnFailure,
	}

	if v := apiObject.ParallelGroup; v != nil {
		tfMap["parallel_group"] = aws.ToString(v)
	}

	if v := apiObject.Parameters; v != nil {
		tfMap[names.AttrParameter] = flattenWorkflowParameters(v)
	}

	if v := apiObject.WorkflowArn; v != nil {
		tfMap["workflow_arn"] = aws.ToString(v)
	}

	return tfMap
}

func flattenWorkflowConfigurations(apiObjects []awstypes.WorkflowConfiguration) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any

	for _, apiObject := range apiObjects {
		if apiObjects == nil {
			continue
		}

		tfList = append(tfList, flattenWorkflowConfiguration(&apiObject))
	}

	return tfList
}

func expandWorkflowParameter(tfMap map[string]any) *awstypes.WorkflowParameter {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.WorkflowParameter{}

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap[names.AttrValue].(string); ok && v != "" {
		// ImageBuilder API quirk
		// Even though Value is a slice, only one element is accepted.
		apiObject.Value = []string{v}
	}

	return apiObject
}

func expandWorkflowParameters(tfList []any) []awstypes.WorkflowParameter {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.WorkflowParameter

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandWorkflowParameter(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}

func expandWorkflowConfiguration(tfMap map[string]any) *awstypes.WorkflowConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.WorkflowConfiguration{}

	if v, ok := tfMap["on_failure"].(string); ok && v != "" {
		apiObject.OnFailure = awstypes.OnWorkflowFailure(v)
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

func expandWorkflowConfigurations(tfList []any) []awstypes.WorkflowConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.WorkflowConfiguration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)
		if !ok {
			continue
		}

		apiObject := expandWorkflowConfiguration(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, *apiObject)
	}

	return apiObjects
}
