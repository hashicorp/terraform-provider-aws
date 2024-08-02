// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func expandParameters(tfList []interface{}) []types.Parameter {
	var apiObjects []types.Parameter

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		if tfMap[names.AttrName].(string) == "" {
			continue
		}

		apiObject := types.Parameter{
			ParameterName:  aws.String(strings.ToLower(tfMap[names.AttrName].(string))),
			ParameterValue: aws.String(tfMap[names.AttrValue].(string)),
		}

		if v, ok := tfMap["apply_method"].(string); ok && v != "" {
			apiObject.ApplyMethod = types.ApplyMethod(strings.ToLower(v))
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenParameters(apiObject []types.Parameter) []interface{} {
	apiObjects := make([]interface{}, 0)

	for _, apiObject := range apiObject {
		if apiObject.ParameterName == nil {
			continue
		}

		tfMap := make(map[string]interface{})

		tfMap["apply_method"] = strings.ToLower(string(apiObject.ApplyMethod))
		tfMap[names.AttrName] = strings.ToLower(aws.ToString(apiObject.ParameterName))

		// Default empty string, guard against nil parameter values.
		tfMap[names.AttrValue] = ""
		if apiObject.ParameterValue != nil {
			tfMap[names.AttrValue] = aws.ToString(apiObject.ParameterValue)
		}

		apiObjects = append(apiObjects, tfMap)
	}

	return apiObjects
}
